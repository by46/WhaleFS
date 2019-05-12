package server

import (
	"github.com/by46/whalefs/model"
)

func (s *Server) GetFileEntity(hash string) (*model.FileMeta, error) {
	entity := &model.FileMeta{}
	if err := s.Meta.Get(hash, entity); err != nil {
		return nil, err
	}
	return entity, nil
}
