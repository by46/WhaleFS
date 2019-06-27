package server

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
)

func (s *Server) GetBucket(name string) (*model.Bucket, error) {
	key := fmt.Sprintf("%s.%s", constant.KeyBucket, name)
	value, exists := s.buckets.Load(key)
	if exists {
		return value.(*model.Bucket), nil
	}
	bucket := new(model.Bucket)
	if err := s.BucketMeta.Get(key, bucket); err != nil {
		return nil, errors.Wrapf(err, "get bucket %s failed", name)
	}
	s.buckets.Store(key, bucket)
	return bucket, nil
}

func (s *Server) parseBucketAndFileKey(value string) (bucket *model.Bucket, key string, size *model.ImageSize) {
	if value == "" {
		return
	}
	value = strings.TrimLeft(value, constant.Separator)
	key = value
	segments := strings.Split(value, constant.Separator)
	if len(segments) == 0 {
		return
	}
	bucketName := segments[0]

	bucket, err := s.GetBucket(bucketName)
	if err != nil {
		return
	}
	// 至少包含/bucket/size/hello.jpg
	if len(segments) >= 3 {
		sizeName := segments[1]
		size := bucket.GetSize(sizeName)
		if size != nil {
			key = strings.Join(append([]string{bucketName}, segments[2:]...), constant.Separator)
		}
	}
	key = "/" + key
	return
}

func (s *Server) clearBucket() {
	s.buckets = &sync.Map{}
}
