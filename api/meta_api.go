package api

import (
	"net/url"
	"strings"

	"gopkg.in/couchbase/gocb.v1"

	"github.com/by46/whalefs/common"

	"github.com/pkg/errors"
)

type metaClient struct {
	*gocb.Bucket
}

func NewMetaClient(connectionString string, password string) common.Meta {
	result, err := url.Parse(connectionString)
	if err != nil {
		panic(errors.Wrapf(err, "initialize meta client failed: %s", connectionString))
	}
	result.Scheme = "http"
	bucketName := strings.Trim(result.Path, "/")
	result.Path = ""

	cluster, err := gocb.Connect(result.Path)
	if err != nil {
		panic(err)
	}

	bucket, err := cluster.OpenBucket(bucketName, password)
	if err != nil {
		panic(err)
	}
	return &metaClient{bucket}
}

func (m *metaClient) Get(key string, value interface{}) error {
	_, err := m.Bucket.Get(key, &value)
	if err != nil && err == gocb.ErrKeyNotFound {
		return common.New(common.CodeFileNotExists)
	}
	return err
}

func (m *metaClient) Set(key string, value interface{}) error {
	_, err := m.Bucket.Insert(key, value, 0)
	return err
}

func (m *metaClient) Exists(key string) (bool, error) {
	var value interface{}
	_, err := m.Bucket.Get(key, &value)
	if err != nil {
		if err == gocb.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *metaClient) SetTTL(key string, value interface{}, ttl int) error {
	// TODO(benjamin)
	return nil
}
