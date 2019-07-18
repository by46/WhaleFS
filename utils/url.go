package utils

import (
	"net/url"
	"path"
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

func Url2FileName(u string) string {
	opt, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return path.Base(opt.Path)
}

func QueryParam(values url.Values, name string) string {
	for key := range values {
		if strings.ToLower(key) == strings.ToLower(name) {
			return values.Get(key)
		}
	}
	return ""
}
