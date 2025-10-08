package repository

import (
	"gorm.io/gorm"

	ba "github.com/1nterdigital/aka-im-wallet/internal/repository/balance_adjustment"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/envelope"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/transfer"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/tx"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet"
	monitoring "github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_monitoring"
	deposit "github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_recharge_request"
	transaction "github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_transaction"
)

type Repository interface {
	Wallet() wallet.Repository
	Envelope() envelope.Repository
	WalletRechargeRequest() deposit.Repository
	WalletTransaction() transaction.Repository
	Transfer() transfer.Repository
	WalletMonitoring() monitoring.WalletMonitoringRepository
	TxRepo() tx.Repository
	BalanceAdjustment() ba.Repository
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Wallet() wallet.Repository {
	return wallet.New(r.db)
}

func (r *repository) Envelope() envelope.Repository {
	return envelope.New(r.db)
}

func (r *repository) WalletRechargeRequest() deposit.Repository {
	return deposit.New(r.db)
}

func (r *repository) WalletTransaction() transaction.Repository {
	return transaction.New(r.db)
}

func (r *repository) Transfer() transfer.Repository {
	return transfer.New(r.db)
}

func (r *repository) WalletMonitoring() monitoring.WalletMonitoringRepository {
	return monitoring.NewWalletMonitoringRepository(r.db)
}

func (r *repository) TxRepo() tx.Repository {
	return tx.NewTx(r.db)
}

func (r *repository) BalanceAdjustment() ba.Repository {
	return ba.New(r.db)
}
