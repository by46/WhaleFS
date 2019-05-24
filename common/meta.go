package common

import (
	"gopkg.in/couchbase/gocb.v1"
)

type Dao interface {
	Get(key string, value interface{}) error
	GetWithCas(key string, value interface{}) (cas uint64, err error)
	Set(key string, value interface{}) error
	Exists(key string) (bool, error)
	SetTTL(key string, value interface{}, ttl uint32) error
	Delete(key string, cas uint64) (err error)
	Query(n1sql string, params interface{}) (gocb.QueryResults, error)
	BulkUpdate(values map[string]interface{}) error
	SubListAppend(key, path string, value interface{}, cas uint64) error
}
