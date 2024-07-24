package test

import (
	"context"
	"github.com/mangohow/mangokit/serialize"
	"github.com/mangohow/mangokit/tools"
	http "github.com/mangohow/mangokit/transport/http"
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

func RegisterGreeterHTTPService(server *http.Server, svc GreeterHTTPService) {
	server.RegisterService(_GreeterHTTPService_serviceDesc, svc)
}

func _Greeter_SayHello_HTTP_Handler(svc interface{}, middleware http.Middleware) http.Middleware {
	return func(ctx context.Context, req interface{}, next http.NextHandler) error {
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

var _GreeterHTTPService_serviceDesc = &http.ServiceDesc{
	HandlerType: (*GreeterHTTPService)(nil),
	Methods: []http.MethodDesc{
		{
			Method:  "GET",
			Path:    "/helloworld/:name",
			Handler: _Greeter_SayHello_HTTP_Handler,
		},
	},
}
