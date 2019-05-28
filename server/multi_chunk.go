package server

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

// 生成multi-chunk上传任务
func (s *Server) uploads(ctx echo.Context) (err error) {
	context := ctx.(*middleware.ExtendContext)
	fileContext := context.FileContext
	bucket := context.FileContext.Bucket

	uploadId := uuid.New().String()
	mimeType := ctx.Request().Header.Get(echo.HeaderContentType)
	if mimeType == "" {
		mimeType = echo.MIMEOctetStream
	}
	uploads := &model.Uploads{
		Bucket:   bucket.Name,
		Key:      context.FileContext.Key,
		UploadId: uploadId,
	}
	if fileContext.ObjectName == "" {
		fileContext.ObjectName = utils.RandomName(mimeType)
		fileContext.Key = fmt.Sprintf("/%s/%s", bucket.Name, fileContext.ObjectName)
		uploads.Key = fileContext.Key
	}
	partMeta := &model.PartMeta{
		Key:      context.FileContext.Key,
		MimeType: mimeType,
		Parts:    []*model.Part{},
	}
	key := fmt.Sprintf("chunks:%s", uploadId)
	if err = s.Meta.SetTTL(key, partMeta, TTLChunk); err != nil {
		return err
	}
	s.Logger.Info("multi-chunk upload : %s", key)
	return ctx.JSON(http.StatusOK, uploads)
}

// 上传multi-chunk分块
func (s *Server) uploadPart(ctx echo.Context) (err error) {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	file := context.FileContext.File

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
		s.saveChunk(ctx, key, entity)
	}

	chunkKey := fmt.Sprintf("chunks:%s", context.FileContext.UploadId)

	part := &model.Part{
		PartNumber: context.FileContext.PartNumber,
		FID:        entity.FID,
		Size:       entity.Size,
	}
	if err = s.Meta.SubListAppend(chunkKey, "parts", part, 0); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, part)
}

// 完成multi-chunk上传任务
func (s *Server) uploadComplete(ctx echo.Context) (err error) {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	var cas uint64

	parts := make([]*model.Part, 0)
	if err = ctx.Bind(&parts); err != nil {
		return err
	}

	uploadId := context.FileContext.UploadId
	key := fmt.Sprintf("chunks:%s", uploadId)

	partMeta := new(model.PartMeta)

	if cas, err = s.Meta.GetWithCas(key, partMeta); err != nil {
		return err
	}

	mapping := make(map[int32]*model.Part)

	for _, part := range partMeta.Parts {
		mapping[part.PartNumber] = part
	}

	meta := &model.FileMeta{
		RawKey:       partMeta.Key,
		MimeType:     partMeta.MimeType,
		LastModified: time.Now().UTC().Unix(),
	}
	for _, part := range parts {
		serverPart, exists := mapping[part.PartNumber]
		if !exists {
			ctx.Response().WriteHeader(http.StatusBadRequest)
			return nil
		}
		meta.Size += serverPart.Size
		meta.FIDs = append(meta.FIDs, serverPart.FID)
	}
	meta.FID = strings.Join(meta.FIDs, FIDSep)
	meta.FIDs = nil
	if err = s.Meta.SetTTL(meta.RawKey, meta, bucket.Basis.TTL.Expiry()); err != nil {
		return err
	}
	_ = s.Meta.Delete(key, cas)
	return ctx.JSON(http.StatusOK, meta)
}

//终止multi-chunk上传任务
func (s *Server) uploadAbort(ctx echo.Context) (err error) {
	context := ctx.(*middleware.ExtendContext)
	chunkKey := fmt.Sprintf("chunks:%s", context.FileContext.UploadId)
	return s.Meta.Delete(chunkKey, 0)
}
