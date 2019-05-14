package server

import (
	"fmt"
	"time"

	"github.com/by46/whalefs/model"
)

func (s *Server) CreateTask(hashKey string, tarFileInfo *model.TarFileEntity) error {
	task := &model.TarTask{
		TarFileInfo:   tarFileInfo,
		InDate:        time.Now().Unix(),
		EditDate:      time.Now().Unix(),
		Status:        model.TASK_PENDING,
		Id:            hashKey,
		TarFileRawKey: fmt.Sprintf("/%s/%s", s.TaskBucketName, tarFileInfo.Name),
	}

	if err := s.TaskMeta.Set(hashKey, task); err != nil {
		return err
	}
	return nil
}
