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

func (s *Server) prepareFileContext(ctx echo.Context) (*model.FileContext, error) {
	fileContext := new(model.FileContext)

	key := ctx.Request().URL.Path
	contentType := ctx.Request().Header.Get(echo.HeaderContentType)
	method := ctx.Request().Method
	if strings.HasPrefix(contentType, echo.MIMEMultipartForm) {
		// 解析表单数据, 主要是通过表单上传
		params := new(model.FormParams)
		if err := ctx.Bind(params); err != nil {
			return nil, errors.WithStack(err)
		}
		// TODO(benjamin): 处理tmp file临时文件close问题
		_, file, err := ctx.Request().FormFile("file")
		if err != nil {
			return nil, err
		}
		fileContext.Override = params.Override
		if params.Key != "" {
			key = params.Key
		}
		if err := fileContext.ParseFileContent(params.Source, file); err != nil {
			return nil, err
		}
	} else if method == http.MethodPost {
		// 解析multi-chunk参数
		values := ctx.Request().URL.Query()
		partNumber := values.Get("partNumber")
		uploadId := values.Get("uploadId")
		if utils.QueryExists(values, "uploads") {
			// 初始化multi-chunk解析参数
			fileContext.Uploads = true
		} else if partNumber != "" && uploadId != "" {
			// 解析单个chunk上传参数
			fileContext.PartNumber = utils.ToInt32(partNumber)
			fileContext.UploadId = uploadId
		} else if uploadId != "" {
			// 完成multi-chunk上传
			fileContext.UploadId = uploadId
		}
	}

	key = utils.PathNormalize(key)

	bucketName := utils.PathSegment(key, 0)
	if bucketName == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "未设置正确设置Bucket名")
	}

	bucket, err := s.GetBucket(bucketName)
	if err != nil {
		return nil, &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  fmt.Sprintf("Bucket %s 不存在", bucketName),
			Internal: errors.WithStack(err),
		}
	}
	fileContext.Key = key
	fileContext.Bucket = bucket
	return fileContext, nil
}

func (s *Server) file(ctx echo.Context) (err error) {
	fileContext, err := s.prepareFileContext(ctx)
	if err != nil {
		return err
	}
	context := &middleware.ExtendContext{ctx, fileContext}
	bucket := fileContext.Bucket

	method := ctx.Request().Method
	switch method {
	case http.MethodHead, http.MethodGet:
		fileContext.ParseImageSize(bucket)
		fileContext.Meta, err = s.GetFileEntity(fileContext.HashKey())
		if err != nil {
			if err == common.ErrKeyNotFound {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}
		if method == http.MethodHead {
			return s.head(context)
		}
		return s.download(context)
	case http.MethodPost:
		if fileContext.Uploads {
			return s.uploads(context)
		}
		if fileContext.UploadId != "" && fileContext.PartNumber != 0 {
			return s.uploadPart(context)
		}
		if fileContext.UploadId != "" {
			return s.uploadComplete(context)
		}
		return s.uploadFile(context)
	case http.MethodPut:

	}
	return echo.ErrMethodNotAllowed
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

func (s *Server) uploadFile(ctx echo.Context) (err error) {
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

	if err = s.validateFile(ctx); err != nil {
		return
	}
	key, entity := s.buildMetaFromChunk(ctx)
	if entity == nil {
		option := &common.UploadOption{
			Collection:  bucket.Basis.Collection,
			Replication: bucket.Basis.Replication,
			TTL:         bucket.Basis.TTL,
		}
		entity, err = s.Storage.Upload(option, file.MimeType, bytes.NewBuffer(file.Content))
		if err != nil {
			return
		}
		if file.IsImage() {
			entity.Width, entity.Height = file.Width, file.Height
		}
		s.saveChunk(ctx, key, entity)
	}

	hash := params.HashKey()
	entity.RawKey = params.Key
	if err = s.Meta.SetTTL(hash, entity, bucket.Basis.TTL.Expiry()); err != nil {
		return
	}
	return ctx.JSON(http.StatusOK, entity)
}

func (s *Server) buildMetaFromChunk(ctx echo.Context) (string, *model.FileMeta) {
	context := ctx.(*middleware.ExtendContext)
	file := context.FileContext.File
	bucket := context.FileContext.Bucket
	entity := new(model.FileMeta)

	if !bucket.Basis.TTL.Empty() {
		return "", nil
	}

	sha1, err := utils.ContentSha1(bytes.NewReader(file.Content))
	if err != nil {
		s.Logger.Warnf("计算文件Sha1值失败 %v", err)
		return "", nil
	}
	sha1 = fmt.Sprintf("%s:%s", bucket.Basis.Collection, sha1)
	chunk := new(model.Chunk)
	if err := s.ChunkDao.Get(sha1, chunk); err != nil {
		return sha1, nil
	}
	entity.FID = chunk.Fid
	entity.MimeType = file.MimeType
	entity.Size = file.Size
	entity.LastModified = time.Now().UTC().Unix()
	entity.Width = chunk.Width
	entity.Height = chunk.Height
	entity.ETag = chunk.Etag
	return sha1, entity
}

func (s *Server) saveChunk(ctx echo.Context, sha1 string, entity *model.FileMeta) {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket

	if !bucket.Basis.TTL.Empty() {
		return
	}

	chunk := new(model.Chunk)
	chunk.Fid = entity.FID
	chunk.Etag = entity.ETag
	if entity.Height > 0 && entity.Width > 0 {
		chunk.Width, chunk.Height = entity.Width, entity.Height
	}
	_ = s.ChunkDao.Set(sha1, chunk)
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
	if err != nil && err == common.ErrFileNotFound {
		ctx.Response().WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
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
