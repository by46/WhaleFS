package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/by46/whalefs/common"

	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/gomemcached"
	"github.com/pkg/errors"
)

var (
	ErrNoEntity = fmt.Errorf("entity not exists")
)

type metaClient struct {
	*couchbase.Bucket
}

func NewMetaClient(connectionString string) common.Meta {
	result, err := url.Parse(connectionString)
	if err != nil {
		panic(errors.Wrapf(err, "initialize meta client failed: %s", connectionString))
	}
	result.Scheme = "http"
	bucketName := strings.Trim(result.Path, "/")
	result.Path = ""

	bucket, err := couchbase.GetBucket(result.String(), "default", bucketName)
	if err != nil {
		panic(err)
	}
	return &metaClient{bucket}
}

func (m *metaClient) Get(key string, value interface{}) error {
	err := m.Bucket.Get(key, &value)
	if err != nil {
		if err2, success := err.(*gomemcached.MCResponse); success {
			if err2.Status == gomemcached.KEY_ENOENT {
				return common.New(common.CodeFileNotExists)
			}
		}
	}
	return err
}

func (m *metaClient) Set(key string, value interface{}) error {
	return m.Bucket.Set(key, 0, value)
}

func (m *metaClient) Exists(key string) (bool, error) {
	return true, nil
}

func (m *metaClient) SetTTL(key string, value interface{}, ttl int) error {
	// TODO(benjamin)
	return nil
}
