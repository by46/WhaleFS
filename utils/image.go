package utils

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func DecodeConfig(mime string, r io.Reader) (image.Config, error) {
	switch mime {
	case "image/jpeg":
		return jpeg.DecodeConfig(r)
	case "image/png":
		return png.DecodeConfig(r)
	case "image/bmp":
		return bmp.DecodeConfig(r)
	case "image/gif":
		return gif.DecodeConfig(r)
	case "image/tiff":
		return tiff.DecodeConfig(r)
	default:
		return image.Config{}, fmt.Errorf("unsupport mime type ")
	}
}
