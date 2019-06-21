package server

import (
	"github.com/by46/whalefs/model"
)

func (s *Server) sendMessage(entity *model.FileEntity, bucket *model.Bucket) {
	if s.Config.Sync.Enable {
		sizes := make([]string, 0)
		for _, size := range bucket.Sizes {
			sizes = append(sizes, size.Name)
		}
		s.rabbitmqCh <- &model.SyncFileEntity{
			Url: entity.Url,
			Sizes:sizes,
		}
	}
}
