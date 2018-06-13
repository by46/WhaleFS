package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/gomemcached"
	"github.com/pkg/errors"
)

var (
	ErrNoEntity = fmt.Errorf("entity not exists")
)

type IMeta interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	SetTTL(key string, value interface{}, ttl int) error
}

type metaClient struct {
	*couchbase.Bucket
}

func NewMetaClient(connectionString string) IMeta {
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
				return ErrNoEntity
			}
		}
	}
	return err
}
func (m *metaClient) Set(key string, value interface{}) error {
	return m.Bucket.Set(key, 0, value)
}
func (m *metaClient) SetTTL(key string, value interface{}, ttl int) error {
	return nil
}
