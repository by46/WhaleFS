package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/by46/whalefs/api"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
)

type Scheduler struct {
	Debug          bool
	Config         *model.Config
	Storage        common.Storage
	Meta           common.Dao
	BucketMeta     common.Dao
	TaskMeta       common.Task
	Logger         common.Logger
	Version        string
	buckets        *sync.Map
	TaskBucketName string
	Cron           *cron.Cron
	HttpClientBase string
	TempFileDir    string
}

func BuildConfig() (*model.Config, error) {
	srvConfig := new(model.Config)
	config := viper.New()
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	config.SetConfigName(strings.ToLower(env))
	config.AddConfigPath("config")
	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := config.Unmarshal(srvConfig); err != nil {
		return nil, err
	}
	return srvConfig, nil
}

func buildLogger(config *model.LogConfig) common.Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.ErrorLevel
	}
	logger.SetLevel(level)
	if err := os.MkdirAll(config.Home, os.ModePerm); err != nil {
		fmt.Printf("Create Log Directory %s error: %v", config.Home, err)
		os.Exit(-1)
	}
	logFilePath := filepath.Join(config.Home, "task_error.log")
	output, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Printf("Open Log file (%s) err: %v", logFilePath, err)
		os.Exit(-1)
	}
	logger.Out = output
	return logger
}

func buildStorage(config *model.StorageConfig) common.Storage {
	return api.NewStorageClient(config.Cluster)
}

func buildMeta(config *model.Config) common.Dao {
	return api.NewMetaClient(config.Meta)
}

func buildBucketMeta(config *model.Config) common.Dao {
	return api.NewMetaClient(config.BucketMeta)
}

func buildTaskMeta(config *model.Config) common.Task {
	return api.NewTaskClient(config.TaskBucket)
}

func NewScheduler() *Scheduler {
	config, err := BuildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}

	logger := buildLogger(config.Log)
	storage := buildStorage(config.Storage)
	meta := buildMeta(config)
	bucketMeta := buildBucketMeta(config)
	taskMeta := buildTaskMeta(config)
	cron := cron.New()

	scheduler := &Scheduler{
		Config:         config,
		Storage:        storage,
		Meta:           meta,
		BucketMeta:     bucketMeta,
		Logger:         logger,
		buckets:        &sync.Map{},
		TaskBucketName: config.TaskFileBucketName,
		TaskMeta:       taskMeta,
		Cron:           cron,
		HttpClientBase: config.HttpClientBase,
		TempFileDir:    config.TempFileDir,
	}
	scheduler.install()
	return scheduler
}

func (s *Scheduler) install() {
	err := s.Cron.AddFunc("@every 1m", func() {
		s.RunPackageFileTask()
	})
	if err != nil {
		panic(err)
	}
}
