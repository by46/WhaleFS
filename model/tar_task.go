package model

const (
	TASK_PENDING = 0
	TASK_SUCCESS = 2
	TASK_RUNNING = 1
	TASK_FAILED  = -1
)

type TarTask struct {
	Id            string         `json:"id"`
	TarFileInfo   *TarFileEntity `json:"tar_file_info"`
	Status        int8           `json:"status"`
	Progress      int8           `json:"progress"`
	TarFileRawKey string         `json:"tar_file_raw_key"`
	ErrorMsg      string         `json:"error_msg"`
	InDate        int64          `json:"in_date"`
	EditDate      int64          `json:"edit_date"`
}
