package common

import (
	"gopkg.in/couchbase/gocb.v1"
)

type Task interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	QueryPendingPkgTask(n1sql string) (gocb.QueryResults, error)
	BulkUpdate(values map[string]interface{}) error
}
