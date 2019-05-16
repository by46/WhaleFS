package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha1(t *testing.T) {
	content := "whale-fs"
	digest, err := Sha1(content)
	assert.Nil(t, err)
	assert.Equal(t, "848a148df383fd663dfae6a75d129ff9682ee86a", digest)
}
