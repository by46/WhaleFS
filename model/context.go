package model

import (
	"io/ioutil"
	"mime/multipart"
	"net/textproto"

	"github.com/labstack/echo"
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
	Uploads     bool
	UploadId    string
	PartNumber  int32
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

func (self *FileContext) ParseFileContentFromRequest(ctx echo.Context) (err error) {
	body := ctx.Request().Body
	if body != nil {
		defer func() {
			_ = body.Close()
		}()
	}
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.WithStack(err)
	}

	file := new(FileContent)
	file.Headers = textproto.MIMEHeader(ctx.Request().Header)
	file.Content = buf
	file.Size = int64(len(buf))
	file.MimeType = ctx.Request().Header.Get(echo.HeaderContentType)
	self.File = file
	return
}

func (self *FileContext) ParseFileContent(url string, formFile *multipart.FileHeader) (err error) {
	if url != "" {
		err = self.parseFileContentFromRemote(url)
	} else if formFile != nil {
		err = self.parseFileContentFromForm(formFile)
	}
	return
}

func (self *FileContext) parseFileContentFromForm(form *multipart.FileHeader) error {
	body, err := form.Open()
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		_ = body.Close()
	}()
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.WithStack(err)
	}
	file := new(FileContent)
	file.Headers = form.Header
	file.Content = buf
	file.Size = int64(len(buf))
	file.MimeType = form.Header.Get(echo.HeaderContentType)
	self.File = file
	return nil
}

func (self *FileContext) parseFileContentFromRemote(source string) error {
	response, err := utils.Get(source, nil)
	if response != nil {
		defer func() {
			_ = response.Close()
		}()
	}
	if err != nil {
		return err
	}
	file := new(FileContent)
	file.Content, err = response.ReadAll()
	if err != nil {
		return errors.WithStack(err)
	}
	file.Size = int64(len(file.Content))
	file.MimeType = response.Header.Get(echo.HeaderContentType)
	self.File = file
	return nil
}

func (self *FileContext) HashKey() string {
	//hash, _ := utils.Sha1(self.Key)
	//return hash
	return self.Key
}

type FormParams struct {
	Key      string `json:"key" form:"key"`
	Source   string `json:"source" form:"source"`
	Override bool   `json:"override" form:"override"`
}
