package httpwrapper

import (
	"context"
)

type Middleware func(ctx context.Context, req interface{}, next NextHandler) error

type methodHandler func(srv interface{}, middleware Middleware) Middleware

type NextHandler func(context.Context, interface{}) error

type ServiceDesc struct {
	HandlerType interface{}
	Methods     []MethodDesc
}

type MethodDesc struct {
	Method  string
	Path    string
	Handler methodHandler
}
