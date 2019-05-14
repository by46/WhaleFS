package utils

import (
	"strings"
)

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
