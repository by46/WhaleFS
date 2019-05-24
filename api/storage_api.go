package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

const (
	MimeSize     = 512
	QueryNameTTL = "ttl"
	FIDSep       = "|"
)

var (
	MimeGuessBuffPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, MimeSize)
		},
	}
)

type FileID struct {
	Count     int    `json:"count,omitempty"`
	FID       string `json:"fid,omitempty"`
	PublicUrl string `json:"publicUrl,omitempty"`
	Url       string `json:"url,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (f *FileID) String() string {
	return fmt.Sprintf("http://%s/%s", f.PublicUrl, f.FID)
}

type LocationEntity struct {
	Url       string `json:"url"`
	PublicUrl string `json:"publicUrl"`
}

func (l *LocationEntity) String() string {
	return fmt.Sprintf(`{Url: "%s", PublicUrl: "%s"}`, l.Url, l.PublicUrl)
}

type VolumeEntity struct {
	VolumeId  string             `json:"volumeId"`
	Locations [] *LocationEntity `json:"locations"`
}

type storageClient struct {
	master []string
}

func NewStorageClient(masters []string) common.Storage {
	return &storageClient{
		master: masters,
	}
}

func (c *storageClient) Download(fid string) (io.Reader, http.Header, error) {
	if strings.Contains(fid, "|") {
		return c.downloadChunks(strings.Split(fid, FIDSep))
	}
	volumeId, _, _, err := parseFileId(fid)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "文件ID%s格式错误, 无法解析", fid)
	}

	entity := c.lookup(volumeId)
	if entity == nil {
		return nil, nil, common.ErrFileNotFound
	}
	responses := make([]*utils.Response, 0)
	defer func() {
		for _, resp := range responses {
			_ = resp.Close()
		}
	}()
	for _, location := range entity.Locations {
		u := fmt.Sprintf("http://%s/%s", location.PublicUrl, fid)
		resp, err := utils.Get(u, nil)
		if resp != nil {
			responses = append(responses, resp)
		}
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}
		return bytes.NewReader(resp.Content), resp.Header, nil
	}
	return nil, nil, common.ErrFileNotFound
}

func (c *storageClient) Upload(option *common.UploadOption, mimeType string, body io.Reader) (*model.FileMeta, error) {
	var size int64
	var preReadSize int
	var err error
	var mimeBuff []byte

	fid, err := c.assign(option)
	if err != nil {
		return nil, err
	}

	if mimeType == "" {
		mimeBuff = MimeGuessBuffPool.Get().([]byte)
		defer MimeGuessBuffPool.Put(mimeBuff)
		preReadSize, err = body.Read(mimeBuff)
		if err != nil && err != io.EOF {
			return nil, errors.Wrap(err, "网络流读取失败")
		}
		mimeType = http.DetectContentType(mimeBuff[:preReadSize])
	}

	buff := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buff)
	h := make(textproto.MIMEHeader)
	h.Set(echo.HeaderContentDisposition, `form-data; name="File"; filename="file.txt"`)
	h.Set(echo.HeaderContentType, mimeType)
	partition, _ := writer.CreatePart(h)

	if mimeBuff != nil {
		if _, err = partition.Write(mimeBuff[:preReadSize]); err != nil {
			return nil, errors.Wrap(err, "构造Form表单失败")
		}
	}

	if size, err = io.Copy(partition, body); err != nil {
		return nil, errors.Wrap(err, "构造Form表单失败")
	}
	if err = writer.Close(); err != nil {
		return nil, errors.Wrap(err, "构造Form表单失败")
	}
	headers := make(http.Header)
	headers.Set(echo.HeaderContentType, writer.FormDataContentType())

	resp, err := utils.Post(fid.String(), headers, buff)
	if resp != nil {
		defer func() {
			_ = resp.Close()
		}()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "上传文件块失败, 上传地址%s", fid.String())
	}

	entity := &model.FileMeta{
		FID:          fid.FID,
		ETag:         strings.Trim(resp.Header.Get(utils.HeaderETag), `"`),
		LastModified: time.Now().UTC().Unix(),
		Size:         size + int64(preReadSize),
		MimeType:     mimeType,
	}
	return entity, nil
}

func (c *storageClient) downloadChunks(fids []string) (io.Reader, http.Header, error) {
	return NewChunksReader(c, fids), nil, nil
}
func (c *storageClient) uploadUrl(option *common.UploadOption, fid FileID) string {
	query := make(url.Values)
	if option.TTL != "" {
		query.Set(QueryNameTTL, option.TTL.String())
	}
	u, _ := url.Parse(fid.String())
	u.RawQuery = query.Encode()
	return u.String()
}

func (c *storageClient) assignUrl(master string, option *common.UploadOption) string {
	u := &url.URL{
		Scheme:   "http",
		Host:     master,
		Path:     "/dir/assign",
		RawQuery: option.Encode(),
	}
	return u.String()
}

func (c *storageClient) assign(option *common.UploadOption) (fid *FileID, err error) {
	responses := make([]*utils.Response, 0)
	defer func() {
		for _, resp := range responses {
			_ = resp.Close()
		}
	}()
	for _, master := range c.master {
		url := c.assignUrl(master, option)
		response, err := utils.Post(url, nil, nil)
		if response != nil {
			responses = append(responses, response)
		}
		if err != nil {
			continue
		}
		fid := &FileID{}
		if err := response.Json(fid); err != nil {
			continue
		}
		return fid, nil
	}
	return nil, errors.Errorf("分配文件ID失败, master节点: %v", c.master)
}

func (c *storageClient) lookup(volumeId uint32) *VolumeEntity {
	responses := make([]*utils.Response, 0)
	defer func() {
		for _, resp := range responses {
			_ = resp.Close()
		}
	}()
	for _, host := range c.master {
		u := fmt.Sprintf("http://%s/dir/lookup?volumeId=%d", host, volumeId)
		response, err := utils.Get(u, nil)
		if response != nil {
			responses = append(responses, response)
		}
		if err != nil {
			continue
		}
		entity := &VolumeEntity{}
		if err := response.Json(entity); err != nil {
			continue
		}
		return entity
	}
	return nil
}

func parseFileId(id string) (volumeId uint32, fileId uint64, cookie uint32, err error) {
	var value uint64
	tmp := strings.Split(id, ",")
	if len(tmp) < 2 || len(tmp[1]) <= 8 {
		return 0, 0, 0, fmt.Errorf("invalid format %s for file id", id)
	}
	value, err = strconv.ParseUint(tmp[0], 10, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	volumeId = uint32(value)

	l := len(tmp[1])
	part1, part2 := tmp[1][:l-8], tmp[1][l-8 : l]
	fileId, err = strconv.ParseUint(part1, 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	value, err = strconv.ParseUint(part2, 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	cookie = uint32(value)
	return

}
