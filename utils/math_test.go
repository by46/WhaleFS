package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToInt32(t *testing.T) {
	actual := ToInt32("8")
	assert.Equal(t, int32(8), actual)
}
