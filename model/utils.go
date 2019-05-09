package model

import (
	"fmt"
	"strings"

	"github.com/by46/whalefs/utils"
)

func normalizePath(value string) string {
	segments := strings.Split(value, "/")
	segments = utils.Filter(segments, func(value string) bool { return value != "" })
	return "/" + strings.Join(segments, "/")
}

func parseBucketName(value string) (string, error) {
	segments := strings.Split(value, "/")
	if len(segments) < 2 {
		return "", fmt.Errorf("invalid bucket name")
	}
	return segments[1], nil
}
