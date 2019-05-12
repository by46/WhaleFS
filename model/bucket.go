package model

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const (
	ExtendKeyMaxAge = "max-age"
)

type ExtendItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Buckets struct {
	Buckets []string `json:"buckets"`
}

type ImageSize struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Mode   string `json:"mode"`
}

type BucketLimit struct {
	MinSize   int64    `json:"min_size"`
	MaxSize   int64    `json:"max_size"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	MimeTypes []string `json:"mime_types"`
}

type Bucket struct {
	Name         string         `json:"name"`
	Expires      int            `json:"expires"` // unit: day
	Extends      [] *ExtendItem `json:"extends"`
	Memo         string         `json:"memo"`
	LastEditDate int64          `json:"last_edit_date"`
	LastEditUser string         `json:"last_edit_user"`
	Sizes        []*ImageSize   `json:"Sizes"`
	Limit        *BucketLimit   `json:"limit"`
	sizesMapping map[string]*ImageSize
	sync.Mutex
}

func (b *Bucket) Key() string {
	return fmt.Sprintf("system.bucket.%s", strings.ToLower(b.Name))
}

func (b *Bucket) MaxAge() int {
	return b.getExtendInt(ExtendKeyMaxAge)
}

func (b *Bucket) getExtend(key string) string {
	if b.Extends == nil {
		return ""
	}
	for _, item := range b.Extends {
		if item.Key == key {
			return item.Value
		}
	}
	return ""
}

func (b *Bucket) getExtendInt(key string) int {
	text := b.getExtend(key)
	if text == "" {
		return 0
	}
	value, _ := strconv.ParseInt(text, 10, 32)
	return int(value)

}

func (b *Bucket) GetSize(name string) *ImageSize {
	if b.Sizes == nil {
		return nil
	}

	if b.sizesMapping == nil {
		b.Lock()
		if b.sizesMapping == nil {
			b.sizesMapping = make(map[string]*ImageSize, len(b.Sizes))
			for _, size := range (b.Sizes) {
				b.sizesMapping[size.Name] = size
			}
		}
		b.Unlock()
	}
	return b.sizesMapping[name]
}
