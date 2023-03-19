package rpc

import (
	"context"
	"errors"
	"github.com/uzziahlin/transport/rpc/compress"
	"github.com/uzziahlin/transport/rpc/compress/gzip"
	"github.com/uzziahlin/transport/rpc/compress/zip"
	"github.com/uzziahlin/transport/rpc/message"
	"github.com/uzziahlin/transport/rpc/serialize"
	"github.com/uzziahlin/transport/rpc/serialize/json"

	"reflect"
	"strconv"
)

type ConstructorOpt func(constructor *ProxyConstructor)

func WithSerializer(serializer serialize.Serializer) ConstructorOpt {
	return func(c *ProxyConstructor) {
		c.serializer = serializer
	}
}

type ProxyConstructor struct {
	serializer  serialize.Serializer
	compressors map[compress.Type]compress.Compressor
}

func NewProxyConstructor(opts ...ConstructorOpt) *ProxyConstructor {
	res := &ProxyConstructor{
		serializer:  &json.Serializer{},
		compressors: make(map[compress.Type]compress.Compressor, 4),
	}

	res.RegisterCompressor(&gzip.Compressor{})
	res.RegisterCompressor(&zip.Compressor{})

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func (p ProxyConstructor) RegisterCompressor(compressor compress.Compressor) {
	p.compressors[compress.Type(compressor.Code())] = compressor
}

// InitProxy 初始化代理 篡改类型属性 要求类型一定是struct，且为一级指针
// 要求service属性为函数类型
// 函数只有两个参数，且第一个参数是context.Context, 第二个参数为真正入参
// 函数有一个返回值
func (p ProxyConstructor) InitProxy(service Service) error {

	proxy := NewRemoteProxy(service.Info().Addr)

	return p.setFuncField(service, proxy)
}

func (p ProxyConstructor) setFuncField(service Service, proxy Proxy) error {
	if service == nil {
		return errors.New("micro：入参不能为nil")
	}

	ptrVal := reflect.ValueOf(service)
	ptrTyp := reflect.TypeOf(service)

	if ptrVal.Kind() != reflect.Pointer && ptrVal.Kind() != reflect.Struct {
		return errors.New("micro：入参必须为结构体指针且为一级指针")
	}

	val := ptrVal.Elem()
	typ := ptrTyp.Elem()

	// 找到类型的函数属性
	for i := 0; i < val.NumField(); i++ {
		fd := val.Field(i)
		fdTyp := typ.Field(i)

		if fd.Kind() != reflect.Func {
			continue
		}

		if fd.CanSet() {

			// 定义函数进行篡改
			fn := func(args []reflect.Value) (results []reflect.Value) {

				// 根据返回值类型，实例化返回值
				res := reflect.New(fdTyp.Type.Out(0).Elem())

				ctx, ok := args[0].Interface().(context.Context)

				if !ok {
					err := errors.New("micro：函数第一个参数必须是context.Context")
					return []reflect.Value{res, reflect.ValueOf(err)}
				}

				// 将参数进行序列化
				arg, err := p.serializer.Serialize(args[1].Interface())

				if err != nil {
					return []reflect.Value{res, reflect.ValueOf(err)}
				}

				// 构造调用信息
				req := &message.Request{
					RequestHeader: message.RequestHeader{
						Header: message.Header{
							Serializer: p.serializer.Code(),
						},
						ServiceName: service.Info().ServiceName,
						MethodName:  fdTyp.Name,
					},
					Data: arg,
				}

				if cTyp, ok := compress.EnableCompress(ctx); ok {
					compressor, ok := p.compressors[cTyp]
					if !ok {
						err = errors.New("micro：找不到对应的压缩算法")
						return []reflect.Value{res, reflect.ValueOf(err)}
					}
					req.Data, err = compressor.Compress(req.Data)
					if err != nil {
						return []reflect.Value{res, reflect.ValueOf(err)}
					}
					req.Compressor = uint8(cTyp)
				}

				meta := make(map[string]string, 4)

				if oneway := isOneway(ctx); oneway {
					meta["sys_oneway"] = "true"
				}

				if deadline, ok := ctx.Deadline(); ok {
					dl := deadline.UnixMilli()
					meta["sys_timeout"] = strconv.FormatInt(dl, 10)
				}

				req.Meta = meta

				var resp *message.Response

				resC := make(chan struct{})

				go func() {
					// 请求服务端，并获得响应
					resp, err = proxy.Invoke(ctx, req)
					select {
					case resC <- struct{}{}:
					default:
					}
				}()

				select {
				case <-ctx.Done():
					return []reflect.Value{res, reflect.ValueOf(ctx.Err())}
				case <-resC:

				}

				if err != nil {
					return []reflect.Value{res, reflect.ValueOf(err)}
				}

				// 将返回结果进行反序列化，构造返回值
				if resData := resp.Data; resData != nil && len(resData) > 0 {
					// 判断用户是否指定压缩算法，有则获取压缩算法进行解压缩操作
					if cTyp, ok := compress.EnableCompress(ctx); ok {
						compressor, ok := p.compressors[cTyp]
						if !ok {
							err = errors.New("micro：找不到对应的压缩算法")
							return []reflect.Value{res, reflect.ValueOf(err)}
						}
						resData, err = compressor.Decompress(resData)
						if err != nil {
							return []reflect.Value{res, reflect.ValueOf(err)}
						}
					}
					err = p.serializer.Deserialize(resData, res.Interface())
				}

				if err != nil {
					return []reflect.Value{res, reflect.ValueOf(err)}
				}

				if resp.Error != "" {
					err = errors.New(resp.Error)
					return []reflect.Value{res, reflect.ValueOf(err)}
				}

				return []reflect.Value{res, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}

			f := reflect.MakeFunc(fd.Type(), fn)
			// 开始篡改属性
			fd.Set(f)
		}

	}

	return nil
}
