package server

import (
	"github.com/labstack/echo"
	"whalefs/model"
	"strings"
	"fmt"
)

// TODO(benjamin): support parameters from header and multipart-form
func (s *Server) parseBucket(ctx echo.Context) (*model.Bucket, error) {
	segments := strings.SplitN(ctx.Request().URL.Path, "/", 3)
	if len(segments) < 3 {
		return nil, fmt.Errorf("invalid url")
	}
	name := strings.ToLower(segments[1])
	entity, exists := s.buckets[name]
	if !exists {
		return nil, fmt.Errorf("bucket or alias not exists")
	}
	return entity, nil
}
