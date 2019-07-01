package utils

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func Sha1(value string) (string, error) {
	return ContentSha1(bytes.NewReader([]byte(value)))
}

func Sha1WithLength(value string, l int) string {
	etag, err := Sha1(value)
	if err != nil {
		return ""
	}
	if len(etag) <= l {
		return etag
	}
	return etag[:l]
}

func ContentSha1(r io.Reader) (string, error) {
	client := sha1.New()
	if _, err := io.Copy(client, r); err != nil {
		return "", errors.WithStack(err)
	}
	return fmt.Sprintf("%x", client.Sum(nil)), nil
}
