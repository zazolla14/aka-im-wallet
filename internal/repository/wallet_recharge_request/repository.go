//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package wallet_recharge_request

import (
	"context"

	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type Repository interface {
	CreateDeposit(
		ctx context.Context, tx *gorm.DB, deposit *entity.WalletRechargeRequest,
	) (depositID int64, err error)
	GetListDeposit(
		ctx context.Context, arg *domain.GetListDepositRequest,
	) (deposit []*entity.WalletRechargeRequest, total int64, err error)
}
