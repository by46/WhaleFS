package model

import (
	"github.com/mholt/binding"
	"mime/multipart"
	"net/http"
)

type FileParams struct {
	Key         string
	BucketName  string
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
