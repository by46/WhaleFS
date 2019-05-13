package middleware

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

type Server interface {
	// get bucket info
	GetBucket(string) (*model.Bucket, error)

	// get meta information
	GetFileEntity(hash string) (*model.FileMeta, error)
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
		return func(ctx echo.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}
			context, success := ctx.(*ExtendContext)
			if !success {
				return next(ctx)
			}

			method := strings.ToLower(ctx.Request().Method)

			fileParams := new(model.FileContext)
			params, err := model.Bind(ctx)
			if err != nil {
				return err
			}

			key := utils.PathNormalize(params.Key)
			fileParams.Key = key
			fileParams.Override = params.Override
			bucketName := utils.PathSegment(key, 0)
			if bucketName == "" {
				return common.New(common.CodeBucketNotExists)
			}

			bucket, err := config.Server.GetBucket(bucketName)
			if err != nil {
				return common.New(common.CodeBucketNotExists)
			}
			fileParams.Bucket = bucket

			if method == "get" || method == "head" {
				fileParams.ParseImageSize(bucket)
				if fileParams.Meta, err = config.Server.GetFileEntity(fileParams.HashKey()); err != nil {
					return err
				}
			} else if method == "post" {
				if err := fileParams.ParseFileContent(params); err != nil {
					return err
				}
			}

			context.FileParams = fileParams
			return next(ctx)
		}
	}
}
