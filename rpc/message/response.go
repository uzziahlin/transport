package message

type respMessage struct {
	message
}

func (m *respMessage) setHeader(header *ResponseHeader) {
	m.message.setHeader(&header.Header)

	m.putString(header.Error)
}

func (m *respMessage) setMessage(resp *Response) {
	m.setHeader(&resp.ResponseHeader)
	m.setBody(resp.Data)
}

func (m *respMessage) getHeader() *ResponseHeader {

	respHeader := &ResponseHeader{
		Header: *(m.message.getHeader()),
	}

	headLen := int(respHeader.HeaderLen)

	mts := m.readToIndex(headLen)

	if mts != nil {
		respHeader.Error = string(mts)
	}

	return respHeader
}

func (m *respMessage) getMessage() *Response {
	header := m.getHeader()
	body := m.getBody(int(header.DataLen))

	if len(body) == 0 {
		body = nil
	}

	return &Response{
		ResponseHeader: *header,
		Data:           body,
	}
}

type ResponseHeader struct {
	Header
	Error string
}

func (r *ResponseHeader) calHeadLen() {
	r.HeaderLen = uint32(r.fixedHeadLen() + len(r.Error))
}

type Response struct {
	ResponseHeader
	Data []byte
}

func (r *Response) calDataLen() {
	r.ResponseHeader.DataLen = uint32(len(r.Data))
}

type ResponseEncoder interface {
	Encode(resp *Response) []byte
	Decode(data []byte) (*Response, error)
}

type DefaultResponseEncoder struct {
}

func (d *DefaultResponseEncoder) Encode(resp *Response) []byte {

	// 先计算协议头和协议体的长度
	resp.calHeadLen()
	resp.calDataLen()

	res := make([]byte, resp.HeaderLen+resp.DataLen)

	msg := respMessage{
		message: message{
			data: res,
		},
	}

	msg.setMessage(resp)

	return res
}

func (d *DefaultResponseEncoder) Decode(data []byte) (*Response, error) {

	respMsg := respMessage{
		message: message{
			data: data,
		},
	}

	return respMsg.getMessage(), nil
}
