package middleware

import (
	"github.com/by46/whalefs/server"
	"github.com/labstack/echo"
)

func InjectContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			c := &server.ExtendContext{context, nil}
			return next(c)
		}
	}
}
