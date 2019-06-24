package common

import (
	"io"
	"net/http"
	"net/url"

	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
)

type UploadOption struct {
	Collection  string
	Replication string
	TTL         model.TTL
}

func (o *UploadOption) Encode() string {
	query := make(url.Values)

	if o.Collection != "" {
		query.Set(constant.QueryNameCollection, o.Collection)
	}
	if o.Replication != "" {
		query.Set(constant.QueryNameReplication, o.Replication)
	}
	if o.TTL != "" {
		query.Set(constant.QueryNameTTL, o.TTL.String())
	}
	return query.Encode()
}

type Storage interface {
	Download(fid string) (io.ReadCloser, http.Header, error)
	Upload(option *UploadOption, mimeType string, body io.Reader) (needle *model.Needle, err error)
}
