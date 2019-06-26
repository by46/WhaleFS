package server

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

const (
	BufferSize   = 4 * 1024 // 4M
	ParamPreview = "preview"
	ParamSize    = "size"
)

var (
	byteBufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, BufferSize)
		},
	}
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
		if params.Source != "" {
			params.Source = utils.UrlDecode(params.Source)
		}
		// TODO(benjamin): 处理tmp file临时文件close问题
		_, file, err := ctx.Request().FormFile("file")
		if err != nil && err != http.ErrMissingFile {
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
			_ = fileContext.ParseFileContentFromRequest(ctx)
		} else if uploadId != "" {
			// 完成multi-chunk上传
			fileContext.UploadId = uploadId
		}
		if utils.QueryExists(values, "check") {
			fileContext.Check = true
		}
	} else if method == http.MethodHead || method == http.MethodGet {
		fileContext.AttachmentName = ctx.QueryParam("attachmentName")
		key = s.legacySupportOSS(ctx, key)
	}

	key = utils.PathNormalize(key)

	bucketName, objectName := utils.PathRemoveSegment(key, 0)
	if bucketName == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "未设置正确设置Bucket名")
	}

	bucket, err := s.getBucketByName(bucketName)
	if err != nil {
		return nil, err
	}
	key = fmt.Sprintf("/%s%s", bucket.Name, objectName)
	fileContext.Key = key
	if len(key) > len(bucketName)+2 {
		fileContext.ObjectName = key[len(bucketName)+2:]
	}
	fileContext.BucketName = bucketName
	fileContext.Bucket = bucket
	return fileContext, nil
}

// 获取Bucket信息, 处理别名的逻辑
func (s *Server) getBucketByName(name string) (*model.Bucket, error) {
	bucket, err := s.GetBucket(name)
	if err != nil {
		return nil, &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  fmt.Sprintf("Bucket %s 不存在", name),
			Internal: errors.WithStack(err),
		}
	}
	if bucket.Basis != nil && bucket.Basis.Alias != "" {
		aliasBucket, err := s.GetBucket(bucket.Basis.Alias)
		if err != nil {
			return nil, &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  fmt.Sprintf("Bucket %s 的别名 %s 不存在", name, bucket.Basis.Alias),
				Internal: errors.WithStack(err),
			}
		}
		bucket = aliasBucket
	}
	return bucket, nil
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
				if bucket.Basis.DefaultImage != "" && utils.IsImageByFileName(fileContext.Key) {
					return s.downloadDefaultImage(context)
				}
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
			if entity, err := s.uploads(context); err != nil {
				return err
			} else {
				return ctx.JSON(http.StatusOK, entity)
			}
		}
		if fileContext.Check {
			if result, err := s.digestCheck(context); err != nil {
				return err
			} else {
				return ctx.JSON(http.StatusOK, result)
			}
		}
		if fileContext.UploadId != "" && fileContext.PartNumber != 0 {
			if entity, err := s.uploadPart(context); err != nil {
				return err
			} else {
				return ctx.JSON(http.StatusOK, entity)
			}
		}
		if fileContext.UploadId != "" {
			parts := make([]*model.Part, 0)
			if err = ctx.Bind(&parts); err != nil {
				return errors.WithStack(err)
			}
			if entity, err := s.uploadComplete(context, parts); err != nil {
				return err
			} else {
				return ctx.JSON(http.StatusOK, entity)
			}
		}
		return s.uploadFile(context)
	case http.MethodDelete:
		if fileContext.UploadId != "" {
			return s.uploadAbort(context)
		}
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
	response.Header().Set(constant.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))
	response.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) delete(ctx echo.Context) (err error) {
	return http.ErrNotSupported
}

func (s *Server) uploadFile(ctx echo.Context) (err error) {
	context := ctx.(*middleware.ExtendContext)
	fileContext := context.FileContext
	var entity *model.FileEntity
	if fileContext.File.Size > constant.ChunkSize {
		entity, err = s.uploadLargeFile(context)
	} else {
		entity, err = s.uploadFileInternal(ctx)
	}
	if err != nil {
		return err
	}
	s.sendMessage(entity, context.FileContext.Bucket)
	return ctx.JSON(http.StatusOK, entity)
}

func (s *Server) uploadFileInternal(ctx echo.Context) (entity *model.FileEntity, err error) {
	context := ctx.(*middleware.ExtendContext)
	fileContext := context.FileContext
	file := context.FileContext.File
	bucket := context.FileContext.Bucket

	if fileContext.ObjectName == "" {
		fileContext.IsRandomName = true
		fileContext.ObjectName = utils.RandomName(file.Extension)
		fileContext.Key = fmt.Sprintf("/%s/%s", bucket.Name, fileContext.ObjectName)
	}

	if file.IsImage() {
		reader := bytes.NewReader(file.Content)
		if config, err := utils.DecodeConfig(file.MimeType, reader); err == nil {
			file.Width, file.Height = config.Width, config.Height
		}
	}

	if err = s.validateFile(ctx); err != nil {
		return
	}
	key, meta := s.buildMetaFromChunk(ctx)
	if meta == nil {
		option := &common.UploadOption{
			Collection:  bucket.Basis.Collection,
			Replication: bucket.Basis.Replication,
			TTL:         bucket.Basis.TTL,
		}
		needle, err := s.Storage.Upload(option, file.MimeType, bytes.NewBuffer(file.Content))
		if err != nil {
			return nil, err
		}
		meta = needle.AsFileMeta()
		if file.IsImage() {
			meta.Width, meta.Height = file.Width, file.Height
		}
		s.saveChunk(ctx, key, meta)
	}

	hash := fileContext.HashKey()
	meta.RawKey = fileContext.Key
	meta.IsRandomName = fileContext.IsRandomName
	meta.WaterMark = file.WaterMark
	if err = s.Meta.SetTTL(hash, meta, bucket.Basis.TTL.Expiry()); err != nil {
		return
	}
	return meta.AsEntity(fileContext.BucketName, file.FileName), nil
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
	chunk.Size = entity.Size
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
		ctx.Response().Header().Add(constant.HeaderExpires, entity.HeaderExpires(*maxAge))
		ctx.Response().Header().Add(constant.HeaderCacheControl, entity.HeaderISOExpires(*maxAge))
	}
	if !s.Config.Debug && s.freshCheck(ctx, entity) {
		ctx.Response().WriteHeader(http.StatusNotModified)
		return
	}

	queryParams := ctx.Request().URL.Query()
	if utils.IsVideo(entity.MimeType) && utils.QueryExists(queryParams, ParamPreview) {
		err := s.fetchPreviewImg(ctx)
		if err != nil {
			return nil
		}
		entity.FID = entity.PreviewImg.FID
		entity.MimeType = entity.PreviewImg.MimeType
		entity.Size = entity.PreviewImg.Size
		if utils.QueryExists(queryParams, ParamSize) {
			size := queryParams.Get(ParamSize)
			context.FileContext.Size = bucket.GetSize(size)
		}
	}

	body, err := s.downloadFile(ctx)
	if err != nil {
		if err == common.ErrFileNotFound {
			if bucket.Basis.DefaultImage != "" && utils.IsImage(entity.MimeType) {
				return s.downloadDefaultImage(ctx)
			}
			ctx.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		return err
	}

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, entity.MimeType)
	response.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", entity.Size))
	response.Header().Set(echo.HeaderLastModified, utils.TimestampToRFC822(entity.LastModified))
	response.Header().Set(constant.HeaderETag, fmt.Sprintf(`"%s"`, entity.ETag))
	if context.FileContext.AttachmentName != "" {
		response.Header().Set(echo.HeaderContentDisposition, utils.Name2Disposition(ctx.Request().UserAgent(), context.FileContext.AttachmentName))
	}

	// support gzip
	if entity.Size >= constant.GzipLimit && entity.IsPlain() && s.shouldGzip(ctx) {
		return s.compress(ctx, body)
	}
	if _, err = io.Copy(response, body); err != nil {
		return errors.WithStack(err)
	}
	return
}

func (s *Server) downloadFile(ctx echo.Context) (io.Reader, error) {
	context := ctx.(*middleware.ExtendContext)
	entity := context.FileContext.Meta

	// 下载缩略图
	if thumbnail := s.downloadThumbnail(ctx); thumbnail != nil {
		return thumbnail, nil
	}

	body, _, err := s.Storage.Download(entity.FID)
	if err != nil {
		return nil, err
	}
	return s.thumbnail(ctx, body)
}

func (s *Server) downloadDefaultImage(ctx echo.Context) error {
	// TODO(benjamin): 考虑是否需要处理http cache的情况
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	r, err := s.downloadFileByFullName(bucket.Basis.DefaultImage)
	if err != nil {
		s.Logger.Errorf("download default image failed %v", err)
		ctx.Response().WriteHeader(http.StatusNotFound)
		return nil
	}
	defer func() {
		_ = r.Close()
	}()
	ctx.Response().Header().Add(constant.HeaderXWhaleFSFlags, constant.FlagDefaultImage)
	_, err = io.Copy(ctx.Response(), r)
	return err
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
	if etag := ctx.Request().Header.Get(constant.HeaderIfNoneMatch); etag != "" {
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("当前上传图片宽高比等于%d:%d, 宽高比必须等于%s", file.Width, file.Height, limit.Ratio))
			}
		}

		if utils.MimeMatch(file.MimeType, limit.MimeTypes) == false {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("只支持%v格式的文件", strings.Join(limit.MimeTypes, ",")))
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

func (s *Server) downloadFileByFullName(fullName string) (io.ReadCloser, error) {
	meta := new(model.FileMeta)
	if err := s.Meta.Get(fullName, meta); err != nil {
		return nil, errors.WithStack(err)
	}
	r, _, err := s.Storage.Download(meta.FID)
	return r, err
}
