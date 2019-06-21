package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/rafaeljesus/rabbus"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

type SyncProducer struct {
	queue  <-chan *model.SyncFileEntity
	config *model.Config
	common.Logger
	recorder *logrus.Logger
	*RabbitMQ
}

func NewProducer(config *model.Config, ch <-chan *model.SyncFileEntity) *SyncProducer {
	logger := utils.BuildLogger(config.Log.Home, config.Log.Level)
	return &SyncProducer{
		queue:    ch,
		config:   config,
		Logger:   logger,
		recorder: buildRecorder(config.Log.Home, "messages.log"),
		RabbitMQ: New(config, logger),
	}
}

func (s *SyncProducer) Run() {
	s.connect(time.Second)
	for {
		select {
		case entity := <-s.queue:
			s.send(entity)
		}
	}
}

func (s *SyncProducer) send(entity *model.SyncFileEntity) {
	content, _ := json.Marshal(entity)
	if s.r != nil {
		msg := rabbus.Message{
			Exchange: s.config.Sync.RabbitMQExchange,
			Kind:     amqp.ExchangeFanout,
			Key:      "",
			Payload:  content,
		}
		s.r.EmitAsync() <- msg
		return
	}
	s.recorder.Error(string(content))
}
