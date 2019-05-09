package middleware

import (
	"github.com/by46/whalefs/model"
	"github.com/labstack/echo"
	"github.com/mholt/binding"
	"strings"
)

func ParseFileParams() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if context, success := ctx.(*ExtendContext); success {
				params := new(model.FileParams)
				switch strings.ToLower(ctx.Request().Method) {
				case "post":
					if err := binding.Bind(ctx.Request(), params); err != nil {
						return echo.ErrBadRequest
					}
				default:
					if err := params.Bind(ctx); err != nil {
						return echo.ErrBadRequest
					}
				}
				context.FileParams = params
			}
			return next(ctx)
		}
	}
}
