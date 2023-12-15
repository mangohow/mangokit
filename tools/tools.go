package tools

import (
	"context"

	"github.com/gin-gonic/gin"
)

func GinCtxFromContext(ctx context.Context) *gin.Context {
	c := ctx.Value("gin-ctx")
	return c.(*gin.Context)
}

