package utils

import (
	"mime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMimeTypeByExtension(t *testing.T) {
	assert.Equal(t, "application/json", MimeTypeByExtension("file.json"))
	assert.Equal(t, "text/plain; charset=utf-8", MimeTypeByExtension("file.txt"))
	assert.Equal(t, "text/html; charset=utf-8", MimeTypeByExtension("file.html"))
	assert.Equal(t, "application/javascript", MimeTypeByExtension("file.js"))
	assert.Equal(t, "text/css; charset=utf-8", MimeTypeByExtension("file.css"))
	assert.Equal(t, "application/pdf", MimeTypeByExtension("file.pdf"))
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", MimeTypeByExtension("file.xlsx"))
	assert.Equal(t, "application/vnd.ms-excel", MimeTypeByExtension("file.xls"))
	assert.Equal(t, "application/msword", MimeTypeByExtension("file.doc"))
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", MimeTypeByExtension("file.docx"))
	assert.Equal(t, "image/jpeg", MimeTypeByExtension("file.jpg"))
	assert.Equal(t, "image/jpeg", MimeTypeByExtension("file.jpeg"))
	assert.Equal(t, "image/png", MimeTypeByExtension("file.png"))
	assert.Equal(t, "image/bmp", MimeTypeByExtension("file.bmp"))
	assert.Equal(t, "text/plain; charset=utf-8", MimeTypeByExtension("file.go"))
	assert.Equal(t, "application/x-sh", MimeTypeByExtension("file.sh"))
	assert.Equal(t, "application/x-sh", MimeTypeByExtension("file.bash"))
	assert.Equal(t, "application/octet-stream", MimeTypeByExtension("file.xxx"))
}

func TestRandomName(t *testing.T) {
	assert.True(t, strings.HasSuffix(RandomName(".jpeg"), ".jpg"))
	assert.True(t, strings.HasSuffix(RandomName(".png"), ".png"))
	assert.True(t, strings.HasSuffix(RandomName(".json"), ".json"))
}

func TestExtensionByMimeType(t *testing.T) {
	assert.Equal(t, ".jpg", ExtensionByMimeType("image/jpeg"))
	assert.Equal(t, "", ExtensionByMimeType("image/unknown"))
}

func TestMimeMatch(t *testing.T) {
	assert.True(t, MimeMatch("text/plain", []string{"text/*"}))
	assert.True(t, MimeMatch("text/html", []string{"text/*"}))
	assert.True(t, MimeMatch("text/html", []string{"text/html", "image/jpeg"}))
	assert.True(t, MimeMatch("text/plain", []string{}))
	assert.True(t, MimeMatch("text/plain", nil))

	assert.False(t, MimeMatch("text/html", []string{"image/*"}))
	assert.False(t, MimeMatch("text/html", []string{"image/*", "video/*"}))
}

func TestParseMediaType(t *testing.T) {
	media, p, err := mime.ParseMediaType("text/plain;charset=utf8")
	assert.Nil(t, err)
	assert.Equal(t, "text/plain", media)
	assert.Equal(t, map[string]string{"charset": "utf8"}, p)
}

func TestMime2Extension(t *testing.T) {
	mimeTypes := []string{"application/json", "application/pdf", "image/jpeg"}
	extensions := Mime2Extension(mimeTypes)
	assert.ElementsMatch(t, []string{".json", ".pdf", ".jpg", ".jpeg", ".jpe"}, extensions)
}
