package server

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/by46/whalefs/model"
)

func (s *Server) GetBucket(name string) (*model.Bucket, error) {
	key := fmt.Sprintf("%s.%s", KeyBucket, name)
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
