// 路径相关工具类
package utils

import (
	"strings"
)

const (
	Separator = "/"
)

func PathSegment(path string, i int) string {
	path = strings.TrimLeft(path, Separator)
	segments := strings.Split(path, Separator)
	if len(segments) <= i {
		return ""
	}
	return segments[i]
}

func PathRemoveSegment(path string, i int) (removed string, result string) {
	segments := strings.Split(strings.TrimLeft(path, Separator), Separator)
	if len(segments) <= i {
		return "", path
	}
	removed = segments[i]
	segments = append(segments[:i], segments[i+1:]...)
	return removed, Separator + strings.Join(segments, Separator)

}

func PathNormalize(path string) string {
	segments := strings.Split(path, Separator)
	segments = Filter(segments, func(s string) bool { return s != "" })
	return Separator + strings.Join(segments, Separator)
}
