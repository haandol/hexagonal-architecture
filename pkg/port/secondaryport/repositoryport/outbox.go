package repositoryport

import (
	"context"

	"github.com/haandol/hexagonal/pkg/dto"
)

type OutboxRepository interface {
	QueryUnsent(ctx context.Context, batchSize int) ([]dto.Outbox, error)
	MarkSentInBatch(ctx context.Context, ids []uint) error
}
