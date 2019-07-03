package utils

import (
	"bufio"
	"mime"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

type mimePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var MimeTypes []*mimePair

func init() {
	loadMime()
}
func loadMime() {
	file := "config/mime.txt"
	if FileExists(file) {
		f, err := os.Open(file)
		if err != nil {
			log.Warnf("load mime type file failed: %v", err)
			return
		}
		defer func() {
			_ = f.Close()
		}()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			segments := strings.SplitN(line, " ", 2)
			if len(segments) != 2 {
				log.Warnf("invalid format %v", line)
				continue
			}
			extension, mimeType := strings.TrimSpace(strings.ToLower(segments[0])), strings.TrimSpace(strings.ToLower(segments[1]))
			MimeTypes = append(MimeTypes, &mimePair{Name: extension, Value: mimeType})
			_ = mime.AddExtensionType(extension, mimeType)
		}
	}
}

func IsImageByFileName(fileName string) bool {
	mimeType := MimeTypeByExtension(fileName)
	return IsImage(mimeType)
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
	mimeType = NormalMimeType(mimeType)
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

func NormalMimeType(mimeType string) string {
	media, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return mimeType
	}
	return media
}

func IsVideo(mimeType string) bool {
	if mimeType == "" {
		return false
	}
	mimeType = strings.ToLower(mimeType)
	return strings.HasPrefix(mimeType, "video/")
}
