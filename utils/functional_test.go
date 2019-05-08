package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter(t *testing.T) {
	segments := []string{"", "benjamin", "", ""}
	actual := Filter(segments, func(value string) bool {
		return value != ""
	})
	assert.Len(t, actual, 1)
}
