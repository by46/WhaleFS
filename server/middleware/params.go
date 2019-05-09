package middleware

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mholt/binding"

	"github.com/by46/whalefs/model"
)

type Server interface {
	// get bucket info
	GetBucket(string) (*model.Bucket, error)

	// get meta information
	GetFileEntity(hash string) (*model.FileEntity, error)
}

type ParseFileParamsConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper

	Server Server
}

func ParseFileParams(config ParseFileParamsConfig) echo.MiddlewareFunc {

	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			if config.Skipper(ctx) {
				return next(ctx)
			}
			if context, success := ctx.(*ExtendContext); success {
				params := new(model.FileParams)
				switch strings.ToLower(ctx.Request().Method) {
				case "post":
					if err := binding.Bind(ctx.Request(), params); err != nil {
						return echo.ErrBadRequest
					}
				case "head", "get":
					if err = params.Bind(ctx); err != nil {
						return echo.ErrBadRequest
					}
					if params.Entity, err = config.Server.GetFileEntity(params.HashKey()); err != nil {
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
