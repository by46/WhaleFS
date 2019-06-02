package api

import (
	"io"

	"github.com/by46/whalefs/common"
)

type ChunksReader struct {
	storage common.Storage
	fids    []string
	reader  io.ReadCloser
}

func NewChunksReader(storage common.Storage, fids []string) io.ReadCloser {
	return &ChunksReader{
		storage: storage,
		fids:    fids,
	}
}

func (r *ChunksReader) Read(p []byte) (n int, err error) {
	for len(r.fids) > 0 {
		if r.reader == nil {
			r.reader, _, err = r.storage.Download(r.fids[0])
			if err != nil {
				return 0, err
			}
		}
		n, err = r.reader.Read(p)
		if err == io.EOF {
			_ = r.reader.Close()
			r.reader = nil
			r.fids = r.fids[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(r.fids) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return
		}
	}
	return 0, io.EOF
}

func (r *ChunksReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}
