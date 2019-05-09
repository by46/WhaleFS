package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBucketName(t *testing.T) {
	name, err := parseBucketName("/benjamin/level1/level2/hello.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "benjamin", name)
}
