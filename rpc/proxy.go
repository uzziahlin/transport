package rpc

import (
	"context"
	"github.com/uzziahlin/transport/rpc/message"
)

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}

func NewRemoteProxy(addr string) *RemoteProxy {
	return &RemoteProxy{
		client:      NewRpcClient(addr),
		reqEncoder:  &message.DefaultRequestEncoder{},
		respEncoder: &message.DefaultResponseEncoder{},
	}
}

type RemoteProxy struct {
	client      Client
	reqEncoder  message.RequestEncoder
	respEncoder message.ResponseEncoder
}

func (r *RemoteProxy) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {

	encodedReq := r.reqEncoder.Encode(req)

	resp, err := r.client.Send(ctx, encodedReq)

	if err != nil {
		return nil, err
	}

	return r.respEncoder.Decode(resp)
}
