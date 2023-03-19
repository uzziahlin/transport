package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseEncoder_EncodeDecode(t *testing.T) {
	testCases := []struct {
		name    string
		encoder ResponseEncoder
		resp    Response
	}{
		{
			name:    "default no Error and no data",
			encoder: &DefaultResponseEncoder{},
			resp: func() Response {
				resp := Response{
					ResponseHeader: ResponseHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
					},
				}
				return resp
			}(),
		},
		{
			name:    "default no data",
			encoder: &DefaultResponseEncoder{},
			resp: func() Response {
				resp := Response{
					ResponseHeader: ResponseHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
						Error: "micro: 程序发生错误了",
					},
				}
				return resp
			}(),
		},
		{
			name:    "default no error",
			encoder: &DefaultResponseEncoder{},
			resp: func() Response {
				resp := Response{
					ResponseHeader: ResponseHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
					},
					Data: []byte("此程序正常返回响应"),
				}
				return resp
			}(),
		},
		{
			name:    "default",
			encoder: &DefaultResponseEncoder{},
			resp: func() Response {
				resp := Response{
					ResponseHeader: ResponseHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
						Error: "micro: 程序发生错误了",
					},
					Data: []byte("此程序正常返回响应"),
				}
				return resp
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bytes := tc.encoder.Encode(&tc.resp)
			resp, err := tc.encoder.Decode(bytes)
			require.NoError(t, err)
			assert.Equal(t, &tc.resp, resp)
		})
	}
}
