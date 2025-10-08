//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package transfer

import (
	"context"
	"time"

	"gorm.io/gorm"

	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type Repository interface {
	Create(transfer *entity.Transfer) error
	Update(ctx context.Context, transfer *entity.Transfer, tx *gorm.DB) error
	FindAllTransferByWalletID(walletID string) ([]entity.Transfer, error)
	FindByTransferID(ctx context.Context, transferID int64) (*entity.Transfer, error)
	CreateTransfer(
		ctx context.Context, tx *gorm.DB, transfer *entity.Transfer,
	) (transferID int64, err error)
	GetEligibleClaimTransfer(
		ctx context.Context, transferID int64, claimerUserID string,
	) (detail *entity.Transfer, err error)
	CountSentTransferInDay(
		ctx context.Context, userID string, now time.Time,
	) (count int64, err error)
	CountClaimedTransferInDay(
		ctx context.Context, userID string, now time.Time,
	) (total int64, err error)
	FetchExpiredTransfers(ctx context.Context, ids []int64) ([]*entity.Transfer, error)
	GetEligibleRefundTransfer(
		ctx context.Context, transferID int64, userID string,
	) (detail *entity.Transfer, err error)
	GetDetailTransfer(
		ctx context.Context, transferID int64, userID string,
	) (detail *entity.Transfer, err error)
}
