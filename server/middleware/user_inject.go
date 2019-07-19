package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/by46/whalefs/constant"
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

			token := ctx.Request().Header.Get(echo.HeaderAuthorization)
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			segments := strings.Fields(token)
			if len(segments) != 2 {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			switch strings.ToLower(segments[0]) {
			case "jwt":
				user, err := config.Server.AuthenticateUser(segments[1])
				if err != nil {
					return err
				}
				ctx.Set(constant.ContextKeyUser, user)
				return next(ctx)
			default:
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
		}
	}
}

type AuthUserConfig struct {
	Skipper middleware.Skipper
	Server  Server
}
