package migration

import (
	"log"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

type Cache struct {
	db *leveldb.DB
}

func NewCache(path string) *Cache {
	if path == "" {
		path = "cache.db"
	}
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalf("打开文件失败 %v", err)
	}
	return &Cache{
		db: db,
	}
}

func (c *Cache) Exists(key string) (exists bool) {
	if key == "" {
		return true
	}
	key = strings.ToLower(key)
	exists, err := c.db.Has([]byte(key), nil)
	return err == nil && exists
}

func (c *Cache) Put(key string) {
	if key == "" {
		return
	}
	key = strings.ToLower(key)
	err := c.db.Put([]byte(key), []byte("1"), nil)
	if err != nil {
		log.Printf("put key error %v", err)
	}
}
