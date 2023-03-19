package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestEncoder_EncodeDecode(t *testing.T) {
	testCases := []struct {
		name    string
		encoder RequestEncoder
		req     Request
	}{
		{
			name:    "default no meta and no data",
			encoder: &DefaultRequestEncoder{},
			req: func() Request {
				req := Request{
					RequestHeader: RequestHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
						ServiceName: "user-service",
						MethodName:  "GetById",
					},
				}

				return req
			}(),
		},
		{
			name:    "default no data",
			encoder: &DefaultRequestEncoder{},
			req: func() Request {
				req := Request{
					RequestHeader: RequestHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
						ServiceName: "user-service",
						MethodName:  "GetById",
						Meta: map[string]string{
							"k1": "v1",
							"k2": "v2",
						},
					},
				}

				return req
			}(),
		},
		{
			name:    "default no meta",
			encoder: &DefaultRequestEncoder{},
			req: func() Request {
				req := Request{
					RequestHeader: RequestHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
						ServiceName: "user-service",
						MethodName:  "GetById",
					},
					Data: []byte("hello world!"),
				}

				return req
			}(),
		},
		{
			name:    "default",
			encoder: &DefaultRequestEncoder{},
			req: func() Request {
				req := Request{
					RequestHeader: RequestHeader{
						Header: Header{
							MessageId:  1234,
							Version:    3,
							Compressor: 2,
							Serializer: 3,
						},
						ServiceName: "user-service",
						MethodName:  "GetById",
						Meta: map[string]string{
							"k1": "v1",
							"k2": "v2",
						},
					},
					Data: []byte("hello world!"),
				}

				return req
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bytes := tc.encoder.Encode(&tc.req)
			req, err := tc.encoder.Decode(bytes)
			require.NoError(t, err)
			assert.Equal(t, &tc.req, req)
		})
	}
}
