package server

import (
	"bytes"
	"image"
	"image/color"
	"io"

	"github.com/disintegration/imaging"
	"github.com/labstack/echo"

	"github.com/by46/whalefs/server/middleware"
)

const (
	ModeFit       = "fit"
	ModeStretch   = "stretch"
	ModeThumbnail = "thumbnail"
)

var (
	ColorTransparency = color.RGBA{255, 255, 255, 0}
)

func (s *Server) resize(ctx echo.Context, r io.Reader) (io.Reader, error) {
	context := ctx.(*middleware.ExtendContext)

	if !context.FileParams.Entity.IsImage() {
		return r, nil
	}

	size := context.FileParams.Size
	if size == nil {
		return r, nil
	}

	img, err := imaging.Decode(r)
	if err != nil {
		return nil, err
	}
	switch size.Mode {
	case ModeFit:
		newImg := imaging.Fit(img, size.Width, size.Height, imaging.Lanczos)
		if !img.Bounds().Eq(newImg.Bounds()) {
			var c color.Color = image.White
			if context.FileParams.Entity.MimeType == "image/png" {
				c = ColorTransparency
			}
			background := imaging.New(size.Width, size.Height, c)
			img = imaging.PasteCenter(background, newImg)
		} else {
			img = newImg
		}
	case ModeStretch:
		img = imaging.Resize(img, size.Width, size.Height, imaging.Lanczos)
	default:
		img = imaging.Thumbnail(img, size.Width, size.Height, imaging.Lanczos)
	}
	return s.encode(ctx, img)
}

func (s *Server) encode(ctx echo.Context, img image.Image) (io.Reader, error) {
	context := ctx.(*middleware.ExtendContext)
	entity := context.FileParams.Entity

	buff := bytes.NewBuffer(nil)

	fmt := imaging.JPEG
	opts := []imaging.EncodeOption{}
	switch entity.MimeType {
	case "image/png":
		fmt = imaging.PNG
	case "image/gif":
		fmt = imaging.GIF
	case "image/bmp":
		fmt = imaging.BMP
	case "image/tiff":
		fmt = imaging.TIFF
	default:
		fmt = imaging.JPEG
		opts = append(opts, imaging.JPEGQuality(75))
	}
	if err := imaging.Encode(buff, img, fmt, opts...); err != nil {
		return nil, err
	}
	entity.Size = int64(buff.Len())
	return buff, nil
}
