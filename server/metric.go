package server

import (
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) metric(ctx echo.Context) error {
	handler := promhttp.Handler()
	handler.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}
