package rabbitmq

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rafaeljesus/rabbus"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
)

type RabbitMQ struct {
	config *model.Config
	common.Logger
	r *rabbus.Rabbus
}

func New(config *model.Config, logger common.Logger) *RabbitMQ {
	return &RabbitMQ{
		config: config,
		Logger: logger,
	}
}
func (s *RabbitMQ) connect(delay time.Duration) {
	if delay != time.Second {
		if delay >= time.Second*120 {
			delay = time.Second
		}
		delayTimer := time.After(delay)
		select {
		case <-delayTimer:
			// delay end
		}
	}
	timeout := time.After(time.Second * 3)
	cbStateChangeFunc := func(name, from, to string) {
		// do something when state is changed
	}
	r, err := rabbus.New(
		s.config.Sync.RabbitMQConnection,
		rabbus.Durable(true),
		rabbus.Attempts(5),
		rabbus.Sleep(time.Second*2),
		rabbus.Threshold(3),
		rabbus.OnStateChange(cbStateChangeFunc),
	)
	if err != nil {
		s.Errorf("connection rabbitmq failed: %v", errors.WithStack(err))
		go s.connect(delay * 2)
		return
	}
	s.r = r

	ctx := context.Background()

	go func() {
		err := r.Run(ctx)
		if err != nil {
			s.Errorf("run rabbitmq failed: %v", errors.WithStack(err))
		}
	}()

	go func() {
		for {
			select {
			case <-r.EmitOk():
			case err := <-r.EmitErr():
				s.Errorf("emit message failed: %v", errors.WithStack(err))
			case <-timeout:
				s.Debug("loop timeout")
			}
		}
	}()
}
