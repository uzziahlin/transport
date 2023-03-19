package rpc

import (
	"context"
	"github.com/uzziahlin/transport/rpc/errs"
	"log"
	"net"
	"time"
)

type Client interface {
	Send(ctx context.Context, data []byte) ([]byte, error)
}

func NewRpcClient(addr string) *DefaultClient {
	pool := &ConnPool[net.Conn]{
		idleConns:   make(chan *Conn[net.Conn], 10),
		maxActive:   20,
		maxIdleTime: 15 * time.Second,
		waitQ:       make(map[uint64]chan *Conn[net.Conn], 16),
		factory: func() net.Conn {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Fatalf("连接创建失败")
				return nil
			}
			return conn
		},
	}

	client := &DefaultClient{
		read: RpcReader,
		pool: pool,
	}

	return client
}

type DefaultClient struct {
	read Reader
	pool Pool[net.Conn]
}

func (r DefaultClient) Send(ctx context.Context, data []byte) ([]byte, error) {
	conn, err := r.pool.Get(ctx)
	defer r.pool.Put(ctx, conn)

	if err != nil {
		return nil, err
	}
	_, err = conn.Write(data)

	if err != nil {
		return nil, err
	}

	if isOneway(ctx) {
		return nil, errs.ErrOneway
	}

	return r.read(conn)
}
