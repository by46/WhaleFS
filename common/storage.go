package common

import (
	"io"
	"net/http"
	"net/url"

	"github.com/by46/whalefs/model"
)

const (
	QueryNameCollection  = "collection"
	QueryNameReplication = "replication"
	QueryNameTTL         = "ttl"
)

type UploadOption struct {
	Collection  string
	Replication string
	TTL         model.TTL
}

func (o *UploadOption) Encode() string {
	query := make(url.Values)

	if o.Collection != "" {
		query.Set(QueryNameCollection, o.Collection)
	}
	if o.Replication != "" {
		query.Set(QueryNameReplication, o.Replication)
	}
	if o.TTL != "" {
		query.Set(QueryNameTTL, o.TTL.String())
	}
	return query.Encode()
}

type Storage interface {
	Download(fid string) (io.Reader, http.Header, error)
	Upload(option *UploadOption, mimeType string, body io.Reader) (needle *model.Needle, err error)
}
