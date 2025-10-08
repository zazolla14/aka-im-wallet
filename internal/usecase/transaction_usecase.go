//go:generate mockgen -source=$GOFILE -destination=$PROJECT_DIR/generated/mock/mock_$GOPACKAGE/$GOFILE

package usecase

import (
	"context"
	"math"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/tx"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_transaction"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

type (
	WalletTransactionSvcImpl struct {
		repo       wallet_transaction.Repository
		walletRepo wallet.Repository
		txRepo     tx.Repository
	}

	WalletTransactionSvc interface {
		CreateTransaction(
			ctx context.Context, req *domain.CreateTransactionReq,
		) (transactionID int64, err error)
		GetListTransaction(
			ctx context.Context, req *domain.GetListTransactionRequest,
		) (resp *domain.GetListTransactionResponse, err error)
	}
)

func NewWalletTransactionUseCase(
	repo wallet_transaction.Repository,
	walletRepo wallet.Repository,
	txRepo tx.Repository,
) WalletTransactionSvc {
	return &WalletTransactionSvcImpl{
		repo:       repo,
		walletRepo: walletRepo,
		txRepo:     txRepo,
	}
}

func (u *WalletTransactionSvcImpl) CreateTransaction(
	ctx context.Context, req *domain.CreateTransactionReq,
) (transactionID int64, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelUsecase)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	if _, err = req.IsValid(); err != nil {
		return transactionID, err
	}

	errTrx := u.txRepo.Do(ctx, func(tx *gorm.DB) error {
		var wallet *entity.Wallet
		wallet, err = u.walletRepo.GetWalletByWalletIDTx(ctx, tx, req.WalletID)
		if err != nil {
			log.ZError(ctx, "while GetWalletByWalletIDTx", err)
			return err
		}

		if req.Entrytype == entity.EntryTypeDebit.String() && math.Abs(req.Amount) > wallet.Balance {
			log.ZError(ctx, "while validate balance user", eerrs.ErrInsufficientBalance)
			return eerrs.ErrInsufficientBalance
		}

		if err = u.walletRepo.UpdateWallet(ctx, tx, &entity.Wallet{
			WalletID:  req.WalletID,
			Balance:   wallet.Balance + req.Amount,
			UpdatedAt: time.Now(),
		}); err != nil {
			log.ZError(ctx, "while update wallet based on transaction", err)
			return err
		}

		transactionID, err = u.repo.CreateTransaction(ctx, tx, &entity.WalletTransaction{
			WalletID:        req.WalletID,
			Amount:          req.Amount,
			TransactionType: entity.TransactionType(req.TransactionType),
			EntryType:       entity.EntryType(req.Entrytype),
			BeforeBalance:   wallet.Balance,
			AfterBalance:    wallet.Balance + req.Amount,
			ReferenceCode:   req.ReferenceCode,
			ImpactedItem:    req.ImpactedItem,
			DescriptionEN:   req.DescriptionEn,
			DescriptionZH:   req.DescriptionZh,
			TransactionDate: time.Now(),
			CreatedBy:       req.CreatedBy,
			IsShown:         true,
			IsActive:        true,
		})
		span.SetAttributes(
			attribute.Int64("walletID", req.WalletID),
			attribute.Float64("amount", req.Amount),
			attribute.String("transactionType", req.TransactionType),
		)

		if err != nil {
			log.ZError(ctx, "while create wallet transaction", err)
			return err
		}

		return nil
	})
	if errTrx != nil {
		log.ZError(ctx, "while do transaction", errTrx)
		err = errTrx
		return transactionID, err
	}

	return transactionID, nil
}

func (u *WalletTransactionSvcImpl) GetListTransaction(
	ctx context.Context, req *domain.GetListTransactionRequest,
) (resp *domain.GetListTransactionResponse, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelUsecase)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	var wallets *entity.Wallet
	wallets, err = u.walletRepo.GetWalletByUserID(ctx, req.UserID)
	if err != nil {
		log.ZError(ctx, "while walletRepo.GetWalletByUserID", err)
		return resp, err
	}
	req.WalletID = wallets.WalletID

	var (
		transactions []*entity.WalletTransaction
		total        int64
	)
	transactions, total, err = u.repo.GetListTransaction(ctx, req)
	if err != nil {
		log.ZError(ctx, "while repo.GetListTransaction", err)
		return resp, err
	}

	resp = &domain.GetListTransactionResponse{
		Page:         req.Page,
		Limit:        req.Limit,
		TotalCount:   total,
		Transactions: dtoWalletTransactions(transactions),
	}

	span.SetAttributes(
		attribute.String("walletID", req.UserID),
		attribute.Int64("walletID", req.WalletID),
		attribute.Int("page", int(req.Page)),
		attribute.Int("limit", int(req.Limit)),
		attribute.Int64("total", total),
	)

	return resp, nil
}
