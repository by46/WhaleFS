package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName2Disposition(t *testing.T) {
	assert.True(t, IsBrowserIE("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:67.0) Gecko/20100101 Firefox/67.0"))
}
