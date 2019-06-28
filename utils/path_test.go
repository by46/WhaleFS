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

func TestPathLastSegment(t *testing.T) {
	assert.Equal(t, "test", PathLastSegment("test"))
	assert.Equal(t, "test", PathLastSegment("/test"))
	assert.Equal(t, "test", PathLastSegment("/benjamin/test"))
}

func TestPathReplace(t *testing.T) {
	assert.Equal(t, "/pdt/p120/file.txt", PathReplace("/pdt/Original/file.txt", 1, "p120"))
	assert.Equal(t, "/pdt/p120/file.txt", PathReplace("pdt/Original/file.txt", 1, "p120"))
}

func TestSubFolderByFileName(t *testing.T) {
	assert.Equal(t, "A/01/1D/514e3e40-8767-4a6d-97da-b6ba23706188.jpg", SubFolderByFileName("514e3e40-8767-4a6d-97da-b6ba23706188.jpg"))
	assert.Equal(t, "A/01/7D/29d3ddb2-5150-467a-93fb-220d74bf4e13.jpg", SubFolderByFileName("29d3ddb2-5150-467a-93fb-220d74bf4e13.jpg"))
	assert.Equal(t, "W/15/9B/68da5d5e-4a7e-43fc-b545-2223a03be705.jpg", SubFolderByFileName("68da5d5e-4a7e-43fc-b545-2223a03be705.jpg"))
	assert.Equal(t, "W/15/0B/15d773f1-b24a-4303-9f7a-702c05a3edb7.jpg", SubFolderByFileName("15d773f1-b24a-4303-9f7a-702c05a3edb7.jpg"))
}
