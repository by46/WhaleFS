package utils

import (
	"io"
	"os"
	"testing"
)

func TestGetFrame(t *testing.T) {
	buffer := GetFrame("/Users/mark.c.jiang/Downloads/hiit.mp4", 1, 0, 0)
	file, err := os.Create("/Users/mark.c.jiang/Downloads/hiit.jpg")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(file, buffer)
	if err != nil {
		panic(err)
	}
}
