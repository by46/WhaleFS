package utils

import (
	"fmt"
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

func errorBinding() error {
	var err Names
	return err
}

type Names []string

func (Names) Error() string {
	return "debugging"
}

func TestErrorBinding(t *testing.T) {
	err := errorBinding()
	if err != nil {
		fmt.Printf("debugging %s\n", err)
	}

	var names []string
	if names != nil {
		fmt.Printf("debugging %s\n", names)
	}

	var err2 []string
	if err2 != nil {
		fmt.Printf("debugging3 %s\n", err2)
	}

	var err3 Names
	fmt.Printf("debugging 4 %v %v\n", err3, err3 == nil)

}
