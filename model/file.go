package model

import (
	"fmt"
	"io"
	"net/textproto"
	"strings"
	"time"

	"github.com/by46/whalefs/utils"
)

type FileMeta struct {
	RawKey       string `json:"raw_key,omitempty"`
	Url          string `json:"url,omitempty"`
	FID          string `json:"fid,omitempty"`
	LastModified int64  `json:"last_modified,omitempty"`
	ETag         string `json:"etag,omitempty"`
	Size         int64  `json:"size,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
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
	if f.MimeType == "" {
		return false
	}

	if strings.HasPrefix(f.MimeType, "text/") {
		return true
	}
	switch f.MimeType {
	default:
		return false
	case "application/javascript", "application/x-javascript":
		return true
	}
}

func (f *FileMeta) IsImage() bool {
	if f.MimeType == "" {
		return false
	}

	return strings.HasPrefix(f.MimeType, "image/")
}

type FileContent struct {
	MimeType string
	Size     int64
	Override bool
	Headers  textproto.MIMEHeader
	Content  io.Reader
}
