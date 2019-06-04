// compatible for legacy api
package server

import (
	"fmt"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

type DownloadZipFile struct {
	ZipFileName string            `json:"zipFileName"`
	Attachments map[string]string `json:"attachments"`
	IsLimit     bool              `json:"islimit"`
}

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
	form, err := ctx.MultipartForm()
	if err != nil || form.File == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "没有文件内容")
	}
	var file *multipart.FileHeader
	for _, value := range form.File {
		if len(value) > 0 {
			file = value[0]
			break
		}
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

func (s *Server) legacyBatchDownload(ctx echo.Context) error {
	downloadZipFile := new(DownloadZipFile)
	if err := ctx.Bind(downloadZipFile); err != nil {
		return err
	}

	var pkgFileItems []model.PkgFileItem
	for k, v := range downloadZipFile.Attachments {
		item := model.PkgFileItem{
			RawKey: k,
			Target: v,
		}

		pkgFileItems = append(pkgFileItems, item)
	}

	packageEntity := &model.PackageEntity{
		Name:  utils.PathLastSegment(downloadZipFile.ZipFileName),
		Items: pkgFileItems,
		Type:  model.Zip,
	}

	err := packageEntity.Validate()
	if err != nil {
		return errors.WithStack(err)
	}

	var totalSize int64
	for _, item := range packageEntity.Items {
		entity, err := s.GetFileEntity(item.RawKey)
		if err != nil {
			return errors.WithStack(err)
		}

		totalSize = totalSize + entity.Size
	}

	if totalSize > s.TaskFileSizeThreshold {
		hashKey, err := utils.Sha1(fmt.Sprintf("/%s/%s", s.TaskBucketName, packageEntity.Name))
		err = s.CreateTask(hashKey, packageEntity)
		err = ctx.Redirect(http.StatusMovedPermanently, "/tasks?key="+hashKey)
		return err
	}

	response := ctx.Response()

	pkgType := packageEntity.GetPkgType()

	if pkgType == utils.Tar {
		response.Header().Set(echo.HeaderContentType, "application/tar")
	} else {
		response.Header().Set(echo.HeaderContentType, "application/zip")
	}

	response.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", packageEntity.GetPkgName()))

	return Package(packageEntity, response, s.GetFileEntity, s.Storage.Download)
}
