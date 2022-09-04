package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/haandol/hexagonal/pkg/dto"
	"github.com/haandol/hexagonal/pkg/message"
	"github.com/haandol/hexagonal/pkg/message/command"
	"github.com/haandol/hexagonal/pkg/message/event"
	"github.com/haandol/hexagonal/pkg/port/secondaryport/producerport"
	"github.com/haandol/hexagonal/pkg/port/secondaryport/repositoryport"
	"github.com/haandol/hexagonal/pkg/util"
)

type FlightService struct {
	producer         producerport.Producer
	flightRepository repositoryport.FlightRepository
}

func NewFlightService(
	producer producerport.Producer,
	flightRepository repositoryport.FlightRepository,
) *FlightService {
	return &FlightService{
		producer:         producer,
		flightRepository: flightRepository,
	}
}

func (s *FlightService) Book(ctx context.Context, cmd *command.BookFlight) error {
	logger := util.GetLogger().With(
		"service", "FlightService",
		"method", "Book",
	)

	logger.Infow("book flight", "command", cmd)

	req := &dto.FlightBooking{
		TripID:   cmd.Body.TripID,
		FlightID: cmd.Body.FlightID,
	}
	boooking, err := s.flightRepository.Book(ctx, req)
	if err != nil {
		logger.Errorf("failed to book flight", "req", req, "err", err.Error())
		return err
	}

	evt := &event.FlightBooked{
		Message: message.Message{
			Name:          "FlightBooked",
			Version:       "1.0.0",
			ID:            uuid.NewString(),
			CorrelationID: cmd.CorrelationID,
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
		Body: event.FlightBookedBody{
			BookingID: boooking.ID,
		},
	}
	v, err := json.Marshal(evt)
	if err != nil {
		logger.Errorf("failed to marshal event", "event", evt, "err", err.Error())
		return err
	}

	if err := s.producer.Produce(ctx, "saga-service", cmd.CorrelationID, v); err != nil {
		return err
	}

	return nil
}

func (s *FlightService) CancelBooking(ctx context.Context, cmd *command.CancelFlightBooking) error {
	logger := util.GetLogger().With(
		"service", "FlightService",
		"method", "CancelBooking",
	)

	logger.Infow("cancel flight booking", "command", cmd)

	if err := s.flightRepository.CancelBooking(ctx, cmd.Body.BookingID); err != nil {
		logger.Errorf("failed to cancel flight booking", "BookingID", cmd.Body.BookingID, "err", err.Error())
		return err
	}

	evt := &event.FlightBookingCanceled{
		Message: message.Message{
			Name:          "FlightBookingCanceled",
			Version:       "1.0.0",
			ID:            uuid.NewString(),
			CorrelationID: cmd.CorrelationID,
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
		Body: event.FlightBookedBody{
			BookingID: cmd.Body.BookingID,
		},
	}
	v, err := json.Marshal(evt)
	if err != nil {
		logger.Errorf("failed to marshal event", "event", evt, "err", err.Error())
		return err
	}

	if err := s.producer.Produce(ctx, "saga-service", cmd.CorrelationID, v); err != nil {
		return err
	}

	return nil
}