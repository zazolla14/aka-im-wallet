package tx

import (
	"context"

	"gorm.io/gorm"
)

type repositoryImpl struct {
	db *gorm.DB
}

func NewTx(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Do(
	ctx context.Context, fn func(tx *gorm.DB) error,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
