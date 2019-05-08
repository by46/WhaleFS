package server

import (
	"fmt"
	"strings"

	"github.com/by46/whalefs/model"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

// TODO(benjamin): support parameters from header and multipart-form
func (s *Server) parseBucket(ctx echo.Context) (*model.Bucket, error) {
	segments := strings.SplitN(ctx.Request().URL.Path, "/", 3)
	if len(segments) < 3 {
		return nil, fmt.Errorf("invalid url")
	}
	name := segments[1]
	entity, err := s.getBucket(name)
	if err != nil {
		return nil, errors.Wrapf(err, "bucket %s not exists", name)
	}
	return entity, nil
}

func (s *Server) getBucket(name string) (*model.Bucket, error) {
	name = strings.ToLower(name)
	key := fmt.Sprintf("system.bucket.%s", name)
	//key, err := common.Sha1(name)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "compute sha1 digest for %s failed", name)
	//}
	//key = key[:10]

	bucket, exists := s.buckets[name]
	if exists {
		return bucket, nil
	}
	bucket = new(model.Bucket)
	if err := s.BucketMeta.Get(key, bucket); err != nil {
		return nil, errors.Wrapf(err, "get bucket %s failed", name)
	}
	// TODO(benjamin): use more effective data structure
	s.buckets[name] = bucket
	return bucket, nil
}
