package model

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/by46/whalefs/common"

	"github.com/mholt/binding"
)

type FileEntity struct {
	RawKey       string `json:"raw_key,omitempty"`
	Url          string `json:"url,omitempty"`
	FID          string `json:"fid,omitempty"`
	LastModified int64  `json:"last_modified,omitempty"`
	ETag         string `json:"etag,omitempty"`
	Size         int64  `json:"size,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
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
	BucketName  string
	ExtractFile bool
	Content     *multipart.FileHeader
}

func (f *FileObject) FieldMap(r *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&f.Key: binding.Field{
			Form:     "key",
			Required: true,
		},
		&f.BucketName: binding.Field{
			Form:     "key",
			Required: true,
			Binder: func(name string, values []string, errors binding.Errors) binding.Errors {
				return errors
			},
		},
		&f.Content: binding.Field{
			Form:     "file",
			Required: true,
		},
	}
}
