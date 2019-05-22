package utils

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TTL string

func (t TTL) Expiry() uint32 {
	return 1
}
func (t TTL) String() string {
	return string(t)
}

type Bucket struct {
	TTL TTL `json:"ttl"`
}

func TestUnmarshal(t *testing.T) {
	content := []byte(`{"ttl":""}`)
	bucket := new(Bucket)
	_ = json.Unmarshal(content, bucket)
	fmt.Printf("%v %v", bucket.TTL, bucket.TTL.Expiry())
	if bucket.TTL == "" {
		fmt.Printf("hello\n")
	} else {
		fmt.Printf("hello1")
	}
}
