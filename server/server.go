package server

import (
	"fmt"
	middleware2 "github.com/by46/whalefs/server/middleware"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/by46/whalefs/api"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Server struct {
	app        *echo.Echo
	Config     *common.Config
	Storage    api.IStorage
	Meta       api.IMeta
	BucketMeta api.IMeta
	Logger     common.ILogger
	Version    string
	buckets    map[string]*model.Bucket
}

func buildConfig() (*common.Config, error) {
	srvConfig := new(common.Config)
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

func buildLogger(config *common.LogConfig) common.ILogger {
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

func buildStorage(config *common.StorageConfig) api.IStorage {
	return api.NewStorageClient(config.Cluster)
}

func buildMeta(config *common.Config) api.IMeta {
	return api.NewMetaClient(config.Meta)
}

func buildBucketMeta(config *common.Config) api.IMeta {
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
	app.Use(middleware.Logger())
	srv := &Server{
		app:        app,
		Config:     config,
		Storage:    storage,
		Meta:       meta,
		BucketMeta: bucketMeta,
		Logger:     logger,
		Version:    VERSION,
	}
	srv.buckets = make(map[string]*model.Bucket)
	srv.install()
	return srv
}

func (s *Server) install() {
	s.app.Use(middleware2.InjectContext())

	s.app.Use(middleware2.ParseFileParams(middleware2.ParseFileParamsConfig{
		s,
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
	s.app.POST("/*", s.upload)
	s.app.GET("/tools", s.tools)
	s.app.GET("/favicon.ico", s.favicon)
}

func (s *Server) error(code int, err error) error {
	s.Logger.Error(err)
	return &echo.HTTPError{
		Code:    code,
		Message: err.Error(),
	}
}

func (s *Server) fatal(err error) error {
	return s.error(http.StatusInternalServerError, err)
}

func (s *Server) objectKey(ctx echo.Context) string {
	uri := ctx.Request().URL.Path
	return strings.ToLower(uri)
}

func (s *Server) hashKey(uri string) (string, error) {
	key := strings.ToLower(uri)
	key = strings.TrimLeft(uri, "/")
	return common.Sha1(key)
}

func (s *Server) ListenAndServe() {
	address := s.Config.Host

	if err := s.app.Start(address); err != nil {
		s.Logger.Fatalf("Listen error %v\n", err)
	}

}
