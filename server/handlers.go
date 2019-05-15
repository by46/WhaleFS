package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

func (s *Server) favicon(ctx echo.Context) error {
	// TODO(benjamin): 添加新鲜度检查
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
	bucket := context.FileContext.Bucket
	entity := context.FileContext.Meta

	if maxAge := bucket.MaxAge(); maxAge != nil {
		ctx.Response().Header().Add(utils.HeaderExpires, entity.HeaderExpires(*maxAge))
		ctx.Response().Header().Add(utils.HeaderCacheControl, entity.HeaderISOExpires(*maxAge))
	}
	if !s.Config.Debug && s.freshCheck(ctx, entity) {
		ctx.Response().WriteHeader(http.StatusNotModified)
		return nil
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
	_, err = io.Copy(response, body)
	return err
}

func (s *Server) tarDownload(ctx echo.Context) error {
	content := ctx.FormValue("content")
	tarFileEntity := new(model.TarFileEntity)
	err := json.Unmarshal([]byte(content), &tarFileEntity)
	if err != nil {
		return err
	}

	var totalSize int64
	for _, item := range tarFileEntity.Items {
		hashKey, err := utils.Sha1(item.RawKey)
		if err != nil {
			return err
		}

		entity, err := s.GetFileEntity(hashKey)
		if err != nil {
			return err
		}

		totalSize = totalSize + entity.Size
	}

	if totalSize > 1024 {
		hashKey, err := utils.Sha1(fmt.Sprintf("/%s/%s", s.TaskBucketName, tarFileEntity.Name))
		if err != nil {
			return err
		}

		err = s.CreateTask(hashKey, tarFileEntity)
		if err != nil {
			return err
		}
		ctx.Redirect(http.StatusMovedPermanently, "/tasks?key="+hashKey)
	}

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, "application/tar")
	response.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", tarFileEntity.Name))

	return Package(tarFileEntity, response, s.GetFileEntity, s.Storage.Download)
}

func (s *Server) upload(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	params := context.FileContext

	if err := s.validateFile(ctx); err != nil {
		return err
	}

	file := params.File
	entity, err := s.Storage.Upload(file.MimeType, bytes.NewBuffer(file.Content))
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
	entity := context.FileContext.Meta

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, utils.TimestampToRFC822(entity.LastModified))
	response.Header().Set(utils.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))
	response.WriteHeader(http.StatusNoContent)
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

func (s *Server) freshCheck(ctx echo.Context, entity *model.FileMeta) bool {
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

func (s *Server) validateFile(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	params := context.FileContext
	limit := context.FileContext.Bucket.Limit
	file := context.FileContext.File

	if limit != nil {
		if limit.MinSize != nil && file.Size < *limit.MinSize {
			return common.New(common.CodeLimit)
		}

		if limit.MaxSize != nil && file.Size > *limit.MaxSize {
			return common.New(common.CodeLimit)
		}

		if file.IsImage() && (limit.Width != nil || limit.Height != nil || limit.Ratio != nil) {
			// TODO(benjamin): 检查图片的宽度和高度
			reader := bytes.NewReader(file.Content)
			config, err := utils.DecodeConfig(file.MimeType, reader)
			if err != nil {
				return err
			}

			if limit.Width != nil && *limit.Width != config.Width {
				return common.New(common.CodeLimit)
			}

			if limit.Height != nil && *limit.Height != config.Height {
				return common.New(common.CodeLimit)
			}
			ratio := utils.RatioEval(limit.Ratio)
			if ratio != nil && utils.Float64Equal(*ratio, float64(config.Width)/float64(config.Height)) == false {
				return common.New(common.CodeLimit)
			}
		}

	}

	hash := params.HashKey()
	if !params.Override {
		if exists, err := s.Meta.Exists(hash); err != nil {
			return err
		} else if exists {
			return common.New(common.CodeForbidden)
		}
	}
	return nil
}

func (s *Server) checkTask(ctx echo.Context) error {
	key := ctx.QueryParam("key")
	if key == "" {
		err := ctx.String(http.StatusBadRequest, "没有指定key")
		if err != nil {
			return err
		}
	}
	var task = model.TarTask{}
	err := s.TaskMeta.Get(key, &task)
	if err != nil {
		return err
	}
	if task.Status == model.TASK_SUCCESS {
		err := ctx.Redirect(http.StatusMovedPermanently, task.TarFileRawKey)
		if err != nil {
			return err
		}
	} else if task.Status == model.TASK_PENDING || task.Status == model.TASK_RUNNING {
		err := ctx.String(http.StatusOK, "文件打包中......")
		if err != nil {
			return err
		}
	} else {
		err := ctx.String(http.StatusInternalServerError, "文件打包失败"+task.ErrorMsg)
		if err != nil {
			return err
		}
	}
	return nil
}
