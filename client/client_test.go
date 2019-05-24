package client

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpClient_MultipartUpload(t *testing.T) {
	f, _ := os.Open("./client.go")
	defer func() {
		_ = f.Close()
	}()
	c := NewClient(&ClientOptions{Base: "http://localhost:8089"})

	entity, err := c.Upload(context.TODO(), &Options{
		Bucket:     "benjamin",
		FileName:   "client3.go",
		Override:   true,
		Content:    f,
		MultiChunk: true,
	})
	assert.Nil(t, err)
	fmt.Printf("%v", entity)
}
