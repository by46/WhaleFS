package utils

import (
	"net/url"
)

func UrlDecode(u string) string {
	n, err := url.PathUnescape(u)
	if err == nil {
		return n
	}
	return u
}
