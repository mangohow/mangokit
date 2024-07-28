package test

import (
	"context"
	"github.com/mangohow/mangokit/transport/http"
)

type GreeterHTTPService interface {
	SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error)
	SayHello1(ctx context.Context, req *HelloRequest) error
	SayHello2(ctx context.Context) (*HelloResponse, error)
	SayHello3(ctx context.Context) error
	Test(ctx context.Context, req *Kinds) error
}

type HelloRequest struct {
	Name string `json:"name" form:"name" param:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

type Kinds struct {
	Kint     int     `json:"kint" param:"kint"`
	Kintp    *int    `json:"kintp" param:"kintp"`
	Kstring  string  `json:"kstring" param:"kstring"`
	Kstringp *string `json:"kstringp" param:"kstringp"`
	Kbool    bool    `json:"kbool" param:"kbool"`
	Kboolp   *bool   `json:"kboolp" param:"kboolp"`
}

func RegisterGreeterHTTPService(server *http.Server, svc GreeterHTTPService) {
	server.RegisterService(_GreeterHTTPService_serviceDesc, svc)
}

func _Greeter_Test_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	in := new(Kinds)
	err := dec(in)
	if err != nil {
		return nil, err
	}

	if middleware == nil {
		return nil, svc.(GreeterHTTPService).Test(ctx, in)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.(GreeterHTTPService).Test(ctx, in)
	}

	return middleware(ctx, in, handler)
}

func _Greeter_SayHello_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	in := new(HelloRequest)
	err := dec(in)
	if err != nil {
		return nil, err
	}

	if middleware == nil {
		return svc.(GreeterHTTPService).SayHello(ctx, in)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.(GreeterHTTPService).SayHello(ctx, in)
	}

	return middleware(ctx, in, handler)
}

func _Greeter_SayHello1_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	in := new(HelloRequest)
	err := dec(in)
	if err != nil {
		return nil, err
	}

	if middleware == nil {
		return nil, svc.(GreeterHTTPService).SayHello1(ctx, in)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.(GreeterHTTPService).SayHello1(ctx, in)
	}

	return middleware(ctx, in, handler)
}

func _Greeter_SayHello2_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	in := new(HelloRequest)
	err := dec(in)
	if err != nil {
		return nil, err
	}

	if middleware == nil {
		return svc.(GreeterHTTPService).SayHello2(ctx)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.(GreeterHTTPService).SayHello2(ctx)
	}

	return middleware(ctx, in, handler)
}

func _Greeter_SayHello3_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	in := new(HelloRequest)
	err := dec(in)
	if err != nil {
		return nil, err
	}

	if middleware == nil {
		return nil, svc.(GreeterHTTPService).SayHello3(ctx)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.(GreeterHTTPService).SayHello3(ctx)
	}

	return middleware(ctx, in, handler)
}

type GreeterHTTPClient interface {
	SayHello(ctx context.Context, request *HelloRequest, opts ...http.CallOption) (*HelloResponse, error)
	SayHello1(ctx context.Context, request *HelloRequest, opts ...http.CallOption) error
	SayHello2(ctx context.Context, opts ...http.CallOption) (*HelloResponse, error)
	SayHello3(ctx context.Context, opts ...http.CallOption) error
	Test(ctx context.Context, request *Kinds, opts ...http.CallOption) error
}

type greeterHTTPClient struct {
	cc *http.Client
}

func NewGreeterHTTPClient(client *http.Client) GreeterHTTPClient {
	return &greeterHTTPClient{cc: client}
}

func (c *greeterHTTPClient) SayHello(ctx context.Context, req *HelloRequest, opts ...http.CallOption) (*HelloResponse, error) {
	reply := new(HelloResponse)
	pattern := "/helloworld/:name"
	path := http.EncodeURL(pattern, req, true)
	_, err := c.cc.Invoke(ctx, "GET", path, req, reply, opts...)
	return reply, err
}

func (c *greeterHTTPClient) SayHello1(ctx context.Context, req *HelloRequest, opts ...http.CallOption) error {
	pattern := "/helloworld1"
	path := http.EncodeURLFromForm(pattern, req)
	_, err := c.cc.Invoke(ctx, "GET", path, req, nil, opts...)
	return err
}

func (c *greeterHTTPClient) SayHello2(ctx context.Context, opts ...http.CallOption) (*HelloResponse, error) {
	reply := new(HelloResponse)
	path := "/helloworld2"
	_, err := c.cc.Invoke(ctx, "GET", path, nil, reply, opts...)
	return reply, err
}

func (c *greeterHTTPClient) SayHello3(ctx context.Context, opts ...http.CallOption) error {
	path := "/helloworld3"
	_, err := c.cc.Invoke(ctx, "GET", path, nil, nil, opts...)
	return err
}

func (c *greeterHTTPClient) Test(ctx context.Context, req *Kinds, opts ...http.CallOption) error {
	pattern := "/test/:kint/:kintp/:kstring/:kstringp/:kbool/:kboolp"
	path := http.EncodeURL(pattern, req, false)
	_, err := c.cc.Invoke(ctx, "GET", path, req, nil, opts...)
	return err
}

var _GreeterHTTPService_serviceDesc = &http.ServiceDesc{
	HandlerType: (*GreeterHTTPService)(nil),
	Methods: []http.MethodDesc{
		{
			Method:  "GET",
			Path:    "/helloworld/:name",
			Handler: _Greeter_SayHello_HTTP_Handler,
		},
		{
			Method:  "GET",
			Path:    "/helloworld1",
			Handler: _Greeter_SayHello1_HTTP_Handler,
		},
		{
			Method:  "GET",
			Path:    "/helloworld2",
			Handler: _Greeter_SayHello2_HTTP_Handler,
		},
		{
			Method:  "GET",
			Path:    "/helloworld3/",
			Handler: _Greeter_SayHello3_HTTP_Handler,
		},
		{
			Method:  "GET",
			Path:    "/test/:kint/:kintp/:kstring/:kstringp/:kbool/:kboolp",
			Handler: _Greeter_Test_HTTP_Handler,
		},
	},
}
