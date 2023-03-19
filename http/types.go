package http

import (
	"context"
	"io"
)

type Header map[string][]string

func (h Header) Set(key, value string) {
	h[key] = []string{value}
}

type Getter interface {
	Get(ctx context.Context, url string) (*Response, error)
}

type Poster interface {
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*Response, error)
}

type Client interface {
	// Send a request and return a response.
	// The caller must close the response body.
	Send(ctx context.Context, req *Request) (*Response, error)

	// Getter Get issues a GET to the specified URL.
	Getter

	Poster

	// Closer Close closes the client, releasing any open resources.
	io.Closer
}

type Request struct {
	Method string
	Url    string
	Header Header
	Body   io.Reader
}

func (r *Request) SetHeader(key, value string) {
	if r.Header == nil {
		r.Header = make(Header)
	}
	r.Header.Set(key, value)
}

type Response struct {
	StatusCode int
	Header     Header
	Body       io.ReadCloser
}
