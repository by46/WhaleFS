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
	content := []byte(`{"ttl":"5d"}`)
	bucket := new(Bucket)
	_ = json.Unmarshal(content, bucket)
	fmt.Printf("%v %v", bucket.TTL, bucket.TTL.Expiry())
}
