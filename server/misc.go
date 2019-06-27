package server

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
)

func (s *Server) sendMessage(entity *model.FileEntity, bucket *model.Bucket) {
	if s.Config.Sync.Enable {
		sizes := make([]string, 0)
		for _, size := range bucket.Sizes {
			sizes = append(sizes, size.Name)
		}
		s.rabbitmqCh <- &model.SyncFileEntity{
			Url:   entity.Url,
			Sizes: sizes,
		}
	}
}

// 用于同步配置信息，包括bucket信息
func (s *Server) SyncConfig(ctx context.Context) {
	ts := &model.Timestamp{}
	err := s.BucketMeta.Get(constant.KeyTimeStamp, ts)
	if err != nil {
		s.Logger.Errorf("get timestamp error %v", errors.WithStack(err))
	}
	for {
		timer := time.After(5 * time.Second)
		select {
		case <-ctx.Done():
			return
		case <-timer:
		}
		current := &model.Timestamp{}
		err = s.BucketMeta.Get(constant.KeyTimeStamp, current)
		if current.BucketUpdate == ts.BucketUpdate {
			continue
		}
		s.clearBucket()
		ts.BucketUpdate = current.BucketUpdate
	}
}
