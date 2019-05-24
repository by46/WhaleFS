package model

import (
	"fmt"
	"net/textproto"
	"time"

	"github.com/by46/whalefs/utils"
)

type FileMeta struct {
	RawKey       string   `json:"raw_key,omitempty"`
	Url          string   `json:"url,omitempty"`
	FID          string   `json:"fid,omitempty"`
	FIDs         []string `json:"fids,omitempty"`
	LastModified int64    `json:"last_modified,omitempty"`
	ETag         string   `json:"etag,omitempty"`
	Size         int64    `json:"size,omitempty"`
	Width        int      `json:"width,omitempty"`
	Height       int      `json:"height,omitempty"`
	MimeType     string   `json:"mime_type,omitempty"`
	ThumbnailFID string   `json:"thumbnail_fid,omitempty"`
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
