package transfer

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

const hoursInDay = 24

type repositoryImpl struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(transfer *entity.Transfer) error {
	if err := r.db.Create(&transfer).Error; err != nil {
		return err
	}
	return nil
}

func (r *repositoryImpl) Update(ctx context.Context, transfer *entity.Transfer, tx *gorm.DB) error {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
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

	span.SetAttributes(
		attribute.Int64("transferID", transfer.TransferID),
	)

	switch tx {
	case nil:
		err = r.db.Model(&entity.Transfer{}).
			WithContext(ctx).
			Where("transfer_id = ?", transfer.TransferID).
			Updates(transfer).Error
	default:
		err = tx.Model(&entity.Transfer{}).
			WithContext(ctx).
			Where("transfer_id = ?", transfer.TransferID).
			Updates(transfer).Error
	}

	if err != nil {
		return err
	}

	return nil
}
func (r *repositoryImpl) FindAllTransferByWalletID(walletID string) ([]entity.Transfer, error) {
	var transfers []entity.Transfer
	if err := r.db.Where("wallet_id = ?", walletID).Find(&transfers).Error; err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *repositoryImpl) FindByTransferID(ctx context.Context, transferID int64) (*entity.Transfer, error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
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

	span.SetAttributes(
		attribute.Int64("transferID", transferID),
	)

	var transfer entity.Transfer
	err = r.db.Where("transfer_id = ?", transferID).
		WithContext(ctx).
		First(&transfer).Error
	if err != nil {
		return &entity.Transfer{}, err
	}

	return &transfer, nil
}

func (r *repositoryImpl) CreateTransfer(
	ctx context.Context, tx *gorm.DB, transfer *entity.Transfer,
) (transferID int64, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	switch tx {
	case nil:
		err = r.db.WithContext(ctx).Create(transfer).Error
	default:
		err = tx.WithContext(ctx).Create(transfer).Error
	}

	if err != nil {
		log.ZError(ctx, "while create transfer", err)
		return 0, err
	}

	span.SetAttributes(
		attribute.Int64("transferID", transfer.TransferID),
	)

	return transfer.TransferID, nil
}

func (r *repositoryImpl) GetEligibleClaimTransfer(
	ctx context.Context, transferID int64, claimerUserID string,
) (detail *entity.Transfer, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	detail = &entity.Transfer{}
	err = r.db.WithContext(ctx).
		Where("transfer_id = ?", transferID).
		Where("to_user_id = ?", claimerUserID).
		Where("status_transfer = ?", entity.StatusTransferPending).
		Where("expired_at > ?", time.Now()).
		Where("is_active IS TRUE").
		First(&detail).Error
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64("transferIDReq", transferID),
		attribute.String("toUserIDReq", claimerUserID),
		attribute.String("statusTransferReq", string(entity.StatusTransferPending)),
		attribute.Int64("transferIDResp", detail.TransferID),
	)

	return detail, nil
}

func (r *repositoryImpl) CountSentTransferInDay(
	ctx context.Context, userID string, now time.Time,
) (total int64, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	var count int64
	startOfDay := now.Truncate(hoursInDay * time.Hour)
	endOfDay := startOfDay.Add(hoursInDay * time.Hour)
	err = r.db.WithContext(ctx).
		Model(&entity.Transfer{}).
		Where("from_user_id = ?", userID).
		Where("created_at > ? AND created_at <= ?", startOfDay, endOfDay).
		Where("is_active = ?", true).
		Count(&count).Error

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.Int64("total", count),
	)

	return count, err
}

func (r *repositoryImpl) CountClaimedTransferInDay(
	ctx context.Context, userID string, now time.Time,
) (total int64, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	var count int64
	startOfDay := now.Truncate(hoursInDay * time.Hour)
	endOfDay := startOfDay.Add(hoursInDay * time.Hour)
	err = r.db.WithContext(ctx).
		Model(&entity.Transfer{}).
		Where("to_user_id = ?", userID).
		Where("claimed_at > ? AND claimed_at <= ?", startOfDay, endOfDay).
		Where("status_transfer = ?", entity.StatusTransferClaimed).
		Where("is_active = ?", true).
		Count(&count).Error

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.Int64("total", count),
	)

	return count, err
}

func (r *repositoryImpl) FetchExpiredTransfers(ctx context.Context, ids []int64) ([]*entity.Transfer, error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
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

	span.SetAttributes(attribute.Int64Slice("transferIDsReq", ids))

	var transfers []*entity.Transfer
	q := r.db.WithContext(ctx).
		Where("expired_at < ?", time.Now()).
		Where("status_transfer = ?", "pending").
		Where("claimed_at IS NULL").
		Where("refunded_at IS NULL").
		Where("is_active = ?", true)

	if len(ids) > 0 {
		q = q.Where("transfer_id IN ?", ids)
	}

	err = q.Find(&transfers).Error
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("total", len(transfers)))

	return transfers, nil
}

func (r *repositoryImpl) GetEligibleRefundTransfer(
	ctx context.Context, transferID int64, userID string,
) (detail *entity.Transfer, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	detail = &entity.Transfer{}

	span.SetAttributes(
		attribute.Int64("transferID", transferID),
		attribute.String("userID", userID),
		attribute.String("statusTransfer", string(entity.StatusTransferPending)),
	)

	q := r.db.WithContext(ctx).
		Where("transfer_id = ?", transferID).
		Where("from_user_id = ?", userID).
		Where("status_transfer = ?", entity.StatusTransferPending).
		Where("expired_at < ?", time.Now()).
		Where("claimed_at IS NULL").
		Where("refunded_at IS NULL").
		Where("is_active = ?", true)

	err = q.First(&detail).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, eerrs.ErrNoEligibleTransferRefund
		}

		return nil, err
	}

	return detail, nil
}

func (r *repositoryImpl) GetDetailTransfer(
	ctx context.Context, transferID int64, userID string,
) (detail *entity.Transfer, err error) {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	detail = &entity.Transfer{}

	span.SetAttributes(
		attribute.Int64("transferID", transferID),
		attribute.String("userID", userID),
	)

	q := r.db.WithContext(ctx).
		Where("transfer_id = ?", transferID).
		Where("(from_user_id = ? OR to_user_id = ?)", userID, userID).
		Where("is_active = ?", true)

	err = q.First(&detail).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, eerrs.ErrTransferNotFound
		}

		return nil, err
	}

	return detail, nil
}
