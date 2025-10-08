package envelope

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	e "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

const hoursInDay = 24

type repositoryImpl struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) GetAllEnvelopesByUserID(userID int64) (resp []*e.Envelope, err error) {
	var envelopes []*e.Envelope
	if err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&envelopes).Error; err != nil {
		return nil, err
	}

	return envelopes, nil
}

func activeEnvelopeQuery(db *gorm.DB, envelopeID int64) *gorm.DB {
	return db.Where("envelope_id = ? AND is_active = ?", envelopeID, true)
}

func (r *repositoryImpl) GetEnvelope(ctx context.Context, envelopeID int64) (resp *e.Envelope, err error) {
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

	span.SetAttributes(attribute.Int64("envelopeID", envelopeID))
	var envelope e.Envelope
	err = activeEnvelopeQuery(r.db.WithContext(ctx), envelopeID).First(&envelope).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, eerrs.ErrEnvelopeNotFound
		}
		return nil, err
	}

	return &envelope, nil
}

func (r *repositoryImpl) CreateEnvelope(ctx context.Context, env *e.Envelope) (err error) {
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

	err = r.db.WithContext(ctx).Create(env).Error
	return err
}

func (r *repositoryImpl) CreateEnvelopeDetails(ctx context.Context, envDetail []*e.EnvelopeDetail) (err error) {
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

	err = r.db.WithContext(ctx).CreateInBatches(envDetail, domain.BatchSizeCreateEnvelope).Error
	return err
}

func (r *repositoryImpl) CheckClaimStatus(ctx context.Context, envelopeID int64, userID string) (resp *e.ClaimStatus, err error) {
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

	var result e.ClaimStatus
	err = activeEnvelopeQuery(r.db.WithContext(ctx), envelopeID).
		Model(&e.EnvelopeDetail{}).
		Select("COUNT(*) AS total_claimed, SUM(CASE WHEN user_id = ? THEN 1 ELSE 0 END) AS user_has_claimed", userID).
		Where("envelope_detail_status = ?", e.EnvelopeClaimed).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64("envelopeID", envelopeID),
		attribute.Int64("claimStatus", result.TotalClaimed),
		attribute.Int64("claimStatus", result.UserHasClaimed),
	)

	return &result, nil
}

func (r *repositoryImpl) CountSentEnvelope(ctx context.Context, userID string, now time.Time) (total int64, err error) {
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
		Model(&e.Envelope{}).
		Where("user_id = ?", userID).
		Where("created_at > ? AND created_at <= ?", startOfDay, endOfDay).
		Where("is_active = ?", true).
		Count(&count).Error

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.Int64("total", count),
	)

	return count, err
}

func (r *repositoryImpl) CountClaimedEnvelope(ctx context.Context, userID string, now time.Time) (total int64, err error) {
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
		Model(&e.EnvelopeDetail{}).
		Where("user_id = ?", userID).
		Where("claimed_at > ? AND claimed_at <= ?", startOfDay, endOfDay).
		Where("envelope_detail_status = ?", e.EnvelopeClaimed).
		Where("is_active = ?", true).
		Count(&count).Error

	span.SetAttributes(
		attribute.String("userID", userID),
		attribute.Int64("total", count),
	)
	return count, err
}

func (r *repositoryImpl) LockEnvelopeByID(ctx context.Context, envelopeID int64) (resp *e.Envelope, err error) {
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

	span.SetAttributes(attribute.Int64("envelopeID", envelopeID))

	var env e.Envelope
	err = activeEnvelopeQuery(r.db.WithContext(ctx), envelopeID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *repositoryImpl) ClaimNextLuckyShare(ctx context.Context, envelopeID int64) (resp *e.EnvelopeDetail, err error) {
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

	span.SetAttributes(attribute.Int64("envelopeID", envelopeID))

	var detail e.EnvelopeDetail
	err = activeEnvelopeQuery(r.db.WithContext(ctx), envelopeID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("envelope_detail_status != ?", e.EnvelopeClaimed).
		Order("created_at ASC").
		Find(&detail).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &detail, nil
}

func (r *repositoryImpl) UpdateEnvelopeDetail(ctx context.Context, detail *e.EnvelopeDetail, tx *gorm.DB) (err error) {
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

	span.SetAttributes(attribute.Int64("envelopeID", detail.EnvelopeDetailID))

	if tx != nil {
		return tx.Save(detail).WithContext(ctx).Error
	}
	return r.db.Save(detail).WithContext(ctx).Error
}

func (r *repositoryImpl) UpdateClaimedAmount(ctx context.Context, envelopeID int64, amount float64) (err error) {
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

	span.SetAttributes(
		attribute.Int64("envelopeID", envelopeID),
		attribute.Float64("claimedAmount", amount),
	)

	err = activeEnvelopeQuery(r.db.WithContext(ctx), envelopeID).
		Model(&e.Envelope{}).
		UpdateColumn("total_amount_claimed", gorm.Expr("total_amount_claimed + ?", amount)).
		Error

	return err
}

func (r *repositoryImpl) GetExpiredUnRefundedEnvelopes(ctx context.Context) (resp []*e.Envelope, err error) {
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
	var envelopes []*e.Envelope
	err = r.db.WithContext(ctx).
		Where("is_active = ? AND total_amount_refunded = ? AND expired_at IS NOT NULL AND expired_at < ?", true, 0, time.Now()).
		Find(&envelopes).Error
	return envelopes, err
}

func (r *repositoryImpl) GetExpiredUnRefundedEnvelopesByID(
	ctx context.Context, envelopeID int64, userID string,
) (resp *e.Envelope, err error) {
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

	var env e.Envelope
	err = r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("total_amount_refunded = ?", 0).
		Where("expired_at IS NOT NULL").
		Where("expired_at < ? ", time.Now()).
		Where("envelope_id = ?", envelopeID).
		Where("user_id = ?", userID).
		Find(&env).Error

	span.SetAttributes(
		attribute.Int64("envelopeID", envelopeID),
		attribute.String("userID", userID),
	)

	return &env, err
}

func (r *repositoryImpl) RefundEnvelope(ctx context.Context, envelopeID int64, refundAmount float64) error {
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
		attribute.Int64("envelopeID", envelopeID),
		attribute.Float64("refundAmount", refundAmount),
	)

	err = activeEnvelopeQuery(r.db.WithContext(ctx), envelopeID).
		Model(&e.Envelope{}).
		Updates(map[string]interface{}{
			"total_amount_refunded": refundAmount,
			"updated_at":            time.Now(),
		}).Error

	return err
}

func (r *repositoryImpl) DeactivateUnclaimedDetails(ctx context.Context, envelopeID int64) error {
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
		attribute.Int64("envelopeID", envelopeID),
	)

	err = activeEnvelopeQuery(r.db.WithContext(ctx), envelopeID).
		Model(&e.EnvelopeDetail{}).
		Where("(user_id IS NULL OR user_id = '') AND envelope_detail_status != ?", e.EnvelopeClaimed).
		Updates(map[string]interface{}{
			"envelope_detail_status": e.EnvelopeRefunded,
			"updated_at":             time.Now(),
			"user_id":                "system",
		}).Error
	return err
}

func (r *repositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo Repository) error) error {
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

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)
		return fn(txRepo)
	})

	return err
}

//nolint:revive // need to comeback later
func (r *repositoryImpl) WithTx(tx *gorm.DB) Repository {
	return &repositoryImpl{db: tx}
}

func (r *repositoryImpl) GetTx() *gorm.DB {
	return r.db
}

func (r *repositoryImpl) Update(ctx context.Context, envelope *e.Envelope, tx *gorm.DB) error {
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

	if tx != nil {
		err = tx.Save(envelope).WithContext(ctx).Error
		return err
	}

	err = r.db.Save(envelope).WithContext(ctx).Error
	return err
}

func (r *repositoryImpl) GetEnvelopeDetail(ctx context.Context, envelopeDetailID int64) (resp *e.EnvelopeDetail, err error) {
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

	span.SetAttributes(
		attribute.Int64("envelopeDetailID", envelopeDetailID),
	)

	var envelopeDetail e.EnvelopeDetail
	err = r.db.Where("envelope_detail_id = ? AND is_active = ?", envelopeDetailID, true).
		WithContext(ctx).
		First(&envelopeDetail).
		Error
	if err != nil {
		return nil, err
	}

	return &envelopeDetail, nil
}

func (r *repositoryImpl) GetEnvelopeDetailsByEnvelopID(ctx context.Context, envelopID int64) (resp []*e.EnvelopeDetail, err error) {
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

	var envelopeDetails []*e.EnvelopeDetail
	err = r.db.Where("envelope_id = ? AND is_active = ?", envelopID, true).
		WithContext(ctx).
		Find(&envelopeDetails).
		Error
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64("envelopeID", envelopID),
		attribute.Int("totalEnvelopeDetail", len(envelopeDetails)),
	)

	return envelopeDetails, nil
}

func (r *repositoryImpl) FetchExpiredEnvelopes(ctx context.Context, ids []int64) (resp []*e.Envelope, err error) {
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

	var envelopes []*e.Envelope
	q := r.db.WithContext(ctx).
		Where("expired_at < ?", time.Now()).
		Where("refunded_at IS NULL").
		Where("total_amount > total_amount_claimed").
		Where("is_active = ?", true)

	if len(ids) > 0 {
		q = q.Where("envelope_id IN ?", ids)
	}

	err = q.Find(&envelopes).Error
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64Slice("envelopeID", ids),
		attribute.Int("totalExpiredEnvelopes", len(envelopes)),
	)

	return envelopes, nil
}
