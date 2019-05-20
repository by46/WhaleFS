package task

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server"
)

const (
	TmpDir = "./tmp/"
)

func (s *Scheduler) RunPackageFileTask() {
	exist, err := pathExists(TmpDir)
	if err != nil {
		panic(fmt.Errorf("System error: %s\n", err))
	}

	if !exist {
		err = os.Mkdir(TmpDir, os.ModePerm)
		if err != nil {
			panic(fmt.Errorf("System error: %s\n", err))
		}
	}

	logger := s.Logger
	storageClient := s.Storage
	metaClient := s.Meta
	taskClient := s.TaskMeta

	logger.Info("Package files task running...")

	tasks, err := taskClient.QueryPendingPkgTask("SELECT id,edit_date,error_msg,in_date,status,package_info,package_raw_key FROM `tasks` WHERE status = 0 ORDER BY in_date LIMIT 5")
	if err != nil {
		logger.Errorf("Select tasks error: %v", err)
		panic(err)
	}

	taskMap := make(map[string]interface{})

	for {
		tmpTask := new(model.PackageTask)
		if tasks.Next(tmpTask) == false {
			break
		}
		tmpTask.Status = model.TASK_RUNNING
		taskMap[tmpTask.Id] = tmpTask
	}
	err = taskClient.BulkUpdate(taskMap)
	if err != nil {
		logger.Errorf("Bulk update tasks status error: %v", err)
		panic(err)
	}

	taskChan := make(chan *model.PackageTask, len(taskMap))

	for _, value := range taskMap {
		pkgTask := value.(*model.PackageTask)
		go func(task *model.PackageTask) {
			defer func() {
				defer func() { taskChan <- task }()
				task.Progress = 100
				err = taskClient.Set(task.Id, task)
				if err != nil {
					logger.Errorf("Task update failed![%v]\n", err)
				}
			}()

			tempFileName := TmpDir + task.PackageInfo.Name

			file, err := os.Create(tempFileName)
			if err != nil {
				errMsg := fmt.Sprintf("Create package file failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				logger.Error(errMsg)
				return
			}
			updateProgress(task, taskClient, 10)

			defer func() {
				err = os.Remove(tempFileName)
				if err != nil {
					logger.Errorf("Remove file error: %v\n", err)
				}
			}()

			err = server.Package(task.PackageInfo, file, getFileEntity(metaClient), downloadFile(storageClient))
			if err != nil {
				errMsg := fmt.Sprintf("Package file failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				logger.Error(errMsg)
				return
			}
			updateProgress(task, taskClient, 40)

			err = file.Close()
			if err != nil {
				errMsg := fmt.Sprintf("Package file close failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				logger.Error(errMsg)
				return
			}

			open, err := os.Open(tempFileName)
			if err != nil {
				errMsg := fmt.Sprintf("Package file open failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				logger.Error(errMsg)
				return
			}
			updateProgress(task, taskClient, 50)

			entity, err := storageClient.Upload("application/tar", open)
			if err != nil {
				errMsg := fmt.Sprintf("Package file upload failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				logger.Error(errMsg)
				return
			}
			updateProgress(task, taskClient, 99)

			entity.RawKey = task.PackageRawKey
			if err := metaClient.Set(task.Id, entity); err != nil {
				errMsg := fmt.Sprintf("Package file save meta failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				logger.Error(errMsg)
				return
			}

			task.Status = model.TASK_SUCCESS
			task.EditDate = time.Now().Unix()
		}(pkgTask)
	}

	for i := 0; i < len(taskMap); i++ {
		completedTask := <-taskChan
		logger.Infof("Task %s completed!\n", completedTask.PackageInfo.Name)
	}

	logger.Info("All tasks completed!\n")
}

func getFileEntity(metaClient common.Dao) func(hash string) (meta *model.FileMeta, e error) {
	return func(hash string) (meta *model.FileMeta, e error) {
		entity := &model.FileMeta{}
		if err := metaClient.Get(hash, entity); err != nil {
			return nil, err
		}
		return entity, nil
	}
}

func downloadFile(storageClient common.Storage) func(url string) (reader io.Reader, headers http.Header, e error) {
	return func(url string) (reader io.Reader, headers http.Header, e error) {
		return storageClient.Download(url)
	}
}

func fillErrorTask(task *model.PackageTask, errMsg string) {
	task.ErrorMsg = errMsg
	task.Status = model.TASK_FAILED
	task.EditDate = time.Now().Unix()
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func updateProgress(task *model.PackageTask, taskClient common.Task, progress int8) {
	tmpTask := new(model.PackageTask)
	task.Progress = progress
	err := taskClient.Set(task.Id, tmpTask)
	if err != nil {
		fmt.Printf("Update progress failed!\n")
	}
}