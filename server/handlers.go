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
	"mime/multipart"
)

func (s *Server) faq(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "<!-- Newegg -->")
}

func (s *Server) download(ctx echo.Context) error {
	uri := ctx.Request().RequestURI
	key := strings.ToLower(uri)
	hash, err := common.Sha1(key)
	if err != nil {
		return s.fatal(err)
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
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, common.TimestampToRFC822(entity.LastModified))
	response.Header().Set(common.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))
	response.WriteHeader(http.StatusOK)
	_, err = io.Copy(response, body)
	return err
}

func (s *Server) upload(ctx echo.Context) error {
	headers := ctx.Request().Header
	body := ctx.Request().Body
	form, err := s.form(ctx)
	if err != nil {
		return s.fatal(err)
	}
	if form != nil {
		headers = http.Header(form.Header)
		body, err = form.Open()
		if err != nil {
			return err
		}

	}
	defer body.Close()
	entity, err := s.Storage.Upload(headers.Get(echo.HeaderContentType), body)
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

func (s *Server) form(ctx echo.Context) (*multipart.FileHeader, error) {
	form, err := ctx.MultipartForm()
	if err != nil {
		return nil, err
	}
	for _, v := range form.File {
		if v == nil {
			continue
		}
		return v[0], nil
	}
	return nil, nil
}
