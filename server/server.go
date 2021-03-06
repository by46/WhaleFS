package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"

	"github.com/by46/whalefs/api"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/rabbitmq"
	middleware2 "github.com/by46/whalefs/server/middleware"
	"github.com/by46/whalefs/utils"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

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
	overlays              *sync.Map // 用于存放少量overlay的缓存数据
	TaskBucketName        string
	TaskFileSizeThreshold int64
	I18nBundle            *i18n.Bundle
	LocalizerMap          map[string]*i18n.Localizer
	rabbitmqCh            chan<- *model.SyncFileEntity
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
	switch config.Type {
	case constant.StorageAliYun:
		oss := config.OSS
		return api.NewAliyunStorageClient(oss.EndPoint, oss.AccessKeyID, oss.AccessKeySecret)
	default:
		return api.NewStorageClient(config.Cluster)
	}
}

func BuildDao(connectionString string) common.Dao {
	return api.NewMetaClient(connectionString)
}

func buildTaskMeta(config *model.Config) common.Task {
	return api.NewTaskClient(config.TaskBucket)
}

func buildI18nBundle() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("i18n/error_message.zh.toml")
	bundle.MustLoadMessageFile("i18n/error_message.en.toml")

	return bundle
}

func buildI18nLocalizer(bundle *i18n.Bundle) map[string]*i18n.Localizer {
	localizerMap := make(map[string]*i18n.Localizer, 2)
	localizerMap["zh"] = i18n.NewLocalizer(bundle, "zh")
	localizerMap["en"] = i18n.NewLocalizer(bundle, "en")
	return localizerMap
}

func NewServer() *Server {
	config, err := BuildConfig()
	if err != nil {
		panic(fmt.Errorf("Load config fatal: %s\n", err))
	}

	logger := utils.BuildLogger(config.Log.Home, config.Log.Level)
	storage := buildStorage(config.Storage)
	meta := BuildDao(config.Meta)
	chuckDao := BuildDao(config.ChunkMeta)
	bucketMeta := BuildDao(config.BucketMeta)
	taskMeta := buildTaskMeta(config)
	bundle := buildI18nBundle()
	localizerMap := buildI18nLocalizer(bundle)
	app := echo.New()
	app.Renderer = &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	srv := &Server{
		app:                   app,
		Config:                config,
		Storage:               storage,
		Meta:                  meta,
		BucketMeta:            bucketMeta,
		ChunkDao:              chuckDao,
		Logger:                logger,
		Version:               constant.VERSION,
		buckets:               &sync.Map{},
		overlays:              &sync.Map{},
		TaskBucketName:        config.TaskFileBucketName,
		TaskMeta:              taskMeta,
		TaskFileSizeThreshold: config.TaskFileSizeThreshold,
		I18nBundle:            bundle,
		LocalizerMap:          localizerMap,
	}
	srv.install()

	if config.Sync.Enable {
		rabbitmqCh := make(chan *model.SyncFileEntity, 100)
		srv.rabbitmqCh = rabbitmqCh
		producer := rabbitmq.NewProducer(config, rabbitmqCh)
		go producer.Run()
	}
	return srv
}

func (s *Server) install() {

	s.app.HTTPErrorHandler = s.HTTPErrorHandler

	s.app.Use(middleware.Recover())

	s.app.Use(middleware.Logger())

	s.app.Use(middleware2.InjectServer())

	s.app.Use(middleware2.NewMetric())

	s.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.HEAD, echo.GET, echo.POST, echo.PUT, echo.DELETE},
		ExposeHeaders: []string{"X-Request-Id"},
		AllowHeaders: []string{
			"X-Request-Id",
			"X-Requested-With",
			"X_Requested_With",
			"X-Requested-LangCode",
			"projectsysno",
			"content-type",
			"Authorization",
			constant.HeaderXWhaleFSToken,
			constant.HeaderVia},
		MaxAge: 60 * 30,
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
	//s.app.GET("/tools", s.tools)
	//s.app.GET("/pkgDownloadTool", s.pkgDownloadTool)
	s.app.GET("/favicon.ico", s.favicon)
	s.app.GET("/tasks", s.checkTask)
	s.app.GET("/metrics", s.metric)
	s.app.HEAD("/status.html", s.status)

	api := s.app.Group("/api", middleware2.InjectUser(middleware2.AuthUserConfig{
		Server: s,
		Skipper: func(context echo.Context) bool {
			return strings.HasPrefix(context.Request().URL.Path, "/api/") == false ||
				context.Request().URL.Path == "/api/login2"
		},
	}))
	api.GET("/users", s.listUser)
	api.POST("/users", s.addUser)
	api.PUT("/users", s.updateUser)
	api.DELETE("/users/*", s.deleteUser)
	api.GET("/users/:id", s.getUser)

	api.GET("/buckets", s.listBucket)
	api.GET("/buckets/:id", s.getBucket)
	api.PUT("/buckets", s.updateBucket)
	api.DELETE("/buckets/:id", s.deleteBucket)
	api.POST("/buckets", s.addBucket)
	api.GET("/bucket-names", s.listBucketNames)

	api.GET("/mimetypes", s.listMimeTypes)

	api.POST("/access-key/", s.createAccessKey)
	api.GET("/access-key/", s.listAccessKey)
	api.POST("/access-key/:key", s.updateAccessKey)
	api.DELETE("/access-key/:key", s.deleteAccessKey)
	api.POST("/access-key/token", s.generateToken)

	s.app.POST("/api/login", s.login)
	s.app.GET("/api/configuration", s.configuration)

	s.app.POST("/UploadHandler.ashx", s.legacyUploadFile)
	s.app.POST("/BatchDownloadHandler.ashx", s.legacyBatchDownload)
	s.app.POST("/BatchDownLoadHandler.ashx", s.legacyBatchDownload)
	s.app.Match(methods, "/DownloadSaveServerHandler.ashx", s.legacyUploadByRemote)
	s.app.GET("/DownloadHandler.ashx", s.legacyDownloadFile)
	s.app.GET("/DownLoadHandler.ashx", s.legacyDownloadFile)
	s.app.Match(methods, "/ApiUploadHandler.ashx", s.legacyApiUpload)
	s.app.Match(methods, "/BatchMergePdfHandler.ashx", s.legacyMergePDF)
	s.app.Match(methods, "/SliceUploadHandler.ashx", s.legacySliceUpload)

	s.app.GET("/", s.home)
	s.app.POST("/", s.uploadByForm)
	s.app.GET("/*", s.downloadByUrl)
	s.app.PUT("/*", s.uploadByBody)
	s.app.POST("/*", s.uploadByChunks)
	s.app.DELETE("/*", s.deleteChunks)
}

func (s *Server) ListenAndServe() {
	address := s.Config.Host

	if err := s.app.Start(address); err != nil {
		s.Logger.Fatalf("Listen error %v\n", err)
	}
}
