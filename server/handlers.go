package server

import (
	"github.com/labstack/echo"
	"net/http"
	"whalefs/common"
	"strings"
)

func (s *Server) faq(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "<!-- Newegg -->")
}

func (s *Server) download(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "download")
}

func (s *Server) upload(ctx echo.Context) error {
	headers := ctx.Request().Header
	entity, err := s.Storage.Upload(headers.Get(echo.HeaderContentType), ctx.Request().Body)
	if err != nil {
		return err
	}
	uri := ctx.Request().RequestURI
	key := strings.ToLower(uri)
	hash, err := common.Sha1(key)
	if err != nil {
		return err
	}
	entity.RawKey = uri
	if err := s.Meta.Set(hash, entity); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, entity)
}

func (s *Server) head(ctx echo.Context) error {
	return ctx.HTML(http.StatusNoContent, "")
}
