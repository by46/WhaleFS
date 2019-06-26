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
)

func (s *Server) generatePreviewImg(ctx echo.Context, chunk io.ReadCloser) (*model.FileMeta, error) {
	defer func() {
		err := chunk.Close()
		if err != nil {
			s.Logger.Error(err)
		}
	}()
	filename := uuid.New().String()
	filename = fmt.Sprintf("/Users/mark.c.jiang/Code/workspace/whalefs/tmp/%s", filename)

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
