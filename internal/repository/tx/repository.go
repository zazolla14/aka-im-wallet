//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package tx

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Do(ctx context.Context, fn func(tx *gorm.DB) error) error
}
