package server

import (
	"github.com/labstack/echo"
	"net/http"
)

func (s *Server) faq(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "<!-- Newegg -->")
}

func (s *Server) download(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "download")
}

func (s *Server) upload(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, 1)
}

func (s *Server) head(ctx echo.Context) error {
	return ctx.HTML(http.StatusNoContent, "")
}
