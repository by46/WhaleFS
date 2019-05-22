package common

import (
	"io"
	"net/http"

	"github.com/by46/whalefs/model"
)

type UploadOption struct {
	Collection  string
	Replication string
	TTL         string
}

type Storage interface {
	Download(fid string) (io.Reader, http.Header, error)
	Upload(option *UploadOption, mimeType string, body io.Reader) (entity *model.FileMeta, err error)
}
