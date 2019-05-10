package model

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

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
	Entity      *FileEntity
	Content     *multipart.FileHeader
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
	size := bucket.getSize(name)
	if size != nil {
		self.Key, self.Size = key, size
	}
}

func (self *FileParams) HashKey() string {
	hash, _ := utils.Sha1(self.Key)
	return hash
}
