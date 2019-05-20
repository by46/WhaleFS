package server

import (
	"bytes"
	"image"
	"image/color"
	"io"

	"github.com/disintegration/imaging"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

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

func (s *Server) thumbnail(ctx echo.Context, r io.Reader) (io.Reader, error) {
	context := ctx.(*middleware.ExtendContext)

	if !context.FileContext.Meta.IsImage() {
		return r, nil
	}

	// 检查是否需要动态切图
	size := context.FileContext.Size
	if size == nil {
		return r, nil
	}

	img, err := s.prepare(ctx, r)
	if err != nil {
		return nil, err
	}

	switch size.Mode {
	case ModeFit:
		newImg := imaging.Fit(img, size.Width, size.Height, imaging.Lanczos)
		if !img.Bounds().Eq(newImg.Bounds()) {
			var c color.Color = image.White
			if context.FileContext.Meta.MimeType == "image/png" {
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

	img = s.overlay(ctx, img)
	return s.encode(ctx, img)
}

func (s *Server) overlay(ctx echo.Context, img image.Image) image.Image {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	overlay := bucket.GetOverlay("")
	if overlay == nil {
		return img
	}

	overlayImage, err := s.downloadOverlay(overlay.Image)
	if err != nil {
		// 如果获取水印出现错误, 就放弃添加水印, 返回原图
		return img
	}

	// 针对水印进行缩放
	ratio := float64(img.Bounds().Dx()) / float64(bucket.Basis.PrepareThumbnailMinWidth)
	width, height := overlayImage.Bounds().Dx(), overlayImage.Bounds().Dy()
	overlayImage = imaging.Resize(overlayImage, int(float64(width)*ratio), int(float64(height)*ratio), imaging.Lanczos)

	pt := overlay.RealPosition(img, overlayImage)
	return imaging.Overlay(img, overlayImage, pt, overlay.Opacity)
}

// 对图片进行预处理, 减少处理时间
func (s *Server) prepare(ctx echo.Context, r io.Reader) (img image.Image, err error) {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	meta := context.FileContext.Meta

	if meta.ThumbnailFID != "" {
		if reader, _, err := s.Storage.Download(meta.ThumbnailFID); err != nil {
			s.Logger.Errorf("下载预处理图片失败 %+v", err)
		} else if img, err := imaging.Decode(reader); err != nil {
			s.Logger.Errorf("解析预处理图片失败 %+v", err)
		} else {
			return img, nil
		}
	}

	if img, err = imaging.Decode(r); err != nil {
		return nil, errors.Wrap(err, "图片解码失败")
	}

	if meta.Width > bucket.Basis.PrepareThumbnailMinWidth {
		img = imaging.Resize(img, bucket.Basis.PrepareThumbnailMinWidth, 0, imaging.Lanczos)
		prepareThumbnail, err := s.encode(ctx, img)
		if err != nil {
			s.Logger.Errorf("生成预处理图片失败 %+v", err)
		} else {
			if prepareThumbnailMeta, err := s.Storage.Upload(meta.MimeType, prepareThumbnail); err != nil {
				s.Logger.Errorf("上传预处理图片失败 %+v", err)
			} else {
				meta.ThumbnailFID = prepareThumbnailMeta.FID
				if err := s.Meta.Set(meta.RawKey, meta); err != nil {
					s.Logger.Errorf("更新文件元数据失败 %+v", err)
				}
			}
		}
	}
	return
}

func (s *Server) downloadOverlay(fid string) (img image.Image, err error) {
	content, _, err := s.Storage.Download(fid)
	if err != nil {
		return
	}
	if img, err = imaging.Decode(content); err != nil {
		return nil, errors.Wrap(err, "图片编码失败")
	}
	return
}

func (s *Server) encode(ctx echo.Context, img image.Image) (io.Reader, error) {
	context := ctx.(*middleware.ExtendContext)
	entity := context.FileContext.Meta

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
		return nil, errors.WithStack(err)
	}
	entity.Size = int64(buff.Len())
	return buff, nil
}
