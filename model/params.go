package model

import (
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mholt/binding"

	"github.com/by46/whalefs/utils"
)

type FileParams struct {
	Key         string
	BucketName  string
	Bucket      *Bucket
	Entity      *FileEntity
	ExtractFile bool
	Content     *multipart.FileHeader
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
		&self.Content: binding.Field{
			Form: "file",
		},
	}
}

func (self *FileParams) ParseKeyAndBucketName(value string) (err error) {
	self.Key = normalizePath(value)
	self.BucketName, err = parseBucketName(self.Key)
	return
}

func (self *FileParams) Bind(ctx echo.Context) (err error) {
	self.Key = normalizePath(ctx.Request().URL.Path)
	self.BucketName, err = parseBucketName(self.Key)
	return
}

func (self *FileParams) HashKey() string {
	hash, _ := utils.Sha1(self.Key)
	return hash
}
