//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package balance_adjustment

import (
	"context"

	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type Repository interface {
	CreateBalanceAdjustment(
		ctx context.Context, tx *gorm.DB, adjustment *entity.BalanceAdjustments,
	) (balanceAdjustmentID int64, err error)
	GetListBalanceAdjustment(
		ctx context.Context, req *domain.GetListbalanceAjustmentRequest,
	) (adjustments []*entity.BalanceAdjustments, total int64, err error)
}
