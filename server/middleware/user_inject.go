package middleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"strings"
)

func InjectUser(config AuthUserConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			authToken := ctx.Request().Header.Get("Authorization")
			authToken = strings.TrimPrefix(authToken, "Bearer ")
			user, err := config.Server.AuthenticateUser(authToken)
			if err != nil {
				return err
			}
			ctx.Set("user", user)
			return next(ctx)
		}
	}
}

type AuthUserConfig struct {
	Skipper middleware.Skipper
	Server  Server
}
