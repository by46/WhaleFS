package api

import (
	"net/url"
	"strings"

	"gopkg.in/couchbase/gocb.v1"

	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
)

type metaClient struct {
	*gocb.Bucket
}

func NewMetaClient(connectionString string) common.Dao {
	result, err := url.Parse(connectionString)
	if err != nil {
		panic(errors.Wrapf(err, "initialize meta client failed: %s", connectionString))
	}
	bucketName := strings.Trim(result.Path, "/")
	user := result.User
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

func (m *metaClient) Get(key string, value interface{}) (err error) {
	if _, err = m.Bucket.Get(key, &value); err == nil {
		return
	} else if err == gocb.ErrKeyNotFound {
		return common.ErrKeyNotFound
	}
	return errors.Wrapf(err, "获取数据失败, key: %s", key)
}

func (m *metaClient) GetWithCas(key string, value interface{}) (uint64, error) {
	cas, err := m.Bucket.Get(key, &value)
	if err == nil {
		return uint64(cas), nil
	} else if err == gocb.ErrKeyNotFound {
		return 0, common.ErrFileNotFound
	}
	return 0, errors.Wrapf(err, "获取数据失败, key: %s", key)
}

func (m *metaClient) Set(key string, value interface{}) error {
	_, err := m.Bucket.Upsert(key, value, 0)
	return errors.Wrapf(err, "设置数据失败, key: %s", key)
}

func (m *metaClient) Exists(key string) (bool, error) {
	var value interface{}
	_, err := m.Bucket.Get(key, &value)
	if err != nil {
		if err == gocb.ErrKeyNotFound {
			return false, nil
		}
		return false, errors.Wrapf(err, "获取数据失败, key: %s", key)
	}
	return true, nil
}

func (m *metaClient) SetTTL(key string, value interface{}, ttl uint32) error {
	_, err := m.Bucket.Upsert(key, value, ttl)
	return errors.Wrapf(err, "设置数据失败, key: %s", key)
}

// TODO(benjamin): 优化couchbase操作, 处理cas不匹配的情况
func (m *metaClient) Delete(key string, cas uint64) (err error) {
	_, err = m.Bucket.Remove(key, gocb.Cas(cas))
	if err == gocb.ErrKeyNotFound {
		return common.ErrFileNotFound
	}
	return
}

func (m *metaClient) Query(n1sql string, params interface{}) (gocb.QueryResults, error) {
	query := gocb.NewN1qlQuery(n1sql)
	return m.ExecuteN1qlQuery(query, params)
}

func (m *metaClient) BulkUpdate(values map[string]interface{}) error {
	var ops []gocb.BulkOp
	for key, value := range values {
		op := gocb.UpsertOp{
			Key:   key,
			Value: value,
		}
		ops = append(ops, &op)
	}
	return m.Bucket.Do(ops)
}

func (m *metaClient) SubListAppend(key, path string, value interface{}, cas uint64) (err error) {
	_, err = m.Bucket.MutateIn(key, 0, 0).ArrayAppend(path, value, false).Execute()
	return errors.WithStack(err)
}
