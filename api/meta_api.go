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
	bucketName := strings.Trim(result.Path, "/")
	user := result.User
	result.Scheme = "http"
	result.User = nil
	result.Path = ""

	cluster, err := gocb.Connect(result.String())
	if err != nil {
		panic(err)
	}
	if password, passwordSet := user.Password(); passwordSet {
		err = cluster.Authenticate(gocb.PasswordAuthenticator{
			Username: user.Username(),
			Password: password,
		})
		if err != nil {
			panic(err)
		}
	}

	bucket, err := cluster.OpenBucket(bucketName, "")
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
	_, err := m.Bucket.Upsert(key, value, 0)
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
