package usecase

import (
	"context"
	"fmt"
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
	ba "github.com/1nterdigital/aka-im-wallet/internal/repository/balance_adjustment"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/tx"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet"
)

type (
	BalanceAdjustmentSvcImpl struct {
		transactionUc WalletTransactionSvc
		repo          ba.Repository
		walletRepo    wallet.Repository
		trxRepo       tx.Repository
	}

	BalanceAdjustmentSvc interface {
		BalanceAdjustmentByAdmin(
			ctx context.Context, arg *domain.BalanceAdjustmentByAdminRequest,
		) (resp *domain.BalanceAdjustmentByAdminResponse, err error)
		GetListBalanceAdjustment(
			ctx context.Context, arg *domain.GetListbalanceAjustmentRequest,
		) (resp *domain.GetListBalanceAdjustmentResponse, err error)
	}
)

func NewBalanceAdjustmentUseCase(
	transactionUc WalletTransactionSvc,
	repo ba.Repository,
	walletRepo wallet.Repository,
	trxRepo tx.Repository,
) BalanceAdjustmentSvc {
	return &BalanceAdjustmentSvcImpl{
		transactionUc: transactionUc,
		repo:          repo,
		walletRepo:    walletRepo,
		trxRepo:       trxRepo,
	}
}

func (s *BalanceAdjustmentSvcImpl) BalanceAdjustmentByAdmin(
	ctx context.Context, arg *domain.BalanceAdjustmentByAdminRequest,
) (resp *domain.BalanceAdjustmentByAdminResponse, err error) {
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

	span.SetAttributes(attribute.String("userIDReq", arg.UserID))

	var targetWallet *entity.Wallet
	targetWallet, err = s.walletRepo.GetWalletByUserID(ctx, arg.UserID)
	if err != nil {
		log.ZError(ctx, "while get wallet target user", err, "UserID", arg.UserID)
		return resp, err
	}

	if arg.Amount < 0 {
		if targetWallet.Balance < math.Abs(arg.Amount) {
			log.ZError(ctx, "while validate balance user", err)
			return resp, fmt.Errorf("insufficient balance user %s", arg.UserID)
		}
	}

	var balanceAdjustmentID int64

	err = s.trxRepo.Do(ctx, func(tx *gorm.DB) error {
		balanceAdjustmentID, err = s.repo.CreateBalanceAdjustment(ctx, tx, &entity.BalanceAdjustments{
			WalletID:    targetWallet.WalletID,
			Amount:      arg.Amount,
			Reason:      arg.Reason,
			Description: arg.Description,
			CreatedAt:   time.Now(),
			CreatedBy:   arg.OperatedBy,
			UpdatedAt:   time.Now(),
			UpdatedBy:   arg.OperatedBy,
			IsActive:    true,
		})
		if err != nil {
			log.ZError(ctx, "while create balance adjustments", err, "request", arg)
			return err
		}

		entryType := entity.EntryTypeCredit
		if arg.Amount < 0 {
			entryType = entity.EntryTypeDebit
		}

		_, err = s.transactionUc.CreateTransaction(ctx, &domain.CreateTransactionReq{
			WalletID:        targetWallet.WalletID,
			Amount:          arg.Amount,
			ImpactedItem:    balanceAdjustmentID,
			TransactionType: string(entity.TransactionTypeSystemAdjustment),
			Entrytype:       string(entryType),
			CreatedBy:       arg.OperatedBy,
			ReferenceCode:   fmt.Sprintf("#BA-%s-%s-%d", nowToStringYYYYMMDD(), arg.UserID, balanceAdjustmentID),
			DescriptionEn:   "Balance Adjustment User" + fmt.Sprintf(" %s %s", arg.UserID, arg.Reason),
			DescriptionZh:   "餘額調整用戶" + fmt.Sprintf(" %s %s", arg.UserID, arg.Reason),
		})
		if err != nil {
			log.ZError(ctx, "while create transaction", err, "request", arg)
			return err
		}

		return nil
	})
	if err != nil {
		log.ZError(ctx, "while create balance adjustments", err, "request", arg)
		return resp, err
	}

	span.SetAttributes(
		attribute.Int64("walletIDResp", targetWallet.WalletID),
		attribute.Float64("amountResp", arg.Amount),
	)

	return &domain.BalanceAdjustmentByAdminResponse{
		BalanceAdjustmentID: balanceAdjustmentID,
	}, nil
}

func (s *BalanceAdjustmentSvcImpl) GetListBalanceAdjustment(
	ctx context.Context, arg *domain.GetListbalanceAjustmentRequest,
) (resp *domain.GetListBalanceAdjustmentResponse, err error) {
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

	if arg.UserID != "" {
		var walletUser *entity.Wallet
		walletUser, err = s.walletRepo.GetWalletByUserID(ctx, arg.UserID)
		if err != nil {
			log.ZError(ctx, "while get wallet target user", err, "UserID", arg.UserID)
			return resp, err
		}

		arg.WalletID = walletUser.WalletID
	}

	var (
		adjustments []*entity.BalanceAdjustments
		total       int64
	)
	adjustments, total, err = s.repo.GetListBalanceAdjustment(ctx, arg)
	if err != nil {
		log.ZError(ctx, "while get list adjustments", err, "arg", arg)
		return resp, err
	}

	resp = &domain.GetListBalanceAdjustmentResponse{
		TotalCount:         total,
		Page:               arg.Page,
		Limit:              arg.Limit,
		BalanceAdjustments: dtoBalanceAdjustments(adjustments),
	}

	span.SetAttributes(
		attribute.String("userID", arg.UserID),
		attribute.Int64("walletID", arg.WalletID),
		attribute.Int("page", int(arg.Page)),
		attribute.Int("int", int(arg.Limit)),
		attribute.Int64("total", total),
	)

	return resp, nil
}
