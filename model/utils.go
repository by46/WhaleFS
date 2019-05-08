package model

import (
	"fmt"
	"strings"

	"github.com/by46/whalefs/utils"
)

func parseBucketName(value string) (string, error) {
	segments := strings.Split(value, "/")
	segments = utils.Filter(segments, func(value string) bool { return value != "" })
	if len(segments) < 1 {
		return "", fmt.Errorf("invalid bucket name")
	}
	return segments[0], nil
}
