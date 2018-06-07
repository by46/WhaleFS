package api

import (
	"github.com/couchbase/go-couchbase"
)

type IMeta interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	SetTTL(key string, value interface{}, ttl int) error
}

type metaClient struct {
	*couchbase.Bucket
}

func NewMetaClient(url, bucketName string) IMeta {
	bucket, err := couchbase.GetBucket(url, "default", bucketName)
	if err != nil {
		panic(err)
	}
	return &metaClient{bucket}
}

func (m *metaClient) Get(key string, value interface{}) error {
	return m.Bucket.Get(key, &value)
}
func (m *metaClient) Set(key string, value interface{}) error {
	return nil
}
func (m *metaClient) SetTTL(key string, value interface{}, ttl int) error {
	return nil
}
