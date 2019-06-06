package server

import (
	"image"
	"io"
	"sync"

	"github.com/disintegration/imaging"
)

type overlayFunc func() image.Image

// 获取用于水印的图标信息
//
func (s *Server) getOverlayByFullName(key string) image.Image {
	if fi, ok := s.overlays.Load(key); ok {
		return fi.(overlayFunc)()
	}

	var wg sync.WaitGroup
	var img image.Image
	var err error
	var r io.ReadCloser

	wg.Add(1)
	fi, loaded := s.overlays.LoadOrStore(key, overlayFunc(func() image.Image {
		wg.Wait()
		return img
	}))
	if loaded {
		return fi.(overlayFunc)()
	}
	r, err = s.downloadFileByFullName(key)
	if err != nil {
		s.Logger.Errorf("下载Overlay图标失败 %s %v", key, err)
	} else {
		if img, err = imaging.Decode(r); err != nil {
			s.Logger.Errorf("图片编码失败 %v", err)
		} else {
			wg.Done()
			s.overlays.Store(key, overlayFunc(func() image.Image { return img }))
			return img
		}
	}
	wg.Done()
	return img
}
