package common

import (
	"gopkg.in/couchbase/gocb.v1"
)

type Dao interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	Exists(key string) (bool, error)
	SetTTL(key string, value interface{}, ttl uint32) error
	Query(n1sql string, params interface{}) (gocb.QueryResults, error)
	BulkUpdate(values map[string]interface{}) error
}
