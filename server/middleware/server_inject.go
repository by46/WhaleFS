package middleware

import (
	"fmt"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
)

func InjectServer() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			res := context.Response()
			res.Header().Set(echo.HeaderServer, fmt.Sprintf("whalefs/%s", common.VERSION))
			return next(context)
		}
	}
}
