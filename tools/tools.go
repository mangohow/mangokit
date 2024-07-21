package tools

import (
	"context"
	"github.com/gin-gonic/gin"
)

func GinCtxFromContext(ctx context.Context) *gin.Context {
	c := ctx.Value("gin-ctx")
	return c.(*gin.Context)
}

func BindVar(ctx context.Context, val interface{}) error {
	c := ctx.Value("gin-ctx").(*gin.Context)
	return c.ShouldBindUri(val)
}
