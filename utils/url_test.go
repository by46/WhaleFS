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
