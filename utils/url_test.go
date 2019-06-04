package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlDecode(t *testing.T) {
	assert.Equal(t, "http://localhost:8080", UrlDecode("http://localhost:8080"))
	assert.Equal(t, "http://localhost:8080/benjamin///hello", UrlDecode("http://localhost:8080/benjamin/%2F/hello"))
	assert.Equal(t, "http://localhost:8080/benjamin?name=/", UrlDecode("http://localhost:8080/benjamin?name=%2F"))
	assert.Equal(t, "http://localhost:8080/benjamin?name=/", UrlDecode("http://localhost:8080/benjamin?name=/"))
	assert.Equal(t, "http://localhost:8080/%2", UrlDecode("http://localhost:8080/%2"))
}

func TestUrl2FileName(t *testing.T) {
	assert.Equal(t, "filename.txt", Url2FileName("http://localhost:80/home/hello/filename.txt"))
	assert.Equal(t, "filename.txt", Url2FileName("http://localhost:80/home/hello/filename.txt?version1=1.2"))
	assert.Equal(t, "filename.txt", Url2FileName("http://localhost:80/home/hello/filename.txt?version1=1.2&ts=1212"))
	assert.Equal(t, "filename.txt", Url2FileName("http://localhost:80/home/hello/filename.txt?version1=1.2&ts=1212#tag1"))

	assert.Equal(t, "filename.txt", Url2FileName("/home/hello/filename.txt"))
	assert.Equal(t, "", Url2FileName("/%2home/hello/filename.txt"))
}
