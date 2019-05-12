package model

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mholt/binding"

	"github.com/by46/whalefs/utils"
)

const (
	Separator = "/"
)

type FileContext struct {
	Key        string
	// 是否允许覆盖已存在文件
	Override    bool
	ExtractFile bool
	Bucket      *Bucket
	Meta        *FileMeta
	File        *FileContent
	Size        *ImageSize
}

// parse image size from path, used to resize picture
func (self *FileContext) ParseImageSize(bucket *Bucket) {
	name, key := utils.PathRemoveSegment(self.Key, 1)
	if name == "" {
		return
	}
	size := bucket.GetSize(name)
	if size != nil {
		self.Key, self.Size = key, size
	}
}

func (self *FileContext) ParseFileContent(params *Params) (err error) {
	// todo(benjamin): upload file from internet source
	file := new(FileContent)
	form := params.Content
	file.Headers = form.Header
	body, err := form.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = body.Close()
	}()
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	file.Content = bytes.NewBuffer(buf)
	file.MimeType = form.Header.Get(echo.HeaderContentType)
	self.File = file
	return
}

func (self *FileContext) HashKey() string {
	hash, _ := utils.Sha1(self.Key)
	return hash
}

type Params struct {
	Key      string
	Override bool
	Content  *multipart.FileHeader
}

func (self *Params) FieldMap(r *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&self.Key: binding.Field{
			Form:     "key",
			Required: true,
		},
		&self.Override: binding.Field{
			Form: "override",
		},
		&self.Content: binding.Field{
			Form: "file",
		},
	}
}

func Bind(ctx echo.Context) (*Params, error) {
	entity := new(Params)

	method := ctx.Request().Method
	if method == "get" || method == "head" {
		values := ctx.Request().URL.Query()
		values.Set("key", ctx.Request().URL.Path)
		ctx.Request().URL.RawQuery = values.Encode()
	}

	err := binding.Bind(ctx.Request(), entity)
	return entity, err
}
