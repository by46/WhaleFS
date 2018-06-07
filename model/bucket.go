package model

import (
	"fmt"
	"strings"
)

type ExtendItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Buckets struct {
	Buckets []string `json:"buckets"`
}

type Bucket struct {
	Name         string         `json:"name"`
	Alias        []string       `json:"alias"`
	Expires      int            `json:"expires"` // unit: day
	Extends      [] *ExtendItem `json:"extends"`
	Memo         string         `json:"memo"`
	LastEditDate int64          `json:"last_edit_date"`
	LastEditUser string         `json:"last_edit_user"`
}

func (b *Bucket) Key() string {
	return fmt.Sprintf("system.bucket.%s", strings.ToLower(b.Name))
}
