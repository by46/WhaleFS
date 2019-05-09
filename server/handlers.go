package server

import (
	"fmt"
	"github.com/by46/whalefs/server/middleware"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
	"github.com/labstack/echo"
)

func (s *Server) favicon(ctx echo.Context) error {
	return ctx.File("static/logo.png")
}

func (s *Server) faq(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "<!-- Newegg -->")
}

func (s *Server) tools(ctx echo.Context) error {
	if ctx.Request().Method == "GET" {
		return ctx.File("templates/tools.html")
	}
	return s.error(http.StatusForbidden, fmt.Errorf("method not implements"))
}

func (s *Server) download(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileParams.Bucket
	entity := context.FileParams.Entity

	maxAge := bucket.MaxAge()
	ctx.Response().Header().Add(utils.HeaderExpires, entity.HeaderExpires(maxAge))
	ctx.Response().Header().Add(utils.HeaderCacheControl, entity.HeaderISOExpires(maxAge))

	if !s.Config.Debug && s.freshCheck(ctx, entity) {
		ctx.Response().WriteHeader(http.StatusNotModified)
		return nil
	}

	body, _, err := s.Storage.Download(entity.FID)
	if err != nil {
		return err
	}

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, utils.TimestampToRFC822(entity.LastModified))
	response.Header().Set(utils.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))

	// support gzip
	if entity.Size >= (5<<10) && entity.IsPlain() && s.shouldGzip(ctx) {
		return s.compress(ctx, body)
	}
	_, err = io.Copy(response, body)
	return err
}

func (s *Server) upload(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	params := context.FileParams

	form := params.Content
	headers := http.Header(form.Header)
	body, err := form.Open()
	if err != nil {
		return err
	}

	defer body.Close()
	entity, err := s.Storage.Upload(headers.Get(echo.HeaderContentType), body)
	if err != nil {
		return err
	}
	hash := params.HashKey()
	entity.RawKey = params.Key
	if err := s.Meta.Set(hash, entity); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, entity)
}

func (s *Server) head(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	entity := context.FileParams.Entity

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, utils.TimestampToRFC822(entity.LastModified))
	response.Header().Set(utils.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))
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

func (s *Server) freshCheck(ctx echo.Context, entity *model.FileEntity) bool {
	headers := ctx.Request().Header
	if since := headers.Get(echo.HeaderIfModifiedSince); since != "" {
		sinceDate, err := utils.RFC822ToTime(since)
		if err != nil {
			s.Logger.Errorf("parse if-modified-since error %v", err)
			return false
		}
		if entity.LastModifiedTime().After(sinceDate) == false {
			return true
		}
	}
	if etag := ctx.Request().Header.Get(utils.HeaderIfNoneMatch); etag != "" {
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
