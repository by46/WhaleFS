// used to sync whale-fs file back into legacy fs
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafaeljesus/rabbus"
	"github.com/sirupsen/logrus"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

var (
	PathSeparator = string(os.PathSeparator)
)

type pathPair struct {
	url  string
	path string
}

type SyncConsumer struct {
	config *model.Config
	common.Logger
	recorder *logrus.Logger
}

func NewConsumer(config *model.Config) *SyncConsumer {
	logger := utils.BuildLogger(config.Log.Home, config.Log.Level)
	return &SyncConsumer{
		config,
		logger,
		buildRecorder(config.Log.Home, "files.log"),
	}
}

func (s *SyncConsumer) Run() {
	conn, err := Dial(s.config.Sync.RabbitMQConnection)
	if err != nil {
		s.Fatalf("connect rabbitmq failed, %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		s.Fatalf("retrieve channel, %v", err)
	}
	d, err := channel.Consume(s.config.Sync.QueueName, "", false, false, false, false, nil)
	if err != nil {
		s.Fatal(err)
	}

	for msg := range d {
		s.write(msg.Body)
		err := msg.Ack(false)
		if err != nil {
			s.Errorf("确认消息失败")
		}
	}
}

func (s *SyncConsumer) Run2() {
	r, err := rabbus.New(s.config.Sync.RabbitMQConnection, rabbus.Durable(true),
		rabbus.Attempts(5),
		rabbus.Sleep(time.Second*2),
		rabbus.Threshold(3))
	if err != nil {
		s.Fatalf("connection rabbitmq failed: %v", err)
	}

	defer func() {
		_ = r.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)

	messages, err := r.Listen(rabbus.ListenConfig{
		Exchange: s.config.Sync.RabbitMQExchange,
		Kind:     rabbus.ExchangeFanout,
		Key:      "",
		Queue:    s.config.Sync.RabbitMQQueue,
	})
	if err != nil {
		s.Fatalf("add listen handler error: %v", err)
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
	if !strings.Contains(entity.Url, constant.Separator) {
		entity.Url = fmt.Sprintf("pdt/Original/%s", entity.Url)
	}
	relativePath := strings.ReplaceAll(entity.Url, constant.Separator, PathSeparator)
	pairs = append(pairs, &pathPair{
		url:  entity.Url,
		path: relativePath,
	})
	if len(entity.Sizes) < 0 {
		return pairs
	}
	for _, size := range entity.Sizes {
		tmp := strings.TrimLeft(utils.PathReplace(entity.Url, 1, size), constant.Separator)
		pairs = append(pairs, &pathPair{
			url:  tmp,
			path: strings.ReplaceAll(tmp, constant.Separator, PathSeparator),
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
		s.recorder.Errorf("exists, %s", pair.url)
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
		s.recorder.Errorf("failed, %s", pair.url)
		return
	}

	if strings.Contains(response.Header.Get(constant.HeaderXWhaleFSFlags), constant.FlagDefaultImage) {
		s.recorder.Errorf("ignore, %s", pair.url)
		return true
	}

	file, err := os.Create(fullPath)
	if err != nil {
		s.Logger.Errorf("create file %s failed: %v", pair.url, err)
		s.recorder.Errorf("denies, %s", pair.url)
		return
	}
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if _, err := io.Copy(file, response); err != nil {
		s.Logger.Errorf("write file %s failed: %v", pair.url, err)
		s.recorder.Error(pair.url)
		return
	}
	return
}