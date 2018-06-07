package model

import (
	"fmt"
	"time"

	"whalefs/common"
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
