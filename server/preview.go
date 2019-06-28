package server

import (
	"bytes"
	"fmt"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"io"
	"os"
)

const (
	BufferSize   = 4 * 1024 * 1024
	ParamPreview = "preview"
	ParamSize    = "size"
	mimeJpeg     = "image/jpeg"
)

func (s *Server) fetchPreviewImg(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	entity := context.FileContext.Meta

	if entity.PreviewImg == nil {
		body, _, err := s.Storage.Download(entity.FID)
		if err != nil {
			return err
		}

		defer func() {
			e := body.Close()
			if e != nil {
				s.Logger.Errorf("close reader failed: %v", err)
			}
		}()

		buffer := bytes.NewBuffer(nil)
		_, err = io.CopyN(buffer, body, BufferSize)
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "读取文件内容失败")
		}

		previewImgChunkMeta, err := s.generatePreviewImg(ctx, buffer)
		if err != nil {
			return err
		}

		entity.PreviewImg = &model.PreviewImgMeta{
			ThumbnailMeta: model.ThumbnailMeta{
				FID:  previewImgChunkMeta.FID,
				Size: previewImgChunkMeta.Size,
			},
			MimeType: previewImgChunkMeta.MimeType,
		}

		if err = s.Meta.SetTTL(entity.RawKey, entity, bucket.Basis.TTL.Expiry()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) generatePreviewImg(ctx echo.Context, chunk io.Reader) (*model.FileMeta, error) {
	filename := fmt.Sprintf("/tmp/%s", uuid.New().String())

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		e := os.Remove(filename)
		if e != nil {
			s.Logger.Errorf("remove file failed: %v", err)
		}
	}()

	_, err = io.Copy(file, chunk)
	if err != nil {
		return nil, err
	}

	buf := utils.GetFrame(filename, 1)

	fileContent := new(model.FileContent)
	fileContent.Content = buf.Bytes()
	fileContent.Size = int64(len(fileContent.Content))
	fileContent.FileName = filename
	fileContent.MimeType = mimeJpeg
	fileContent.Digest, err = utils.ContentSha1(bytes.NewReader(fileContent.Content))
	if err != nil {
		return nil, errors.WithMessage(err, "文件内容摘要错误")
	}

	context := ctx.(*middleware.ExtendContext)
	context.FileContext.File = fileContent
	bucket := context.FileContext.Bucket

	key, entity := s.buildMetaFromChunk(ctx)
	if entity == nil {
		option := &common.UploadOption{
			Collection:  bucket.Basis.Collection,
			Replication: bucket.Basis.Replication,
			TTL:         bucket.Basis.TTL,
		}
		needle, err := s.Storage.Upload(option, mimeJpeg, buf)
		if err != nil {
			return nil, err
		}
		entity = needle.AsFileMeta()
		s.saveChunk(ctx, key, entity)
	}

	return entity, err
}
