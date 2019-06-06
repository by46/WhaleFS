package server

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"

	"github.com/disintegration/imaging"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
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
		// TODO(benjamin): 如果有错误产生, 就返回原图, 尽量避免错误产生
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
	buff, err := s.encode(ctx, img)
	if err == nil {
		s.uploadThumbnail(ctx, bytes.NewReader(buff.Bytes()))
	}
	return buff, err
}

func (s *Server) overlay(ctx echo.Context, img image.Image) image.Image {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket
	meta := context.FileContext.Meta

	var overlay *model.ImageOverlay
	if meta.WaterMark != "" {
		overlay = &model.ImageOverlay{
			Image:    meta.WaterMark,
			Opacity:  0.8,
			Position: model.PositionBottomRight,
		}
	} else {
		overlay = bucket.GetOverlay("")
	}
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
			// TODO(benjamin): 把所有缩略图生成到临时collection中
			option := &common.UploadOption{
				Collection:  s.Config.Basis.CollectionTmp,
				Replication: ReplicationNo,
			}
			if prepareThumbnailMeta, err := s.Storage.Upload(option, meta.MimeType, prepareThumbnail); err != nil {
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

// 下载overlay的图片
// TODO(benjamin): 需要优化, 可以在加载bucket的时候, 就加载overlay文件内容
func (s *Server) downloadOverlay(name string) (img image.Image, err error) {
	fullName := fmt.Sprintf("/%s/overlay/%s", s.Config.Basis.BucketHome, name)
	img = s.getOverlayByFullName(fullName)
	if img == nil {
		return nil, errors.New("读取文件失败")
	}
	return
}

func (s *Server) downloadThumbnail(ctx echo.Context) io.Reader {
	context := ctx.(*middleware.ExtendContext)
	size := context.FileContext.Size
	meta := context.FileContext.Meta

	if size == nil {
		return nil
	}

	if thumbnailMeta, exists := meta.Thumbnails[size.Name]; exists {
		r, _, err := s.Storage.Download(thumbnailMeta.FID)
		if err != nil {
			s.Logger.Warnf("下载预处理图片失败: %v", err)
			return nil
		}
		meta.Size = thumbnailMeta.Size
		return r
	}
	return nil
}

// 上传生成好的缩略图到tmp collection
func (s *Server) uploadThumbnail(ctx echo.Context, r io.Reader) {
	context := ctx.(*middleware.ExtendContext)
	meta := context.FileContext.Meta
	size := context.FileContext.Size

	option := &common.UploadOption{
		Collection:  CollectionNameTmp,
		Replication: "000",
	}
	needle, err := s.Storage.Upload(option, meta.MimeType, r)
	if err != nil {
		s.Logger.Warnf("上传缩略图失败原图:%s,错误:%v", meta.RawKey, err)
		return
	}
	thumbnailMeta := &model.ThumbnailMeta{
		FID:  needle.FID,
		ETag: needle.ETag,
		Size: needle.Size,
	}
	// TODO(benjamin): save meta
	if err := s.Meta.SubSet(meta.RawKey, fmt.Sprintf("thumbnails.%s", size.Name), thumbnailMeta, 0); err != nil {
		s.Logger.Warnf("更新缩略图失败 %s, %v", meta.RawKey, err)
	} else {
		meta.Thumbnails[size.Name] = thumbnailMeta
	}
}

func (s *Server) encode(ctx echo.Context, img image.Image) (*bytes.Buffer, error) {
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
