package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/disintegration/imaging"
	"github.com/hhrutter/pdfcpu/pkg/api"
	pdf "github.com/hhrutter/pdfcpu/pkg/pdfcpu"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

type Chunk struct {
	FID  string `json:"fid"`
	ETag string `json:"etag"`
}

type Sizes map[string]*Chunk

type A struct {
	Name  string `json:"name"`
	Sizes Sizes  `json:"sizes"`
}

func TestJson(t *testing.T) {
	content := `{"name":"benjamin", "sizes":{"p200":{"fid":"1,21231","etag":"etag1"}, "p400":{"fid":"2,21231","etag":"etag2"}}}`
	a := new(A)
	err := json.Unmarshal([]byte(content), a)
	assert.Nil(t, err)
	assert.Equal(t, &Chunk{FID: "1,21231", ETag: "etag1"}, a.Sizes["p200"])
	assert.Equal(t, &A{Name: "benjamin", Sizes: map[string]*Chunk{"p200": {FID: "1,21231", ETag: "etag1"}, "p400": {FID: "2,21231", ETag: "etag2"}}}, a)

	content = `{"name":"benjamin", "sizes":{}}`
	a = new(A)
	err = json.Unmarshal([]byte(content), a)
	assert.Nil(t, err)
	assert.Equal(t, &A{Name: "benjamin", Sizes: make(Sizes)}, a)
}

func TestPDFMerge(t *testing.T) {
	rr := make([]pdf.ReadSeekerCloser, 0)
	files := []string{"../sample/raft.pdf", "../sample/Beaver.pdf"}
	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			log.Fatalf("open file %s failed %v\n", name, err)
		}
		rr = append(rr, f)
	}

	defer func() {
		for _, closer := range rr {
			_ = closer.Close()
		}
	}()

	config := pdf.NewDefaultConfiguration()
	config.Cmd = pdf.MERGE

	ctx, err := api.MergeContexts(rr, config)
	assert.Nil(t, err)

	out, err := os.Create("../sample/merge2.pdf")
	assert.Nil(t, err)
	defer func() { _ = out.Close() }()

	_ = api.WriteContext(ctx, out)
}

func TestSyncMap(t *testing.T) {
	type NameFunc func() string
	m := &sync.Map{}
	value, loaded := m.LoadOrStore("name", NameFunc(func() string {
		fmt.Printf("debugging in")
		return "benjamin"
	}))

	assert.Equal(t, "benjamin", value)
	assert.False(t, loaded)

	value, loaded = m.LoadOrStore("name", NameFunc(func() string {
		fmt.Printf("debugging in")
		return "benjamin"
	}))

	assert.Equal(t, "benjamin", value)
	assert.True(t, loaded)
}

func TestRange(t *testing.T) {
	sizes := make([]string, 0)
	for _, size := range sizes {
		fmt.Printf("hello %s", size)
	}
	sizes = nil
	for _, size := range sizes {
		fmt.Printf("hello %s", size)
	}
}

func TestPanicAndRecover(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("error: %v", err)
		}
	}()

	var a *A
	//panic(a)
	fmt.Printf("tricky: %v %v", a == nil, a.Name)
}

func TestImage(t *testing.T) {
	content, err := ioutil.ReadFile("sample/file.png")
	assert.Nil(t, err)
	img, err := imaging.Decode(bytes.NewReader(content))
	assert.Nil(t, err)
	newImg := imaging.Fit(img, 160, 160, imaging.Lanczos)
	_ = imaging.Save(newImg, "sample/file-fit.png")

	newImg = imaging.Thumbnail(img, 160, 160, imaging.Lanczos)
	_ = imaging.Save(newImg, "sample/file-thumbnail.png")

	newImg = imaging.Resize(img, 160, 160, imaging.Lanczos)
	_ = imaging.Save(newImg, "sample/file-resize.png")

	newImg = imaging.Resize(img, 160, 0, imaging.Lanczos)
	_ = imaging.Save(newImg, "sample/file-resize-height.png")

	newImg = imaging.Resize(img, 0, 160, imaging.Lanczos)
	_ = imaging.Save(newImg, "sample/file-resize-width.png")
}

func TestTimestamp2(t *testing.T) {
	fmt.Printf("%v", time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
}
