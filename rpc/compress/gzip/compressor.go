package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/uzziahlin/transport/rpc/compress"
	"io"
)

type Compressor struct {
}

func (c Compressor) Compress(data []byte) ([]byte, error) {
	buf := bytes.Buffer{}
	w := gzip.NewWriter(&buf)

	_, err := w.Write(data)

	err = w.Close()

	if err != nil {
		return nil, fmt.Errorf("数据压缩错误, %w", err)
	}

	return buf.Bytes(), nil
}

func (c Compressor) Decompress(data []byte) ([]byte, error) {
	out := bytes.Buffer{}

	in := bytes.Buffer{}
	in.Write(data)

	r, err := gzip.NewReader(&in)
	err = r.Close()

	if err != nil {
		return nil, err
	}

	_, err = io.Copy(&out, r)

	if err != nil {
		return nil, fmt.Errorf("数据解压缩错误, %w", err)
	}

	return out.Bytes(), nil
}

func (c Compressor) Code() uint8 {
	return uint8(compress.GZIP)
}
