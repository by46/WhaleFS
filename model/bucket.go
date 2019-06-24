package model

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/by46/whalefs/constant"
)

var (
	PositionTopRight    = &ImageOverlayPosition{new(int), nil, nil, new(int)}
	PositionTopLeft     = &ImageOverlayPosition{new(int), new(int), nil, nil}
	PositionBottomRight = &ImageOverlayPosition{nil, nil, new(int), new(int)}
	PositionBottomLeft  = &ImageOverlayPosition{nil, new(int), new(int), nil}
	UnitMapping         = map[string]uint64{
		"m": 60,
		"h": 60 * 60,
		"d": 24 * 60 * 60,
		"w": 7 * 24 * 60 * 60,
		"M": 30 * 24 * 60 * 60,
		"y": 365 * 24 * 60 * 60,
	}
)

type ExtendItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TTL string

func (t TTL) Empty() bool {
	return t == ""
}

func (t TTL) Expiry() uint32 {
	value := string(t)
	if value == "" {
		return 0
	}
	l := len(value)
	n, unit := value[:l-1], value[l-1:]
	count, _ := strconv.ParseUint(n, 10, 4)
	if second, exists := UnitMapping[unit]; exists {
		return uint32(count * second)
	}
	return 0
}

func (t TTL) String() string {
	return string(t)
}

type Buckets struct {
	Buckets []string `json:"buckets"`
}

type Basis struct {
	Alias            string `json:"alias"`
	Collection       string `json:"collection"`
	Replication      string `json:"replication"`
	TTL              TTL    `json:"ttl"`
	Expires          *int   `json:"expires"` // unit: second
	DefaultImage     string `json:"default_image"`
	PrepareThumbnail string `json:"prepare_thumbnail"`
	// 触发进行图片预处理的最小宽度
	PrepareThumbnailMinWidth int `json:"prepare_thumbnail_min_width"`
}

type ImageOverlayPosition struct {
	Top    *int `json:"top"`
	Left   *int `json:"left"`
	Bottom *int `json:"bottom"`
	Right  *int `json:"right"`
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
	case constant.OverlayPositionTopLeft:
		o.Position = PositionTopLeft
	case constant.OverlayPositionTopRight:
		o.Position = PositionTopRight
	case constant.OverlayPositionBottomLeft:
		o.Position = PositionBottomLeft
	case constant.OverlayPositionBottomRight:
		o.Position = PositionBottomRight
	default:
		content := []byte(o.PositionString)
		o.Position = new(ImageOverlayPosition)
		if err := json.Unmarshal(content, o.Position); err != nil {
			o.Position = nil
		}
	}
}

func (o *ImageOverlay) RealPosition(background, img image.Image) image.Point {
	width := float64(background.Bounds().Dx() - img.Bounds().Dx())
	height := float64(background.Bounds().Dy() - img.Bounds().Dy())
	position := o.Position
	if position.Top != nil && position.Left != nil {
		return image.Point{
			X: int(math.Min(float64(*position.Left), width)),
			Y: int(math.Min(float64(*position.Top), height)),
		}
	}
	if position.Top != nil && position.Right != nil {
		return image.Point{
			X: int(math.Max(0.0, float64(background.Bounds().Dx()-img.Bounds().Dx()-*position.Right))),
			Y: int(math.Min(float64(*position.Top), height)),
		}
	}
	if position.Bottom != nil && position.Left != nil {
		return image.Point{
			X: int(math.Min(float64(*position.Left), width)),
			Y: int(math.Max(0.0, float64(background.Bounds().Dy()-img.Bounds().Dy()-*position.Bottom))),
		}
	}
	// if position.Bottom >= 0 && position.Right >= 0
	return image.Point{
		X: int(math.Max(0.0, float64(background.Bounds().Dx()-img.Bounds().Dx()-*position.Right))),
		Y: int(math.Max(0.0, float64(background.Bounds().Dy()-img.Bounds().Dy()-*position.Bottom))),
	}
}

type ImageSize struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Mode   string `json:"mode"`
}

type BucketLimit struct {
	MinSize   *int64   `json:"min_size"`
	MaxSize   *int64   `json:"max_size"`
	Width     *int     `json:"width"`
	Height    *int     `json:"height"`
	Ratio     string   `json:"ratio"`
	MimeTypes []string `json:"mime_types"`
}

type Bucket struct {
	Name           string                   `json:"name"`
	Memo           string                   `json:"memo"`
	Basis          *Basis                   `json:"basis"`
	Extends        [] *ExtendItem           `json:"extends"`
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

func (b *Bucket) MaxAge() *int {
	// TODO(benjamin): 处理最大过期时间
	return b.Basis.Expires
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
	if len(b.Sizes) == 0 {
		return nil
	}

	if b.sizesMapping == nil {
		b.Lock()
		defer b.Unlock()
		if b.sizesMapping == nil {
			b.sizesMapping = make(map[string]*ImageSize, len(b.Sizes))
			for _, size := range b.Sizes {
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

func (b *Bucket) HasSizes() bool {
	return len(b.Sizes) != 0
}
