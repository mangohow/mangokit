package http

import (
	"context"
)

type Handler func(ctx context.Context, req interface{}) (resp interface{}, err error)

type Middleware func(ctx context.Context, req interface{}, handler Handler) (interface{}, error)

type methodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, middleware Middleware) (interface{}, error)

type ServiceDesc struct {
	HandlerType interface{}
	Methods     []MethodDesc
}

type MethodDesc struct {
	Method  string
	Path    string
	Handler methodHandler
}
