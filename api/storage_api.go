package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/by46/whalefs/common"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
	"github.com/labstack/echo"
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
	*http.Client
}

func NewStorageClient(masters []string) common.Storage {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	return &storageClient{
		master: masters,
		Client: client,
	}
}

func (c *storageClient) Download(fid string) (io.ReadCloser, http.Header, error) {
	volumeId, _, _, err := parseFileId(fid)
	if err != nil {
		return nil, nil, fmt.Errorf("parse file id error %v", err)
	}

	entity, err := c.lookup(volumeId)
	if err != nil {
		return nil, nil, fmt.Errorf("lookup volume error %v", err)
	}
	for _, location := range entity.Locations {
		response, err := c.Get(fmt.Sprintf("http://%s/%s", location.PublicUrl, fid))
		if err != nil {
			continue
		}
		return response.Body, response.Header, nil
	}
	return nil, nil, fmt.Errorf("download file error for all location %v", entity.Locations)
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
		FID:          fid.FID,
		ETag:         strings.Trim(response.Header.Get(utils.HeaderETag), `"`),
		LastModified: time.Now().UTC().Unix(),
		Size:         size + int64(preReadSize),
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

func (c *storageClient) lookup(volumeId uint32) (*VolumeEntity, error) {
	for _, host := range c.master {
		url := fmt.Sprintf("http://%s/dir/lookup?volumeId=%d", host, volumeId)
		response, err := c.Get(url)
		if err != nil {
			continue
		}
		data, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			continue
		}

		entity := &VolumeEntity{}
		if err := json.Unmarshal(data, entity); err != nil {
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
