package http

import (
	"context"
	"io"
	"net/http"
)

func NewDefaultClient() Client {
	return &DefaultClient{
		client: &http.Client{},
	}
}

type DefaultClient struct {
	client *http.Client
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
		Body:   body,
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