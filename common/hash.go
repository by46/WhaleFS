package common

import (
	"bytes"
	"crypto/sha1"
	"io"
	"fmt"
)

func Sha1(value string) (string, error) {
	buf := bytes.NewBuffer([]byte(value))
	client := sha1.New()
	if _, err := io.Copy(client, buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", client.Sum(nil)), nil
}
