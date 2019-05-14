package api

import (
	"gopkg.in/couchbase/gocb.v1"

	"github.com/by46/whalefs/common"
)

type taskClient struct {
	*gocb.Bucket
}

func NewTaskClient(connectionString string) common.Task {
	bucket := common.OpenBucket(connectionString)
	return &taskClient{bucket}
}

func (m *taskClient) Get(key string, value interface{}) error {
	_, err := m.Bucket.Get(key, &value)
	if err != nil && err == gocb.ErrKeyNotFound {
		return common.New(common.CodeFileNotExists)
	}
	return err
}

func (m *taskClient) Set(key string, value interface{}) error {
	_, err := m.Bucket.Upsert(key, value, 0)
	return err
}

func (m *taskClient) QueryPendingTarTask(n1sql string) (gocb.QueryResults, error) {
	query := gocb.NewN1qlQuery(n1sql)
	return m.ExecuteN1qlQuery(query, nil)
}
