// Code generated by protoc-gen-go-gin. DO NOT EDIT.
// versions:
// - protoc-gen-go-gin v1.0.0
// - protoc             v5.26.1
// source: test_gen_gin/test.proto

package test

import (
	"context"

	"github.com/mangohow/mangokit/serialize"
	"github.com/mangohow/mangokit/tools"
	http "github.com/mangohow/mangokit/transport/http"
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

func _Greeter_SayHello_HTTP_Handler(svc interface{}, middleware http.Middleware) http.Middleware {
	return func(ctx context.Context, req interface{}, next http.NextHandler) error {
		in := new(GreeterRequest)
		err := http.BindVar(ctx, in)
		if err != nil {
			return err
		}

		handler := func(ctx context.Context, req interface{}) error {
			ctxt := tools.GinCtxFromContext(ctx)
			reply, err := svc.(GreeterHTTPService).SayHello(ctx, in)
			if err != nil {
				return err
			}
			ctxt.JSON(http.StatusOK, serialize.Response{Data: reply})

			return nil
		}

		if middleware == nil {
			return handler(ctx, in)
		}

		return middleware(ctx, in, handler)
	}
}

func _Greeter_SayHelloEmptyRequest_HTTP_Handler(svc interface{}, middleware http.Middleware) http.Middleware {
	return func(ctx context.Context, req interface{}, next http.NextHandler) error {
		handler := func(ctx context.Context, req interface{}) error {
			ctxt := tools.GinCtxFromContext(ctx)
			reply, err := svc.(GreeterHTTPService).SayHelloEmptyRequest(ctx)
			if err != nil {
				return err
			}
			ctxt.JSON(http.StatusOK, serialize.Response{Data: reply})

			return nil
		}

		if middleware == nil {
			return handler(ctx, nil)
		}

		return middleware(ctx, nil, handler)
	}
}

func _Greeter_SayHelloEmptyResponse_HTTP_Handler(svc interface{}, middleware http.Middleware) http.Middleware {
	return func(ctx context.Context, req interface{}, next http.NextHandler) error {
		in := new(GreeterRequest)
		err := http.BindVar(ctx, in)
		if err != nil {
			return err
		}

		handler := func(ctx context.Context, req interface{}) error {
			ctxt := tools.GinCtxFromContext(ctx)
			err := svc.(GreeterHTTPService).SayHelloEmptyResponse(ctx, in)
			if err != nil {
				return err
			}
			ctxt.Status(http.StatusOK)

			return nil
		}

		if middleware == nil {
			return handler(ctx, in)
		}

		return middleware(ctx, in, handler)
	}
}

func _Greeter_SayHelloEmpty_HTTP_Handler(svc interface{}, middleware http.Middleware) http.Middleware {
	return func(ctx context.Context, req interface{}, next http.NextHandler) error {
		handler := func(ctx context.Context, req interface{}) error {
			ctxt := tools.GinCtxFromContext(ctx)
			err := svc.(GreeterHTTPService).SayHelloEmpty(ctx)
			if err != nil {
				return err
			}
			ctxt.Status(http.StatusOK)

			return nil
		}

		if middleware == nil {
			return handler(ctx, nil)
		}

		return middleware(ctx, nil, handler)
	}
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
