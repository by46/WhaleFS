package api

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type aliyunStorageClient struct {
	EndPoint        string
	AccessKeyID     string
	AccessKeySecret string
	*oss.Client
}

func NewAliyunStorageClient(endPoints, accessKeyId, accessKeySecret string) common.Storage {
	client, err := oss.New(endPoints, accessKeyId, accessKeySecret)
	if err != nil {
		panic(errors.WithStack(err))
	}

	return &aliyunStorageClient{
		EndPoint:        endPoints,
		AccessKeyID:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		Client:          client,
	}
}

func (c *aliyunStorageClient) Download(fid string) (io.ReadCloser, http.Header, error) {
	if strings.Contains(fid, "|") {
		return c.downloadChunks(strings.Split(fid, constant.FIDSep))
	}

	terms := strings.SplitN(fid, ":", 2)
	collection, fid := terms[0], terms[1]
	bucket, _ := c.Bucket(collection)

	body, err := bucket.GetObject(fid)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return body, nil, nil
}

func (c *aliyunStorageClient) Upload(option *common.UploadOption, mimeType string, body io.Reader) (*model.Needle, error) {

	size := int64(0)
	switch v := body.(type) {
	case *bytes.Buffer:
		size = int64(v.Len())
	default:
		size = int64(0)
	}

	bucket, _ := c.Bucket(option.Collection)

	if option.Digest == "" {
		option.Digest, _ = utils.ContentSha1(body)
	}

	err := bucket.PutObject(option.Digest, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &model.Needle{
		FID:          fmt.Sprintf("%s:%s", option.Collection, option.Digest),
		Mime:         mimeType,
		ETag:         option.Digest,
		LastModified: time.Now().UTC().Unix(),
		Size:         size,
	}, nil
}

func (c *aliyunStorageClient) downloadChunks(fids []string) (io.ReadCloser, http.Header, error) {
	return NewChunksReader(c, fids), nil, nil
}
