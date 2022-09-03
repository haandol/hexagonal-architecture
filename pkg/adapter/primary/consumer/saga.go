package consumer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/haandol/hexagonal/pkg/message"
	"github.com/haandol/hexagonal/pkg/message/command"
	"github.com/haandol/hexagonal/pkg/port/primaryport/consumerport"
	"github.com/haandol/hexagonal/pkg/service"
	"github.com/haandol/hexagonal/pkg/util"
)

type SagaConsumer struct {
	*KafkaConsumer
	sagaService *service.SagaService
}

func NewSagaConsumer(
	kafkaConsumer *KafkaConsumer,
	sagaService *service.SagaService,
) *SagaConsumer {
	return &SagaConsumer{
		KafkaConsumer: kafkaConsumer,
		sagaService:   sagaService,
	}
}

func (c *SagaConsumer) Init() {
	logger := util.GetLogger().With(
		"module", "SagaConsumer",
		"func", "Init",
	)

	if err := c.RegisterHandler(c.Handle); err != nil {
		logger.Fatalw("Failed to register handler", "err", err.Error())
	}
}

func (c *SagaConsumer) Handle(ctx context.Context, r *consumerport.Message) error {
	logger := util.GetLogger().With(
		"module", "SagaConsumer",
		"func", "Handle",
	)

	msg := &message.Message{}
	if err := json.Unmarshal(r.Value, msg); err != nil {
		logger.Errorw("Failed to unmarshal command", "err", err.Error())
	}

	logger.Infow("Received command", "command", msg)

	switch msg.Name {
	case "StartSaga":
		cmd := &command.StartSaga{}
		if err := json.Unmarshal(r.Value, cmd); err != nil {
			logger.Errorw("Failed to unmarshal command", "err", err.Error())
		}
		return c.sagaService.Start(ctx, cmd)
	case "EndSaga":
		cmd := &command.EndSaga{}
		if err := json.Unmarshal(r.Value, cmd); err != nil {
			logger.Errorw("Failed to unmarshal command", "err", err.Error())
		}
		return c.sagaService.End(ctx, cmd)
	case "AbortSaga":
		cmd := &command.AbortSaga{}
		if err := json.Unmarshal(r.Value, cmd); err != nil {
			logger.Errorw("Failed to unmarshal command", "err", err.Error())
		}
		return c.sagaService.Abort(ctx, cmd)
	default:
		logger.Errorw("unknown command", "message", msg)
		return errors.New("unknown command")
	}
}
