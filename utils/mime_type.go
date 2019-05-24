package utils

import (
	"mime"
	"path"
	"strings"

	"github.com/labstack/echo"
)

func init() {
	_ = mime.AddExtensionType(".go", "text/plain; charset=utf-8")
}

func IsImage(mime string) bool {
	if mime == "" {
		return false
	}
	mime = strings.ToLower(mime)
	return strings.HasPrefix(mime, "image/")
}

func IsPlain(mime string) bool {
	if mime == "" {
		return false
	}
	mime = strings.ToLower(mime)
	if strings.HasPrefix(mime, "text/") {
		return true
	}
	switch mime {
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
