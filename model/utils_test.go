package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBucketName(t *testing.T) {
	name, err := parseBucketName("/benjamin/level1/level2/hello.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "benjamin", name)

	name, err = parseBucketName("//benjamin/level1/level2/hello.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "benjamin", name)

	name, err = parseBucketName("///benjamin//level1/level2/hello.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "benjamin", name)
}
