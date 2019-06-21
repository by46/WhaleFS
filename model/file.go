package model

import (
	"fmt"
	"net/textproto"
	"path"
	"strings"
	"time"

	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/utils"
)

// 用于存储缩略图信息
type ThumbnailMeta struct {
	FID  string `json:"fid,omitempty"`
	ETag string `json:"etag,omitempty"`
	Size int64  `json:"size,omitempty"`
}

type Thumbnails map[string]*ThumbnailMeta

// 用于存储在数据库中的文件元数据信息
type FileMeta struct {
	RawKey       string     `json:"raw_key,omitempty"`
	Url          string     `json:"url,omitempty"`
	FID          string     `json:"fid,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	ThumbnailFID string     `json:"thumbnail_fid,omitempty"`
	ETag         string     `json:"etag,omitempty"`
	LastModified int64      `json:"last_modified,omitempty"`
	Size         int64      `json:"size,omitempty"`
	Width        int        `json:"width,omitempty"`
	Height       int        `json:"height,omitempty"`
	Thumbnails   Thumbnails `json:"thumbnails,omitempty"`
	IsRandomName bool       `json:"is_random_name,omitempty"`
	WaterMark    string     `json:"water_mark,omitempty"`
	Bucket       string     `json:"bucket,omitempty"`
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

func (f *FileMeta) AsEntity(bucketName, fileName string) *FileEntity {
	aliasBucketName, objectName := utils.PathRemoveSegment(f.RawKey, 0)
	key := fmt.Sprintf("%s%s", aliasBucketName, objectName)
	if constant.BucketPdt == aliasBucketName {
		key = strings.TrimLeft(objectName, constant.Separator)
	} else if f.IsRandomName {
		key = fmt.Sprintf("%s/Original%s", aliasBucketName, objectName)
	}
	return &FileEntity{
		Key:      key,
		Url:      key,
		Title:    path.Base(key),
		Original: fileName,
		Message:  "上传成功",
		State:    "SUCCESS",
		Size:     f.Size,
	}
}

// 用于记录上传文件内容
type FileContent struct {
	FileName  string
	Digest    string
	MimeType  string
	Size      int64
	Override  bool
	Headers   textproto.MIMEHeader
	Content   []byte
	Width     int
	Height    int
	Extension string
	WaterMark string // 记录上传时设定的水印
}

func (f *FileContent) IsImage() bool {
	return utils.IsImage(f.MimeType)
}

// 上传接口返回的api
type FileEntity struct {
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	Url      string `json:"url"`      // legacy property
	Title    string `json:"title"`    // legacy property
	Message  string `json:"message"`  // legacy property
	State    string `json:"state"`    // legacy property
	Original string `json:"original"` // legacy property
}
