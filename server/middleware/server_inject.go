package middleware

import (
	"fmt"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/constant"
)

func InjectServer() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			res := context.Response()
			res.Header().Set(echo.HeaderServer, fmt.Sprintf("whalefs/%s", constant.VERSION))
			res.Header().Set(constant.HeaderVia, fmt.Sprintf("whalefs/%s", constant.VERSION))
			return next(context)
		}
	}
}
