package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/haandol/hexagonal/pkg/dto"
	"github.com/haandol/hexagonal/pkg/message"
	"github.com/haandol/hexagonal/pkg/message/command"
	"github.com/haandol/hexagonal/pkg/message/event"
	"github.com/haandol/hexagonal/pkg/util"
)

type CarProducer struct {
	*KafkaProducer
}

func NewCarProducer(kafkaProducer *KafkaProducer) *CarProducer {
	return &CarProducer{
		KafkaProducer: kafkaProducer,
	}
}

func (p *CarProducer) PublishCarBooked(ctx context.Context, corrID string, d dto.CarBooking) error {
	evt := &event.CarBooked{
		Message: message.Message{
			Name:          "CarBooked",
			Version:       "1.0.0",
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
		Body: event.CarBookedBody{
			BookingID: d.ID,
		},
	}
	v, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	if err := p.Produce(ctx, "saga-service", corrID, v); err != nil {
		return err
	}

	return nil
}

func (p *CarProducer) PublishCarBookingCanceled(ctx context.Context, corrID string, d dto.CarBooking) error {
	evt := &event.CarBookingCanceled{
		Message: message.Message{
			Name:          "CarBookingCanceled",
			Version:       "1.0.0",
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
		Body: event.CarBookedBody{
			BookingID: d.ID,
		},
	}
	v, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	if err := p.Produce(ctx, "saga-service", corrID, v); err != nil {
		return err
	}

	return nil
}

func (p *CarProducer) PublishAbortSaga(ctx context.Context, corrID string, tripID uint, reason string) error {
	cmd := &command.AbortSaga{
		Message: message.Message{
			Name:          "AbortSaga",
			Version:       "1.0.0",
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
		Body: command.AbortSagaBody{
			TripID: tripID,
			Reason: reason,
			Source: "car",
		},
	}
	if err := util.ValidateStruct(cmd); err != nil {
		return err
	}
	v, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	if err := p.Produce(ctx, "saga-service", corrID, v); err != nil {
		return err
	}

	return nil
}
