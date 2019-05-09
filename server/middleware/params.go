package middleware

import (
	"github.com/by46/whalefs/model"
	"github.com/labstack/echo"
	"github.com/mholt/binding"
	"strings"
)

type Server interface {
	GetBucket(string) (*model.Bucket, error)
}

type ParseFileParamsConfig struct {
	Server Server
}

func ParseFileParams(config ParseFileParamsConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
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
				if params.Bucket, err = config.Server.GetBucket(params.BucketName); err != nil {
					return echo.ErrBadRequest
				}
				context.FileParams = params
			}
			return next(ctx)
		}
	}
}
