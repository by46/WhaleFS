package server

import (
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/model"
)

// TODO(benjamin): support parameters from header and multipart-form
func (s *Server) parseBucket(ctx echo.Context) (*model.Bucket, error) {
	var path string
	switch strings.ToLower(ctx.Request().Method) {
	case "get":
		path = ctx.Request().URL.Path
	default:
		path = ctx.Request().Form.Get("key")
	}
	segments := strings.SplitN(path, "/", 3)
	if len(segments) < 3 {
		return nil, fmt.Errorf("invalid url")
	}
	name := segments[1]
	entity, err := s.GetBucket(name)
	if err != nil {
		return nil, errors.Wrapf(err, "bucket %s not exists", name)
	}
	return entity, nil
}

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
