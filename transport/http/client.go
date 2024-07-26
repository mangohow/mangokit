package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	neturl "net/url"
)

// Client http client
type Client struct {
	client *http.Client
	config config
}

type config struct {
	host         string
	transport    http.RoundTripper
	interceptors []Interceptor
}

// Interceptor 拦截器
type Interceptor func() error

type ClientOption func(*config)

func NewClient(options ...ClientOption) (*Client, error) {
	c := &Client{}
	for _, option := range options {
		option(&c.config)
	}

	c.client.Transport = c.config.transport

	if c.config.host == "" {
		return nil, errors.New("host is required, use WithHost option to set host")
	}

	return c, nil
}

func WithHost(host string) ClientOption {
	return func(c *config) {
		c.host = host
	}
}

func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *config) {
		c.transport = transport
	}
}

func WithInterceptors(interceptors ...Interceptor) ClientOption {
	return func(c *config) {
		c.interceptors = append(c.interceptors, interceptors...)
	}
}

type beforeCallInfo struct {
	contentType string
	header      http.Header
	value       interface{}
}

type afterCallInfo struct {
	resp   interface{}
	status int
}

// CallOption 在请求被调用前和调用后执行的handler
type CallOption interface {
	Before(call *beforeCallInfo)

	After(call *afterCallInfo)
}

type EmptyCallOptions struct{}

func (EmptyCallOptions) Before(call *beforeCallInfo) {}
func (EmptyCallOptions) After(call *afterCallInfo)   {}

type contentTypeCallOption struct {
	EmptyCallOptions
	contentType string
}

func (c contentTypeCallOption) Before(info *beforeCallInfo) {
	info.contentType = c.contentType
}

// ContentType 为请求设置content type
func ContentType(contentType string) CallOption {
	return contentTypeCallOption{contentType: contentType}
}

type headerCallOption struct {
	EmptyCallOptions
	headers http.Header
}

func (c headerCallOption) Before(info *beforeCallInfo) {
	info.header = c.headers
}

// Headers 为请求设置headers
func Headers(headers http.Header) CallOption {
	return headerCallOption{headers: headers}
}

// Invoke 先执行全局拦截器，再执行CallOption中的before，最后再发起请求
func (c *Client) Invoke(ctx context.Context, method, path string, req, resp interface{}, opts ...CallOption) (status int, err error) {
	var url string

	url, err = neturl.JoinPath(c.config.host, path)
	if err != nil {
		return
	}

	bco := &beforeCallInfo{
		header: make(http.Header),
		value:  req,
	}
	for _, opt := range opts {
		opt.Before(bco)
	}

	var bodyReader io.Reader
	if req != nil {
		bodyBytes, err := json.Marshal(req)
		if err != nil {
			return
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}
	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return
	}

	if bco.contentType == "" {
		request.Header.Set("Content-Type", "application/json")
	} else {
		request.Header.Set("Content-Type", bco.contentType)
	}
	for k, v := range bco.header {
		request.Header.Set(k, v[0])
	}

	response, err := c.client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	status = response.StatusCode

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	if resp != nil {
		if err = json.Unmarshal(respBytes, resp); err != nil {
			return
		}
	}

	aco := &afterCallInfo{
		resp:   resp,
		status: status,
	}

	for _, opt := range opts {
		opt.After(aco)
	}

	return
}
