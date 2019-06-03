package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathNormalize(t *testing.T) {
	assert.Equal(t, "/benjamin/hello.jpg", PathNormalize("/benjamin/hello.jpg"))
	assert.Equal(t, "/benjamin/hello.jpg", PathNormalize("/benjamin///hello.jpg"))
	assert.Equal(t, "/benjamin/level/hello.jpg", PathNormalize("/benjamin//level/hello.jpg"))
	assert.Equal(t, "/benjamin/hellO.jpg", PathNormalize("/benjamin/hellO.jpg"))
	assert.Equal(t, "/benjamin/hello.jpg", PathNormalize("benjamin/hello.jpg"))
}

func TestPathSegment(t *testing.T) {
	assert.Equal(t, "benjamin", PathSegment("/benjamin/hello.jpg", 0))
	assert.Equal(t, "p200", PathSegment("/benjamin/p200/hello.jpg", 1))
	assert.Equal(t, "hello.jpg", PathSegment("/benjamin/hello.jpg", 1))
	assert.Equal(t, "hello.jpg", PathSegment("benjamin/hello.jpg", 1))
	assert.Equal(t, "", PathSegment("benjamin/hello.jpg", 2))
}

func TestPathRemoveSegment(t *testing.T) {
	actualRemoved, actualResult := PathRemoveSegment("/benjamin/level1/level/hello.jpg", 1)
	assert.Equal(t, "/benjamin/level/hello.jpg", actualResult)
	assert.Equal(t, "level1", actualRemoved)

	actualRemoved, actualResult = PathRemoveSegment("benjamin/level1/level/hello.jpg", 1)
	assert.Equal(t, "/benjamin/level/hello.jpg", actualResult)
	assert.Equal(t, "level1", actualRemoved)

	actualRemoved, actualResult = PathRemoveSegment("/benjamin/level1/level/hello.jpg", 4)
	assert.Equal(t, "/benjamin/level1/level/hello.jpg", actualResult)
	assert.Equal(t, "", actualRemoved)
}

func TestNameWithoutExtension(t *testing.T) {
	assert.Equal(t, "name", NameWithoutExtension("name"))
	assert.Equal(t, "name", NameWithoutExtension("name.txt"))
	assert.Equal(t, "name.tar", NameWithoutExtension("name.tar.gz"))
	assert.Equal(t, "中国", NameWithoutExtension("中国.txt"))
}
