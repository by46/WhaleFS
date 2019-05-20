package client

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/utils"
)

const (
	BufferSize = 4 * 1024 // 4M
)

type Client interface {
	// 上传小文件
	Upload(context.Context, *Options) (*FileEntity, error)
}

type httpClient struct {
	base string
}

func NewClient(options *ClientOptions) Client {
	return &httpClient{base: options.Base}
}

// TODO(benjamin): 完善
func (c *httpClient) Upload(ctx context.Context, options *Options) (*FileEntity, error) {
	buff := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buff)

	if err := writer.WriteField("key", options.key()); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := writer.WriteField("override", options.getOverride()); err != nil {
		return nil, errors.WithStack(err)
	}

	h := make(textproto.MIMEHeader)
	h.Set(echo.HeaderContentDisposition, fmt.Sprintf(`form-data; name="file"; filename="%s"`, options.FileName))
	partition, _ := writer.CreatePart(h)
	tmp := make([]byte, BufferSize)
	for {
		n, err := options.Content.Read(tmp)
		if err != nil {
			return nil, errors.Wrap(err, "读取文件内容失败")
		}
		_, err = partition.Write(tmp[:n])
		if err != nil {
			return nil, errors.Wrap(err, "构建上传表单错误")
		}
		if n < BufferSize {
			break
		}
	}
	if err := writer.Close(); err != nil {
		return nil, errors.Wrap(err, "关闭表单失败")
	}
	headers := make(http.Header)
	headers.Set(echo.HeaderContentType, writer.FormDataContentType())
	headers.Set(echo.HeaderContentLength, fmt.Sprintf("%d", buff.Len()))

	resp, err := utils.Post(c.uploadUrl(), headers, buff)
	if resp != nil {
		defer func() {
			_ = resp.Close()
		}()
	}
	if err != nil {
		return nil, errors.Wrap(err, "上传文件失败")
	}
	entity := new(FileEntity)
	err = resp.Json(entity)
	return entity, errors.WithStack(err)
}

func (c *httpClient) uploadUrl() string {
	return c.base
}
