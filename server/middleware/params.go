package middleware

import (
	"github.com/labstack/echo/middleware"

	"github.com/by46/whalefs/model"
)

type Server interface {
	// get bucket info
	GetBucket(string) (*model.Bucket, error)

	// get meta information
	GetFileEntity(hash string, isRemoveOriginal bool) (*model.FileMeta, error)

	AuthenticateUser(authToken string) (*model.User, error)
}

type ParseFileParamsConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper

	Server Server
}
