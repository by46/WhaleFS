package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	segments := []string{"", "benjamin", "", ""}
	actual := Filter(segments, func(value string) bool {
		return value != ""
	})
	assert.Len(t, actual, 1)
}

func TestExists(t *testing.T) {
	assert.True(t, Exists([]string{"image/png", "image/jpeg"}, "image/jpeg"))
	assert.False(t, Exists([]string{"image/png", "image/jpeg"}, "image/jpeg1"))
}
