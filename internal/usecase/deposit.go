package usecase

import (
	"context"
	"fmt"
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
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_recharge_request"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_transaction"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
	"github.com/1nterdigital/aka-im-wallet/pkg/util/convert"
)

type (
	WalletRechargeRequestSvcImpl struct {
		trx                         *gorm.DB
		transactionUc               WalletTransactionSvc
		repo                        wallet_recharge_request.Repository
		walletTransactionRepository wallet_transaction.Repository
		walletRepository            wallet.Repository
		trxRepo                     tx.Repository
	}

	WalletRechargeRequestSvc interface {
		ProcessDepositByAdmin(
			ctx context.Context, arg *domain.ProcessDepositByAdminRequest,
		) (resp *domain.ProcessDepositByAdminResponse, err error)
		GetListDeposit(
			ctx context.Context, arg *domain.GetListDepositRequest,
		) (resp *domain.GetListDepositResponse, err error)
	}
)

func NewWalletRechargeRequestUseCase(
	trx *gorm.DB,
	transactionUc WalletTransactionSvc,
	repo wallet_recharge_request.Repository,
	walletTransactionRepo wallet_transaction.Repository,
	walletRepo wallet.Repository,
	trxRepo tx.Repository,
) WalletRechargeRequestSvc {
	return &WalletRechargeRequestSvcImpl{
		trx:                         trx,
		transactionUc:               transactionUc,
		repo:                        repo,
		walletTransactionRepository: walletTransactionRepo,
		walletRepository:            walletRepo,
		trxRepo:                     trxRepo,
	}
}

func (s *WalletRechargeRequestSvcImpl) ProcessDepositByAdmin(
	ctx context.Context, arg *domain.ProcessDepositByAdminRequest,
) (resp *domain.ProcessDepositByAdminResponse, err error) {
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

	span.SetAttributes(attribute.String("userID", arg.UserID))

	var targetWallet *entity.Wallet
	targetWallet, err = s.walletRepository.GetWalletByUserID(ctx, arg.UserID)
	if err != nil {
		log.ZError(ctx, "while get wallet target user", err, "UserID", arg.UserID)
		return resp, err
	}

	var depositID int64

	err = s.trxRepo.Do(ctx, func(tx *gorm.DB) error {
		depositID, err = s.repo.CreateDeposit(ctx, tx, &entity.WalletRechargeRequest{
			WalletID:      targetWallet.WalletID,
			Amount:        arg.Amount,
			Description:   arg.Description,
			StatusRequest: entity.StatusRequestApproved,
			ApprovedAt:    convert.PtrTime(time.Now()),
			OperatedBy:    &arg.OperatedBy,
			CreatedBy:     arg.OperatedBy,
			UpdatedBy:     arg.OperatedBy,
			IsActive:      true,
		})
		if err != nil {
			log.ZError(ctx, "while create deposit", err, "request", arg)
			return err
		}

		_, err = s.transactionUc.CreateTransaction(ctx, &domain.CreateTransactionReq{
			WalletID:        targetWallet.WalletID,
			Amount:          arg.Amount,
			ImpactedItem:    depositID,
			TransactionType: string(entity.TransactionTypeDeposit),
			Entrytype:       string(entity.EntryTypeCredit),
			CreatedBy:       arg.OperatedBy,
			ReferenceCode:   fmt.Sprintf("#DEP-%s-%s-%d", nowToStringYYYYMMDD(), arg.UserID, depositID),
			DescriptionEn:   "Deposit user " + arg.UserID,
			DescriptionZh:   "存款用戶 " + arg.UserID,
		})
		if err != nil {
			log.ZError(ctx, "while create transaction", err, "request", arg)
			return err
		}
		span.SetAttributes(
			attribute.Int64("walletID", targetWallet.WalletID),
			attribute.Float64("amount", arg.Amount),
		)

		return nil
	})
	if err != nil {
		log.ZError(ctx, "while process deposit", err, "request", arg)
		return resp, err
	}

	return &domain.ProcessDepositByAdminResponse{
		WalletRechargeRequestID: depositID,
	}, nil
}

func (s *WalletRechargeRequestSvcImpl) GetListDeposit(
	ctx context.Context, arg *domain.GetListDepositRequest,
) (resp *domain.GetListDepositResponse, err error) {
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
		walletUser, err = s.walletRepository.GetWalletByUserID(ctx, arg.UserID)
		if err != nil {
			log.ZError(ctx, "while get wallet target user", err, "UserID", arg.UserID)
			return resp, err
		}

		arg.WalletID = walletUser.WalletID
	}

	if arg.StatusRequest != "" && !entity.StatusRequest(arg.StatusRequest).IsValid() {
		log.ZError(ctx, "while validate status", eerrs.ErrInvalidStatusRequest, "status", arg.StatusRequest)
		return resp, eerrs.ErrInvalidStatusRequest
	}

	var (
		deposits []*entity.WalletRechargeRequest
		total    int64
	)
	deposits, total, err = s.repo.GetListDeposit(ctx, arg)
	if err != nil {
		log.ZError(ctx, "while get list deposit", err, "arg", arg)
		return resp, err
	}

	resp = &domain.GetListDepositResponse{
		TotalCount: total,
		Page:       arg.Page,
		Limit:      arg.Limit,
		Deposits:   dtoDeposits(deposits),
	}

	span.SetAttributes(
		attribute.String("userID", arg.UserID),
		attribute.Int64("walletID", arg.WalletID),
		attribute.Int64("total", resp.TotalCount),
		attribute.Int("amount", int(arg.Page)),
		attribute.Int("amount", int(arg.Limit)),
	)

	return resp, nil
}
