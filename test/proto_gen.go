package test

import (
	"context"
	"github.com/mangohow/mangokit/serialize"
	"github.com/mangohow/mangokit/tools"
	"github.com/mangohow/mangokit/transport/httpwrapper"
	"net/http"
)

type GreeterHTTPService interface {
	SayHello(ctx context.Context, request *HelloRequest) (*HelloResponse, error)
}

type HelloRequest struct {
	Name string `json:"name" uri:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

func RegisterGreeterHTTPService(server *httpwrapper.Server, svc GreeterHTTPService) {
	server.RegisterService(_GreeterHTTPService_serviceDesc, svc)
}

func _Greeter_SayHello_HTTP_Handler(svc interface{}, middleware httpwrapper.Middleware) httpwrapper.Middleware {
	return func(ctx context.Context, req interface{}, next httpwrapper.NextHandler) error {
		in := new(HelloRequest)
		err := tools.BindVar(ctx, in)
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

var _GreeterHTTPService_serviceDesc = &httpwrapper.ServiceDesc{
	HandlerType: (*GreeterHTTPService)(nil),
	Methods: []httpwrapper.MethodDesc{
		{
			Method:  "GET",
			Path:    "/helloworld/:name",
			Handler: _Greeter_SayHello_HTTP_Handler,
		},
	},
}
