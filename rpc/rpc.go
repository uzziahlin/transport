package rpc

import (
	"encoding/binary"
	"io"
)

const (
	rpcHeadLenBytes = 4
	rpcDataLenBytes = 4
)

type Reader func(r io.Reader) ([]byte, error)

func RpcReader(r io.Reader) ([]byte, error) {

	// 先读前边8个字节，以确定后续还需要读取多长的字节

	head, headLen, err := readUint32(r)

	if err != nil {
		return nil, err
	}

	data, dataLen, err := readUint32(r)

	if err != nil {
		return nil, err
	}

	res := make([]byte, int(headLen+dataLen))

	cur := res

	copy(cur, head)

	cur = cur[rpcHeadLenBytes:]

	copy(cur, data)

	cur = cur[rpcDataLenBytes:]

	_, err = r.Read(cur)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func readUint32(r io.Reader) ([]byte, uint32, error) {
	lens := make([]byte, 4)

	_, err := r.Read(lens)

	if err != nil {
		return nil, 0, err
	}

	return lens, binary.BigEndian.Uint32(lens), nil
}
