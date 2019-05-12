package model

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/mholt/binding"

	"github.com/by46/whalefs/utils"
)

const (
	Separator = "/"
)

type FileParams struct {
	Key        string
	BucketName string
	// 是否允许覆盖已存在文件
	Override    bool
	ExtractFile bool
	Bucket      *Bucket
	Entity      *FileMeta
	Content     *multipart.FileHeader
	File        *FileContent
	Size        *ImageSize
}

func (self *FileParams) FieldMap(r *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&self.Key: binding.Field{
			Form: "key",
			Binder: func(name string, values []string, errors binding.Errors) binding.Errors {
				err := self.ParseKeyAndBucketName(values[0])
				if err != nil {
					errors.Add([]string{name}, binding.TypeError, err.Error())
				}
				return errors
			},
		},
		&self.Override: binding.Field{
			Form: "override",
		},
		&self.Content: binding.Field{
			Form: "file",
		},
	}
}

func (self *FileParams) ParseKeyAndBucketName(value string) (err error) {
	self.Key = utils.PathNormalize(strings.ToLower(value))
	self.BucketName = utils.PathSegment(self.Key, 0)
	if self.BucketName == "" {
		return fmt.Errorf("invalid bucket name")
	}
	return
}

// parse image size from path, used to resize picture
func (self *FileParams) ParseImageSize(bucket *Bucket) {
	name, key := utils.PathRemoveSegment(self.Key, 1)
	if name == "" {
		return
	}
	size := bucket.GetSize(name)
	if size != nil {
		self.Key, self.Size = key, size
	}
}

func (self *FileParams) ParseFileContent(params *Params) (err error) {
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

func (self *FileParams) HashKey() string {
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
