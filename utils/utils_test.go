package utils

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
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

func TestFile(t *testing.T) {
	f, _ := os.Open("path.go")
	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	fmt.Printf("n= %v, err = %v", n, err)
	n, err = f.Read(buf)
	fmt.Printf("n= %v, err = %v", n, err)
}

func TestJson(t *testing.T) {
	content := []byte(`[{"ttl":"5d"}]`)
	ttls := make([]*Bucket, 0)
	err := json.Unmarshal(content, &ttls)
	fmt.Printf("%v %v", ttls[0].TTL, err)
}

func TestIoEOF(t *testing.T) {
	fmt.Printf("%v", mime.TypeByExtension(".jsoN"))
}
