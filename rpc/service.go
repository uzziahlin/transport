package rpc

type Service interface {
	Info() ServiceInfo
}

type ServiceInfo struct {
	ServiceName string
	Addr        string
}
