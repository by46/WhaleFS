package common

import (
	"io"
	"net/http"

	"github.com/by46/whalefs/model"
)

type Storage interface {
	Download(url string) (io.Reader, http.Header, error)
	Upload(mimeType string, body io.Reader) (entity *model.FileMeta, err error)
}
