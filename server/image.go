package server

import (
	"bytes"
	"io"

	"github.com/disintegration/imaging"
)

func (s *Server) resize(r io.ReadCloser) (io.Reader, error) {
	img, err := imaging.Decode(r)
	if err != nil {
		return nil, err
	}
	img = imaging.Resize(img, 200, 0, imaging.Lanczos)
	buff := bytes.NewBuffer(nil)
	if err := imaging.Encode(buff, img, imaging.JPEG); err != nil {
		return nil, err
	}
	return buff, nil
}
