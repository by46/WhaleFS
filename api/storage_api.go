package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
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
	volumeId, _, _, err := parseFileId(fid)
	if err != nil {
		return nil, nil, fmt.Errorf("parse file id error %v", err)
	}

	entity, err := c.lookup(volumeId)
	if err != nil {
		return nil, nil, fmt.Errorf("lookup volume error %v", err)
	}
	for _, location := range entity.Locations {
		url := fmt.Sprintf("http://%s/%s", location.PublicUrl, fid)
		resp, err := utils.Get(url, nil)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}
		return bytes.NewReader(resp.Content), resp.Header, nil
	}
	return nil, nil, fmt.Errorf("download file error for all location %v", entity.Locations)
}

func (c *storageClient) Upload(mimeType string, body io.Reader) (*model.FileMeta, error) {
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
	if err = writer.Close(); err != nil {
		return nil, err
	}
	headers := make(http.Header)
	headers.Set(echo.HeaderContentType, writer.FormDataContentType())

	resp, err := utils.Post(fid.VolumeUrl(), headers, buff)
	if err != nil {
		return nil, err
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

func (c *storageClient) assign() (fid *FileID, err error) {
	for _, master := range c.master {
		url := fmt.Sprintf("http://%s/dir/assign", master)
		response, err := utils.Post(url, nil, nil)
		if err != nil {
			continue
		}
		fid := &FileID{}
		if err := response.Json(fid); err != nil {
			continue
		}
		return fid, nil
	}
	return nil, fmt.Errorf("assign fid from %v failed", c.master)
}

func (c *storageClient) lookup(volumeId uint32) (*VolumeEntity, error) {
	for _, host := range c.master {
		url := fmt.Sprintf("http://%s/dir/lookup?volumeId=%d", host, volumeId)
		response, err := utils.Get(url, nil)
		if err != nil {
			continue
		}
		entity := &VolumeEntity{}
		if err := response.Json(entity); err != nil {
			continue
		}
		return entity, nil

	}
	return nil, fmt.Errorf("lookup volume info %v failed", c.master)
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
