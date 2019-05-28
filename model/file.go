package model

import (
	"fmt"
	"net/textproto"
	"strings"
	"time"

	"github.com/by46/whalefs/utils"
)

const (
	ProductBucketName = "pdt"
)

// 用于存储在数据库中的文件元数据信息
type FileMeta struct {
	RawKey       string `json:"raw_key,omitempty"`
	Url          string `json:"url,omitempty"`
	FID          string `json:"fid,omitempty"`
	LastModified int64  `json:"last_modified,omitempty"`
	ETag         string `json:"etag,omitempty"`
	Size         int64  `json:"size,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	ThumbnailFID string `json:"thumbnail_fid,omitempty"`
}

func (f *FileMeta) LastModifiedTime() time.Time {
	return time.Unix(f.LastModified, 0).UTC()
}

func (f *FileMeta) HeaderISOExpires(cacheMaxAge int) string {
	return fmt.Sprintf("max-age=%d, must-revalidate", cacheMaxAge)
}

func (f *FileMeta) HeaderExpires(cacheMaxAge int) string {
	duration := uint64(cacheMaxAge) * uint64(time.Second)
	expired := time.Now().Add(time.Duration(duration)).UTC()
	return utils.TimeToRFC822(expired)
}

func (f *FileMeta) HeaderETag() string {
	return fmt.Sprintf(`"%s"`, f.ETag)
}

func (f *FileMeta) HeaderLastModified() string {
	return utils.TimeToRFC822(f.LastModifiedTime())
}

func (f *FileMeta) IsPlain() bool {
	return utils.IsPlain(f.MimeType)
}

func (f *FileMeta) IsImage() bool {
	return utils.IsImage(f.MimeType)
}

func (f *FileMeta) AsEntity(bucketName, aliasBucketName string) *FileEntity {
	_, objectName := utils.PathRemoveSegment(f.RawKey, 0)
	key := fmt.Sprintf("%s%s", bucketName, objectName)
	if ProductBucketName == aliasBucketName {
		key = strings.TrimLeft(objectName, Separator)
	}
	return &FileEntity{
		Key:  key,
		Size: f.Size,
	}
}

// 用于记录上传文件内容
type FileContent struct {
	MimeType string
	Size     int64
	Override bool
	Headers  textproto.MIMEHeader
	Content  []byte
	Width    int
	Height   int
}

func (f *FileContent) IsImage() bool {
	return utils.IsImage(f.MimeType)
}

// 上传接口返回的api
type FileEntity struct {
	Key        string `json:"key"`
	ObjectName string `json:"objectName"`
	Size       int64  `json:"size"`
}
