package message

import (
	"strings"
)

const (
	itemSplitter = '\n'
	kvSplitter   = '\r'
)

type reqMessage struct {
	message
}

func (m *reqMessage) setHeader(header *RequestHeader) {
	m.message.setHeader(&header.Header)

	m.putString(header.ServiceName)

	m.putByte(itemSplitter)

	m.putString(header.MethodName)

	m.putByte(itemSplitter)

	for k, v := range header.Meta {
		m.putString(k)
		// kv中间写入分隔符，方便后边解析报文
		m.putByte(kvSplitter)
		m.putString(v)
		m.putByte(itemSplitter)
	}
}

func (m *reqMessage) setMessage(req *Request) {
	m.setHeader(&req.RequestHeader)
	m.setBody(req.Data)
}

func (m *reqMessage) getHeader() *RequestHeader {

	reqHeader := &RequestHeader{
		Header: *(m.message.getHeader()),
	}

	headLen := int(reqHeader.HeaderLen)

	reqHeader.ServiceName = string(m.readToByte(itemSplitter, headLen))

	reqHeader.MethodName = string(m.readToByte(itemSplitter, headLen))

	mts := m.readToIndex(headLen)

	if mts != nil && len(mts) > 0 {
		m := make(map[string]string, 4)

		metas := string(mts)

		seps := strings.Split(metas, string(itemSplitter))

		for _, sep := range seps {
			if sep == "" {
				continue
			}
			kv := strings.Split(sep, string(kvSplitter))
			m[kv[0]] = kv[1]
		}

		reqHeader.Meta = m
	}

	return reqHeader
}

func (m *reqMessage) getMessage() *Request {
	header := m.getHeader()
	body := m.getBody(int(header.DataLen))

	if len(body) == 0 {
		body = nil
	}

	return &Request{
		RequestHeader: *header,
		Data:          body,
	}
}

type RequestHeader struct {
	Header
	ServiceName string
	MethodName  string
	Meta        map[string]string
}

func (r *RequestHeader) calHeadLen() {
	res := r.fixedHeadLen() +
		len(r.ServiceName) + 1 +
		len(r.MethodName) + 1

	for k, v := range r.Meta {
		// +1是因为k后边跟着kv分隔符
		res += len(k) + 1

		// +1是因为v后边跟着\n分隔符
		res += len(v) + 1
	}

	r.HeaderLen = uint32(res)
}

type Request struct {
	RequestHeader
	Data []byte
}

func (r *Request) IsOneway() bool {
	oneway, ok := r.Meta["sys_oneway"]
	return ok && oneway == "true"
}

func (r *Request) calDataLen() {
	r.RequestHeader.DataLen = uint32(len(r.Data))
}

type RequestEncoder interface {
	Encode(request *Request) []byte
	Decode(data []byte) (*Request, error)
}

type DefaultRequestEncoder struct {
}

func (d *DefaultRequestEncoder) Encode(req *Request) []byte {

	// 是否在这里计算协议头长度和协议体长度
	req.calHeadLen()
	req.calDataLen()

	res := make([]byte, req.HeaderLen+req.DataLen)

	msg := reqMessage{
		message: message{
			data: res,
		},
	}

	msg.setMessage(req)

	return res
}

func (d *DefaultRequestEncoder) Decode(data []byte) (*Request, error) {

	reqMsg := reqMessage{
		message: message{
			data: data,
		},
	}

	return reqMsg.getMessage(), nil
}
