package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	StatusOK = http.StatusOK

	StatusMovedPermanently = http.StatusMovedPermanently

	StatusBadRequest   = http.StatusBadRequest
	StatusUnauthorized = http.StatusUnauthorized
	StatusForbidden    = http.StatusForbidden
	StatusNotFound     = http.StatusNotFound

	StatusInternalServerError = http.StatusInternalServerError
	StatusNotImplemented      = http.StatusNotImplemented
	StatusBadGateway          = http.StatusBadGateway
	StatusServiceUnavailable  = http.StatusServiceUnavailable
)

func BindVar(ctx context.Context, val interface{}) error {
	c := ctx.Value("gin-ctx").(*gin.Context)
	return c.ShouldBind(val)
}
