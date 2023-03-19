package rpc

import "net"

type ConnHandler func(conn net.Conn)

func NewServer(addr string, handler ConnHandler) *Server {
	server := &Server{
		addr:    addr,
		handler: handler,
	}

	return server
}

type Server struct {
	addr    string
	handler ConnHandler
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)

	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			return err
		}

		go func() {
			s.handler(conn)
		}()
	}
}
