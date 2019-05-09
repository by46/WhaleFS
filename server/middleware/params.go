package middleware

import (
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server"
	"github.com/labstack/echo"
	"github.com/mholt/binding"
	"strings"
)

func ParseFileParams() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if context, success := ctx.(*server.ExtendContext); success {
				switch strings.ToLower(ctx.Request().Method) {
				case "post":
					params := new(model.FileParams)
					if err := binding.Bind(ctx.Request(), params); err != nil {
						return echo.ErrBadRequest
					}
					context.FileParams = params
				}
			}
			return next(ctx)
		}
	}
}
