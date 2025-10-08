package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	d "github.com/1nterdigital/aka-im-wallet/internal/domain"
	e "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/envelope"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/tx"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/kafka"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
	"github.com/1nterdigital/aka-im-wallet/pkg/helper"
)

type (
	EnvelopeSvcImpl struct {
		expiredEnvelopePublisher *kafka.Producer
		walletTransactionUc      WalletTransactionSvc
		walletUC                 WalletSvc
		txRepo                   tx.Repository
		envelopeRepo             envelope.Repository
		walletRepo               wallet.Repository
	}

	EnvelopeSvc interface {
		GetAllEnvelopesByUserID(userID int64) (resp []*d.EnvelopeDomain, err error)
		FindOne(ctx context.Context, envelopeID int64) (resp *e.Envelope, err error)
		CreateWalletTransaction(ctx context.Context, tx *d.TransactionAction, env *e.Envelope) (err error)
		CreateEnvelope(ctx context.Context, req *d.EnvelopeCreateRequest) (resp *d.EnvelopeCreateResponse, err error)
		Claim(ctx context.Context, req *d.EnvelopeClaimRequest) (resp *d.EnvelopeClaimResponse, err error)
		AutoRefund(ctx context.Context) (err error)
		RefundByID(ctx context.Context, userID string, envID int64) (err error)
		Refund(ctx context.Context, env *e.Envelope) (resp *e.Envelope, err error)
		ProcessExpiredEnvelopes(
			ctx context.Context, envelopes []*d.MsgKafkaExpiredEnvelope,
		) (failedRefund []*d.MsgKafkaExpiredEnvelope, err error)
		FetchExpiredEnvelopes(ctx context.Context, envelopeIDs []int64) (resp []*d.MsgKafkaExpiredEnvelope, err error)
		ProcessManualRefund(ctx context.Context, envelopeIDs []int64) (err error)
	}
)

func NewEnvelopeUseCase(
	expiredEnvelopePublisher *kafka.Producer,
	walletUC WalletSvc,
	walletTransactionUc WalletTransactionSvc,
	txRepo tx.Repository,
	envelopeRepo envelope.Repository,
	walletRepo wallet.Repository,
) EnvelopeSvc {
	return &EnvelopeSvcImpl{
		expiredEnvelopePublisher: expiredEnvelopePublisher,
		walletUC:                 walletUC,
		walletTransactionUc:      walletTransactionUc,
		txRepo:                   txRepo,
		envelopeRepo:             envelopeRepo,
		walletRepo:               walletRepo,
	}
}

func (uc *EnvelopeSvcImpl) GetAllEnvelopesByUserID(userID int64) (resp []*d.EnvelopeDomain, err error) {
	envelopes, err := uc.envelopeRepo.GetAllEnvelopesByUserID(userID)
	if err != nil {
		return nil, err
	}
	domainEnvelopes := make([]*d.EnvelopeDomain, len(envelopes))
	for i, newEnvelope := range envelopes {
		domainEnvelopes[i] = &d.EnvelopeDomain{
			TotalAmount: newEnvelope.TotalAmount,
		}
	}
	return domainEnvelopes, nil
}

func (uc *EnvelopeSvcImpl) FindOne(ctx context.Context, envelopeID int64) (resp *e.Envelope, err error) {
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

	var env *e.Envelope
	env, err = uc.envelopeRepo.GetEnvelope(ctx, envelopeID)
	if err != nil {
		log.ZError(ctx, "while get envelope", err, "envelopeID", envelopeID)
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64("envelopeID", envelopeID),
	)

	return env, nil
}

func (uc *EnvelopeSvcImpl) CreateWalletTransaction(ctx context.Context, act *d.TransactionAction, env *e.Envelope) (err error) {
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

	format, ok := e.EnvelopeDescFormat[act.Action]
	if !ok {
		return eerrs.ErrUnsupportedAction(act.Action)
	}
	reqTx := &d.CreateTransactionReq{
		WalletID:        act.WalletID,
		TransactionType: act.TransactionType,
		Entrytype:       act.EntryType,
		Amount:          act.Amount,
		DescriptionEn:   fmt.Sprintf(format.En, act.UserID, env.EnvelopeID, act.Amount),
		DescriptionZh:   fmt.Sprintf(format.Zh, act.UserID, env.EnvelopeID, act.Amount),
		ReferenceCode:   fmt.Sprintf("#%s-%s-%s-%d", format.RefPrefix, nowToStringYYYYMMDD(), act.UserID, env.EnvelopeID),
		ImpactedItem:    env.EnvelopeID,
		CreatedBy:       env.CreatedBy,
	}
	_, err = uc.walletTransactionUc.CreateTransaction(ctx, reqTx)
	if err != nil {
		log.ZError(ctx, "while get createTransaction", err, "envelopeID", env.EnvelopeID)
		return err
	}

	span.SetAttributes(
		attribute.Int64("walletID", act.WalletID),
		attribute.String("transactionType", act.TransactionType),
		attribute.String("entryType", act.EntryType),
		attribute.Float64("amount", act.Amount),
	)

	return nil
}

func (uc *EnvelopeSvcImpl) validateCreateEnvelope(ctx context.Context, walletUser *e.Wallet, req *d.EnvelopeCreateRequest) (err error) {
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

	if walletUser.UserID != req.UserID {
		return eerrs.ErrUnauthorizedUserID
	}
	if walletUser.Balance < req.TotalAmount {
		return eerrs.ErrAmountExceedsWalletBalance
	}
	var countToday int64
	countToday, err = uc.envelopeRepo.CountSentEnvelope(ctx, req.UserID, time.Now())
	if err != nil {
		log.ZError(ctx, "while get countSentEnvelope", err, "userID", req.UserID)
		return err
	}
	if countToday >= e.MaxEnvelopeSendPerDay {
		return eerrs.ErrSendingDailyLimit
	}
	return nil
}

func (uc *EnvelopeSvcImpl) CreateEnvelope(
	ctx context.Context, req *d.EnvelopeCreateRequest,
) (resp *d.EnvelopeCreateResponse, err error) {
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

	var (
		types        = req.EnvelopeType
		amount       = req.TotalAmount
		remarks      = req.Remarks
		totalClaimer = req.TotalClaimer
		userID       = req.UserID
		walletID     = req.WalletID
	)
	var walletDB *e.Wallet
	walletDB, err = uc.walletRepo.FindByWalletID(walletID)
	if err != nil {
		log.ZError(ctx, "while get findByWalletID", err, "walletID", walletID)
		return nil, err
	}
	err = uc.validateCreateEnvelope(ctx, walletDB, req)
	if err != nil {
		log.ZError(ctx, "while get validateCreateEnvelope", err, "userID", req.UserID)
		return nil, err
	}
	dayDuration := 24 * time.Hour
	newEnvelope := &e.Envelope{
		UserID:              userID,
		WalletID:            walletID,
		TotalAmount:         amount,
		TotalAmountClaimed:  0,
		TotalAmountRefunded: 0,
		Remarks:             remarks,
		MaxNumReceived:      totalClaimer,
		EnvelopeType:        types,
		ExpiredAt:           d.GenerateExpiredAt(dayDuration),
		IsActive:            true,
		CreatedAt:           time.Now(),
		CreatedBy:           userID,
	}
	err = uc.createEnvelopeTx(ctx, newEnvelope, req.ToUserID)
	if err != nil {
		log.ZError(ctx, "while get createEnvelopeTx", err, "userID", req.UserID)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.Int64("walletID", walletID),
		attribute.Float64("totalAmount", amount),
		attribute.Int("totalClaimer", totalClaimer),
		attribute.String("type", types),
		attribute.String("toUserId", req.ToUserID),
	)

	result := &d.EnvelopeCreateResponse{
		EnvelopeID:          newEnvelope.EnvelopeID,
		UserID:              newEnvelope.UserID,
		WalletID:            newEnvelope.WalletID,
		TotalAmount:         newEnvelope.TotalAmount,
		TotalAmountClaimed:  newEnvelope.TotalAmountClaimed,
		TotalAmountRefunded: newEnvelope.TotalAmountRefunded,
		MaxNumReceived:      newEnvelope.MaxNumReceived,
		EnvelopeType:        newEnvelope.EnvelopeType,
		Remarks:             newEnvelope.Remarks,
		ExpiredAt:           newEnvelope.ExpiredAt,
		RefundedAt:          newEnvelope.RefundedAt,
		IsActive:            newEnvelope.IsActive,
		CreatedAt:           newEnvelope.CreatedAt,
		CreatedBy:           newEnvelope.CreatedBy,
		UpdatedAt:           newEnvelope.UpdatedAt,
		UpdatedBy:           newEnvelope.UpdatedBy,
		DeletedAt:           newEnvelope.DeletedAt,
		DeletedBy:           newEnvelope.DeletedBy,
	}

	span.SetAttributes(
		attribute.Int64("envelopeID", newEnvelope.EnvelopeID),
		attribute.String("createdAt", newEnvelope.CreatedAt.Format(time.RFC3339)),
	)

	return result, nil
}

func (uc *EnvelopeSvcImpl) createEnvelopeTx(ctx context.Context, envelopeTx *e.Envelope, toUserID string) (err error) {
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

	return uc.envelopeRepo.WithTransaction(ctx, func(txRepo envelope.Repository) error {
		var (
			amounts []float64
			txType  e.TransactionType
		)
		txType, err = e.GetTransactionTypeByEnvelopeType(envelopeTx.EnvelopeType)
		if err != nil {
			log.ZError(ctx, "while get validateCreateEnvelope", err, "envelopeID", envelopeTx.EnvelopeID)
			return err
		}
		amounts, err = splitAmounts(envelopeTx)
		if err != nil {
			log.ZError(ctx, "while get splitAmounts", err, "envelopeID", envelopeTx.EnvelopeID)
			return err
		}
		err = txRepo.CreateEnvelope(ctx, envelopeTx)
		if err != nil {
			log.ZError(ctx, "while get createEnvelope", err, "envelopeID", envelopeTx.EnvelopeID)
			return err
		}
		envOpts := d.EnvelopeDetailOption{
			Envelope:   envelopeTx,
			Amounts:    amounts,
			ReceiverID: toUserID,
		}
		err = createEnvelopeDetails(ctx, txRepo, envOpts)
		if err != nil {
			log.ZError(ctx, "while get createEnvelopeDetails", err, "envelopeID", envelopeTx.EnvelopeID)
			return err
		}
		transactionAction := &d.TransactionAction{
			Action:          e.ActionCreate,
			WalletID:        envelopeTx.WalletID,
			TransactionType: string(txType),
			EntryType:       string(e.EntryTypeDebit),
			Amount:          -(envelopeTx.TotalAmount),
			UserID:          envelopeTx.UserID,
		}
		err = uc.CreateWalletTransaction(ctx, transactionAction, envelopeTx)
		if err != nil {
			log.ZError(ctx, "while get createWalletTransaction", err, "envelopeID", envelopeTx.EnvelopeID)
			return err
		}

		span.SetAttributes(
			attribute.Int64("envelopeId", envelopeTx.EnvelopeID),
			attribute.Float64("totalAmount", envelopeTx.TotalAmount),
			attribute.String("toUserId", toUserID),
		)

		return nil
	})
}

func createEnvelopeDetails(ctx context.Context, txRepo envelope.Repository, opts d.EnvelopeDetailOption) error {
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

	var details []*e.EnvelopeDetail
	now := time.Now()
	for _, amount := range opts.Amounts {
		detail := &e.EnvelopeDetail{
			EnvelopeID:           opts.Envelope.EnvelopeID,
			Amount:               amount,
			EnvelopeDetailStatus: e.EnvelopePending,
			IsActive:             true,
			CreatedAt:            now,
			CreatedBy:            opts.Envelope.CreatedBy,
		}
		if e.EnvelopeType(opts.Envelope.EnvelopeType) == e.EnvelopeTypeSingle {
			detail.UserID = opts.ReceiverID
		}
		details = append(details, detail)
	}
	span.SetAttributes(
		attribute.Int64("envelopeID", opts.Envelope.EnvelopeID),
	)

	return txRepo.CreateEnvelopeDetails(ctx, details)
}

func splitAmounts(env *e.Envelope) (amounts []float64, err error) {
	switch env.EnvelopeType {
	case string(e.EnvelopeTypeFixed):
		return equalSplit(env.TotalAmount, env.MaxNumReceived), nil
	case string(e.EnvelopeTypeLucky):
		return luckySplit(env.TotalAmount, env.MaxNumReceived), nil
	case string(e.EnvelopeTypeSingle):
		return []float64{env.TotalAmount}, nil
	default:
		return nil, eerrs.ErrUnsupportedEnvelopeType
	}
}

func luckySplit(total float64, count int) (amounts []float64) {
	shares := make([]float64, 0, count)
	remain := total
	var (
		sum                 float64
		minimumShareInCents float64 = 1
		centsMultiplier     float64 = 100
	)

	minimumShare := minimumShareInCents / centsMultiplier

	for i := range count - 1 {
		_max := (remain / float64(count-i)) * 2
		//nolint:gosec // weak random is fine here since it's not for security
		part := rand.Float64()*(_max-minimumShare) + minimumShare
		if part > remain {
			part = remain
		}
		part = math.Floor(part*centsMultiplier) / centsMultiplier
		shares = append(shares, part)
		sum += part
		remain -= part
	}

	last := total - sum
	last = math.Round(last*centsMultiplier) / centsMultiplier
	shares = append(shares, last)

	return shares
}

func equalSplit(total float64, count int) (amounts []float64) {
	var centsMultiplier float64 = 100
	shares := make([]float64, count)
	share := math.Floor((total/float64(count))*centsMultiplier) / centsMultiplier
	remaining := total

	for i := range count - 1 {
		shares[i] = share
		remaining -= share
	}
	shares[count-1] = math.Round(remaining*centsMultiplier) / centsMultiplier
	return shares
}

func (uc *EnvelopeSvcImpl) validateClaimEnvelope(
	ctx context.Context, req *d.EnvelopeClaimRequest, env *e.Envelope, walletUser *e.Wallet,
) (err error) {
	var (
		funcName   = tracer.GetFullFunctionPath()
		t          = otel.Tracer(tracer.LevelUsecase)
		envelopeID = req.EnvelopeID
		userID     = req.UserID
	)
	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	if walletUser.UserID != userID {
		return eerrs.ErrUnauthorizedUserID
	}
	now := time.Now()
	if env.ExpiredAt != nil && now.After(*env.ExpiredAt) {
		return eerrs.ErrExpiredEnvelope
	}
	var claimCountToday int64
	claimCountToday, err = uc.envelopeRepo.CountClaimedEnvelope(ctx, req.UserID, time.Now())
	if err != nil {
		log.ZError(ctx, "while get countClaimedEnvelope", err, "userID", userID)
		return err
	}
	if claimCountToday > e.MaxEnvelopeClaimPerDay {
		return eerrs.ErrClaimingDailyLimit
	}
	var status *e.ClaimStatus
	status, err = uc.envelopeRepo.CheckClaimStatus(ctx, envelopeID, userID)
	if err != nil {
		log.ZError(ctx, "while get checkClaimStatus", err, "userID", userID, "envelopeID", env.EnvelopeID)
		return err
	}
	span.SetAttributes(
		attribute.Int64("envelopeId", req.EnvelopeID),
		attribute.String("userId", req.UserID),
		attribute.Int("maxNumReceived", env.MaxNumReceived),
		attribute.Int64("claimCountToday", claimCountToday),
		attribute.Int64("totalClaimed", status.TotalClaimed),
		attribute.Int64("userHasClaimed", status.UserHasClaimed),
	)

	if status.TotalClaimed >= int64(env.MaxNumReceived) {
		return eerrs.ErrAllEnvelopeHasBeenClaimed
	}
	if status.UserHasClaimed > 0 {
		return eerrs.ErrUserAlreadyClaimedThisEnvelope
	}
	return nil
}

func (uc *EnvelopeSvcImpl) Claim(
	ctx context.Context, req *d.EnvelopeClaimRequest,
) (resp *d.EnvelopeClaimResponse, err error) {
	var (
		funcName      = tracer.GetFullFunctionPath()
		t             = otel.Tracer(tracer.LevelUsecase)
		claimedDetail *d.EnvelopeClaimResponse
		envelopeID    = req.EnvelopeID
		userID        = req.UserID
		walletID      = req.WalletID
	)
	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	errTrx := uc.envelopeRepo.WithTransaction(ctx, func(txRepo envelope.Repository) error {
		lockedEnv, errs := txRepo.LockEnvelopeByID(ctx, envelopeID)
		if errs != nil {
			log.ZError(ctx, "while get lockEnvelopeByID", errs, "userID", userID, "envelopeID", envelopeID)
			return errs
		}
		if !lockedEnv.IsActive {
			return eerrs.ErrEnvelopeNotActive
		}
		walletDB, errs := uc.walletRepo.FindByWalletID(walletID)
		if errs != nil {
			log.ZError(ctx, "while get findByWalletID", errs, "userID", userID, "walletID", walletID, "envelopeID", envelopeID)
			return errs
		}
		errs = uc.validateClaimEnvelope(ctx, req, lockedEnv, walletDB)
		if errs != nil {
			log.ZError(ctx, "while get validateClaimEnvelope", errs, "userID", userID, "envelopeID", envelopeID)
			return errs
		}

		detail, errs := txRepo.ClaimNextLuckyShare(ctx, envelopeID)
		if errs != nil {
			log.ZError(ctx, "while get claimNextLuckyShare", errs, "userID", userID, "envelopeID", envelopeID)
			return errs
		}
		if e.EnvelopeType(lockedEnv.EnvelopeType) == e.EnvelopeTypeSingle {
			if detail.UserID != userID {
				return eerrs.ErrUnauthorizedClaimer
			}
		}
		if detail == nil {
			return eerrs.ErrNoMoreSharesToClaim
		}

		now := time.Now()
		detail.UserID = userID
		detail.EnvelopeDetailStatus = e.EnvelopeClaimed
		detail.ClaimedAt = &now
		detail.UpdatedAt = now
		detail.UpdatedBy = userID

		errs = txRepo.UpdateEnvelopeDetail(ctx, detail, nil)
		if errs != nil {
			log.ZError(ctx, "while get updateEnvelopeDetail", errs, "userID", userID, "envelopeID", envelopeID)
			return errs
		}

		errs = txRepo.UpdateClaimedAmount(ctx, envelopeID, detail.Amount)
		if errs != nil {
			log.ZError(ctx, "while get updateClaimedAmount", errs, "userID", userID, "envelopeID", envelopeID)
			return errs
		}

		txType, errs := e.GetTransactionTypeByEnvelopeType(lockedEnv.EnvelopeType)
		if errs != nil {
			log.ZError(ctx, "while get getTransactionTypeByEnvelopeType", errs, "userID", userID, "envelopeID", envelopeID)
			return errs
		}

		transactionAction := &d.TransactionAction{
			Action:          e.ActionClaim,
			WalletID:        walletID,
			TransactionType: string(txType),
			EntryType:       string(e.EntryTypeCredit),
			Amount:          detail.Amount,
			UserID:          userID,
		}

		errs = uc.CreateWalletTransaction(ctx, transactionAction, lockedEnv)
		if errs != nil {
			log.ZError(ctx, "while get createWalletTransaction", errs, "userID", userID, "envelopeID", envelopeID)
			return errs
		}

		claimedDetail = &d.EnvelopeClaimResponse{
			EnvelopeDetailID:     detail.EnvelopeDetailID,
			EnvelopeID:           detail.EnvelopeID,
			Amount:               detail.Amount,
			UserID:               detail.UserID,
			EnvelopeDetailStatus: detail.EnvelopeDetailStatus.String(),
			ClaimedAt:            detail.ClaimedAt,
			IsActive:             detail.IsActive,
			CreatedAt:            detail.CreatedAt,
			CreatedBy:            detail.CreatedBy,
			UpdatedAt:            detail.UpdatedAt,
			UpdatedBy:            detail.UpdatedBy,
			DeletedAt:            detail.DeletedAt,
			DeletedBy:            detail.DeletedBy,
			Envelope:             detail.Envelope,
			Wallet:               detail.Wallet,
		}
		span.SetAttributes(
			attribute.Int64("envelopeId", envelopeID),
			attribute.String("userId", userID),
			attribute.Int64("walletId", walletID),
			attribute.String("envelopeType", lockedEnv.EnvelopeType),
			attribute.String("transactionType", string(txType)),
			attribute.Float64("claimAmount", detail.Amount),
			attribute.String("claimStatus", detail.EnvelopeDetailStatus.String()),
		)

		return nil
	})

	if errTrx != nil {
		err = errTrx
		return nil, err
	}

	return claimedDetail, nil
}

func (uc *EnvelopeSvcImpl) AutoRefund(ctx context.Context) (err error) {
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

	var envelopes []*e.Envelope
	envelopes, err = uc.envelopeRepo.GetExpiredUnRefundedEnvelopes(ctx)
	if err != nil {
		return err
	}
	for _, env := range envelopes {
		_, err = uc.Refund(ctx, env)
		if err != nil {
			log.ZError(ctx, "while get Refund", err, "userID", env.UserID, "envelopeID", env.EnvelopeID)
			return eerrs.ErrRefund(env.EnvelopeID, env.UserID, err)
		}
		span.SetAttributes(
			attribute.String("userID", env.UserID),
			attribute.Int64("envelopeID", env.EnvelopeID),
		)
	}
	return nil
}

func (uc *EnvelopeSvcImpl) RefundByID(ctx context.Context, userID string, envID int64) (err error) {
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

	var env *e.Envelope
	env, err = uc.envelopeRepo.GetExpiredUnRefundedEnvelopesByID(ctx, envID, userID)
	if err != nil {
		log.ZError(ctx, "while get getExpiredUnRefundedEnvelopesByID", err, "userID", userID, "envelopeID", envID)
		return err
	}
	if reflect.ValueOf(env).IsZero() {
		return eerrs.ErrEnvelopeNotFound
	}
	_, err = uc.Refund(ctx, env)
	if err != nil {
		log.ZError(ctx, "while get Refund", err, "userID", userID, "envelopeID", envID)
		return eerrs.ErrRefund(env.EnvelopeID, env.UserID, err)
	}

	span.SetAttributes(
		attribute.String("userId", userID),
		attribute.Int64("envelopeId", envID),
	)

	return nil
}

func (uc *EnvelopeSvcImpl) Refund(ctx context.Context, env *e.Envelope) (resp *e.Envelope, err error) {
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

	remaining := env.TotalAmount - env.TotalAmountClaimed
	if remaining <= 0 {
		return nil, eerrs.ErrNoRemainingAmountToRefund
	}

	errTx := uc.envelopeRepo.WithTransaction(ctx, func(txRepo envelope.Repository) error {
		if err = txRepo.RefundEnvelope(ctx, env.EnvelopeID, remaining); err != nil {
			log.ZError(ctx, "while get refundEnvelope", err, "userID", env.UserID, "envelopeID", env.EnvelopeID)
			return err
		}
		if err = txRepo.DeactivateUnclaimedDetails(ctx, env.EnvelopeID); err != nil {
			log.ZError(ctx, "while get deactiveUnclaimedDetails", err, "userID", env.UserID, "envelopeID", env.EnvelopeID)
			return eerrs.ErrDeactivateDetails(err)
		}
		transactionAction := &d.TransactionAction{
			Action:          e.ActionRefund,
			WalletID:        env.WalletID,
			TransactionType: string(e.TransactionTypeRefundEnvelope),
			EntryType:       string(e.EntryTypeCredit),
			Amount:          remaining,
			UserID:          env.UserID,
		}
		err = uc.CreateWalletTransaction(ctx, transactionAction, env)
		if err != nil {
			return err
		}

		return nil
	})

	span.SetAttributes(
		attribute.Int64("envelopeId", env.EnvelopeID),
		attribute.String("userId", env.UserID),
		attribute.Int64("walletId", env.WalletID),
		attribute.Float64("totalAmount", env.TotalAmount),
		attribute.Float64("claimedAmount", env.TotalAmountClaimed),
		attribute.Float64("remainingAmount", remaining),
	)

	if errTx != nil {
		err = errTx
		return nil, err
	}
	return env, nil
}

func (uc *EnvelopeSvcImpl) ProcessExpiredEnvelopes(
	ctx context.Context, envelopes []*d.MsgKafkaExpiredEnvelope,
) (failedRefund []*d.MsgKafkaExpiredEnvelope, err error) {
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
	for idx := range envelopes {
		operatedBy := envelopes[idx].OperatedBy
		envelopes[idx].Counter += 1
		if envelopes[idx].Counter >= retryAttemptsProcess {
			continue
		}
		span.SetAttributes(attribute.Int64("envelopeID", envelopes[idx].EnvelopeID))

		var env *e.Envelope
		env, err = uc.envelopeRepo.GetEnvelope(ctx, envelopes[idx].EnvelopeID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.ZWarn(ctx, "while GetEnvelope", err,
					"envelopeID", envelopes[idx].EnvelopeID,
					"operatedBy", operatedBy,
				)
				failedRefund = append(failedRefund, envelopes[idx])
			}
			continue
		}

		span.SetAttributes(attribute.String("userID", env.UserID))

		var walletUser *e.Wallet
		walletUser, err = uc.walletRepo.GetWalletByUserID(ctx, env.UserID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.ZWarn(ctx, "while GetWalletByUserID", err,
					"userID", env.UserID,
					"operatedBy", operatedBy,
				)
				failedRefund = append(failedRefund, envelopes[idx])
			}
			continue
		}

		if env.RefundedAt != nil {
			continue
		}

		var envelopesDetails []*e.EnvelopeDetail
		envelopesDetails, err = uc.envelopeRepo.GetEnvelopeDetailsByEnvelopID(ctx, env.EnvelopeID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.ZWarn(ctx, "while GetEnvelopeDetailsByEnvelopID", err,
					"userID", env.UserID,
					"envelopeID", env.EnvelopeID,
					"operatedBy", operatedBy,
				)
				failedRefund = append(failedRefund, envelopes[idx])
			}
			continue
		}

		txErr := uc.txRepo.Do(ctx, func(tx *gorm.DB) error {
			var errx error
			refundedAt := time.Now()
			env.RefundedAt = &refundedAt
			env.UpdatedAt = refundedAt
			env.UpdatedBy = operatedBy
			env.TotalAmountRefunded = env.TotalAmount - env.TotalAmountClaimed
			errx = uc.envelopeRepo.Update(ctx, env, tx)
			if errx != nil {
				log.ZError(ctx, "while Update envelope", errx,
					"userID", env.UserID,
					"envelopeID", env.EnvelopeID,
					"operatedBy", operatedBy,
				)
				return errx
			}

			ctx = context.WithValue(ctx, d.KeyOperatedBy, operatedBy)
			errx = uc.processRefundEnvelopeDetail(ctx, tx, envelopesDetails, refundedAt)
			if errx != nil {
				log.ZError(ctx, "while processRefundEnvelopeDetail", errx,
					"userID", env.UserID,
					"envelopeID", env.EnvelopeID,
					"operatedBy", operatedBy,
				)
				return errx
			}

			_, errx = uc.walletTransactionUc.CreateTransaction(ctx, &d.CreateTransactionReq{
				WalletID:        walletUser.WalletID,
				ImpactedItem:    env.EnvelopeID,
				TransactionType: string(e.TransactionTypeRefundTransfer),
				Entrytype:       string(e.EntryTypeDebit),
				Amount:          env.TotalAmount - env.TotalAmountClaimed,
				DescriptionEn:   fmt.Sprintf(d.TransactionRefundEnvelopeEN, env.EnvelopeID),
				DescriptionZh:   fmt.Sprintf(d.TransactionRefundEnvelopeZN, env.EnvelopeID),
				ReferenceCode:   fmt.Sprintf(d.FormatReferenceCodeRefundEnvelope, env.EnvelopeID, walletUser.WalletID),
				CreatedBy:       operatedBy,
			})
			if errx != nil {
				log.ZWarn(ctx, "while CreateTransaction", errx,
					"userID", env.UserID,
					"envelopeID", env.EnvelopeID,
					"operatedBy", operatedBy,
				)
				return errx
			}

			span.SetAttributes(attribute.Float64("amount", env.TotalAmount-env.TotalAmountClaimed))

			return nil
		})
		if txErr != nil {
			log.ZError(ctx, "while do trx refund transfer", txErr)
			failedRefund = append(failedRefund, envelopes[idx])
			err = txErr
			continue
		}
	}

	return failedRefund, nil
}

func (uc *EnvelopeSvcImpl) processRefundEnvelopeDetail(
	ctx context.Context, db *gorm.DB, envelopeDetails []*e.EnvelopeDetail, refundedAt time.Time,
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

	operatedBy := helper.ChainString(ctx.Value(d.KeyOperatedBy).(string), d.KafkaProducerOperator)
	for _, detail := range envelopeDetails {
		if detail.EnvelopeDetailStatus != e.EnvelopePending {
			continue
		}
		detail.UpdatedAt = refundedAt
		detail.UpdatedBy = operatedBy
		detail.EnvelopeDetailStatus = e.EnvelopeRefunded

		err = uc.envelopeRepo.UpdateEnvelopeDetail(ctx, detail, db)
		if err != nil {
			return err
		}
		span.SetAttributes(
			attribute.Int64("envelopeDetailID", detail.EnvelopeDetailID),
			attribute.Float64("envelopeDetailID", detail.Amount),
		)
	}
	return nil
}

func (uc *EnvelopeSvcImpl) FetchExpiredEnvelopes(ctx context.Context, envelopeIDs []int64) ([]*d.MsgKafkaExpiredEnvelope, error) {
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

	envelopes, err := uc.envelopeRepo.FetchExpiredEnvelopes(ctx, envelopeIDs)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var msg []*d.MsgKafkaExpiredEnvelope
	for idx := range envelopes {
		msg = append(msg, &d.MsgKafkaExpiredEnvelope{
			EnvelopeID: envelopes[idx].EnvelopeID,
		})
	}
	span.SetAttributes(attribute.Int("expiredEnvelopesTotal", len(envelopes)))

	return msg, nil
}

func (uc *EnvelopeSvcImpl) ProcessManualRefund(ctx context.Context, envelopeIDs []int64) error {
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

	envelopes, err := uc.FetchExpiredEnvelopes(ctx, envelopeIDs)
	if err != nil {
		return err
	}

	for idx := range envelopes {
		span.SetAttributes(attribute.Int64("envelopeID", envelopes[idx].EnvelopeID))
		envelopes[idx].OperatedBy = helper.ChainString(ctx.Value(d.KeyOperatedBy).(string), d.KafkaProducerOperator)
		_, _, err := uc.expiredEnvelopePublisher.SendMessage(ctx, "manualTriggerRefund", envelopes[idx])
		if err != nil {
			log.ZError(ctx, "while manual refund SendMessage", err, "envelopID", envelopes[idx].EnvelopeID)
		}
	}

	return nil
}
