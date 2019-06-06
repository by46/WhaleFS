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
	_ = mime.AddExtensionType(".bash", "application/x-sh")
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

func ExtensionByMimeType(mimeType string) string {
	extension, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(extension) == 0 {
		return ""
	}
	return extension[0]
}

func RandomName(extension string) string {
	name := uuid.New().String()
	if extension != "" {
		name = name + extension
	}
	return name
}

func MimeMatch(mimeType string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	mimeType = strings.ToLower(mimeType)
	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			mainType := strings.Split(pattern, "/")[0]
			if strings.HasPrefix(mimeType, strings.ToLower(mainType)) {
				return true
			}
		} else if strings.ToLower(pattern) == mimeType {
			return true
		}
	}
	return false
}
