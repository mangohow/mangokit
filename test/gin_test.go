package test

import (
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestGin(t *testing.T) {
	router := gin.Default()
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	router.GET("/get", Handler)
	router.GET("/get/:username/:password", Handler)
	router.POST("/post", Handler)

	server.ListenAndServe()
}

// 1、/get/:username/:password 参数保存在ctx.Param中
// 2、/get?username=aabb&password=ccdd 从ctx.Request.RUL.Query()获取
// 3、/post json数据从body获取
func Handler(ctx *gin.Context) {
	in := new(User)
	if err := ctx.ShouldBindQuery(in); err != nil {
		slog.Error("bind query", "err=", err)
	}

	if err := ctx.ShouldBindUri(in); err != nil {
		slog.Error("bind uri", "err=", err)
	}

	if err := ctx.ShouldBind(in); err != nil {
		slog.Error("should bind", "err=", err)
	}

	slog.Info("user", "user", in)
}

func TestByteToUpper(t *testing.T) {
	c := byte('a')
	arr := make([]byte, 26)
	for i := 0; i < 26; i++ {
		arr[i] = c + byte(i)
		arr[i] = arr[i] & '_'
	}

	fmt.Println(string(arr))
}
