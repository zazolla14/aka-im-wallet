package balance_adjustment

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

func (r *repositoryImpl) CreateBalanceAdjustment(
	ctx context.Context, tx *gorm.DB, adjustment *entity.BalanceAdjustments,
) (balanceAdjustmentID int64, err error) {
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
		err = r.db.WithContext(ctx).Create(adjustment).Error
	default:
		err = tx.WithContext(ctx).Create(adjustment).Error
	}

	if err != nil {
		log.ZError(ctx, "while create balance adjustment", err)
		return balanceAdjustmentID, err
	}

	return adjustment.BalanceAdjustmentID, nil
}

func (r *repositoryImpl) GetListBalanceAdjustment(
	ctx context.Context, arg *domain.GetListbalanceAjustmentRequest,
) (adjustments []*entity.BalanceAdjustments, total int64, err error) {
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
		Model(&entity.BalanceAdjustments{}).
		Where("is_active IS TRUE AND deleted_at IS NULL")

	if arg.WalletID > 0 {
		query = query.Where("wallet_id = ?", arg.WalletID)
	}

	if arg.FilterDateBy != "" {
		query = query.Where(
			fmt.Sprintf("%s BETWEEN ? AND ?", arg.FilterDateBy),
			arg.StartDate, arg.EndDate,
		)
	}

	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	query.Order(fmt.Sprintf("%s %s", arg.SortBy, arg.SortOrder))
	offset := (arg.Page - 1) * arg.Limit
	query.Limit(int(arg.Limit)).Offset(int(offset))

	err = query.Find(&adjustments).Error
	if err != nil {
		log.ZError(ctx, "while repositoryImpl GetListBalanceAdjustment", err)
		return nil, 0, err
	}
	span.SetAttributes(attribute.Int64("total", total))

	return adjustments, total, nil
}
