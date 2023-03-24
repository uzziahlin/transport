package http

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func NewDefaultClient(opts ...ClientOption) Client {

	c := &DefaultClient{
		client: &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.proxyUrl != nil {
		c.client.Transport = &http.Transport{
			Proxy: http.ProxyURL(c.proxyUrl),
		}
	}

	return c
}

type ClientOption func(*DefaultClient)

func WithProxy(proxyUrl *url.URL) ClientOption {
	return func(c *DefaultClient) {
		c.proxyUrl = proxyUrl
	}
}

type DefaultClient struct {
	client   *http.Client
	proxyUrl *url.URL
}

func (c DefaultClient) Get(ctx context.Context, url string) (*Response, error) {

	return c.Send(ctx, &Request{
		Method: http.MethodGet,
		Url:    url,
	})
}

func (c DefaultClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*Response, error) {

	req := &Request{
		Method: http.MethodPost,
		Url:    url,
		Body:   io.NopCloser(body),
	}

	req.Header.Set("Content-Type", contentType)

	return c.Send(ctx, req)
}

func (c DefaultClient) Send(ctx context.Context, req *Request) (*Response, error) {
	request, err := http.NewRequestWithContext(ctx, req.Method, req.Url, req.Body)

	if err != nil {
		return nil, err
	}

	request.Header = http.Header(req.Header)

	resp, err := c.client.Do(request)

	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Header:     Header(resp.Header),
		Body:       resp.Body,
	}, nil
}

func (c DefaultClient) Close() error {

	c.client.CloseIdleConnections()

	return nil
}
