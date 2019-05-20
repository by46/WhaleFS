package server

import (
	"fmt"
	"time"

	"github.com/by46/whalefs/model"
)

func (s *Server) CreateTask(hashKey string, pkgFileInfo *model.PackageEntity) error {
	task := &model.PackageTask{
		PackageInfo:   pkgFileInfo,
		InDate:        time.Now().Unix(),
		EditDate:      time.Now().Unix(),
		Status:        model.TASK_PENDING,
		Id:            hashKey,
		PackageRawKey: fmt.Sprintf("/%s/%s", s.TaskBucketName, pkgFileInfo.Name),
	}

	if err := s.TaskMeta.Set(hashKey, task); err != nil {
		return err
	}
	return nil
}
