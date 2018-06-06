package api

import (
	"io"
	"strings"
)

type IStorage interface {
	Download(url string) (io.Reader, error)
	Upload(mimeType string, body io.Reader) (url string, err error)
}

type storageClient struct {
	master []string
}

func NewStorageClient(master string) IStorage {
	masters := strings.Split(master, ",")
	return &storageClient{
		master: masters,
	}
}

func (c *storageClient) Download(url string) (io.Reader, error) {
	return nil, nil
}

func (c *storageClient) Upload(mimeType string, body io.Reader) (url string, err error) {
	return
}
