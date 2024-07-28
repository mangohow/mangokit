// Code generated by protoc-gen-go-gin. DO NOT EDIT.
// versions:
// - protoc-gen-go-gin v1.0.0
// - protoc             v5.26.1
// source: test_gen_gin/test.proto

package test

import (
	"context"

	"github.com/mangohow/mangokit/transport/http"
)

type GreeterHTTPService interface {
	SayHello(context.Context, *GreeterRequest) (*GreeterResponse, error)
	SayHelloEmptyRequest(context.Context) (*GreeterResponse, error)
	SayHelloEmptyResponse(context.Context, *GreeterRequest) error
	SayHelloEmpty(context.Context) error
}

func RegisterGreeterHTTPService(server *http.Server, svc GreeterHTTPService) {
	server.RegisterService(_GreeterHTTPService_serviceDesc, svc)
}

func _Greeter_SayHello_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	in := new(GreeterRequest)
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

func _Greeter_SayHelloEmptyRequest_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	if middleware == nil {
		return svc.(GreeterHTTPService).SayHelloEmptyRequest(ctx)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.(GreeterHTTPService).SayHelloEmptyRequest(ctx)
	}

	return middleware(ctx, nil, handler)

}

func _Greeter_SayHelloEmptyResponse_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	in := new(GreeterRequest)
	err := dec(in)
	if err != nil {
		return nil, err
	}

	if middleware == nil {
		return nil, svc.(GreeterHTTPService).SayHelloEmptyResponse(ctx, in)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.(GreeterHTTPService).SayHelloEmptyResponse(ctx, in)
	}

	return middleware(ctx, in, handler)

}

func _Greeter_SayHelloEmpty_HTTP_Handler(svc interface{}, ctx context.Context, dec func(interface{}) error, middleware http.Middleware) (interface{}, error) {
	if middleware == nil {
		return nil, svc.(GreeterHTTPService).SayHelloEmpty(ctx)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.(GreeterHTTPService).SayHelloEmpty(ctx)
	}

	return middleware(ctx, nil, handler)

}

var _GreeterHTTPService_serviceDesc = &http.ServiceDesc{
	HandlerType: (*GreeterHTTPService)(nil),
	Methods: []http.MethodDesc{
		{
			Method:  "GET",
			Path:    "/hello/:name",
			Handler: _Greeter_SayHello_HTTP_Handler,
		},
		{
			Method:  "GET",
			Path:    "/hello1",
			Handler: _Greeter_SayHelloEmptyRequest_HTTP_Handler,
		},
		{
			Method:  "GET",
			Path:    "/hello2/:name",
			Handler: _Greeter_SayHelloEmptyResponse_HTTP_Handler,
		},
		{
			Method:  "GET",
			Path:    "/hello3",
			Handler: _Greeter_SayHelloEmpty_HTTP_Handler,
		},
	},
}
