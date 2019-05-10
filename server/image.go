package server

import (
	"bytes"
	"io"

	"github.com/disintegration/imaging"
	"github.com/labstack/echo"

	"github.com/by46/whalefs/server/middleware"
)

func (s *Server) resize(ctx echo.Context, r io.Reader) (io.Reader, error) {
	context := ctx.(*middleware.ExtendContext)
	entity := context.FileParams.Entity
	size := context.FileParams.Size

	if size == nil {
		return r, nil
	}

	img, err := imaging.Decode(r)
	if err != nil {
		return nil, err
	}
	img = imaging.Resize(img, size.Width, size.Height, imaging.Lanczos)
	buff := bytes.NewBuffer(nil)
	if err := imaging.Encode(buff, img, imaging.JPEG); err != nil {
		return nil, err
	}
	entity.Size = int64(buff.Len())
	return buff, nil
}
