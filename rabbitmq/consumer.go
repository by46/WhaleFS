// used to sync whale-fs file back into legacy fs
package rabbitmq

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rafaeljesus/rabbus"
	"github.com/sirupsen/logrus"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

var (
	PathSeparator     = string(os.PathSeparator)
	ReReplaceOriginal = regexp.MustCompile("(^.*?)[\\//]Original[\\//](.*)")
)

type pathPair struct {
	url  string
	path string
}

type SyncConsumer struct {
	config *model.Config
	common.Logger
	recorder *logrus.Logger
	*RabbitMQ
}

func NewConsumer(config *model.Config) *SyncConsumer {
	logger := utils.BuildLogger(config.Log.Home, config.Log.Level)
	return &SyncConsumer{
		config:   config,
		Logger:   logger,
		recorder: buildRecorder(config.Log.Home, "files.log"),
		RabbitMQ: New(config, logger),
	}
}

func (s *SyncConsumer) Run() {
	s.connect(time.Second)
	timer := time.After(time.Second * 5)
	for s.r == nil {
		select {
		case <-timer:
			if s.r == nil {
				s.Errorf("connect rabbit failed")
			}
		}
	}
	messages, err := s.r.Listen(rabbus.ListenConfig{
		Exchange: s.config.Sync.RabbitMQExchange,
		Kind:     rabbus.ExchangeFanout,
		Key:      "",
		Queue:    s.config.Sync.RabbitMQQueue,
	})
	if err != nil {
		s.Fatalf("add listen handler error: %v", errors.WithStack(err))
	}
	defer close(messages)

	for message := range messages {
		s.write(message.Body)
		err := message.Ack(false)
		if err != nil {
			s.Errorf("ack message failed: %v", err)
		}
	}
}

func (s *SyncConsumer) normalUrl(entity *model.SyncFileEntity) []*pathPair {
	if strings.TrimSpace(entity.Url) == "" {
		return nil
	}
	pairs := make([]*pathPair, 0)
	url := entity.Url
	relativePath := strings.ReplaceAll(entity.Url, constant.Separator, PathSeparator)
	if !strings.Contains(url, constant.Separator) {
		url = path.Join("pdt", "Original", url)
		relativePath = filepath.Join("pdt", "Original", utils.SubFolderByFileName(entity.Url))
	}
	pairs = append(pairs, &pathPair{
		url:  url,
		path: relativePath,
	})
	if len(entity.Sizes) < 0 {
		return pairs
	}
	for _, size := range entity.Sizes {
		tmp := strings.Replace(url, "/Original/", fmt.Sprintf("/%s/", size), 1)
		//tmp2 := ReReplaceOriginal.ReplaceAllString(relativePath, fmt.Sprintf("$1/%s/$2", size))
		tmp2 := strings.Replace(relativePath, "Original", size, 1)
		pairs = append(pairs, &pathPair{
			url:  tmp,
			path: tmp2,
		})
	}
	return pairs
}

func (s *SyncConsumer) write(content []byte) {
	entity := new(model.SyncFileEntity)
	err := json.Unmarshal(content, entity)
	if err != nil {
		msg := content
		if len(msg) > 1024 {
			msg = msg[:1024]
		}
		s.Errorf("反序列化消息失败:%v, message: %v", err, string(msg))
		return
	}

	pairs := s.normalUrl(entity)
	skip := false
	for _, pair := range pairs {
		if skip {
			s.recorder.Errorf("ignore, %s", pair.url)
			continue
		}
		skip = s.writeFile(pair)
	}

}

func (s *SyncConsumer) writeFile(pair *pathPair) (skip bool) {
	fullPath := filepath.Join(s.config.Sync.LegacyFSRoot, pair.path)
	if utils.FileExists(fullPath) {
		s.recorder.Errorf("exists, %s %s", pair.url, pair.path)
		return
	}

	parentPath := filepath.Dir(fullPath)
	err := os.MkdirAll(parentPath, os.ModePerm)
	if err != nil {
		s.recorder.Error(pair.url)
		return
	}
	url := fmt.Sprintf("http://%s/%s", s.config.Sync.DFSHost, pair.url)
	response, err := utils.Get(url, nil)

	if err != nil {
		s.recorder.Error(pair.url)
		return
	}
	defer func() {
		_ = response.Close()
	}()

	if response.StatusCode != http.StatusOK {
		s.recorder.Errorf("failed, %s %s", pair.url, pair.path)
		return
	}

	if strings.Contains(response.Header.Get(constant.HeaderXWhaleFSFlags), constant.FlagDefaultImage) {
		s.recorder.Errorf("ignore, %s %s", pair.url, pair.path)
		return true
	}

	file, err := os.Create(fullPath)
	if err != nil {
		s.Logger.Errorf("create file %s failed: %v", pair.url, err)
		s.recorder.Errorf("denies, %s %s", pair.url, pair.path)
		return
	}
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if _, err := io.Copy(file, response); err != nil {
		s.Logger.Errorf("write file %s failed: %v", pair.url, err)
		s.recorder.Error("error, %s %s", pair.url, pair.path)
		return
	}
	return
}
