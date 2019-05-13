package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/by46/whalefs/api"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	middleware2 "github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

type Server struct {
	Debug      bool
	Config     *model.Config
	Storage    common.Storage
	Meta       common.Meta
	BucketMeta common.Meta
	Logger     common.Logger
	Version    string
	app        *echo.Echo
	buckets    *sync.Map
}

func buildConfig() (*model.Config, error) {
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
	logFilePath := filepath.Join(config.Home, "error.log")
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

func buildMeta(config *model.Config) common.Meta {
	return api.NewMetaClient(config.Meta)
}

func buildBucketMeta(config *model.Config) common.Meta {
	return api.NewMetaClient(config.BucketMeta)
}

func NewServer() *Server {
	config, err := buildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}

	logger := buildLogger(config.Log)
	storage := buildStorage(config.Storage)
	meta := buildMeta(config)
	bucketMeta := buildBucketMeta(config)

	app := echo.New()

	srv := &Server{
		app:        app,
		Config:     config,
		Storage:    storage,
		Meta:       meta,
		BucketMeta: bucketMeta,
		Logger:     logger,
		Version:    VERSION,
		buckets:    &sync.Map{},
	}
	srv.install()
	return srv
}

func (s *Server) install() {

	s.app.HTTPErrorHandler = s.HTTPErrorHandler

	s.app.Use(middleware.Logger())

	s.app.Use(middleware2.InjectContext())

	s.app.Use(middleware2.ParseFileParams(middleware2.ParseFileParamsConfig{
		Server: s,
		Skipper: func(context echo.Context) bool {
			url := strings.ToLower(context.Request().URL.Path)
			return url == "/tools" ||
				url == "/tardownload" ||
				url == "/favicon.ico"
		},
	}))

	s.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.HEAD, echo.GET, echo.POST},
		ExposeHeaders: []string{"X-Request-Id"},
		AllowHeaders:  []string{"X-Request-Id"},
		MaxAge:        60 * 30,
	}))

	s.app.GET("/faq.htm", s.faq)
	s.app.GET("/*", s.download)
	s.app.HEAD("/*", s.head)
	s.app.POST("/upload", s.upload)
	s.app.POST("/tarDownload", s.tarDownload)
	s.app.GET("/tools", s.tools)
	s.app.GET("/favicon.ico", s.favicon)
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
