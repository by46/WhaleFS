package server

import (
	"github.com/by46/whalefs/model"
)

func (s *Server) sendMessage(entity *model.FileEntity) {
	if s.Config.Sync.Enable {
		s.rabbitmqCh <- &model.SyncFileEntity{
			Url: entity.Url,
		}
	}
}
