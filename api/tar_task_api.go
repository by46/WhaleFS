package api

import (
	"gopkg.in/couchbase/gocb.v1"

	"github.com/by46/whalefs/common"
)

type taskClient struct {
	common.Meta
}

func NewTaskClient(connectionString string) common.Task {
	meta := NewMetaClient(connectionString)
	return &taskClient{meta}
}

func (m *taskClient) QueryPendingTarTask(n1sql string) (gocb.QueryResults, error) {
	return m.Query(n1sql, nil)
}