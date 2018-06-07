package server

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"whalefs/api"
	"whalefs/common"
	"whalefs/model"

	"github.com/labstack/echo"
)

func (s *Server) faq(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "<!-- Newegg -->")
}

func (s *Server) download(ctx echo.Context) error {
	hash, err := s.hashKey(ctx)
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

	if !s.Config.Debug && s.freshCheck(ctx, entity) {
		// TODO(benjamin): process max age form configuration
		maxAge := 200
		ctx.Response().Header().Add(common.HeaderExpires, entity.HeaderExpires(maxAge))
		ctx.Response().Header().Add(common.HeaderCacheControl, entity.HeaderISOExpires(maxAge))
		ctx.Response().WriteHeader(http.StatusNotModified)
		return nil
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
	if err != nil && err != http.ErrNotMultipart {
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
	hash, err := s.hashKey(ctx)
	if err != nil {
		return err
	}
	entity.RawKey = s.objectKey(ctx)
	if err := s.Meta.Set(hash, entity); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, entity)
}

func (s *Server) head(ctx echo.Context) error {
	key, err := s.hashKey(ctx)
	if err != nil {
		return s.fatal(err)
	}
	entity := &model.FileEntity{}
	if err := s.Meta.Get(key, entity); err != nil {
		return s.fatal(err)
	}

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, common.TimestampToRFC822(entity.LastModified))
	response.Header().Set(common.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))
	response.WriteHeader(http.StatusOK)
	return nil
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

func (srv *Server) freshCheck(ctx echo.Context, entity *model.FileEntity) bool {
	headers := ctx.Request().Header
	if since := headers.Get(echo.HeaderIfModifiedSince); since != "" {
		sinceDate, err := common.RFC822ToTime(since)
		if err != nil {
			srv.Logger.Errorf("parse if-modified-since error %v", err)
			return false
		}
		if entity.LastModifiedTime().After(sinceDate) == false {
			return true
		}
	}
	if etag := ctx.Request().Header.Get(common.HeaderIfNoneMatch); etag != "" {
		for _, value := range strings.Split(etag, ",") {
			value = strings.TrimSpace(value)
			value = strings.Trim(value, `"`)
			if value == entity.ETag {
				return true
			}
		}
	}
	return false
}
