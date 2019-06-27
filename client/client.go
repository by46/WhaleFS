package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/utils"
)

const (
	BufferSize = 4 * 1024 * 1024  // 4M
	ChunkSize  = 16 * 1024 * 1024 // 16M
)

var (
	byteBufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, BufferSize)
		},
	}
	headers = http.Header{echo.HeaderContentType: []string{echo.MIMEApplicationJSON}}
)

type Uploads struct {
	UploadId string `json:"uploadId"`
}

type Part struct {
	PartNumber int32  `json:"partNumber"`
	ETag       string `json:"ETag"`
}

type ChunkReader struct {
	R      io.Reader // underlying reader
	Size   int64     // chunk size
	remain int64
}

func NewChunkReader(r io.Reader, size int64) *ChunkReader {
	return &ChunkReader{
		R:      r,
		Size:   size,
		remain: size,
	}
}

func (c *ChunkReader) Read(p []byte) (n int, err error) {
	if c.remain <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > c.Size {
		p = p[0:c.Size]
	}
	n, err = c.R.Read(p)
	c.remain -= int64(n)
	return
}

type Client interface {
	// 上传小文件
	Upload(context.Context, *Options) (*FileEntity, error)
}

type httpClient struct {
	base string
}

func NewClient(options *ClientOptions) Client {
	options.Base = strings.ToLower(options.Base)
	if strings.HasPrefix(options.Base, "http://") == false {
		options.Base = fmt.Sprintf("http://%s", options.Base)
	}
	return &httpClient{base: options.Base}
}

func (c *httpClient) multiChunkUpload(ctx context.Context, options *Options) (*FileEntity, error) {
	responses := make([]*utils.Response, 0)
	defer func() {
		for _, response := range responses {
			if response != nil {
				_ = response.Close()
			}
		}
	}()
	h := http.Header{
		echo.HeaderContentType: []string{utils.MimeTypeByExtension(options.FileName)},
	}
	resp, err := utils.Post(c.initMultiChunkUploadUrl(options.key()), h, nil)
	responses = append(responses, resp)
	if err != nil {
		return nil, errors.Wrap(err, "上传文件失败")
	}
	uploads := new(Uploads)
	if err := resp.Json(uploads); err != nil {
		return nil, errors.Wrap(err, "初始化Multi-Chunk上传失败")
	}
	parts := make([]*Part, 0)
	partNumber := int32(1)
	count := ChunkSize
	chunk := make([]byte, ChunkSize)
	for count == ChunkSize {
		select {
		case <-ctx.Done():
			// TODO(benjamin): abort multi-chunk
			return nil, ErrAbort
		default:
			count, err = options.Content.Read(chunk)
			if err != nil && err == io.EOF {
				break
			} else if err != nil {
				return nil, errors.WithStack(err)
			}
			u := c.chunkUploadUrl(options.key(), uploads.UploadId, partNumber)
			resp, err := utils.Post(u, nil, bytes.NewReader(chunk[:count]))
			responses = append(responses, resp)
			if err != nil {
				return nil, errors.Wrap(err, "上传文件失败")
			}
			part := new(Part)
			if err := resp.Json(part); err != nil {
				return nil, errors.Wrap(err, "上传文件失败")
			}
			parts = append(parts, part)
			partNumber += 1
		}
	}
	u := c.completeUploadUrl(options.key(), uploads.UploadId)
	content := bytes.NewBuffer(nil)
	if err = json.NewEncoder(content).Encode(parts); err != nil {
		return nil, errors.Wrap(err, "上传文件失败")
	}
	resp, err = utils.Post(u, headers, content)
	responses = append(responses, resp)
	if err != nil {
		return nil, errors.Wrap(err, "上传文件失败")
	}
	entity := new(FileEntity)
	err = resp.Json(entity)
	return entity, errors.WithStack(err)
}

func (c *httpClient) singleUpload(ctx context.Context, options *Options) (*FileEntity, error) {
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
	tmp := byteBufferPool.Get().([]byte)
	defer byteBufferPool.Put(tmp)
	for {
		n, err := options.Content.Read(tmp)
		if err != nil && err != io.EOF {
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

// TODO(benjamin): 完善
func (c *httpClient) Upload(ctx context.Context, options *Options) (*FileEntity, error) {
	if options.MultiChunk {
		return c.multiChunkUpload(ctx, options)
	}
	return c.singleUpload(ctx, options)
}

func (c *httpClient) uploadUrl() string {
	return c.base
}
func (c *httpClient) initMultiChunkUploadUrl(filename string) string {
	query := make(url.Values)
	query.Set("uploads", "")
	u, _ := url.Parse(c.base)
	u.Path = filename
	u.RawQuery = query.Encode()
	return u.String()
}

func (c *httpClient) chunkUploadUrl(filename, uploadId string, partNumber int32) string {
	query := make(url.Values)
	query.Set("uploadId", uploadId)
	query.Set("partNumber", fmt.Sprintf("%v", partNumber))
	u, _ := url.Parse(c.base)
	u.Path = filename
	u.RawQuery = query.Encode()
	return u.String()
}

func (c *httpClient) completeUploadUrl(filename, uploadId string) string {
	query := make(url.Values)
	query.Set("uploadId", uploadId)
	u, _ := url.Parse(c.base)
	u.Path = filename
	u.RawQuery = query.Encode()
	return u.String()
}
