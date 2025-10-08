//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package wallet

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type Repository interface {
	FindByUserID(userID string) (resp *entity.Wallet, err error)
	Create(wallet *entity.Wallet) (err error)
	FindByWalletID(walletID int64) (resp *entity.Wallet, err error)
	GetWalletByWalletIDTx(
		ctx context.Context, tx *gorm.DB, walletID int64,
	) (wallet *entity.Wallet, err error)
	UpdateWallet(ctx context.Context, tx *gorm.DB, wallet *entity.Wallet) (err error)
	GetWalletByUserID(ctx context.Context, userID string) (wallet *entity.Wallet, err error)
	CreateWallet(
		ctx context.Context, wallet *entity.Wallet,
	) (walletID int64, err error)
}
