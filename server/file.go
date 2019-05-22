package server

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

func (s *Server) head(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	entity := context.FileContext.Meta

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, utils.TimestampToRFC822(entity.LastModified))
	response.Header().Set(utils.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))
	response.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) upload(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	params := context.FileContext
	file := context.FileContext.File
	bucket := context.FileContext.Bucket

	if file.IsImage() {
		reader := bytes.NewReader(file.Content)
		if config, err := utils.DecodeConfig(file.MimeType, reader); err == nil {
			file.Width, file.Height = config.Width, config.Height
		}
	}

	if err := s.validateFile(ctx); err != nil {
		return err
	}

	sha1, err := utils.ContentSha1(bytes.NewReader(file.Content))
	if err != nil {
		return err
	}
	chunk := new(model.Chunk)
	entity := new(model.FileMeta)
	if err := s.ChunkDao.Get(sha1, chunk); err == nil {
		entity.FID = chunk.Fid
		entity.MimeType = file.MimeType
		entity.Size = file.Size
		entity.LastModified = time.Now().UTC().Unix()
		entity.Width = chunk.Width
		entity.Height = chunk.Height
	} else {
		option := &common.UploadOption{
			Collection:  bucket.Basis.Collection,
			Replication: bucket.Basis.Replication,
			TTL:         bucket.Basis.TTL,
		}
		entity, err = s.Storage.Upload(option, file.MimeType, bytes.NewBuffer(file.Content))
		if err != nil {
			return err
		}
		chunk.Fid = entity.FID
		chunk.Etag = entity.ETag
		if file.IsImage() {
			chunk.Width, chunk.Height = file.Width, file.Height
			entity.Width, entity.Height = file.Width, file.Height
		}
		if err = s.ChunkDao.Set(sha1, chunk); err != nil {
			return err
		}

	}

	hash := params.HashKey()
	entity.RawKey = params.Key
	if err := s.Meta.Set(hash, entity); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, entity)
}

func (s *Server) download(ctx echo.Context) (err error) {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	entity := context.FileContext.Meta

	if maxAge := bucket.MaxAge(); maxAge != nil {
		ctx.Response().Header().Add(utils.HeaderExpires, entity.HeaderExpires(*maxAge))
		ctx.Response().Header().Add(utils.HeaderCacheControl, entity.HeaderISOExpires(*maxAge))
	}
	if !s.Config.Debug && s.freshCheck(ctx, entity) {
		ctx.Response().WriteHeader(http.StatusNotModified)
		return
	}

	body, _, err := s.Storage.Download(entity.FID)
	if err != nil {
		return err
	}

	body, err = s.thumbnail(ctx, body)
	if err != nil {
		return err
	}

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, utils.TimestampToRFC822(entity.LastModified))
	response.Header().Set(utils.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))

	// support gzip
	if entity.Size >= GzipLimit && entity.IsPlain() && s.shouldGzip(ctx) {
		return s.compress(ctx, body)
	}
	if _, err = io.Copy(response, body); err != nil {
		return errors.WithStack(err)
	}
	return
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

func (s *Server) freshCheck(ctx echo.Context, entity *model.FileMeta) bool {
	headers := ctx.Request().Header
	if since := headers.Get(echo.HeaderIfModifiedSince); since != "" {
		sinceDate, err := utils.RFC822ToTime(since)
		if err != nil {
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

func (s *Server) validateFile(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	params := context.FileContext
	limit := context.FileContext.Bucket.Limit
	file := context.FileContext.File

	if limit != nil {
		if limit.MinSize != nil && file.Size < *limit.MinSize {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("当前上传文件大小等于%d, 不能小于下限阈值%d", file.Size, *limit.MinSize))
		}

		if limit.MaxSize != nil && file.Size > *limit.MaxSize {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("当前上传文件大小等于%d, 不能大于上限阈值%d", file.Size, *limit.MaxSize))
		}

		if file.IsImage() && (limit.Width != nil || limit.Height != nil || limit.Ratio != "") {
			if limit.Width != nil && *limit.Width != file.Width {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("当前上传图片宽度等于%d, 宽度必须等于%d", file.Width, *limit.Width))
			}

			if limit.Height != nil && *limit.Height != file.Height {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("当前上传图片高度等于%d, 高度必须等于%d", file.Height, *limit.Height))
			}
			ratio := utils.RatioEval(limit.Ratio)
			if ratio != nil && utils.Float64Equal(*ratio, float64(file.Width)/float64(file.Height)) == false {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("当前上传图片宽高比等于%d:%d, 宽高比必须等于%d", file.Width, file.Height, limit.Ratio))
			}
		}
	}

	hash := params.HashKey()
	if !params.Override {
		if exists, err := s.Meta.Exists(hash); err != nil {
			return errors.Wrap(err, "获取文件内容错误")
		} else if exists {
			return echo.NewHTTPError(http.StatusForbidden, "当前文件已经存在, 不允许覆盖")
		}
	}
	return nil
}
