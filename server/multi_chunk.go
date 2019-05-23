package server

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
)

// 生成multi-chunk上传任务
func (s *Server) uploads(ctx echo.Context) (err error) {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket

	uploadId := uuid.New().String()
	uploads := &model.Uploads{
		Bucket:   bucket.Name,
		Key:      context.FileContext.Key,
		UploadId: uploadId,
	}
	partMeta := &model.PartMeta{
		Key:   context.FileContext.Key,
		Parts: []*model.Part{},
	}
	key := fmt.Sprintf("chunks:%s", uploadId)
	if err = s.Meta.SetTTL(key, partMeta, TTLChunk); err != nil {
		return err
	}
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

	_ = fmt.Sprintf("chunks:%s", context.FileContext.UploadId)

	part := &model.Part{
		ETag: key,
	}
	return s.Meta.SubListAppend(key, "parts", part, 0)
}

// 完成multi-chunk上传任务
func (s *Server) uploadComplete(ctx echo.Context) (err error) {
	return nil
}

//终止multi-chunk上传任务
func (s *Server) uploadAbort(ctx echo.Context) (err error) {
	return
}
