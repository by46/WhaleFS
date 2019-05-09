package model

import (
	"mime/multipart"
	"net/http"

	"github.com/by46/whalefs/utils"
	"github.com/labstack/echo"
	"github.com/mholt/binding"
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
			Form:     "key",
			Required: true,
			Binder: func(name string, values []string, errors binding.Errors) binding.Errors {
				var err error
				self.Key = normalizePath(values[0])
				self.BucketName, err = parseBucketName(self.Key)
				if err != nil {
					errors.Add([]string{name}, binding.TypeError, err.Error())
				}
				return errors
			},
		},
		&self.Content: binding.Field{
			Form:     "file",
			Required: true,
		},
	}
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
