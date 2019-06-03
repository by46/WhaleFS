package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"

	"github.com/by46/whalefs/api"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

type Server struct {
	Debug                 bool
	Config                *model.Config
	Storage               common.Storage
	Meta                  common.Dao
	BucketMeta            common.Dao
	ChunkDao              common.Dao
	TaskMeta              common.Task
	Logger                common.Logger
	Version               string
	app                   *echo.Echo
	buckets               *sync.Map
	TaskBucketName        string
	TaskFileSizeThreshold int64
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

func buildStorage(config *model.StorageConfig) common.Storage {
	return api.NewStorageClient(config.Cluster)
}

func buildDao(connectionString string) common.Dao {
	return api.NewMetaClient(connectionString)
}

func buildTaskMeta(config *model.Config) common.Task {
	return api.NewTaskClient(config.TaskBucket)
}

func NewServer() *Server {
	config, err := BuildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}

	logger := utils.BuildLogger(config.Log.Home, config.Log.Level)
	storage := buildStorage(config.Storage)
	meta := buildDao(config.Meta)
	chuckDao := buildDao(config.ChunkMeta)
	bucketMeta := buildDao(config.BucketMeta)
	taskMeta := buildTaskMeta(config)
	app := echo.New()

	srv := &Server{
		app:                   app,
		Config:                config,
		Storage:               storage,
		Meta:                  meta,
		BucketMeta:            bucketMeta,
		ChunkDao:              chuckDao,
		Logger:                logger,
		Version:               VERSION,
		buckets:               &sync.Map{},
		TaskBucketName:        config.TaskFileBucketName,
		TaskMeta:              taskMeta,
		TaskFileSizeThreshold: config.TaskFileSizeThreshold,
	}
	srv.install()
	return srv
}

func (s *Server) install() {

	s.app.HTTPErrorHandler = s.HTTPErrorHandler

	s.app.Use(middleware.Logger())

	//s.app.Use(middleware2.InjectContext())

	//s.app.Use(middleware2.ParseFileParams(middleware2.ParseFileParamsConfig{
	//	Server: s,
	//	Skipper: func(context echo.Context) bool {
	//		url := strings.ToLower(context.Request().URL.Path)
	//		return url == "/tools" ||
	//			url == "/packagedownload" ||
	//			url == "/pkgdownloadtool" ||
	//			url == "/tasks" ||
	//			url == "/metric" ||
	//			url == "/favicon.ico"
	//	},
	//}))

	s.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.HEAD, echo.GET, echo.POST},
		ExposeHeaders: []string{"X-Request-Id"},
		AllowHeaders:  []string{"X-Request-Id"},
		MaxAge:        60 * 30,
	}))
	methods := []string{
		http.MethodHead,
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	s.app.GET("/faq.htm", s.faq)
	s.app.POST("/packageDownload", s.packageDownload)
	s.app.GET("/tools", s.tools)
	s.app.GET("/pkgDownloadTool", s.pkgDownloadTool)
	s.app.GET("/favicon.ico", s.favicon)
	s.app.GET("/tasks", s.checkTask)
	s.app.GET("/metrics", s.metric)
	s.app.POST("/demo", s.demo)
	s.app.POST("/UploadHandler.ashx", s.legacyUploadFile)
	s.app.Match(methods, "/DownloadSaveServerHandler.ashx", s.legacyUploadByRemote)
	s.app.GET("/DownloadHandler.ashx", s.legacyDownloadFile)
	s.app.POST("/ApiUploadHandler.ashx", s.legacyApiUpload)
	s.app.Match(methods, "/*", s.file)
}

func (s *Server) hashKey(uri string) (string, error) {
	key := strings.ToLower(uri)
	key = strings.TrimLeft(uri, "/")
	return utils.Sha1(key)
}

func (s *Server) ListenAndServe() {
	address := s.Config.Host

	if err := s.app.Start(address); err != nil {
		s.Logger.Fatalf("Listen error %v\n", err)
	}

}
