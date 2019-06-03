// compatible for legacy api
package server

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

const (
	ResultCodeSuccess    = "1000";
	ResultCodeFailed     = "1001";
	ResultMessageSuccess = "调用成功";
	ResultMessageFailed  = "调用失败"
)

func (s *Server) legacyFormFile(ctx echo.Context) (file *multipart.FileHeader, err error) {
	form, err := ctx.MultipartForm()
	if err != nil || form.File == nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "没有文件内容")
	}
	for _, value := range form.File {
		if len(value) > 0 {
			file = value[0]
			break
		}
	}
	return
}

// UploadHandler.ashx
func (s *Server) legacyUploadFile(ctx echo.Context) error {
	bucketName := ctx.QueryParam("appName")
	if bucketName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "未设置正确设置Bucket名")
	}
	bucket, err := s.getBucketByName(bucketName)
	if err != nil {
		return err
	}
	fileContext := &model.FileContext{
		Bucket:     bucket,
		BucketName: bucketName,
	}
	file, err := s.legacyFormFile(ctx)
	if err != nil {
		return err
	}
	if file == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "没有文件内容")
	}
	if err := fileContext.ParseFileContent("", file); err != nil {
		return err
	}
	context := &middleware.ExtendContext{ctx, fileContext}

	return s.uploadFile(context)
}

// DownloadSaveServerHandler.ashx
func (s *Server) legacyUploadByRemote(ctx echo.Context) error {
	bucketName := ctx.QueryParam("appName")
	if bucketName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "未设置正确设置Bucket名")
	}
	bucket, err := s.getBucketByName(bucketName)
	if err != nil {
		return err
	}
	fileContext := &model.FileContext{
		Bucket:     bucket,
		BucketName: bucketName,
	}
	source := ctx.QueryParam("FileUrl")
	if source != "" {
		return echo.NewHTTPError(http.StatusBadRequest, "FileUrl不能为空")
	}
	if err := fileContext.ParseFileContent(source, nil); err != nil {
		return err
	}
	context := &middleware.ExtendContext{ctx, fileContext}
	return s.uploadFile(context)
}

// DownloadHandler.ashx
func (s *Server) legacyDownloadFile(ctx echo.Context) (err error) {
	key := ctx.QueryParam("FilePath")
	attachmentName := ctx.QueryParam("FileName")
	//shouldMark := utils.ToBool(ctx.QueryParam("Mark"))

	if utils.IsRemote(key) {
		return s.legacyDownloadFileByRemote(ctx)
	}
	bucket, key, size := s.parseBucketAndFileKey(key)

	if bucket == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "未设置正确设置Bucket名")
	}
	fileContext := &model.FileContext{
		Bucket:         bucket,
		BucketName:     bucket.Name,
		Size:           size,
		Key:            key,
		AttachmentName: attachmentName,
	}
	fileContext.Meta, err = s.GetFileEntity(fileContext.HashKey())
	if err != nil {
		if err == common.ErrKeyNotFound {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}
	context := &middleware.ExtendContext{ctx, fileContext}
	return s.download(context)
}

func (s *Server) legacyDownloadFileByRemote(ctx echo.Context) error {
	source := ctx.QueryParam("FilePath")
	fileContext := &model.FileContext{
		AttachmentName: ctx.QueryParam("FileName"),
	}
	if err := fileContext.ParseFileContent(source, nil); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "文件不存在")
	}
	file := fileContext.File
	ctx.Response().Header().Set(echo.HeaderContentType, file.MimeType)
	ctx.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", file.Size))

	if fileContext.AttachmentName != "" {
		ctx.Response().Header().Set(echo.HeaderContentDisposition, utils.Name2Disposition(ctx.Request().UserAgent(), fileContext.AttachmentName))
	}
	_, err := ctx.Response().Write(file.Content)
	return err
}

// ApiUploadHandler.ashx
func (s *Server) legacyApiUpload(ctx echo.Context) error {
	result := &model.ResponseInfo{
		Code:    ResultCodeFailed,
		Message: ResultMessageFailed,
		Data:    &model.UploadResut{},
	}
	bucketName := "tender"
	bucket, err := s.getBucketByName(bucketName)
	if err != nil {
		return err
	}
	fileContext := &model.FileContext{
		Bucket:     bucket,
		BucketName: bucket.Name,
		Key:        fmt.Sprintf("/%s/", bucketName),
	}
	file, err := s.legacyFormFile(ctx)
	if err != nil {
		return err
	}
	if err := fileContext.ParseFileContent("", file); err != nil {
		return err
	}

	context := &middleware.ExtendContext{ctx, fileContext}
	entity, err := s.uploadFileInternal(context)
	if err == nil {
		result.Code = ResultCodeSuccess
		result.Message = ResultMessageSuccess
		result.Data.Name = utils.NameWithoutExtension(entity.Original)
		result.Data.Extension = strings.Replace(filepath.Ext(entity.Original), ".", "", 1)
		result.Data.Url = s.legacyBuildDownloadUrl(ctx, entity.Key, entity.Original)
	}
	return ctx.JSON(http.StatusOK, result)
}

func (s *Server) legacyBuildDownloadUrl(ctx echo.Context, filePath, fileName string) string {
	return fmt.Sprintf("%s://%s/%s?attachmentName=%s", ctx.Scheme(), ctx.Request().Host, filePath, url.PathEscape(fileName))
}