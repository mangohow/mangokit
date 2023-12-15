package httpwrapper

import (
	"context"
	"net/http"

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
	fnt       DecodeFieldNameType
}

type HandlerFunc func(ctx *Context) error

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

func WithDecodeFieldNameType(t DecodeFieldNameType) Option {
	return func(s *Server) {
		s.fnt = t
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

func (s *Server) handlersConvert(handlers ...HandlerFunc) (hs []gin.HandlerFunc) {
	hs = make([]gin.HandlerFunc, 0, len(handlers))
	for i := 0; i < len(handlers); i++ {
		handler := handlers[i]
		hs = append(hs, func(ctx *gin.Context) {
			c := ctxPool.Get().(*Context)
			c.Context = ctx
			c.fnt = s.fnt
			err := handler(c)
			if err != nil {
				s.errorFunc(ctx, err, s.log)
			}

			c.clear()
			putPool(c)
		})
	}

	return
}

func (s *Server) Handle(method, relativePath string, handlers ...HandlerFunc) {
	s.router.Handle(method, relativePath, s.handlersConvert(handlers...)...)
}

func (s *Server) GET(relativePath string, handlers ...HandlerFunc) {
	s.Handle(http.MethodGet, relativePath, handlers...)
}

func (s *Server) POST(relativePath string, handlers ...HandlerFunc) {
	s.Handle(http.MethodPost, relativePath, handlers...)
}

func (s *Server) DELETE(relativePath string, handlers ...HandlerFunc) {
	s.Handle(http.MethodDelete, relativePath, handlers...)
}

func (s *Server) PATCH(relativePath string, handlers ...HandlerFunc) {
	s.Handle(http.MethodPatch, relativePath, handlers...)
}

func (s *Server) PUT(relativePath string, handlers ...HandlerFunc) {
	s.Handle(http.MethodPut, relativePath, handlers...)
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
