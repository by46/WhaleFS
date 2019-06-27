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
	"io/ioutil"
	"os"
	"sync"
)

const (
	BufferSize   = 4 * 1024 * 1024
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

func (s *Server) fetchPreviewImg(ctx echo.Context) error {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	entity := context.FileContext.Meta

	if entity.PreviewImg == nil {
		body, _, err := s.Storage.Download(entity.FID)
		if err != nil {
			return err
		}
		defer func() { body.Close() }()

		tmp := byteBufferPool.Get().([]byte)
		defer byteBufferPool.Put(tmp)

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
	filename := uuid.New().String()
	filename = fmt.Sprintf("/tmp/%s", filename)

	file, err := os.Create(filename)
	defer func() {
		err = os.Remove(filename)
	}()
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, chunk)
	if err != nil {
		return nil, err
	}

	buf := utils.GetFrame(filename, 1)

	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket

	allBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	context.FileContext.File = new(model.FileContent)
	context.FileContext.File.Content = allBytes
	context.FileContext.File.Size = int64(len(context.FileContext.File.Content))
	context.FileContext.File.FileName = filename
	context.FileContext.File.MimeType = "image/jpeg"
	context.FileContext.File.Digest, err = utils.ContentSha1(bytes.NewReader(context.FileContext.File.Content))

	if err != nil {
		return nil, errors.WithMessage(err, "文件内容摘要错误")
	}

	key, entity := s.buildMetaFromChunk(ctx)
	if entity == nil {
		option := &common.UploadOption{
			Collection:  bucket.Basis.Collection,
			Replication: bucket.Basis.Replication,
			TTL:         bucket.Basis.TTL,
		}
		needle, err := s.Storage.Upload(option, "image/jpeg", bytes.NewBuffer(allBytes))
		if err != nil {
			return nil, err
		}
		entity = needle.AsFileMeta()
		s.saveChunk(ctx, key, entity)
	}

	return entity, err
}
