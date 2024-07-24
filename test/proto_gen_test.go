package test

import (
	"context"
	"fmt"
	"github.com/mangohow/mangokit/transport/http"
	"testing"
	"time"
)

type FakeGreeterService struct{}

func (f FakeGreeterService) SayHello(ctx context.Context, request *HelloRequest) (*HelloResponse, error) {
	fmt.Println("SayHello called")
	return &HelloResponse{Message: "hello " + request.Name}, nil
}

func TestProtoGen(t *testing.T) {
	server := http.New(http.WithAddr(":80"))
	service := FakeGreeterService{}
	RegisterGreeterHTTPService(server, service)
	server.Start()
}

func TestProtoGenWithMiddleware(t *testing.T) {
	server := http.New(http.WithAddr(":80"))
	service := FakeGreeterService{}
	server.Middleware(func(ctx context.Context, req interface{}, next http.NextHandler) error {
		defer printRunTime(time.Now())()
		v := req.(*HelloRequest)
		fmt.Println(v)
		fmt.Println("middleware1 enter")
		err := next(ctx, req)
		fmt.Println("middleware1 exit")

		return err
	})

	server.Middleware(func(ctx context.Context, req interface{}, next http.NextHandler) error {
		fmt.Println("middleware2 enter")
		v := req.(*HelloRequest)
		if v.Name == "test" {
			return nil
		}

		err := next(ctx, req)

		fmt.Println("middleware2 exit")

		return err
	})

	RegisterGreeterHTTPService(server, service)

	server.Start()
}

func printRunTime(t time.Time) func() {
	return func() {
		fmt.Println(time.Now().Sub(t).String())
	}
}
