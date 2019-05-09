package server

import (
	"github.com/by46/whalefs/model"
)

func (s *Server) GetFileEntity(hash string) (*model.FileEntity, error) {
	entity := &model.FileEntity{}
	if err := s.Meta.Get(hash, entity); err != nil {
		return nil, err
	}
	return entity, nil
}
