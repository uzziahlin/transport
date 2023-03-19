package rpc

import (
	"context"
	"errors"
	"github.com/uzziahlin/transport/rpc/compress"
	"github.com/uzziahlin/transport/rpc/compress/gzip"
	"github.com/uzziahlin/transport/rpc/compress/zip"
	"github.com/uzziahlin/transport/rpc/errs"
	"github.com/uzziahlin/transport/rpc/message"
	"github.com/uzziahlin/transport/rpc/serialize"
	"github.com/uzziahlin/transport/rpc/serialize/json"
	"github.com/uzziahlin/transport/rpc/serialize/proto"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"
)

func NewEndPoint(addr string) *EndPoint {

	jsonS := &json.Serializer{}
	protoS := &proto.Serializer{}

	gzipC := &gzip.Compressor{}
	zipC := &zip.Compressor{}

	ep := &EndPoint{
		read: RpcReader,
		serializers: map[uint8]serialize.Serializer{
			jsonS.Code():  jsonS,
			protoS.Code(): protoS,
		},
		compressors: map[uint8]compress.Compressor{
			gzipC.Code(): gzipC,
			zipC.Code():  zipC,
		},
		services:    make(map[string]reflectionStub, 16),
		reqEncoder:  &message.DefaultRequestEncoder{},
		respEncoder: &message.DefaultResponseEncoder{},
	}

	server := NewServer(addr, ep.handler)

	ep.server = server

	return ep
}

type EndPoint struct {
	services    map[string]reflectionStub
	serializers map[uint8]serialize.Serializer
	compressors map[uint8]compress.Compressor
	read        Reader
	server      *Server
	reqEncoder  message.RequestEncoder
	respEncoder message.ResponseEncoder
}

func (e *EndPoint) Register(service Service) {
	serviceName := service.Info().ServiceName

	e.services[serviceName] = reflectionStub{
		s:           service,
		value:       reflect.ValueOf(service),
		serializers: e.serializers,
		compressors: e.compressors,
	}
}

func (e *EndPoint) RegisterSerializer(serializer serialize.Serializer) {
	e.serializers[serializer.Code()] = serializer
}

func (e *EndPoint) RegisterCompressor(compressor compress.Compressor) {
	e.compressors[compressor.Code()] = compressor
}

func (e *EndPoint) handler(conn net.Conn) {

	for {
		// 解码出请求信息
		data, err := e.read(conn)

		var (
			res    *message.Response
			req    *message.Request
			cancel context.CancelFunc = func() {}
		)

		ctx := context.Background()

		if err != nil {
			log.Printf("请求数据解码错误")
			goto RESP
		}

		// 反序列化请求调用信息，应该是Request结构
		req, err = e.reqEncoder.Decode(data)

		if err != nil {
			log.Printf("请求数据反序列化错误")
			goto RESP
		}

		if meta := req.Meta; meta != nil {
			dl, ok := meta["sys_timeout"]
			if ok {
				deadline, err := strconv.ParseInt(dl, 10, 64)
				if err != nil {
					log.Printf("时间格式不对")
					goto RESP
				}
				// 如果上游服务带了deadline，说明链路有过期时间，应该重建context
				ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(deadline))
			}
		}

		res, err = e.Invoke(ctx, req)

		if ctx.Err() != nil {
			log.Printf("请求超时了，不用给客户端返回了")
			err = ctx.Err()
		}

		cancel()

	RESP:
		if err != nil {
			log.Printf(err.Error())
			res.Error = err.Error()
		}

		resp := e.respEncoder.Encode(res)

		_, err = conn.Write(resp)

		if err != nil {
			log.Printf("响应出错了")
			continue
		}

	}
}

func (e *EndPoint) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	// 根据调用信息获取服务
	service := e.services[req.ServiceName]

	// 通过反射获取服务的相关方法信息
	res, err := service.invoke(ctx, req)

	var errStr string

	if err != nil {
		if errors.Is(err, errs.ErrOneway) {
			return nil, err
		}
		errStr = err.Error()
	}

	return &message.Response{
		ResponseHeader: message.ResponseHeader{
			Error: errStr,
		},
		Data: res,
	}, nil

}

func (e *EndPoint) Startup() error {
	return e.server.Start()
}

type reflectionStub struct {
	s           Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer
	compressors map[uint8]compress.Compressor
}

func (r *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	method := r.value.MethodByName(req.MethodName)

	var err error

	if cTyp := req.Compressor; cTyp != 0 {
		compressor, ok := r.compressors[cTyp]
		if !ok {
			return nil, errors.New("找不到相应的压缩算法支持")
		}
		req.Data, err = compressor.Decompress(req.Data)
		if err != nil {
			log.Printf("请求数据解压缩失败")
			return nil, err
		}
	}

	serializer := r.serializers[req.Serializer]

	resPtr := reflect.New(method.Type().In(1).Elem())

	err = serializer.Deserialize(req.Data, resPtr.Interface())

	if err != nil {
		log.Printf("请求数据反序列化错误")
		return nil, err
	}

	in := []reflect.Value{reflect.ValueOf(ctx), resPtr}

	// 调用服务并获得相应
	results := method.Call(in)

	if req.IsOneway() {
		return nil, errs.ErrOneway
	}

	var data []byte

	if results[0].Interface() != nil {
		data, err = serializer.Serialize(results[0].Interface())

		if err != nil {
			log.Printf("结果序列化出错了")
			return nil, err
		}

		if cTyp := req.Compressor; cTyp != 0 {
			compressor, ok := r.compressors[cTyp]
			if !ok {
				return nil, errors.New("找不到相应的压缩算法支持")
			}
			data, err = compressor.Compress(data)
			if err != nil {
				log.Printf("请求数据解压缩失败")
				return nil, err
			}
		}
	}

	if results[1].Interface() != nil {
		log.Printf("调用出错了")
		return data, results[1].Interface().(error)
	}

	return data, nil
}
