//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package wallet_transaction

import (
	"context"

	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type Repository interface {
	FindAllByWalletID(walletID int64) ([]entity.WalletTransaction, error)
	FindByWalletTransactionID(walletTransactionID int64) (entity.WalletTransaction, error)
	CreateTransaction(
		ctx context.Context, tx *gorm.DB, transaction *entity.WalletTransaction,
	) (transactionID int64, err error)
	GetListTransaction(
		ctx context.Context, req *domain.GetListTransactionRequest,
	) (trans []*entity.WalletTransaction, count int64, err error)
}
