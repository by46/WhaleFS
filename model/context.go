package model

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/utils"
)

type FileContext struct {
	Key              string
	ObjectName       string // 去掉Bucket之后的Key路径
	BucketName       string
	AttachmentName   string // 用于浏览器中保存时的别名
	UploadId         string
	Override         bool // 是否允许覆盖已存在文件
	IsRandomName     bool // 是否自动生成文件名
	ExtractFile      bool
	Uploads          bool
	Check            bool
	PartNumber       int32
	IsRemoveOriginal bool
	IsDownload       bool
	Bucket           *Bucket
	Meta             *FileMeta
	File             *FileContent
	Size             *ImageSize
}

// parse image size from path, used to resize picture
func (f *FileContext) ParseImageSize(bucket *Bucket) {
	name, key := utils.PathRemoveSegment(f.Key, 1)
	if name == "" {
		return
	}
	size := bucket.GetSize(name)
	if size != nil {
		f.Key, f.Size = key, size
	}
}

func (f *FileContext) ParseFileContentFromRequest(ctx echo.Context) (err error) {
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
	f.File, err = f.buildFileContent(buf, textproto.MIMEHeader(ctx.Request().Header), ctx.Request().URL.Path)
	return err
}

func (f *FileContext) ParseFileContent(url string, formFile *multipart.FileHeader) (err error) {
	if url != "" {
		return f.parseFileContentFromRemote(url)
	}
	if formFile != nil {
		return f.parseFileContentFromForm(formFile)
	}
	return fmt.Errorf("error")
}

func (f *FileContext) parseFileContentFromForm(form *multipart.FileHeader) error {
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
	f.File, err = f.buildFileContent(buf, form.Header, form.Filename)
	return err
}

func (f *FileContext) parseFileContentFromRemote(source string) error {
	response, err := utils.Get(source, nil)
	if response != nil {
		defer func() {
			_ = response.Close()
		}()
	}
	if err != nil {
		return err
	}
	buf, err := response.ReadAll()
	if err != nil {
		return errors.WithStack(err)
	}
	fileName := utils.Url2FileName(source)
	f.File, err = f.buildFileContent(buf, textproto.MIMEHeader(response.Header), fileName)
	return err
}

func (f *FileContext) ParseFileContentFromBytes(buf []byte) (err error) {
	f.File, err = f.buildFileContent(buf, nil, "")
	return
}

func (f *FileContext) buildFileContent(buf []byte, headers textproto.MIMEHeader, filename string) (file *FileContent, err error) {
	file = new(FileContent)
	file.Headers = headers
	file.Content = buf
	file.Size = int64(len(buf))
	filename = strings.Trim(filename, "\"")
	extension := filepath.Ext(filename)
	if filename != "" && extension != "" && extension != ".ashx" {
		file.FileName = filename
		file.Extension = extension
		file.MimeType = utils.MimeTypeByExtension(filename)
	} else {
		file.MimeType = http.DetectContentType(buf)
		file.Extension = utils.ExtensionByMimeType(file.MimeType)
	}
	file.Digest, err = utils.ContentSha1(bytes.NewReader(buf))
	if err != nil {
		return nil, errors.WithMessage(err, "文件内容摘要错误")
	}
	return
}

func (f *FileContext) HashKey() string {
	//hash, _ := utils.Sha1(f.Key)
	//return hash
	return f.Key
}

type FormParams struct {
	Key      string `json:"key" form:"key"`
	Source   string `json:"source" form:"source"`
	Override bool   `json:"override" form:"override"`
	Token    string `json:"token" form:"token"`
}
