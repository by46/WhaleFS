// compatible for legacy api
package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/hhrutter/pdfcpu/pkg/api"
	"github.com/hhrutter/pdfcpu/pkg/pdfcpu"
	pdf "github.com/hhrutter/pdfcpu/pkg/pdfcpu"
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

type ReturnData struct {
	Status  int8   `json:"status"`
	Message string `json:"message"`
	Url     string `json:"url"`
	Path    string `json:"path"`
}

const (
	ResultCodeSuccess    = "1000"
	ResultCodeFailed     = "1001"
	ResultMessageSuccess = "调用成功"
	ResultMessageFailed  = "调用失败"
	AppName              = "appName"
	ActionCancel         = "cancel"
)

var (
	ReInteger = regexp.MustCompile("^[0-9]+$")
)

// UploadHandler.ashx
func (s *Server) legacyUploadFile(ctx echo.Context) error {
	bucketName := utils.Params(ctx, AppName)
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
	bucketName := utils.Params(ctx, AppName)
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
	source := utils.Params(ctx, "FileUrl")
	if source == "" {
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

func (s *Server) legacyBatchDownload(ctx echo.Context) error {
	downloadZipFile := new(DownloadZipFile)
	if err := ctx.Bind(downloadZipFile); err != nil {
		s.Logger.Error(err)
		return returnMessage(ctx, "参数错误")
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
		s.Logger.Error(err)
		return returnMessage(ctx, "参数错误")
	}

	var totalSize int64
	for _, item := range packageEntity.Items {
		entity, err := s.GetFileEntity(item.RawKey)
		if err != nil {
			if err == common.ErrKeyNotFound {
				return returnMessage(ctx, "文件不存在")
			}
			return errors.WithStack(err)
		}

		totalSize = totalSize + entity.Size
	}

	if downloadZipFile.IsLimit && totalSize > s.TaskFileSizeThreshold {
		return returnMessage(ctx, "文件太大，请分供应商下载或到详情页面下载")
	}

	response := ctx.Response()

	response.Header().Set(echo.HeaderContentType, "application/zip")

	response.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", packageEntity.GetPkgName()))

	return Package(packageEntity, response, s.GetFileEntity, s.Storage.Download)
}

func returnMessage(ctx echo.Context, msg string) error {
	returnData := ReturnData{
		Status:  0,
		Url:     "",
		Path:    "",
		Message: msg,
	}
	err := ctx.JSON(200, returnData)
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

// BatchMergePdfHandler.ashx
func (s *Server) legacyMergePDF(ctx echo.Context) error {
	if err := ctx.Request().ParseForm(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "解析输入参数失败")
	}
	pdfFiles := ctx.Request().Form.Get("pdfFilePaths")
	if pdfFiles == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "pdfFilePaths未设置")
	}
	items := make([]pdfcpu.ReadSeekerCloser, 0)
	files := utils.Split(pdfFiles, ",")
	for _, file := range files {
		reader, err := s.legacyDownloadFileByFile(ctx, file)
		if err != nil {
			return err
		}
		items = append(items, reader)
	}

	config := pdf.NewDefaultConfiguration()
	config.Cmd = pdf.MERGE

	mergeContext, err := api.MergeContexts(items, config)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "合并PDF失败")
	}
	buf := bytes.NewBuffer(nil)
	if err := api.WriteContext(mergeContext, buf); err != nil {
		return errors.WithStack(err)
	}

	bucket, err := s.getBucketByName("tmp")
	if err != nil {
		return err
	}
	fileContext := &model.FileContext{
		BucketName: bucket.Name,
		Bucket:     bucket,
		File: &model.FileContent{
			Content:   buf.Bytes(),
			Size:      int64(buf.Len()),
			MimeType:  "application/pdf",
			FileName:  "merge.pdf",
			Extension: ".pdf",
		},
	}
	context := &middleware.ExtendContext{ctx, fileContext}
	entity, err := s.uploadFileInternal(context)
	if err != nil {
		return err
	}
	return ctx.String(http.StatusOK, entity.Key)
}

// SliceUploadHandler.ashx
func (s *Server) legacySliceUpload(ctx echo.Context) error {
	appName := strings.ToLower(ctx.QueryParam(AppName))
	identity := strings.TrimSpace(ctx.QueryParam("FileIdentity"))
	if identity == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "缺少FileIdentity")
	}
	if appName == "" {
		if ActionCancel == strings.TrimSpace(ctx.QueryParam("Action")) {
			return s.legacySliceUploadAbort(ctx, identity)
		}
		return s.legacySliceUploadChunk(ctx, identity)
	}
	return s.legacySliceUploadComplete(ctx, appName, identity)
}

func (s *Server) legacySliceUploadChunk(ctx echo.Context, identity string) error {
	positionValue := strings.TrimSpace(ctx.QueryParam("startPosition"))

	if ReInteger.MatchString(positionValue) == false {
		return echo.NewHTTPError(http.StatusBadRequest, "StartPosition未设置")
	}
	position := utils.ToInt32(positionValue)

	key, _ := utils.Sha1(identity)
	key = fmt.Sprintf("chunks:%s", key)
	fileContext := new(model.FileContext)

	if err := fileContext.ParseFileContentFromRequest(ctx); err != nil {
		return errors.WithMessage(err, "读取文件失败")
	}

	meta := &model.PartMeta{Parts: make(model.Parts, 0)}
	if err := s.Meta.Get(key, meta); err != nil {
		if err = s.Meta.SetTTL(key, meta, TTLChunk); err != nil {
			return errors.WithStack(err)
		}
	}
	opt := &common.UploadOption{
		Collection:  s.Config.Basis.CollectionShare,
		Replication: ReplicationOne,
	}
	needle, err := s.Storage.Upload(opt, echo.MIMEOctetStream, bytes.NewReader(fileContext.File.Content))
	if err != nil {
		return err
	}
	part := &model.Part{
		FID:        needle.FID,
		Size:       needle.Size,
		PartNumber: position,
	}
	if err := s.Meta.SubListAppend(key, "parts", part, 0); err != nil {
		return errors.WithMessage(err, "更新文件Parts失败")
	}
	return ctx.String(http.StatusOK, identity)
}

func (s *Server) legacySliceUploadComplete(ctx echo.Context, appName, identity string) error {
	bucket, err := s.getBucketByName(appName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "未正确设置Bucket")
	}
	key, _ := utils.Sha1(identity)
	key = fmt.Sprintf("chunks:%s", key)
	partMeta := &model.PartMeta{}
	if err := s.Meta.Get(key, partMeta); err != nil {
		return errors.WithMessage(err, "获取文件信息失败")
	}

	fileName := ctx.QueryParam("FileName")
	relativePath := ctx.QueryParam("RelativePath")
	relativePath = strings.ReplaceAll(relativePath, "\\", model.Separator)
	fileKey := path.Join(model.Separator, appName, relativePath, fileName)

	meta := partMeta.AsFileMeta()
	meta.RawKey = fileKey
	if err := s.Meta.SetTTL(fileKey, meta, bucket.Basis.TTL.Expiry()); err != nil {
		return errors.WithMessage(err, "设置文件内容失败")
	}
	if err := s.Meta.Delete(key, 0); err != nil {
		return errors.WithMessage(err, "删除临时文件失败")
	}
	return ctx.JSON(http.StatusOK, meta.AsEntity(appName, fileName))
}

func (s *Server) legacySliceUploadAbort(ctx echo.Context, identity string) error {
	key, _ := utils.Sha1(identity)
	key = fmt.Sprintf("chunks:%s", key)
	return s.Meta.Delete(key, 0)
}

// end SliceUploadHandler.ashx

func (s *Server) legacyDownloadFileByFile(ctx echo.Context, key string) (*utils.PDFFile, error) {
	_, key, _ = s.parseBucketAndFileKey(key)
	meta, err := s.GetFileEntity(key)
	if err != nil {
		return nil, err
	}
	reader, _, err := s.Storage.Download(meta.FID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = reader.Close()
	}()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return utils.NewPDFFile(content), nil
}

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

func (s *Server) legacyBuildDownloadUrl(ctx echo.Context, filePath, fileName string) string {
	return fmt.Sprintf("%s://%s/%s?attachmentName=%s", ctx.Scheme(), ctx.Request().Host, filePath, url.PathEscape(fileName))
}
