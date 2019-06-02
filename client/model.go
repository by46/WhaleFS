package client

import (
	"fmt"
	"io"
)

type Options struct {
	Bucket     string
	FileName   string
	Content    io.Reader
	Override   bool
	MultiChunk bool
}

func (o *Options) key() string {
	return fmt.Sprintf("%s/%s", o.Bucket, o.FileName)
}
func (o *Options) getOverride() string {
	if o.Override {
		return "1"
	}
	return "0"
}

type FileEntity struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

type ClientOptions struct {
	Base string
}
