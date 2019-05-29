package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Chunk struct {
	FID  string `json:"fid"`
	ETag string `json:"etag"`
}

type Sizes map[string]*Chunk

type A struct {
	Name  string `json:"name"`
	Sizes Sizes  `json:"sizes"`
}

func TestJson(t *testing.T) {
	content := `{"name":"benjamin", "sizes":{"p200":{"fid":"1,21231","etag":"etag1"}, "p400":{"fid":"2,21231","etag":"etag2"}}}`
	a := new(A)
	err := json.Unmarshal([]byte(content), a)
	assert.Nil(t, err)
	assert.Equal(t, &Chunk{FID: "1,21231", ETag: "etag1"}, a.Sizes["p200"])
	assert.Equal(t, &A{Name: "benjamin", Sizes: map[string]*Chunk{"p200": {FID: "1,21231", ETag: "etag1"}, "p400": {FID: "2,21231", ETag: "etag2"}}}, a)

	content = `{"name":"benjamin", "sizes":{}}`
	a = new(A)
	err = json.Unmarshal([]byte(content), a)
	assert.Nil(t, err)
	assert.Equal(t, &A{Name: "benjamin", Sizes: make(Sizes)}, a)
}
