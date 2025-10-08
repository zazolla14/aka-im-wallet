package usecase

import (
	"time"

	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

func dtoWalletTransactions(dbs []*entity.WalletTransaction) (transactions []*domain.WalletTransaction) {
	for _, db := range dbs {
		transactions = append(transactions, &domain.WalletTransaction{
			WalletTransactionID: db.WalletTransactionID,
			WalletID:            db.WalletID,
			Amount:              db.Amount,
			TransactionType:     string(db.TransactionType),
			EntryType:           string(db.EntryType),
			BeforeBalance:       db.BeforeBalance,
			AfterBalance:        db.AfterBalance,
			ReferenceCode:       db.ReferenceCode,
			DescriptionEn:       db.DescriptionEN,
			DescriptionZh:       db.DescriptionZH,
			ImpactedItem:        db.ImpactedItem,
			IsShown:             db.IsShown,
			TransactionDate:     db.TransactionDate,
		})
	}

	return transactions
}

func dtoWalletDetail(dbs *entity.Wallet) *domain.Wallet {
	return &domain.Wallet{
		ID:        dbs.WalletID,
		UserID:    dbs.UserID,
		Balance:   dbs.Balance,
		CreatedAt: dbs.CreatedAt,
		CreatedBy: dbs.CreatedBy,
	}
}

func nowToStringYYYYMMDD() string {
	return time.Now().Format("20060102")
}

func dtoDeposits(dbs []*entity.WalletRechargeRequest) (deposits []*domain.Deposit) {
	for _, db := range dbs {
		deposits = append(deposits, &domain.Deposit{
			DepositID:     db.WalletRechargeRequestID,
			WalletID:      db.WalletID,
			Amount:        db.Amount,
			StatusRequest: string(db.StatusRequest),
			Description:   db.Description,
			CreatedAt:     db.CreatedAt,
			ApprovedAt:    db.ApprovedAt,
			OperatedBy:    db.OperatedBy,
			CreatedBy:     db.CreatedBy,
		})
	}

	return deposits
}

func dtoBalanceAdjustments(dbs []*entity.BalanceAdjustments) (adjustments []*domain.BalanceAdjustment) {
	for _, db := range dbs {
		adjustments = append(adjustments, &domain.BalanceAdjustment{
			BalanceAdjustmentID: db.BalanceAdjustmentID,
			WalletID:            db.WalletID,
			Amount:              db.Amount,
			Reason:              db.Reason,
			Description:         db.Description,
			CreatedAt:           db.CreatedAt,
			CreatedBy:           db.CreatedBy,
			UpdatedAt:           db.UpdatedAt,
			UpdatedBy:           db.UpdatedBy,
		})
	}

	return adjustments
}

func dtoTransfer(db *entity.Transfer) (transfer *domain.Transfer) {
	return &domain.Transfer{
		TransferID:     db.TransferID,
		FromUserID:     db.FromUserID,
		ToUserID:       db.ToUserID,
		Amount:         db.Amount,
		StatusTransfer: string(db.StatusTransfer),
		Remark:         db.Remark,
		ExpiredAt:      db.ExpiredAt,
		RefundedAt:     db.RefundedAt,
		ClaimedAt:      db.ClaimedAt,
		CreatedAt:      db.CreatedAt,
		CreatedBy:      db.CreatedBy,
		UpdatedAt:      db.UpdatedAt,
		UpdatedBy:      db.UpdatedBy,
	}
}
