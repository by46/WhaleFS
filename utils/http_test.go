package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderCopy(t *testing.T) {
	dst, src := make(http.Header), make(http.Header)
	dst.Set("content-type", "application/json")
	src.Set("content-length", "120")
	headerCopy(dst, src)
	expect := http.Header(map[string][]string{
		"Content-Type":   []string{"application/json"},
		"Content-Length": []string{"120"},
	})
	assert.EqualValues(t, expect, headerCopy(dst, src))

	dst, src = make(http.Header), make(http.Header)
	dst.Set("content-type", "application/json")
	dst.Set("content-length", "120")
	src.Set("content-length", "120")
	headerCopy(dst, src)
	expect = http.Header(map[string][]string{
		"Content-Type":   []string{"application/json"},
		"Content-Length": []string{"120"},
	})
	assert.EqualValues(t, expect, headerCopy(dst, src))
}
