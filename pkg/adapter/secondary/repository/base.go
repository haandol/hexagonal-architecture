package repository

import (
	"context"
	"errors"

	"github.com/haandol/hexagonal/pkg/constant"
	"gorm.io/gorm"
)

type BaseRepository struct {
	DB *gorm.DB
}

func (r BaseRepository) WithContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(constant.TX("tx")).(*gorm.DB); ok {
		return tx
	} else {
		return r.DB.WithContext(ctx)
	}
}

func (r BaseRepository) TXBegin(ctx context.Context) (context.Context, error) {
	if _, ok := ctx.Value(constant.TX("tx")).(*gorm.DB); ok {
		return ctx, errors.New("transaction already exists")
	}

	tx := r.DB.Begin()
	if tx.Error != nil {
		return ctx, tx.Error
	}

	return context.WithValue(ctx, constant.TX("tx"), tx), nil
}

func (r BaseRepository) TXCommit(ctx context.Context) error {
	if tx, ok := ctx.Value(constant.TX("tx")).(*gorm.DB); ok {
		return tx.Commit().Error
	}

	return errors.New("no transaction found")
}

func (r BaseRepository) TXRollback(ctx context.Context) error {
	if tx, ok := ctx.Value(constant.TX("tx")).(*gorm.DB); ok {
		return tx.Rollback().Error
	}

	return errors.New("no transaction found")
}
