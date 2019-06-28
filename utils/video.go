package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
)

func GetFrame(filename string, index int) *bytes.Buffer {
	buf := new(bytes.Buffer)

	//cmd := exec.Command("/Users/mark.c.jiang/ffmpeg", "-y", "-i", filename, "-ss", "00:00:01", "-vframes", strconv.Itoa(index), "-s", fmt.Sprintf("%dx%d", width, height), "-f", "image2", "-")
	cmd := exec.Command("ffmpeg", "-y", "-i", filename, "-ss", "00:00:01", "-vframes", strconv.Itoa(index), "-f", "image2", "-")

	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("could not generate frame: %v", err))
	}

	return buf
}
