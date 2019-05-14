package common

import (
	"gopkg.in/couchbase/gocb.v1"
)

type Task interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	QueryPendingTarTask(n1sql string) (gocb.QueryResults, error)
}
