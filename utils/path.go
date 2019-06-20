// 路径相关工具类
package utils

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	Separator = "/"
)

func PathReplace(path string, i int, new string) string {
	path2 := strings.TrimLeft(path, Separator)
	segments := strings.Split(path2, Separator)
	if len(segments) <= i {
		return path
	}
	segments[i] = new
	return Separator + strings.Join(segments, Separator)
}
func PathSegment(path string, i int) string {
	path = strings.TrimLeft(path, Separator)
	segments := strings.Split(path, Separator)
	if len(segments) <= i {
		return ""
	}
	return segments[i]
}

func PathLastSegment(path string) string {
	path = strings.TrimLeft(path, Separator)
	segments := strings.Split(path, Separator)
	if len(segments) <= 1 {
		return path
	}
	return segments[len(segments)-1]
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
	path = strings.ToLower(path)
	segments := strings.Split(path, Separator)
	segments = Filter(segments, func(s string) bool { return s != "" })
	if len(segments) > 2 {
		if segments[1] == "original" {
			segments = append(segments[:1], segments[2:]...)
		}
	}
	return Separator + strings.Join(segments, Separator)
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func NameWithoutExtension(name string) string {
	name = filepath.Base(name)
	ext := filepath.Ext(name)
	if ext != "" {
		name = name[:len(name)-len(ext)]
	}
	return name
}
