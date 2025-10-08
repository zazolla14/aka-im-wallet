package usecase

import (
	"context"
	"errors"
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
	"github.com/1nterdigital/aka-im-wallet/internal/repository/transfer"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/tx"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/kafka"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
	"github.com/1nterdigital/aka-im-wallet/pkg/helper"
	"github.com/1nterdigital/aka-im-wallet/pkg/util/convert"
)

type (
	TransferSvcImpl struct {
		expiredTransferPublisher *kafka.Producer
		transactionUc            WalletTransactionSvc
		repo                     transfer.Repository
		walletRepo               wallet.Repository
		txRepo                   tx.Repository
	}

	TransferSvc interface {
		CreateTransfer(
			ctx context.Context, arg *domain.CreateTransferRequest,
		) (resp *domain.CreateTransferResponse, err error)
		validateTransfer(
			ctx context.Context, arg *domain.CreateTransferRequest,
		) (sourceWalletID int64, err error)
		ProcessExpiredTransfers(
			ctx context.Context, transfers []*domain.MsgKafkaExpiredTransfer,
		) (failedRefund []*domain.MsgKafkaExpiredTransfer, err error)
		ClaimTransfer(ctx context.Context, arg *domain.ClaimTransferRequest) (err error)
		FetchExpiredTransfers(
			ctx context.Context, transferIDs []int64,
		) ([]*domain.MsgKafkaExpiredTransfer, error)
		ProcessManualRefund(ctx context.Context, transferIDs []int64) error
		RefundTransfer(ctx context.Context, arg *domain.RefundTransferReq) (err error)
		GetDetailTransfer(
			ctx context.Context, transferID int64, userID string,
		) (tranfer *domain.Transfer, err error)
	}
)

func NewTransferUseCase(
	expiredTransferPublisher *kafka.Producer,
	transactionUc WalletTransactionSvc,
	repo transfer.Repository,
	walletRepo wallet.Repository,
	txRepo tx.Repository,
) TransferSvc {
	return &TransferSvcImpl{
		expiredTransferPublisher: expiredTransferPublisher,
		transactionUc:            transactionUc,
		repo:                     repo,
		walletRepo:               walletRepo,
		txRepo:                   txRepo,
	}
}

func (s *TransferSvcImpl) CreateTransfer(
	ctx context.Context, arg *domain.CreateTransferRequest,
) (resp *domain.CreateTransferResponse, err error) {
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

	var sourceWalletID int64
	sourceWalletID, err = s.validateTransfer(ctx, arg)
	if err != nil {
		log.ZError(ctx, "while validate transfer", err, "request", arg)
		return resp, err
	}

	dayDuration := time.Hour * 24
	expiredAt := time.Now().Add(dayDuration)

	var transferID int64
	errTrx := s.txRepo.Do(ctx, func(tx *gorm.DB) error {
		transferID, err = s.repo.CreateTransfer(ctx, tx, &entity.Transfer{
			FromUserID:     arg.FromUserID,
			ToUserID:       arg.ToUserID,
			Amount:         arg.Amount,
			StatusTransfer: entity.StatusTransferPending,
			ExpiredAt:      &expiredAt,
			Remark:         arg.Remark,
			IsActive:       true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})

		span.SetAttributes(
			attribute.String("fromUserID", arg.FromUserID),
			attribute.String("toUserID", arg.ToUserID),
			attribute.Float64("amount", arg.Amount),
			attribute.String("createdAt", time.Now().Format(time.RFC3339)),
		)

		if err != nil {
			log.ZError(ctx, "while create transfer", err, "request", arg)
			return err
		}

		_, err = s.transactionUc.CreateTransaction(ctx, &domain.CreateTransactionReq{
			WalletID:        sourceWalletID,
			Amount:          -1 * arg.Amount,
			Entrytype:       entity.EntryTypeDebit.String(),
			TransactionType: entity.TransactionTypeTransfer.String(),
			ReferenceCode:   fmt.Sprintf("#TRF-%s-%s-%d", nowToStringYYYYMMDD(), arg.FromUserID, transferID),
			ImpactedItem:    transferID,
			DescriptionEn:   "Transfer to " + arg.ToUserID,
			DescriptionZh:   "转账至 " + arg.ToUserID, // need to double check
			CreatedBy:       arg.FromUserID,
		})
		if err != nil {
			log.ZError(ctx, "while create transaction", err, "request", arg)
			return err
		}

		span.SetAttributes(
			attribute.Int64("sourceWalletID", sourceWalletID),
		)

		return nil
	})
	if errTrx != nil {
		log.ZError(ctx, "while txRepo.Do", err, "request", arg)
		err = errTrx
		return resp, err
	}

	return &domain.CreateTransferResponse{
		TransferID: transferID,
	}, nil
}

func (s *TransferSvcImpl) validateTransfer(
	ctx context.Context, arg *domain.CreateTransferRequest,
) (sourceWalletID int64, err error) {
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

	var sourceWallet *entity.Wallet
	sourceWallet, err = s.walletRepo.GetWalletByUserID(ctx, arg.FromUserID)
	if err != nil {
		log.ZError(ctx, "while get wallet source user", err, "FromUserID", arg.FromUserID)
		return sourceWalletID, eerrs.ErrWalletNotFound
	}

	var countToday int64
	countToday, err = s.repo.CountSentTransferInDay(ctx, arg.FromUserID, time.Now())
	if err != nil {
		log.ZError(ctx, "while countSentTransferInDay", err, "FromUserID", arg.FromUserID)
		return sourceWalletID, err
	}
	if countToday >= entity.MaxTransferSendPerDay {
		return sourceWalletID, eerrs.ErrSendingTransferDailyLimit
	}

	if sourceWallet.Balance < arg.Amount {
		log.ZError(ctx, "while validate balance source user", eerrs.ErrInsufficientBalance, "FromUserID", arg.FromUserID)
		return sourceWalletID, eerrs.ErrInsufficientBalance
	}

	_, err = s.walletRepo.GetWalletByUserID(ctx, arg.ToUserID)
	if err != nil {
		log.ZError(ctx, "while get wallet target user", err, "ToUserID", arg.ToUserID)
		return sourceWalletID, eerrs.ErrReceiverWalletNotFound
	}

	span.SetAttributes(attribute.Int64("walletID", sourceWallet.WalletID))

	return sourceWallet.WalletID, nil
}

func (s *TransferSvcImpl) ProcessExpiredTransfers(
	ctx context.Context, transfers []*domain.MsgKafkaExpiredTransfer,
) (failedRefund []*domain.MsgKafkaExpiredTransfer, err error) {
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

	retryAttemptsProcess := 4

	for idx := range transfers {
		operatedBy := transfers[idx].OperatedBy
		transfers[idx].Counter += 1
		if transfers[idx].Counter >= retryAttemptsProcess {
			continue
		}

		var transferDetail *entity.Transfer
		transferDetail, err = s.repo.FindByTransferID(ctx, transfers[idx].TransferID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.ZError(ctx, "while FindByTransferID", err,
					"transferID", transfers[idx].TransferID,
					"operatedBy", operatedBy,
				)
				failedRefund = append(failedRefund, transfers[idx])
			}
			continue
		}
		span.SetAttributes(attribute.Int64("transferID", transfers[idx].TransferID))

		var walletUser *entity.Wallet
		walletUser, err = s.walletRepo.GetWalletByUserID(ctx, transferDetail.FromUserID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.ZWarn(ctx, "while GetWalletByUserID", err,
					"userID", transferDetail.FromUserID,
					"operatedBy", operatedBy,
				)
				failedRefund = append(failedRefund, transfers[idx])
			}
			continue
		}

		span.SetAttributes(attribute.Int64("walletID", walletUser.WalletID))
		span.SetAttributes(attribute.String("userID", walletUser.UserID))

		if transferDetail.RefundedAt != nil {
			continue
		}

		txErr := s.txRepo.Do(ctx, func(tx *gorm.DB) error {
			var errx error
			refundedAt := time.Now()
			transferDetail.RefundedAt = &refundedAt
			transferDetail.UpdatedAt = refundedAt
			transferDetail.UpdatedBy = operatedBy
			transferDetail.StatusTransfer = entity.StatusTransferRefunded
			errx = s.repo.Update(ctx, transferDetail, tx)
			if errx != nil {
				log.ZError(ctx, "while Update transfer", errx,
					"userID", transferDetail.FromUserID,
					"transferID", transferDetail.TransferID,
					"operatedBy", operatedBy,
				)
				return errx
			}

			_, errx = s.transactionUc.CreateTransaction(ctx, &domain.CreateTransactionReq{
				WalletID:        walletUser.WalletID,
				ImpactedItem:    transferDetail.TransferID,
				TransactionType: string(entity.TransactionTypeRefundTransfer),
				Entrytype:       string(entity.EntryTypeDebit),
				Amount:          transferDetail.Amount,
				DescriptionEn:   fmt.Sprintf(domain.TransactionRefundTransferEN, transferDetail.TransferID),
				DescriptionZh:   fmt.Sprintf(domain.TransactionRefundTransferZN, transferDetail.TransferID),
				ReferenceCode:   fmt.Sprintf(domain.FormatReferenceCodeRefundTransfer, transferDetail.TransferID, walletUser.WalletID),
				CreatedBy:       operatedBy,
			})
			if errx != nil {
				log.ZError(ctx, "while CreateTransaction", errx,
					"userID", transferDetail.FromUserID,
					"transferID", transferDetail.TransferID,
				)
				return errx
			}
			span.SetAttributes(attribute.Float64("amount", transferDetail.Amount))

			return nil
		})
		if txErr != nil {
			log.ZError(ctx, "while do trx refund transfer", txErr)
			failedRefund = append(failedRefund, transfers[idx])
			err = txErr
			continue
		}
	}

	return failedRefund, nil
}

func (s *TransferSvcImpl) validateClaimTransfer(
	ctx context.Context, userID string,
) (err error) {
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

	span.SetAttributes(attribute.String("userID", userID))

	var claimCountToday int64
	claimCountToday, err = s.repo.CountClaimedTransferInDay(ctx, userID, time.Now())
	if err != nil {
		log.ZError(ctx, "while get countClaimedTransferInDay", err, "claimerUserID", userID)
		return err
	}
	if claimCountToday > entity.MaxTransferClaimPerDay {
		return eerrs.ErrClaimTransferDailyLimit
	}
	return nil
}

func (s *TransferSvcImpl) ClaimTransfer(
	ctx context.Context, arg *domain.ClaimTransferRequest,
) (err error) {
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

	var claimerWallet *entity.Wallet
	claimerWallet, err = s.walletRepo.GetWalletByUserID(ctx, arg.ClaimerUserID)
	if err != nil {
		log.ZError(ctx, "while get wallet claimer user", err, "claimerUserID", arg.ClaimerUserID)
		return eerrs.ErrWalletNotFound
	}

	err = s.validateClaimTransfer(ctx, arg.ClaimerUserID)
	if err != nil {
		log.ZError(ctx, "while validate claim transfer", err, "claimerUserID", arg.ClaimerUserID)
		return err
	}

	var transferDetail *entity.Transfer
	transferDetail, err = s.repo.GetEligibleClaimTransfer(ctx, arg.TransferID, arg.ClaimerUserID)
	if err != nil {
		log.ZError(ctx, "while get eligible claim transfer", err, "transferID", arg.TransferID)
		return eerrs.ErrNoEligibleTransfer
	}

	errTrx := s.txRepo.Do(ctx, func(tx *gorm.DB) error {
		claimAt := time.Now()
		transferDetail.ClaimedAt = &claimAt
		transferDetail.UpdatedAt = claimAt
		transferDetail.UpdatedBy = arg.OperateBy
		transferDetail.StatusTransfer = entity.StatusTransferClaimed

		err = s.repo.Update(ctx, transferDetail, tx)
		if err != nil {
			log.ZError(ctx, "while update transfer", err, "transferID", arg.TransferID)
			return err
		}

		_, err = s.transactionUc.CreateTransaction(ctx, &domain.CreateTransactionReq{
			WalletID:        claimerWallet.WalletID,
			ImpactedItem:    transferDetail.TransferID,
			TransactionType: string(entity.TransactionTypeTransfer),
			Entrytype:       string(entity.EntryTypeCredit),
			Amount:          transferDetail.Amount,
			DescriptionEn:   "Claim transfer from " + transferDetail.FromUserID,
			DescriptionZh:   "领取转账 " + transferDetail.FromUserID,
			ReferenceCode:   fmt.Sprintf("#TRF-%s-%s-%d", nowToStringYYYYMMDD(), arg.OperateBy, transferDetail.TransferID),
			CreatedBy:       arg.OperateBy,
		})
		if err != nil {
			log.ZError(ctx, "while create transaction", err, "transferID", arg.TransferID)
			return err
		}
		span.SetAttributes(
			attribute.Int64("walletID", claimerWallet.WalletID),
			attribute.Float64("amount", transferDetail.Amount),
		)

		return nil
	})
	if errTrx != nil {
		log.ZError(ctx, "while do trx claim transfer", errTrx)
		err = errTrx
		return err
	}

	return nil
}

func (s *TransferSvcImpl) FetchExpiredTransfers(
	ctx context.Context, transferIDs []int64,
) ([]*domain.MsgKafkaExpiredTransfer, error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelUsecase)
		err      error
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	span.SetAttributes(attribute.Int64Slice("transferIDs", transferIDs))
	var transfers []*entity.Transfer
	transfers, err = s.repo.FetchExpiredTransfers(ctx, transferIDs)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var msg []*domain.MsgKafkaExpiredTransfer
	for idx := range transfers {
		msg = append(msg, &domain.MsgKafkaExpiredTransfer{
			TransferID: transfers[idx].TransferID,
		})
	}

	return msg, nil
}

func (s *TransferSvcImpl) ProcessManualRefund(ctx context.Context, transferIDs []int64) error {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelUsecase)
		err      error
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	span.SetAttributes(attribute.Int64Slice("transferIDs", transferIDs))
	var transfers []*domain.MsgKafkaExpiredTransfer
	transfers, err = s.FetchExpiredTransfers(ctx, transferIDs)
	if err != nil {
		return err
	}

	for idx := range transfers {
		transfers[idx].OperatedBy = helper.ChainString(ctx.Value(domain.KeyOperatedBy).(string), domain.KafkaProducerOperator)
		_, _, err = s.expiredTransferPublisher.SendMessage(ctx, "manualTriggerRefund", transfers[idx])
		if err != nil {
			log.ZError(ctx, "while manual refund SendMessage", err, "transferID", transfers[idx].TransferID)
		}
	}

	return nil
}

func (s *TransferSvcImpl) RefundTransfer(
	ctx context.Context, arg *domain.RefundTransferReq,
) (err error) {
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

	var walletUser *entity.Wallet
	walletUser, err = s.walletRepo.GetWalletByUserID(ctx, arg.UserID)
	if err != nil {
		log.ZError(ctx, "while get wallet user", err, "UserID", arg.UserID)
		return err
	}
	span.SetAttributes(
		attribute.Int64("argTransferID", arg.TransferID),
		attribute.String("argUserID", arg.UserID),
	)

	var transferDetail *entity.Transfer
	transferDetail, err = s.repo.GetEligibleRefundTransfer(ctx, arg.TransferID, arg.UserID)
	if err != nil {
		log.ZError(ctx, "while get eligible refund transfer", err, "transferID", arg.TransferID)
		return err
	}

	errTrx := s.txRepo.Do(ctx, func(tx *gorm.DB) error {
		transferDetail.RefundedAt = convert.PtrTime(time.Now())
		transferDetail.UpdatedAt = time.Now()
		transferDetail.UpdatedBy = arg.OperatedBy
		transferDetail.StatusTransfer = entity.StatusTransferRefunded

		err = s.repo.Update(ctx, transferDetail, tx)
		if err != nil {
			log.ZError(ctx, "while update transfer", err, "transferID", arg.TransferID)
			return err
		}

		_, err = s.transactionUc.CreateTransaction(ctx, &domain.CreateTransactionReq{
			WalletID:        walletUser.WalletID,
			ImpactedItem:    transferDetail.TransferID,
			TransactionType: string(entity.TransactionTypeRefundTransfer),
			Entrytype:       string(entity.EntryTypeCredit),
			Amount:          transferDetail.Amount,
			DescriptionEn:   fmt.Sprintf("Refund Transfer from %s", transferDetail.ToUserID),
			DescriptionZh:   fmt.Sprintf("退款转账 %s", transferDetail.ToUserID),
			ReferenceCode:   fmt.Sprintf("#RTRF-%s-%s-%d", nowToStringYYYYMMDD(), arg.OperatedBy, transferDetail.TransferID),
			CreatedBy:       arg.OperatedBy,
		})
		if err != nil {
			log.ZError(ctx, "while create transaction", err, "transferID", arg.TransferID)
			return err
		}
		span.SetAttributes(
			attribute.Int64("refundTransferID", transferDetail.TransferID),
			attribute.Float64("refundAmount", transferDetail.Amount),
		)

		return nil
	})
	if errTrx != nil {
		log.ZError(ctx, "while do trx claim transfer", errTrx)
		err = errTrx
		return err
	}

	return nil
}

func (s *TransferSvcImpl) GetDetailTransfer(
	ctx context.Context, transferID int64, userID string,
) (tranfer *domain.Transfer, err error) {
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

	span.SetAttributes(
		attribute.Int64("transferID", transferID),
		attribute.String("userID", userID),
	)

	var detail *entity.Transfer
	detail, err = s.repo.GetDetailTransfer(ctx, transferID, userID)
	if err != nil {
		log.ZError(ctx, "while get detail transfer", err, "transferID", transferID, "userID", userID)
		return nil, err
	}

	return dtoTransfer(detail), nil
}
