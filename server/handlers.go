package server

import (
	"github.com/labstack/echo"
	"net/http"
	"whalefs/common"
	"strings"
	"whalefs/model"
	"io"
	"whalefs/api"
	"fmt"
)

func (s *Server) faq(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "<!-- Newegg -->")
}

func (s *Server) download(ctx echo.Context) error {
	uri := ctx.Request().RequestURI
	key := strings.ToLower(uri)
	hash, err := common.Sha1(key)
	if err != nil {
		return err
	}
	entity := &model.FileEntity{}
	if err := s.Meta.Get(hash, entity); err != nil {
		if err == api.ErrNoEntity {
			return s.error(http.StatusNotFound, fmt.Errorf("file not found"))
		}
		return s.error(http.StatusInternalServerError, err)
	}
	body, _, err := s.Storage.Download(entity.Url)
	if err != nil {
		return err
	}
	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.WriteHeader(http.StatusOK)
	_, err = io.Copy(response, body)
	return err
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
