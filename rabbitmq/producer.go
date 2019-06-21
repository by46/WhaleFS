package rabbitmq

import (
	"encoding/json"

	"github.com/pkg/errors"
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
	ch       *Channel
	recorder *logrus.Logger
}

func NewProducer(config *model.Config, ch <-chan *model.SyncFileEntity) *SyncProducer {
	logger := utils.BuildLogger(config.Log.Home, config.Log.Level)
	return &SyncProducer{
		queue:    ch,
		config:   config,
		Logger:   logger,
		recorder: buildRecorder(config.Log.Home, "messages.log"),
	}
}

func (s *SyncProducer) Run() {
	s.ch = s.channel()
	for {
		select {
		case entity := <-s.queue:
			s.send(entity)
		}
	}
}

func (s *SyncProducer) send(entity *model.SyncFileEntity) {
	var err error
	content, _ := json.Marshal(entity)
	if s.ch != nil {
		err = s.ch.Publish("",
			s.config.Sync.QueueName,
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        content,
			})
		if err == nil {
			return
		}
		s.Errorf("send message failed: %v", err)
	}
	s.recorder.Error(string(content))
}

func (s *SyncProducer) channel() *Channel {
	conn, err := Dial(s.config.Sync.RabbitMQConnection)
	if err != nil {
		s.Errorf("connect rabbitmq failed, %v", errors.WithStack(err))
		return nil
	}

	channel, err := conn.Channel()
	if err != nil {
		s.Errorf("retrieve channel, %v", errors.WithStack(err))
		return nil
	}
	_, err = channel.QueueDeclare(s.config.Sync.QueueName, true, false, false, false, nil)
	if err != nil {
		s.Errorf("declare queue failed %v", errors.WithStack(err))
		return nil
	}
	return channel
}
