package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/by46/whalefs/api"
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

func init() {
	rootCmd.AddCommand(tarFileDownloadCmd)
}

func executeTask(cmd *cobra.Command, args []string) {
	config, err := server.BuildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}

	storageClient := api.NewStorageClient(config.Storage.Cluster)
	metaClient := api.NewMetaClient(config.Meta)
	taskClient := api.NewTaskClient(config.TaskBucket)

	tasks, err := taskClient.QueryPendingTarTask("SELECT id,edit_date,error_msg,in_date,status,tar_file_info,tar_file_raw_key FROM `tasks` WHERE status = 0")
	if err != nil {
		panic(err)
	}

	var task model.TarTask
	for tasks.Next(&task) {
		//go func(task *model.TarTask) {
		//
		//}(task.(*model.TarTask))
		file, err := os.Create(task.TarFileInfo.Name)
		if err != nil {
			panic(err)
		}

		err = server.Package(task.TarFileInfo, file,
			func(hash string) (meta *model.FileMeta, e error) {
				entity := &model.FileMeta{}
				if err := metaClient.Get(hash, entity); err != nil {
					return nil, err
				}
				return entity, nil
			},
			func(url string) (reader io.Reader, headers http.Header, e error) {
				return storageClient.Download(url)
			})
		if err != nil {
			panic(err)
		}

		err = file.Close()
		if err != nil {
			panic(err)
		}

		open, err := os.Open(task.TarFileInfo.Name)
		if err != nil {
			panic(err)
		}

		entity, err := storageClient.Upload("application/tar", open)
		if err != nil {
			panic(err)
		}

		entity.RawKey = task.TarFileRawKey
		if err := metaClient.Set(task.Id, entity); err != nil {
			panic(err)
		}

		task.Status = model.TASK_SUCCESS
		task.EditDate = time.Now().Unix()

		err = taskClient.Set(task.Id, task)
		if err != nil {
			panic(err)
		}
	}
}
