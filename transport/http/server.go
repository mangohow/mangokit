package http

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/mangokit/errors"
	"github.com/mangohow/mangokit/serialize"
	"github.com/sirupsen/logrus"
)

type Server struct {
	server *http.Server
	router *gin.Engine
	addr   string

	log       *logrus.Logger
	errorFunc EncodeErrorFunc

	middlewares []Middleware

	ctx context.Context
}

type HandlerFunc func(c *gin.Context) error

// EncodeErrorFunc 错误处理函数
type EncodeErrorFunc func(ctx *gin.Context, err error, log *logrus.Logger)

// DefaultEncodeErrorFunc 默认错误处理函数
func DefaultEncodeErrorFunc(ctx *gin.Context, err error, log *logrus.Logger) {
	e, ok := err.(errors.Error)
	if !ok {
		e = errors.FromError(errors.UnknownCode, errors.DefaultStatus, errors.UnknownReason, errors.UnknownMessage, err)
	}

	ctx.JSON(int(e.HttpStatus()), serialize.Response{
		Error: e,
	})
	log.Error(e.Error())

	return
}

type Option func(s *Server)

func WithAddr(addr string) Option {
	return func(s *Server) {
		if addr == "" {
			addr = ":8080"
		}
		s.addr = addr
	}
}

func WithRouter(router *gin.Engine) Option {
	return func(s *Server) {
		s.router = router
	}
}

func WithEncodeErrorFunc(fn EncodeErrorFunc) Option {
	return func(s *Server) {
		s.errorFunc = fn
	}
}

func WithLogger(log *logrus.Logger) Option {
	return func(s *Server) {
		s.log = log
	}
}

func WithContext(ctx context.Context) Option {
	return func(s *Server) {
		s.ctx = ctx
	}
}

func New(opts ...Option) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}

	if s.router == nil {
		s.router = gin.Default()
	}

	s.server = &http.Server{
		Handler: s.router,
	}

	if s.errorFunc == nil {
		s.errorFunc = DefaultEncodeErrorFunc
	}

	if s.log == nil {
		s.log = logrus.StandardLogger()
	}

	if s.ctx == nil {
		s.ctx = context.Background()
	}

	if s.addr == "" {
		s.addr = ":8000"
	}
	s.server.Addr = s.addr

	if s.log == nil {
		s.log = logrus.StandardLogger()
	}

	return s
}

func (s *Server) HttpServer() *http.Server {
	return s.server
}

func (s *Server) GinEngine() *gin.Engine {
	return s.router
}

func (s *Server) RegisterService(sd *ServiceDesc, srv interface{}) {
	if srv != nil {
		ht := reflect.TypeOf(sd.HandlerType).Elem()
		st := reflect.TypeOf(srv)
		if !st.Implements(ht) {
			s.log.Fatalf("handler type %v not implement %v", st, ht)
		}
	}

	s.register(sd, srv)
}

func (s *Server) register(sd *ServiceDesc, srv interface{}) {
	for _, d := range sd.Methods {
		handler := d.Handler
		s.handle(d.Method, d.Path, handler(srv, chainHandler(s.middlewares)))
	}
}

func chainHandler(middlewares []Middleware) Middleware {
	if len(middlewares) == 0 {
		return nil
	}

	return func(ctx context.Context, req interface{}, handler NextHandler) error {
		return middlewares[0](ctx, req, getChainMiddleware(middlewares, 0, handler))
	}
}

func getChainMiddleware(middlewares []Middleware, cur int, handler NextHandler) NextHandler {
	if cur >= len(middlewares)-1 {
		return handler
	}

	return func(ctx context.Context, req interface{}) error {
		return middlewares[cur+1](ctx, req, getChainMiddleware(middlewares, cur+1, handler))
	}
}

func (s *Server) handle(method, relativePath string, handler Middleware) {
	s.router.Handle(method, relativePath, s.handlerConvert(handler))
}

func (s *Server) handlerConvert(handler Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(s.ctx, "gin-ctx", c)
		err := handler(ctx, nil, nil)
		if err != nil && s.errorFunc != nil {
			s.errorFunc(c, err, s.log)
		}
	}
}

func (s *Server) Middleware(middleware ...Middleware) {
	s.middlewares = append(s.middlewares, middleware...)
}

func (s *Server) Start() error {
	s.log.Info("server listen at ", s.addr)
	err := s.server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
