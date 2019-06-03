package utils

import (
	"net/url"
	"strings"
)

func UrlDecode(u string) string {
	n, err := url.PathUnescape(u)
	if err == nil {
		return n
	}
	return u
}

func IsRemote(u string) bool {
	u = strings.ToLower(u)
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "ftp://")
}
