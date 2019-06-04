package utils

import (
	"bytes"
)

type PDFFile struct {
	buf *bytes.Reader
}

func NewPDFFile(buf []byte) *PDFFile {
	return &PDFFile{
		buf: bytes.NewReader(buf),
	}
}

func (f *PDFFile) Read(s []byte) (int, error) {
	return f.buf.Read(s)
}

func (f *PDFFile) Close() error {
	return nil
}

func (f *PDFFile) Seek(o int64, w int) (pos int64, err error) {
	return f.buf.Seek(o, w)
}
