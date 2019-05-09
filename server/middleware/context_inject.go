package middleware

import (
	"github.com/labstack/echo"
)

func InjectContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			c := &ExtendContext{context, nil}
			return next(c)
		}
	}
}
