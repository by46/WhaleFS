package model

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/mholt/binding"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/utils"
)

const (
	Separator = "/"
)

type FileContext struct {
	Key string
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
	if params.Source != "" {
		self.File, err = self.parseFileContentFromRemote(params.Source)
	} else {
		self.File, err = self.parseFileContentFromForm(params.Content)
	}
	return
}

func (self *FileContext) parseFileContentFromForm(form *multipart.FileHeader) (*FileContent, error) {
	body, err := form.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = body.Close()
	}()
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	file := new(FileContent)
	file.Headers = form.Header
	file.Content = buf
	file.MimeType = form.Header.Get(echo.HeaderContentType)
	return file, nil
}

func (self *FileContext) parseFileContentFromRemote(source string) (*FileContent, error) {
	body, headers, err := utils.Get(source)
	if err != nil {
		return nil, err
	}
	file := new(FileContent)
	file.Content = body
	file.Size = int64(len(body))
	file.MimeType = headers.Get(echo.HeaderContentType)
	return file, nil
}

func (self *FileContext) HashKey() string {
	//hash, _ := utils.Sha1(self.Key)
	//return hash
	return self.Key
}

type Params struct {
	Key      string
	Source   string
	Override bool
	Content  *multipart.FileHeader
}

func (self *Params) FieldMap(r *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&self.Key: binding.Field{
			Form:     "key",
			Required: true,
		},
		&self.Source: binding.Field{
			Form: "source",
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

	method := strings.ToLower(ctx.Request().Method)
	if method == "get" || method == "head" {
		values := ctx.Request().URL.Query()
		values.Set("key", ctx.Request().URL.Path)
		ctx.Request().URL.RawQuery = values.Encode()
	}

	err := binding.Bind(ctx.Request(), entity)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return entity, nil
}
