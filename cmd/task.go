package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/by46/whalefs/api"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server"
)

var (
	tarFileDownloadCmd = &cobra.Command{
		Use:   "task",
		Short: "",
		Run:   executeTask,
	}
)

const (
	TmpDir = "./tmp/"
)

func init() {
	rootCmd.AddCommand(tarFileDownloadCmd)
}

func executeTask(cmd *cobra.Command, args []string) {
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

	config, err := server.BuildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}

	storageClient := api.NewStorageClient(config.Storage.Cluster)
	metaClient := api.NewMetaClient(config.Meta)
	taskClient := api.NewTaskClient(config.TaskBucket)

	tasks, err := taskClient.QueryPendingPkgTask("SELECT id,edit_date,error_msg,in_date,status,tar_file_info,tar_file_raw_key FROM `tasks` WHERE status = 0 ORDER BY in_date LIMIT 5")
	if err != nil {
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
		panic(err)
	}

	taskChan := make(chan *model.PackageTask, len(taskMap))

	for _, value := range taskMap {
		tarTask := value.(*model.PackageTask)
		go func(task *model.PackageTask) {
			defer func() {
				defer func() { taskChan <- task }()
				task.Progress = 100
				err = taskClient.Set(task.Id, task)
				if err != nil {
					fmt.Printf("Task update failed![%v]\n", err)
				}
			}()

			tempFileName := TmpDir + task.PackageInfo.Name

			file, err := os.Create(tempFileName)
			if err != nil {
				errMsg := fmt.Sprintf("Create tar file failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				return
			}
			updateProgress(task, taskClient, 10)

			defer func() {
				err = os.Remove(tempFileName)
				if err != nil {
					fmt.Errorf("Remove file error: %s\n", err)
				}
			}()

			err = server.Package(task.PackageInfo, file, getFileEntity(metaClient), downloadFile(storageClient))
			if err != nil {
				errMsg := fmt.Sprintf("Package file failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				return
			}
			updateProgress(task, taskClient, 40)

			err = file.Close()
			if err != nil {
				errMsg := fmt.Sprintf("Tar file close failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				return
			}

			open, err := os.Open(tempFileName)
			if err != nil {
				errMsg := fmt.Sprintf("Tar file open failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				return
			}
			updateProgress(task, taskClient, 50)

			entity, err := storageClient.Upload("application/tar", open)
			if err != nil {
				errMsg := fmt.Sprintf("Tar file upload failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				return
			}
			updateProgress(task, taskClient, 99)

			entity.RawKey = task.PackageRawKey
			if err := metaClient.Set(task.Id, entity); err != nil {
				errMsg := fmt.Sprintf("Tar file save meta failed. %s", err.Error())
				fillErrorTask(task, errMsg)
				return
			}

			task.Status = model.TASK_SUCCESS
			task.EditDate = time.Now().Unix()
		}(tarTask)
	}

	for i := 0; i < len(taskMap); i++ {
		completedTask := <-taskChan
		fmt.Printf("Task %s completed!\n", completedTask.PackageInfo.Name)
	}

	fmt.Printf("All tasks completed!\n")
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
