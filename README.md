# 基于go语言的http客户端及rpc框架实现

## 1. http客户端
http客户端目前仅有基于go sdk的实现，用户也扩展自己的实现，只要实现了`Client`接口即可。
使用实例如下：
```go
package main

import (
    "context"
    "github.com/uzziahlin/transport/http"
)

func main() {
	client := http.NewDefaultClient()
	resp, err := client.Send(context.TODO(), &http.Request{
		Method: "GET",
		Url: "https://www.baidu.com?p1=v1&p2=v2",
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"name":"test"}`),
	})
	if err != nil {
		// todo handle error
	}
	defer resp.Body.Close()

	// todo handle response
}
```
或者使用更加易用的Get或者Post方法， 其中Get方法使用如下：
```go
package main

import (
    "context"
    "github.com/uzziahlin/transport/http"
)

func main() {
	client := http.NewDefaultClient()
	resp, err := client.Get(context.TODO(), "https://www.baidu.com?p1=v1&p2=v2")
	if err != nil {
		// todo handle error
	}
	defer resp.Body.Close()

	// todo handle response
}
```

## 2. rpc框架
提供一套易用的api，使用户可以简单使用rpc框架， 支持压缩、oneway调用语义等功能。
### 2.1 rpc客户端
`rpc客户端`主要采用篡改结构体的函数属性实现，并且要代理的结构体必须实现Service接口

```go
package main

import (
	"context"
	"github.com/uzziahlin/transport/rpc"
)

func main() {
    constructor := rpc.NewProxyConstructor()

    s := &ClientServiceImpl{}
	
    constructor.InitProxy(s)

    resp, err := s.Action(context.TODO(), &ActionReq{
        Name: "test",
    })
	
    if err != nil {
        // todo handle error
    }
	
    // todo handle response

}

type ClientServiceImpl struct {
	Action func(ctx context.Context, req *ActionReq) (*ActionResp, error)
}

func (s ServiceImpl) Info() rpc.ServiceInfo {
	return rpc.ServiceInfo{
		Name: "test",
		Addr: "localhost:8080",
	}
}

type ActionReq struct {
	Name string `json:"name"`
}

type ActionResp struct {
	Code int `json:"code"`
}

```

### 2.2 rpc服务端
rpc服务端使用实例如下：
```go
package main

import (
    "context"
    "github.com/uzziahlin/transport/rpc"
)

func main() {
    ep := rpc.NewEndpoint("localhost:8080")
    
    ep.Register(&ServerServiceImpl{})
    
    ep.Startup()
}

type ServerServiceImpl struct {
    
}

func (s ServiceImpl) Info() rpc.ServiceInfo {
    return rpc.ServiceInfo{
        Name: "test",
    }
}

func (s ServerServiceImpl) Action(ctx context.Context, req *ActionReq) (*ActionResp, error) {
    return &ActionResp{
        Code: 0,
    }, nil
}
```
