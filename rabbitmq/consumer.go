// used to sync whale-fs file back into legacy fs
package rabbitmq

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/by46/whalefs/common"
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
func (s *SyncConsumer) normalUrl(entity *model.SyncFileEntity) []*pathPair {
	pairs := make([]*pathPair, 0)
	if !strings.Contains(entity.Url, model.Separator) {
		entity.Url = fmt.Sprintf("pdt/Original/%s", entity.Url)
	}
	relativePath := strings.ReplaceAll(entity.Url, model.Separator, PathSeparator)
	pairs = append(pairs, &pathPair{
		url:  entity.Url,
		path: relativePath,
	})
	if len(entity.Sizes) < 0 {
		return pairs
	}
	for _, size := range entity.Sizes {
		tmp := strings.TrimLeft(utils.PathReplace(entity.Url, 1, size), model.Separator)
		pairs = append(pairs, &pathPair{
			url:  tmp,
			path: strings.ReplaceAll(tmp, model.Separator, PathSeparator),
		})
	}
	return pairs
}

func (s *SyncConsumer) write(content []byte) {
	entity := new(model.SyncFileEntity)
	err := json.Unmarshal(content, entity)
	if err != nil {
		s.Errorf("反序列化消息失败 %v", err)
		return
	}

	pairs := s.normalUrl(entity)
	for _, pair := range pairs {
		s.writeFile(pair)
	}

}
func (s *SyncConsumer) writeFile(pair *pathPair) {
	fullPath := filepath.Join(s.config.Sync.LegacyFSRoot, pair.path)
	if utils.FileExists(fullPath) {
		// TODO(benjamin): process exists
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

	file, err := os.Create(fullPath)
	if err != nil {
		s.Logger.Errorf("create file %s failed: %v", pair.url, err)
		s.recorder.Error(pair.url)
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
}
