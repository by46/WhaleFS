package api

import (
	"io"
	"strings"
	"net/http"
	"time"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"mime/multipart"
	"net/textproto"
	"github.com/labstack/echo"
	"whalefs/model"
	"whalefs/common"
	"sync"
)

const (
	MimeSize = 512
)

var (
	MimeGuessBuffPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, MimeSize)
		},
	}
)

type IStorage interface {
	Download(url string) (io.ReadCloser, http.Header, error)
	Upload(mimeType string, body io.Reader) (entity *model.FileEntity, err error)
}

type FileID struct {
	Count     int    `json:"count,omitempty"`
	FID       string `json:"fid,omitempty"`
	PublicUrl string `json:"publicUrl,omitempty"`
	Url       string `json:"url,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (f *FileID) VolumeUrl() string {
	return fmt.Sprintf("http://%s/%s", f.PublicUrl, f.FID)
}

type storageClient struct {
	master []string
	*http.Client
}

func NewStorageClient(master string) IStorage {
	masters := strings.Split(master, ",")
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	return &storageClient{
		master: masters,
		Client: client,
	}
}

func (c *storageClient) Download(url string) (io.ReadCloser, http.Header, error) {
	response, err := c.Get(url)
	if err != nil {
		return nil, nil, err
	}
	return response.Body, response.Header, nil
}

func (c *storageClient) Upload(mimeType string, body io.Reader) (*model.FileEntity, error) {
	var size int64
	var preReadSize int
	var err error
	var mimeBuff []byte

	fid, err := c.assign()
	if err != nil {
		return nil, err
	}

	if mimeType == "" {
		mimeBuff = MimeGuessBuffPool.Get().([]byte)
		defer MimeGuessBuffPool.Put(mimeBuff)
		preReadSize, err = body.Read(mimeBuff)
		if err != nil && err != io.EOF {
			return nil, err
		}
		mimeType = http.DetectContentType(mimeBuff[:preReadSize])
	}

	buff := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buff)
	h := make(textproto.MIMEHeader)
	// TODO(benjamin): process filename and guess mimeType
	h.Set(echo.HeaderContentDisposition, `form-data; name="File"; filename="file.txt"`)
	h.Set(echo.HeaderContentType, mimeType)
	partition, _ := writer.CreatePart(h)

	if mimeBuff != nil {
		if _, err = partition.Write(mimeBuff[:preReadSize]); err != nil {
			return nil, err
		}
	}

	if size, err = io.Copy(partition, body); err != nil {
		return nil, err
	}
	writer.Close()
	response, err := c.Post(fid.VolumeUrl(), writer.FormDataContentType(), bytes.NewBuffer(buff.Bytes()))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK && response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("upload content error, code %s", response.Status)
	}
	entity := &model.FileEntity{
		Url:          fid.VolumeUrl(),
		ETag:         strings.Trim(response.Header.Get(common.HeaderETag), `"`),
		LastModified: time.Now().UTC().Unix(),
		Size:         size,
		MimeType:     mimeType,
	}
	return entity, nil
}

func (c *storageClient) assign() (fid *FileID, err error) {
	for _, master := range c.master {
		url := fmt.Sprintf("http://%s/dir/assign", master)
		response, err := c.Post(url, "", nil)
		if err != nil {
			continue
		}
		data, err := ioutil.ReadAll(response.Body)
		response.Body.Close()

		if err != nil {
			continue
		}
		fid := &FileID{}
		if err := json.Unmarshal(data, fid); err != nil {
			continue
		}
		return fid, nil
	}
	return nil, fmt.Errorf("assign fid from %v failed", c.master)
}
