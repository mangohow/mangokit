package http

import (
	"context"
	stderr "errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/mangokit/errors"
	"github.com/mangohow/mangokit/serialize"
	"github.com/sirupsen/logrus"
)

const (
	ParamKey = "param"
	FormKey  = "form"
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
		s.handle(d.Method, d.Path, func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			return handler(srv, ctx, reqDecoder(ctx), chainHandler(s.middlewares))
		})
	}
}

func chainHandler(middlewares []Middleware) Middleware {
	if len(middlewares) == 0 {
		return nil
	}

	return func(ctx context.Context, req interface{}, handler Handler) (interface{}, error) {
		return middlewares[0](ctx, req, getChainMiddleware(middlewares, 0, handler))
	}
}

func getChainMiddleware(middlewares []Middleware, cur int, handler Handler) Handler {
	if cur >= len(middlewares)-1 {
		return handler
	}

	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return middlewares[cur+1](ctx, req, getChainMiddleware(middlewares, cur+1, handler))
	}
}

func (s *Server) handle(method, relativePath string, handler Handler) {
	s.router.Handle(method, relativePath, s.handlerConvert(handler))
}

func (s *Server) handlerConvert(handler Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(s.ctx, "gin-ctx", c)
		resp, err := handler(ctx, nil)
		if err != nil && s.errorFunc != nil {
			s.errorFunc(c, err, s.log)
			return
		}
		c.JSON(http.StatusOK, resp)
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

func BindVar(ctx context.Context, val interface{}) error {
	c := ctx.Value("gin-ctx").(*gin.Context)
	if err := bindParam(c, val); err != nil {
		return err
	}
	return c.ShouldBind(val)
}

// 解析路径/xxx/:yyy 中的参数
func bindParam(ctx *gin.Context, val interface{}) error {
	if len(ctx.Params) == 0 || !strings.Contains(ctx.FullPath(), ":") {
		return nil
	}

	params := ctx.Params
	v := reflect.ValueOf(val)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	if v.Type().Kind() != reflect.Struct {
		return stderr.New("must be struct")
	}

	n := v.NumField()
	for i := 0; i < n; i++ {
		fieldInfo := t.Field(i)
		fieldVal := v.Field(i)
		key := fieldInfo.Tag.Get(ParamKey)
		if key == "" {
			continue
		}
		val := params.ByName(key)
		if val == "" {
			continue
		}

		fieldType := fieldInfo.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fieldType))
			}
			fieldVal = fieldVal.Elem()
		}

		switch fieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			fieldVal.SetInt(n)
		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			fieldVal.SetFloat(n)
		case reflect.String:
			fieldVal.SetString(val)
		case reflect.Bool:
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			fieldVal.SetBool(b)
		default:
		}
	}

	return nil
}

func reqDecoder(ctx context.Context) func(interface{}) error {
	return func(req interface{}) error {
		return BindVar(ctx, req)
	}
}
