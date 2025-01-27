package service

import (
	"context"

	"github.com/haandol/hexagonal/internal/adapter/secondary/repository"
	"github.com/haandol/hexagonal/internal/dto"
	"github.com/haandol/hexagonal/internal/instrument"
	"github.com/haandol/hexagonal/internal/message/command"
	"github.com/haandol/hexagonal/internal/port/secondaryport/repositoryport"
	"github.com/haandol/hexagonal/pkg/o11y"
	"github.com/haandol/hexagonal/pkg/util"
)

type FlightService struct {
	flightRepository repositoryport.FlightRepository
}

func NewFlightService(
	flightRepository *repository.FlightRepository,
) *FlightService {
	return &FlightService{
		flightRepository: flightRepository,
	}
}

func (s *FlightService) Book(ctx context.Context, cmd *command.BookFlight) error {
	logger := util.LoggerFromContext(ctx).WithGroup("FlightService.Book")

	ctx, span := o11y.BeginSubSpan(ctx, "Book")
	defer span.End()

	req := &dto.FlightBooking{
		TripID:   cmd.Body.TripID,
		FlightID: cmd.Body.FlightID,
	}
	if err := s.flightRepository.Book(ctx, req, cmd); err != nil {
		instrument.RecordBookFlightError(logger, span, err, cmd)

		go func(reason string) {
			if err := s.flightRepository.PublishAbortSaga(ctx,
				cmd.CorrelationID, cmd.ParentID, cmd.Body.TripID, reason,
			); err != nil {
				instrument.RecordPublishAbortSagaError(logger, span, err, cmd)
			}
		}(err.Error())

		return err
	}

	return nil
}

func (s *FlightService) CancelBooking(ctx context.Context, cmd *command.CancelFlightBooking) error {
	logger := util.LoggerFromContext(ctx).WithGroup("FlightService.CancelBooking")

	ctx, span := o11y.BeginSubSpan(ctx, "CancelBooking")
	defer span.End()

	if err := s.flightRepository.CancelBooking(ctx, cmd); err != nil {
		instrument.RecordCancelFlightBookingError(logger, span, err, cmd)
		return err
	}

	return nil
}
