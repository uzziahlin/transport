package message

import (
	"bytes"
	"encoding/binary"
)

type Header struct {
	HeaderLen  uint32
	DataLen    uint32
	MessageId  uint32
	Version    uint8
	Compressor uint8
	Serializer uint8
}

func (h *Header) fixedHeadLen() int {
	return 15
}

// message 对报文的抽象，提供了写入和读取报文的操作
type message struct {
	data   []byte
	offset int
}

// putUint32 将uint32数据写入报文
func (m *message) putUint32(data uint32) {
	binary.BigEndian.PutUint32(m.data[m.offset:m.offset+4], data)
	// m.data = m.data[4:]
	m.offset += 4
}

// putUint8 将uint8数据写入报文
func (m *message) putUint8(data uint8) {
	m.data[m.offset] = data
	// m.data = m.data[1:]
	m.offset += 1
}

func (m *message) putString(data string) {
	m.putBytes([]byte(data))
}

func (m *message) putBytes(data []byte) {
	copy(m.data[m.offset:], data)
	// m.data = m.data[len(data):]
	m.offset += len(data)
}

func (m *message) putByte(data byte) {
	m.putUint8(data)
}

func (m *message) uint32() uint32 {
	res := binary.BigEndian.Uint32(m.data[m.offset : m.offset+4])
	// m.data = m.data[4:]
	m.offset += 4
	return res
}

func (m *message) uint8() uint8 {
	res := m.data[m.offset]
	// m.data = m.data[1:]
	m.offset += 1
	return res
}

func (m *message) bytes(size int) []byte {
	res := m.data[m.offset : m.offset+size]
	// m.data = m.data[size:]
	m.offset += size
	return res
}

func (m *message) readToIndex(idx int) []byte {

	if idx <= m.offset {
		return nil
	}

	return m.bytes(idx - m.offset)

}

func (m *message) readToByte(splitter byte, end ...int) []byte {

	res := make([]byte, 0)

	r := len(m.data)

	if end != nil && len(end) > 0 {

		r = end[0]

		if r <= m.offset {
			return res
		}
	}

	idx := bytes.IndexByte(m.data[m.offset:r], splitter)

	if idx == -1 {
		return res
	}

	res = m.data[m.offset : m.offset+idx]

	m.offset += idx + 1

	return res
}

func (m *message) string(size int) string {
	return string(m.bytes(size))
}

// setHeader 往报文中写入头部信息
func (m *message) setHeader(header *Header) {
	m.putUint32(header.HeaderLen)

	m.putUint32(header.DataLen)

	m.putUint32(header.MessageId)

	m.putUint8(header.Version)

	m.putUint8(header.Compressor)

	m.putUint8(header.Serializer)
}

// setHeader 往报文中写入消息体信息
func (m *message) setBody(body []byte) {
	m.putBytes(body)
}

// getHeader 从报文中读取头部信息
func (m *message) getHeader() *Header {

	res := &Header{
		HeaderLen:  m.uint32(),
		DataLen:    m.uint32(),
		MessageId:  m.uint32(),
		Version:    m.uint8(),
		Compressor: m.uint8(),
		Serializer: m.uint8(),
	}

	return res
}

// getBody 从报文中读取消息体信息
func (m *message) getBody(size int) []byte {
	return m.bytes(size)
}
