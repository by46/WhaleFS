package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"WhaleFS/api"
	"WhaleFS/common"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Server struct {
	app     *echo.Echo
	Config  *common.Config
	Storage api.IStorage
	Logger  common.ILogger
	Version string
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

func buildLogger(config *common.Config) common.ILogger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.ErrorLevel
	}
	logger.SetLevel(level)
	if err := os.MkdirAll(config.LogHome, os.ModePerm); err != nil {
		fmt.Printf("Create Log Directory %s error: %v", config.LogHome, err)
		os.Exit(-1)
	}
	logFilePath := filepath.Join(config.LogHome, "error.log")
	output, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Printf("Open Log file (%s) err: %v", logFilePath, err)
		os.Exit(-1)
	}
	logger.Out = output
	return logger
}

func buildStorage(config *common.Config) api.IStorage {
	return api.NewStorageClient(config.Master)
}

func NewServer() *Server {

	config, err := buildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}

	logger := buildLogger(config)
	storage := buildStorage(config)

	app := echo.New()
	app.Use(middleware.Logger())
	srv := &Server{
		app:     app,
		Config:  config,
		Storage: storage,
		Logger:  logger,
		Version: VERSION,
	}
	srv.install()
	return srv
}

func (s *Server) install() {

	s.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.GET, echo.POST},
		ExposeHeaders: []string{"X-Request-Id"},
		AllowHeaders:  []string{"X-Request-Id"},
		MaxAge:        60 * 30,
	}))

	s.app.GET("/engine/faq.htm", s.faq)
}
func (s *Server) ListenAndServe() {
	address := s.Config.Host

	if err := s.app.Start(address); err != nil {
		s.Logger.Fatalf("Listen error %v\n", err)
	}

}
