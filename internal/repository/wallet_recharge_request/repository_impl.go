package wallet_recharge_request

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type repositoryImpl struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) CreateDeposit(
	ctx context.Context, tx *gorm.DB, deposit *entity.WalletRechargeRequest,
) (depositID int64, err error) {
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
		err = r.db.WithContext(ctx).Create(deposit).Error
	default:
		err = tx.WithContext(ctx).Create(deposit).Error
	}

	if err != nil {
		log.ZError(ctx, "while create deposit", err)
		return depositID, err
	}

	span.SetAttributes(
		attribute.Int64("walletID", deposit.WalletID),
		attribute.Float64("amount", deposit.Amount),
		attribute.Int64("walletRechargeRequestID", deposit.WalletRechargeRequestID),
	)

	return deposit.WalletRechargeRequestID, nil
}

func (r *repositoryImpl) GetListDeposit(
	ctx context.Context, arg *domain.GetListDepositRequest,
) (deposits []*entity.WalletRechargeRequest, total int64, err error) {
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

	query := r.db.WithContext(ctx).
		Model(&entity.WalletRechargeRequest{}).
		Where("is_active IS TRUE AND deleted_at IS NULL")

	if arg.WalletID > 0 {
		query = query.Where("wallet_id = ?", arg.WalletID)
		span.SetAttributes(attribute.Int64("walletID", arg.WalletID))
	}

	if arg.FilterDateBy != "" {
		query = query.Where(
			fmt.Sprintf("%s BETWEEN ? AND ?", arg.FilterDateBy),
			arg.StartDate, arg.EndDate,
		)
		span.SetAttributes(attribute.String("filterDateBy", arg.FilterDateBy))
	}

	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	span.SetAttributes(
		attribute.String("sortBy", arg.SortBy),
		attribute.String("sortOder", arg.SortOrder),
		attribute.Int("page", int(arg.Page)),
		attribute.Int("limit", int(arg.Limit)),
	)

	query.Order(fmt.Sprintf("%s %s", arg.SortBy, arg.SortOrder))
	offset := (arg.Page - 1) * arg.Limit
	query.Limit(int(arg.Limit)).Offset(int(offset))

	err = query.Find(&deposits).Error
	if err != nil {
		log.ZError(ctx, "while repositoryImpl GetListDeposit", err)
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int64("total", total))
	return deposits, total, nil
}
