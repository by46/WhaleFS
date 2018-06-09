package model

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"whalefs/common"

	"github.com/mholt/binding"
	"strings"
)

type FileEntity struct {
	RawKey       string `json:"raw_key"`
	Url          string `json:"url"`
	LastModified int64  `json:"last_modified"`
	ETag         string `json:"etag"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
}

func (f *FileEntity) LastModifiedTime() time.Time {
	return time.Unix(f.LastModified, 0).UTC()
}

func (f *FileEntity) HeaderISOExpires(cacheMaxAge int) string {
	return fmt.Sprintf("max-age=%d, must-revalidate", cacheMaxAge)
}

func (f *FileEntity) HeaderExpires(cacheMaxAge int) string {
	duration := uint64(cacheMaxAge) * uint64(time.Second)
	expired := time.Now().Add(time.Duration(duration)).UTC()
	return common.TimeToRFC822(expired)
}

func (f *FileEntity) HeaderETag() string {
	return fmt.Sprintf(`"%s"`, f.ETag)
}

func (f *FileEntity) HeaderLastModified() string {
	return common.TimeToRFC822(f.LastModifiedTime())
}

func (f *FileEntity) IsPlain() bool {
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

type FileObject struct {
	Key         string
	ExtractFile bool
	Content     *multipart.FileHeader
}

func (f *FileObject) FieldMap(r *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&f.Key: binding.Field{
			Form:     "key",
			Required: true,
		},
		&f.Content: binding.Field{
			Form:     "file",
			Required: true,
		},
	}
}
