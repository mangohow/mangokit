package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mangohow/mangokit/transport/http"
	"strings"
	"testing"
	"time"
)

type FakeGreeterService struct{}

func (s FakeGreeterService) Test(ctx context.Context, request *Kinds) error {
	fmt.Println(request)

	return nil
}

func (s FakeGreeterService) SayHello1(ctx context.Context, request *HelloRequest) error {
	fmt.Println("SayHello1 called")
	fmt.Println(request)
	return nil
}

func (s FakeGreeterService) SayHello2(ctx context.Context) (*HelloResponse, error) {
	fmt.Println("SayHello2 called")
	return &HelloResponse{Message: "hello"}, nil
}

func (s FakeGreeterService) SayHello3(ctx context.Context) error {
	fmt.Println("SayHello3 called")
	return nil
}

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
	server.Middleware(func(ctx context.Context, req interface{}, next http.Handler) (interface{}, error) {
		defer printRunTime(time.Now())()
		fmt.Println(req)
		fmt.Println("middleware1 enter")
		defer fmt.Println("middleware1 exit")

		return next(ctx, req)
	})

	server.Middleware(func(ctx context.Context, req interface{}, next http.Handler) (interface{}, error) {
		c := http.GinCtxFromContext(ctx)
		fmt.Println("fullpath:", c.FullPath())
		fmt.Println("middleware2 enter")
		defer fmt.Println("middleware2 exit")

		if !strings.Contains(c.FullPath(), "helloworld") {
			return next(ctx, req)
		}

		v := req.(*HelloRequest)
		if v.Name == "test" {
			return nil, nil
		}

		return next(ctx, req)
	})

	RegisterGreeterHTTPService(server, service)

	server.Start()
}

func printRunTime(t time.Time) func() {
	return func() {
		fmt.Println(time.Now().Sub(t).String())
	}
}

type printCallOption struct {
}

func (p printCallOption) Before(info *http.BeforeCallInfo) {
	info.ContentType = "application/json"
	fmt.Println("before call")
	bytes, _ := json.MarshalIndent(info, "", "    ")
	fmt.Println("req:", string(bytes))
}

func (p printCallOption) After(info *http.AfterCallInfo) {
	fmt.Println("after call")
	fmt.Println("status:", info.Status)
	fmt.Println("resp:", info.Resp)
}

func PrintCallOption() printCallOption {
	return printCallOption{}
}

func intp(i int) *int {
	return &i
}

func stringp(s string) *string {
	return &s
}

func boolp(b bool) *bool {
	return &b
}

func TestHTTPClient(t *testing.T) {
	cli, _ := http.NewClient(http.WithEndpoint("http://127.0.0.1:80"))
	client := NewGreeterHTTPClient(cli)
	_, err := client.SayHello(context.Background(), &HelloRequest{Name: "tom"}, PrintCallOption())
	if err != nil {
		t.Fatalf("SayHello error=%v", err)
	}
	err = client.SayHello1(context.Background(), &HelloRequest{Name: "tom"}, PrintCallOption())
	if err != nil {
		t.Fatalf("SayHello1 error=%v", err)
	}
	_, err = client.SayHello2(context.Background(), PrintCallOption())
	if err != nil {
		t.Fatalf("SayHello2 error=%v", err)
	}
	err = client.SayHello3(context.Background(), PrintCallOption())
	if err != nil {
		t.Fatalf("SayHello3 error=%v", err)
	}
	err = client.Test(context.Background(), &Kinds{
		Kint:     100,
		Kintp:    intp(200),
		Kstring:  "hello",
		Kstringp: stringp("world"),
		Kbool:    true,
		Kboolp:   boolp(true),
	}, PrintCallOption())
	if err != nil {
		t.Fatalf("Test error=%v", err)
	}
}
