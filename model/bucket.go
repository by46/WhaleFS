package model

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"
	"sync"
)

const (
	ExtendKeyMaxAge            = "max-age"
	OverlayPositionTopRight    = "TopRight"
	OverlayPositionTopLeft     = "TopLeft"
	OverlayPositionBottomRight = "BottomRight"
	OverlayPositionBottomLeft  = "BottomLeft"
)

var (
	PositionTopRight    = &ImageOverlayPosition{0, 0, -1, -1}
	PositionTopLeft     = &ImageOverlayPosition{0, 0, -1, -1}
	PositionBottomRight = &ImageOverlayPosition{0, 0, -1, -1}
	PositionBottomLeft  = &ImageOverlayPosition{0, 0, -1, -1}
)

type ExtendItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Buckets struct {
	Buckets []string `json:"buckets"`
}

type ImageOverlayPosition struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Bottom int `json:"bottom"`
	Right  int `json:"right"`
}

type ImageOverlay struct {
	Name           string                `json:"name"`
	Default        bool                  `json:"default"`
	PositionString string                `json:"position"`
	Image          string                `json:"image"`
	Opacity        float64               `json:"opacity"`
	Position       *ImageOverlayPosition `json:"-"`
}

func (o *ImageOverlay) Init() {
	switch o.PositionString {
	case OverlayPositionTopLeft:
		o.Position = PositionTopLeft
	case OverlayPositionTopRight:
		o.Position = PositionTopRight
	case OverlayPositionBottomLeft:
		o.Position = PositionBottomLeft
	case OverlayPositionBottomRight:
		o.Position = PositionBottomRight
	default:
		content := []byte(o.PositionString)
		o.Position = new(ImageOverlayPosition)
		if err := json.Unmarshal(content, o.Position); err != nil {
			// TODO(benjamin): log error message
		}
	}
}

func (o *ImageOverlay) RealPosition(background, img image.Image) image.Point {
	width := float64(background.Bounds().Dx() - img.Bounds().Dx())
	height := float64(background.Bounds().Dy() - img.Bounds().Dy())
	position := o.Position
	if position.Top >= 0 && position.Left >= 0 {
		return image.Point{
			X: int(math.Min(float64(position.Left), width)),
			Y: int(math.Min(float64(position.Top), height)),
		}
	}
	if position.Top >= 0 && position.Right >= 0 {
		return image.Point{
			X: int(math.Min(0.0, float64(background.Bounds().Dx()-img.Bounds().Dx()-position.Right))),
			Y: int(math.Min(float64(position.Top), height)),
		}
	}
	if position.Bottom >= 0 && position.Left >= 0 {
		return image.Point{
			X: int(math.Min(float64(position.Left), width)),
			Y: int(math.Min(0.0, float64(background.Bounds().Dy()-img.Bounds().Dy()-position.Bottom))),
		}
	}
	// if position.Bottom >= 0 && position.Right >= 0
	return image.Point{
		X: int(math.Min(0.0, float64(background.Bounds().Dx()-img.Bounds().Dx()-position.Right))),
		Y: int(math.Min(0.0, float64(background.Bounds().Dy()-img.Bounds().Dy()-position.Bottom))),
	}
}

type ImageSize struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Mode   string `json:"mode"`
}

type BucketLimit struct {
	MinSize   int64    `json:"min_size"`
	MaxSize   int64    `json:"max_size"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	MimeTypes []string `json:"mime_types"`
}

type Bucket struct {
	Name           string                   `json:"name"`
	Expires        int                      `json:"expires"` // unit: day
	Extends        [] *ExtendItem           `json:"extends"`
	Memo           string                   `json:"memo"`
	LastEditDate   int64                    `json:"last_edit_date"`
	LastEditUser   string                   `json:"last_edit_user"`
	Sizes          []*ImageSize             `json:"Sizes"`
	Limit          *BucketLimit             `json:"limit"`
	Overlays       []*ImageOverlay          `json:"overlays"`
	Overlay        *ImageOverlay            `json:"-"`
	overlayMapping map[string]*ImageOverlay `json:"-"`
	sizesMapping   map[string]*ImageSize
	sync.Mutex
}

func (b *Bucket) Key() string {
	return fmt.Sprintf("system.bucket.%s", strings.ToLower(b.Name))
}

func (b *Bucket) MaxAge() int {
	return b.getExtendInt(ExtendKeyMaxAge)
}

func (b *Bucket) getExtend(key string) string {
	if b.Extends == nil {
		return ""
	}
	for _, item := range b.Extends {
		if item.Key == key {
			return item.Value
		}
	}
	return ""
}

func (b *Bucket) getExtendInt(key string) int {
	text := b.getExtend(key)
	if text == "" {
		return 0
	}
	value, _ := strconv.ParseInt(text, 10, 32)
	return int(value)

}

// 获取图片切片信息
func (b *Bucket) GetSize(name string) *ImageSize {
	if b.Sizes == nil {
		return nil
	}

	if b.sizesMapping == nil {
		b.Lock()
		defer b.Unlock()
		if b.sizesMapping == nil {
			b.sizesMapping = make(map[string]*ImageSize, len(b.Sizes))
			for _, size := range (b.Sizes) {
				b.sizesMapping[size.Name] = size
			}
		}
	}
	return b.sizesMapping[name]
}

// 获取水印信息, 用于图片添加水印
func (b *Bucket) GetOverlay(name string) *ImageOverlay {
	if b.Overlays == nil {
		return nil
	}

	if b.overlayMapping == nil {
		b.Lock()
		defer b.Unlock()
		if b.overlayMapping == nil {
			b.overlayMapping = make(map[string]*ImageOverlay, len(b.Overlays))
			for _, overlay := range b.Overlays {
				if overlay.Default {
					b.Overlay = overlay
				}
				overlay.Init()
				b.overlayMapping[overlay.Name] = overlay
			}
		}
	}

	if name != "" {
		return b.overlayMapping[name]
	} else {
		return b.Overlay
	}
}
