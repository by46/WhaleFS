package common

import (
	"io"
	"net/http"

	"github.com/by46/whalefs/model"
)

type Storage interface {
	Download(url string) (io.ReadCloser, http.Header, error)
	Upload(mimeType string, body io.Reader) (entity *model.FileEntity, err error)
}
