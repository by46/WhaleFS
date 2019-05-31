package migration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestLevelDB(t *testing.T) {
	db, err := leveldb.OpenFile("../cache.db", nil)
	assert.Nil(t, err)
	defer func() {
		_ = db.Close()
	}()
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("key: %v, value: %v\n", string(key), string(value))
	}
	iter.Release()
}
