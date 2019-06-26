package server

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
)

var (
	ChunkBuffPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, constant.ChunkSize)
		},
	}
)

func (s *Server) uploadLargeFile(ctx echo.Context) (entity *model.FileEntity, err error) {
	var n int64
	var part *model.Part

	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	content := bytes.NewReader(context.FileContext.File.Content)

	fileContext := &model.FileContext{
		Bucket:     bucket,
		Key:        context.FileContext.Key,
		ObjectName: context.FileContext.ObjectName,
		BucketName: context.FileContext.BucketName,
	}
	r, _ := http.NewRequest("POST", "", nil)
	r.Header.Set(echo.HeaderContentType, context.FileContext.File.MimeType)
	fakeCtx := s.app.NewContext(r, nil)
	fakeContext := &middleware.ExtendContext{fakeCtx, fileContext}
	uploads, err := s.uploads(fakeContext)
	if err != nil {
		return nil, err
	}
	fileContext.UploadId = uploads.UploadId
	parts := make(model.Parts, 0)
	for {
		fileContext.PartNumber += 1
		writer := bytes.NewBuffer(nil)
		n, err = io.CopyN(writer, content, constant.ChunkSize)
		if err != nil && err != io.EOF {
			break
		}
		content := writer.Bytes()
		err = fileContext.ParseFileContentFromBytes(content)
		if err != nil {
			break
		}
		part, err = s.uploadPart(fakeContext)
		if err != nil {
			break
		}
		parts = append(parts, part)
		if n < constant.ChunkSize {
			break
		}
	}
	if err != nil {
		err2 := s.uploadAbort(fakeContext)
		s.Logger.Errorf("abort multipart upload failed, upload id: %s, %v", uploads.UploadId, err2)
		return nil, err
	}
	return s.uploadComplete(fakeContext, parts)
}
