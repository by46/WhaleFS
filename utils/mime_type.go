package utils

import (
	"mime"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo"
)

func init() {
	_ = mime.AddExtensionType(".go", "text/plain; charset=utf-8")
}

func IsImage(mimeType string) bool {
	if mimeType == "" {
		return false
	}
	mimeType = strings.ToLower(mimeType)
	return strings.HasPrefix(mimeType, "image/")
}

func IsPlain(mimeType string) bool {
	if mimeType == "" {
		return false
	}
	mimeType = strings.ToLower(mimeType)
	if strings.HasPrefix(mimeType, "text/") {
		return true
	}
	switch mimeType {
	case "application/javascript", "application/x-javascript":
		return true
	default:
		return false
	}
}

func MimeTypeByExtension(filename string) string {
	t := mime.TypeByExtension(path.Ext(filename))
	if t == "" {
		return echo.MIMEOctetStream
	}
	return t
}

func RandomName(mimeType string) string {
	name := uuid.New().String()
	extensions, _ := mime.ExtensionsByType(mimeType)
	if extensions != nil {
		name = name + extensions[0]
	}
	return name
}
