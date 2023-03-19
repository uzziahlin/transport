package rpc

import (
	"context"
	"errors"
	"github.com/uzziahlin/transport/rpc/compress"
	"github.com/uzziahlin/transport/rpc/errs"
	"github.com/uzziahlin/transport/rpc/proto/user_service/gen"
	"github.com/uzziahlin/transport/rpc/serialize/proto"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestProxyConstructor_InitProxy(t *testing.T) {

	endpoint := NewEndPoint(":8081")

	endpoint.Register(&UserServiceImpl{})

	go func() {
		err := endpoint.Startup()
		require.NoError(t, err)
	}()

	time.Sleep(3 * time.Second)

	constructor := NewProxyConstructor()

	userService := &UserService{}

	err := constructor.InitProxy(userService)

	require.NoError(t, err)

	resp, err := userService.GetById(context.Background(), &UserReq{
		Id: "this is the user id",
	})

	require.NoError(t, err)

	// fmt.Print(resp.Content)
	t.Log(resp.Content)
}

func TestProxyConstructor_Oneway(t *testing.T) {

	endpoint := NewEndPoint(":8081")

	endpoint.Register(&UserServiceImpl{})

	go func() {
		err := endpoint.Startup()
		require.NoError(t, err)
	}()

	time.Sleep(3 * time.Second)

	constructor := NewProxyConstructor()

	userService := &UserService{}

	err := constructor.InitProxy(userService)

	require.NoError(t, err)

	ctx := OnewayContext(context.Background())

	resp, err := userService.GetById(ctx, &UserReq{
		Id: "this is the user id",
	})

	assert.Equal(t, errs.ErrOneway, err)

	// fmt.Print(resp.Content)
	t.Log(resp.Content)
}

func TestProxyConstructor_Timeout(t *testing.T) {

	endpoint := NewEndPoint(":8081")

	endpoint.Register(&UserServiceTimeout{})

	go func() {
		err := endpoint.Startup()
		require.NoError(t, err)
	}()

	time.Sleep(3 * time.Second)

	constructor := NewProxyConstructor()

	userService := &UserService{}

	err := constructor.InitProxy(userService)

	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := userService.GetById(ctx, &UserReq{
		Id: "this is the user id",
	})

	assert.Equal(t, context.DeadlineExceeded, err)

	// fmt.Print(resp.Content)
	t.Log(resp.Content)
}

func TestProxyConstructor_Err(t *testing.T) {

	endpoint := NewEndPoint(":8081")

	service := &UserServiceErr{}

	service.Msg = "this is the msg"
	service.Err = "this is the err"

	endpoint.Register(service)

	go func() {
		err := endpoint.Startup()
		require.NoError(t, err)
	}()

	time.Sleep(3 * time.Second)

	constructor := NewProxyConstructor()

	userService := &UserService{}

	err := constructor.InitProxy(userService)

	require.NoError(t, err)

	resp, err := userService.GetById(context.Background(), &UserReq{
		Id: "this is the user id",
	})

	assert.Equal(t, errors.New("this is the err"), err)

	// fmt.Print(resp.Content)
	t.Log(resp.Content)
}

func TestProxyConstructor_WithProtoSerializer(t *testing.T) {

	endpoint := NewEndPoint(":8081")

	endpoint.RegisterSerializer(&proto.Serializer{})

	service := &UserServiceProto{}

	service.Msg = "this is the msg"
	service.Err = "this is the err"

	endpoint.Register(service)

	go func() {
		err := endpoint.Startup()
		require.NoError(t, err)
	}()

	time.Sleep(3 * time.Second)

	constructor := NewProxyConstructor(WithSerializer(&proto.Serializer{}))

	userService := &UserService{}

	err := constructor.InitProxy(userService)

	require.NoError(t, err)

	resp, err := userService.GetByIdProto(context.Background(), &gen.UserReq{
		Id: "this is the user id",
	})

	assert.Equal(t, errors.New("this is the err"), err)

	// fmt.Print(resp.Content)
	t.Log(resp.Msg)
}

func TestProxyConstructor_WithCompressor(t *testing.T) {

	endpoint := NewEndPoint(":8081")

	service := &UserServiceProto{}

	service.Msg = "this is the msg"
	service.Err = "this is the err"

	endpoint.Register(service)

	go func() {
		err := endpoint.Startup()
		require.NoError(t, err)
	}()

	time.Sleep(3 * time.Second)

	constructor := NewProxyConstructor(WithSerializer(&proto.Serializer{}))

	userService := &UserService{}

	err := constructor.InitProxy(userService)

	require.NoError(t, err)

	ctx := compress.Context(context.Background(), compress.GZIP)

	resp, err := userService.GetByIdProto(ctx, &gen.UserReq{
		Id: "this is the user id",
	})

	assert.Equal(t, errors.New("this is the err"), err)

	// fmt.Print(resp.Content)
	t.Log(resp.Msg)
}

type UserService struct {
	GetById func(ctx context.Context, req *UserReq) (*UserResp, error)

	GetByIdProto func(ctx context.Context, req *gen.UserReq) (*gen.UserResp, error)
}

type UserReq struct {
	Id string
}

type UserResp struct {
	Content string
}

func (u UserService) Info() ServiceInfo {
	return ServiceInfo{
		ServiceName: "user-service",
		Addr:        "localhost:8081",
	}
}

type UserServiceImpl struct {
}

func (u *UserServiceImpl) GetById(ctx context.Context, req *UserReq) (*UserResp, error) {
	id := req.Id

	return &UserResp{
		Content: "response: " + id,
	}, nil
}

func (u *UserServiceImpl) Info() ServiceInfo {
	return ServiceInfo{
		ServiceName: "user-service",
	}
}

type UserServiceTimeout struct {
	Msg string
	Err string
}

func (u *UserServiceTimeout) GetById(ctx context.Context, req *UserReq) (*UserResp, error) {
	time.Sleep(3 * time.Second)
	return &UserResp{
		Content: u.Msg,
	}, errors.New(u.Err)
}

func (u *UserServiceTimeout) Info() ServiceInfo {
	return ServiceInfo{
		ServiceName: "user-service",
	}
}

type UserServiceErr struct {
	Msg string
	Err string
}

func (u *UserServiceErr) GetById(ctx context.Context, req *UserReq) (*UserResp, error) {
	return &UserResp{
		Content: u.Msg,
	}, errors.New(u.Err)
}

func (u *UserServiceErr) Info() ServiceInfo {
	return ServiceInfo{
		ServiceName: "user-service",
	}
}

type UserServiceProto struct {
	Msg string
	Err string
}

func (u *UserServiceProto) GetByIdProto(ctx context.Context, req *gen.UserReq) (*gen.UserResp, error) {
	return &gen.UserResp{
		Msg: u.Msg,
	}, errors.New(u.Err)
}

func (u *UserServiceProto) Info() ServiceInfo {
	return ServiceInfo{
		ServiceName: "user-service",
	}
}
