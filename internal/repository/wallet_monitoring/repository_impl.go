package wallet_monitoring

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type walletMonitoringRepositoryImpl struct {
	db *gorm.DB
}

func NewWalletMonitoringRepository(db *gorm.DB) WalletMonitoringRepository {
	return &walletMonitoringRepositoryImpl{
		db: db,
	}
}

// GetDashboardTransactionVolume TODO: implement caching
func (r *walletMonitoringRepositoryImpl) GetDashboardTransactionVolume(
	ctx context.Context,
) (resp []*entity.TransactionCount, err error) {
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

	var results []*entity.TransactionCount

	query := r.db.WithContext(ctx).
		Model(&entity.WalletTransaction{}).
		Select("transaction_type, entry_type, COUNT(*) as count").
		Where("is_active = ? AND deleted_at IS NULL", true).
		Group("transaction_type, entry_type")

	err = query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("total", len(results)))

	return results, nil
}

func (r *walletMonitoringRepositoryImpl) GetListTransactionMonitoring(
	ctx context.Context,
	req *domain.GetListTransactionMonitoringRequest,
) (resp []*entity.WalletTransaction, count int64, err error) {
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

	baseQuery := r.db.WithContext(ctx).
		Model(&entity.WalletTransaction{}).
		Where("is_active = ? AND deleted_at IS NULL", true)

	span.SetAttributes(
		attribute.String("fromTransactionDateReq", req.FromTransactionDate.Format(time.RFC3339)),
		attribute.String("toTransactionDateReq", req.ToTransactionDate.Format(time.RFC3339)),
		attribute.String("transactionTypeReq", req.TransactionType),
		attribute.String("entryTypeReq", req.EntryType),
		attribute.String("referenceCodeReq", req.ReferenceCode),
		attribute.Float64("fromAmountReq", req.FromAmount),
		attribute.Float64("toAmountReq", req.ToAmount),
		attribute.Int64("walletIDReq", req.WalletID),
	)

	query := applyFilters(baseQuery, req)

	var total int64
	countQuery := query.Session(&gorm.Session{})
	err = countQuery.Count(&total).Error
	if err != nil {
		return nil, 0, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to count transactions: %v", err)).Wrap()
	}

	if total == 0 {
		return nil, 0, nil
	}

	var transactions []*entity.WalletTransaction
	offset := (req.Page - 1) * req.Limit

	err = query.
		Order("transaction_date DESC, wallet_transaction_id DESC").
		Limit(int(req.Limit)).
		Offset(int(offset)).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to fetch transactions: %v", err)).Wrap()
	}

	span.SetAttributes(
		attribute.Int64("totalCount", total),
		attribute.Int("page", int(req.Page)),
		attribute.Int("limit", int(req.Limit)),
	)

	return transactions, total, nil
}

// applyFilters applies all filters to the query
func applyFilters(
	query *gorm.DB,
	req *domain.GetListTransactionMonitoringRequest,
) *gorm.DB {
	if !req.FromTransactionDate.IsZero() && !req.ToTransactionDate.IsZero() {
		query = query.Where("transaction_date BETWEEN ? AND ?",
			req.FromTransactionDate.Format("2006-01-02 00:00:00"),
			req.ToTransactionDate.Format("2006-01-02 23:59:59"))
	} else if !req.FromTransactionDate.IsZero() {
		query = query.Where("transaction_date >= ?", req.FromTransactionDate.Format("2006-01-02 00:00:00"))
	} else if !req.ToTransactionDate.IsZero() {
		query = query.Where("transaction_date <= ?", req.ToTransactionDate.Format("2006-01-02 23:59:59"))
	}

	if req.TransactionType != "" {
		query = query.Where("transaction_type = ?", req.TransactionType)
	}

	if req.EntryType != "" {
		query = query.Where("entry_type = ?", req.EntryType)
	}

	if req.FromAmount > 0 && req.ToAmount > 0 {
		query = query.Where("amount BETWEEN ? AND ?", req.FromAmount, req.ToAmount)
	} else if req.FromAmount > 0 {
		query = query.Where("amount >= ?", req.FromAmount)
	} else if req.ToAmount > 0 {
		query = query.Where("amount <= ?", req.ToAmount)
	}

	if req.WalletID > 0 {
		query = query.Where("wallet_id = ?", req.WalletID)
	}

	if req.ReferenceCode != "" {
		query = query.Where("reference_code = ?", req.ReferenceCode)
	}

	if req.WalletTransactionID != "" {
		query = query.Where("wallet_transaction_id = ?", req.WalletTransactionID)
	}

	return query
}

func (r *walletMonitoringRepositoryImpl) GetListEnvelope(
	ctx context.Context,
	req *domain.GetListEnvelopeRequest,
) (envs []*entity.Envelope, total int64, err error) {
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

	baseQuery := r.db.WithContext(ctx).
		Model(&entity.Envelope{}).
		Where("is_active = ? AND deleted_at IS NULL", true)

	span.SetAttributes(
		attribute.Int("page", int(req.Page)),
		attribute.Int("limit", int(req.Limit)),
		attribute.Float64("fromTotalAmount", req.FromTotalAmount),
		attribute.Float64("toTotalAmount", req.ToTotalAmount),
		attribute.String("userID", req.UserID),
		attribute.String("envelopeType", req.EnvelopeType),
		attribute.Int("maxNumReceived", int(req.MaxNumReceived)),
		attribute.Int64("fromExpiredAt", req.FromExpiredAt),
		attribute.Int64("toExpiredAt", req.ToExpiredAt),
		attribute.Int64("fromRefundedAt", req.FromRefundedAt),
		attribute.Int64("toRefundedAt", req.ToRefundedAt),
	)

	query := applyEnvelopeFilters(baseQuery, req)

	countQuery := query.Session(&gorm.Session{})
	err = countQuery.Count(&total).Error
	if err != nil {
		return nil, 0, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to count envelopes: %v", err)).Wrap()
	}

	if total == 0 {
		return nil, 0, nil
	}

	var envelopes []*entity.Envelope
	offset := (req.Page - 1) * req.Limit

	err = query.
		Order("created_at DESC, envelope_id DESC").
		Limit(int(req.Limit)).
		Offset(int(offset)).
		Find(&envelopes).Error

	if err != nil {
		return nil, 0, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to fetch envelopes: %v", err)).Wrap()
	}

	span.SetAttributes(
		attribute.Int64("total", total),
	)

	return envelopes, total, nil
}

// applyEnvelopeFilters applies all filters to the query
func applyEnvelopeFilters(query *gorm.DB, req *domain.GetListEnvelopeRequest) *gorm.DB {
	query = applyUserEnvelopeFilters(query, req)
	query = applyAmountEnvelopeFilters(query, req)
	query = applyExpiredEnvelopeFilters(query, req)
	query = applyRefundedEnvelopeFilters(query, req)

	return query
}

// applyEnvelopeFilters applies all filters to the query
func applyUserEnvelopeFilters(query *gorm.DB, req *domain.GetListEnvelopeRequest) *gorm.DB {
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	if req.MaxNumReceived > 0 {
		query = query.Where("max_num_received = ?", req.MaxNumReceived)
	}

	if req.EnvelopeType != "" {
		query = query.Where("envelope_type = ?", req.EnvelopeType)
	}

	return query
}

// applyEnvelopeFilters applies all filters to the query
func applyAmountEnvelopeFilters(query *gorm.DB, req *domain.GetListEnvelopeRequest) *gorm.DB {
	if req.FromTotalAmount > 0 && req.ToTotalAmount > 0 {
		query = query.Where("total_amount BETWEEN ? AND ?", req.FromTotalAmount, req.ToTotalAmount)
	} else if req.FromTotalAmount > 0 {
		query = query.Where("total_amount >= ?", req.FromTotalAmount)
	} else if req.ToTotalAmount > 0 {
		query = query.Where("total_amount <= ?", req.ToTotalAmount)
	}

	return query
}

// applyEnvelopeFilters applies all filters to the query
func applyExpiredEnvelopeFilters(query *gorm.DB, req *domain.GetListEnvelopeRequest) *gorm.DB {
	return buildTimeRangeFilter(query, "expired_at", req.FromExpiredAt, req.ToExpiredAt)
}

// applyEnvelopeFilters applies all filters to the query
func applyRefundedEnvelopeFilters(query *gorm.DB, req *domain.GetListEnvelopeRequest) *gorm.DB {
	return buildTimeRangeFilter(query, "refunded_at", req.FromRefundedAt, req.ToRefundedAt)
}

func (r *walletMonitoringRepositoryImpl) GetEnvelopeDetail(
	ctx context.Context,
	req *domain.GetEnvelopeDetailRequest,
) (env *entity.Envelope, envDetails []*entity.EnvelopeDetail, err error) {
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

	var (
		envelope *entity.Envelope
		details  []*entity.EnvelopeDetail
	)

	span.SetAttributes(attribute.Int64("envelopeIDReq", req.EnvelopeID))

	baseQuery := r.db.WithContext(ctx).
		Where("envelope_id = ? AND is_active = ? AND deleted_at IS NULL", req.EnvelopeID, true)

	err = baseQuery.First(&envelope).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errs.ErrArgs.WithDetail(fmt.Sprintf("envelope not found: %v", err)).Wrap()
			return envelope, details, err
		}
		err = errs.ErrArgs.WithDetail(fmt.Sprintf("failed to fetch envelope: %v", err)).Wrap()
		return envelope, details, err
	}

	detailsQuery := r.applyEnvelopeDetailsQuery(ctx, req)

	err = detailsQuery.
		Order("created_at DESC, envelope_detail_id DESC").
		Find(&details).Error
	if err != nil {
		err = errs.ErrArgs.WithDetail(fmt.Sprintf("failed to fetch envelope details: %v", err)).Wrap()
		return envelope, details, err
	}

	return envelope, details, nil
}

// applyEnvelopeDetailsQuery builds the query for envelope details with filters
func (r *walletMonitoringRepositoryImpl) applyEnvelopeDetailsQuery(
	ctx context.Context,
	req *domain.GetEnvelopeDetailRequest,
) *gorm.DB {
	var (
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelRepository)
	)

	ctx, span := t.Start(ctx, funcName)
	defer func() {
		span.End()
	}()

	span.SetAttributes(attribute.Int64("envelopeIDReq", req.EnvelopeID))

	query := r.db.WithContext(ctx).
		Model(&entity.EnvelopeDetail{}).
		Where("envelope_id = ? AND is_active = ? AND deleted_at IS NULL", req.EnvelopeID, true)

	if req.DetailStatus != "" {
		query = query.Where("envelope_detail_status = ?", req.DetailStatus)
		span.SetAttributes(attribute.String("detailStatus", req.DetailStatus))
	}

	if req.Claimed != nil {
		if *req.Claimed {
			query = query.Where("claimed_at IS NOT NULL")
		} else {
			query = query.Where("claimed_at IS NULL")
		}
	}

	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
		span.SetAttributes(attribute.String("userID", req.UserID))
	}

	return query
}

func (r *walletMonitoringRepositoryImpl) GetTransferHistory(
	ctx context.Context,
	req *domain.GetTransferHistoryRequest,
) (resp []*entity.Transfer, count int64, err error) {
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

	baseQuery := r.db.WithContext(ctx).
		Model(&entity.Transfer{}).
		Where("is_active = ? AND deleted_at IS NULL", true)

	query := applyTransferFilters(baseQuery, req)

	var total int64
	countQuery := query.Session(&gorm.Session{})
	err = countQuery.Count(&total).Error
	if err != nil {
		return nil, 0, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to count transfers: %v", err)).Wrap()
	}

	if total == 0 {
		return nil, 0, nil
	}

	var transfers []*entity.Transfer
	offset := (req.Page - 1) * req.Limit

	err = query.
		Order("created_at DESC, transfer_id DESC").
		Limit(int(req.Limit)).
		Offset(int(offset)).
		Find(&transfers).Error

	if err != nil {
		return nil, 0, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to fetch transfers: %v", err)).Wrap()
	}

	span.SetAttributes(attribute.Int64("total", total))

	return transfers, total, nil
}

// applyTransferFilters applies all filters to the query
func applyTransferFilters(query *gorm.DB, req *domain.GetTransferHistoryRequest) *gorm.DB {
	query = applyUserFilters(query, req)
	query = applyAmountFilters(query, req)
	query = applyExpiredFilters(query, req)
	query = applyClaimedFilters(query, req)

	return query
}

// applyTransferFilters applies all filters to the query
func applyAmountFilters(query *gorm.DB, req *domain.GetTransferHistoryRequest) *gorm.DB {
	if req.FromAmount > 0 && req.ToAmount > 0 {
		query = query.Where("amount BETWEEN ? AND ?", req.FromAmount, req.ToAmount)
	} else if req.FromAmount > 0 {
		query = query.Where("amount >= ?", req.FromAmount)
	} else if req.ToAmount > 0 {
		query = query.Where("amount <= ?", req.ToAmount)
	}

	return query
}

// applyTransferFilters applies all filters to the query
func applyExpiredFilters(query *gorm.DB, req *domain.GetTransferHistoryRequest) *gorm.DB {
	return buildTimeRangeFilter(query, "expired_at", req.FromExpiredAt, req.ToExpiredAt)
}

// applyTransferFilters applies all filters to the query
func applyClaimedFilters(query *gorm.DB, req *domain.GetTransferHistoryRequest) *gorm.DB {
	return buildTimeRangeFilter(query, "claimed_at", req.FromClaimedAt, req.ToClaimedAt)
}

func buildTimeRangeFilter(query *gorm.DB, column string, from, to int64) *gorm.DB {
	if from > 0 && to > 0 {
		fromTime := time.Unix(from, 0)
		toTime := time.Unix(to, 0)
		query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", column), fromTime, toTime)
	} else if from > 0 {
		fromTime := time.Unix(from, 0)
		query = query.Where(fmt.Sprintf("%s >= ?", column), fromTime)
	} else if to > 0 {
		toTime := time.Unix(to, 0)
		query = query.Where(fmt.Sprintf("%s <= ?", column), toTime)
	}

	return query
}

// applyTransferFilters applies all filters to the query
func applyUserFilters(
	query *gorm.DB,
	req *domain.GetTransferHistoryRequest,
) *gorm.DB {
	if req.FromUserID != "" {
		query = query.Where("from_user_id = ?", req.FromUserID)
	}

	if req.ToUserID != "" {
		query = query.Where("to_user_id = ?", req.ToUserID)
	}

	if req.StatusTransfer != "" {
		query = query.Where("status_transfer = ?", req.StatusTransfer)
	}

	return query
}

func (r *walletMonitoringRepositoryImpl) GetTop10Users(
	ctx context.Context,
	req *domain.GetTop10UsersRequest,
) (users []*entity.UserStatResult, s string, err error) {
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

	ctx, cancel := context.WithTimeout(ctx, time.Duration(domain.TimeoutContextIn30Seconds)*time.Second)
	defer cancel()

	// Build time condition and value
	var (
		timeCondition string
		timeValue     string
		queryArgs     []interface{}
	)
	timeCondition, timeValue, queryArgs, err = buildTimeCondition(req.Scope)
	if err != nil {
		return nil, "", errs.ErrArgs.WithDetail(fmt.Sprintf("failed to build time condition: %v", err)).Wrap()
	}

	// Build and execute query
	query := applyTop10UsersQuery(timeCondition, req.SortByTotalAmount)
	span.SetAttributes(
		attribute.Bool("sortByTotalAmount", req.SortByTotalAmount),
		attribute.String("scope", req.Scope),
	)

	var results []*entity.UserStatResult

	dbQuery := r.db.WithContext(ctx).Raw(query, queryArgs...)
	err = dbQuery.Scan(&results).Error
	if err != nil {
		return nil, "", errs.ErrArgs.WithDetail(fmt.Sprintf("failed to execute query: %v", err)).Wrap()
	}

	return results, timeValue, nil
}

// buildTimeCondition constructs the WHERE clause and parameters based on scope
func buildTimeCondition(scope string) (q, s string, c []interface{}, e error) {
	now := time.Now()

	switch scope {
	case domain.ScopeAllTime.String():
		return "", "all_time", []interface{}{}, nil

	case domain.ScopeYear.String():
		startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return "AND transaction_date >= ?", fmt.Sprintf("%d", now.Year()), []interface{}{startOfYear}, nil

	case domain.ScopeMonth.String():
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		return "AND transaction_date >= ? AND transaction_date < ?",
			now.Format("Jan"),
			[]interface{}{startOfMonth, endOfMonth}, nil

	default:
		return "", "", nil, errs.ErrArgs.WithDetail(fmt.Sprintf("unsupported scope: %s", scope)).Wrap()
	}
}

// applyTop10UsersQuery constructs the SQL query for getting top 10 users
func applyTop10UsersQuery(timeCondition string, sortByTotalAmount bool) (q string) {
	orderBy := "transaction_frequency"
	if sortByTotalAmount {
		orderBy = "(total_credit + total_debit)"
	}

	return fmt.Sprintf(`
		WITH user_stats AS (
			SELECT 
				wallet_id,
				SUM(CASE WHEN entry_type = 'credit' THEN amount ELSE 0 END) as total_credit,
				SUM(CASE WHEN entry_type = 'debit' THEN amount ELSE 0 END) as total_debit,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'transfer' THEN 1 ELSE 0 END) as credit_transfer,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'envelope_fixed' THEN 1 ELSE 0 END) as credit_envelope_fixed,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'envelope_lucky' THEN 1 ELSE 0 END) as credit_envelope_lucky,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'envelope_single' THEN 1 ELSE 0 END) as credit_envelope_single,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'deposit' THEN 1 ELSE 0 END) as credit_deposit,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'refund_envelope' THEN 1 ELSE 0 END) as credit_refund_envelope,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'refund_transfer' THEN 1 ELSE 0 END) as credit_refund_transfer,
				SUM(CASE WHEN entry_type = 'credit' AND transaction_type = 'system_adjustment' THEN 1 ELSE 0 END) as credit_system_adjustment,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'transfer' THEN 1 ELSE 0 END) as debit_transfer,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'envelope_fixed' THEN 1 ELSE 0 END) as debit_envelope_fixed,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'envelope_lucky' THEN 1 ELSE 0 END) as debit_envelope_lucky,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'envelope_single' THEN 1 ELSE 0 END) as debit_envelope_single,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'deposit' THEN 1 ELSE 0 END) as debit_deposit,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'refund_envelope' THEN 1 ELSE 0 END) as debit_refund_envelope,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'refund_transfer' THEN 1 ELSE 0 END) as debit_refund_transfer,
				SUM(CASE WHEN entry_type = 'debit' AND transaction_type = 'system_adjustment' THEN 1 ELSE 0 END) as debit_system_adjustment,
				COUNT(*) as transaction_frequency
			FROM wallet_transactions 
			WHERE is_active = true AND deleted_at IS NULL %s
			GROUP BY wallet_id
		)
		SELECT * FROM user_stats
		ORDER BY %s DESC
		LIMIT 10
	`, timeCondition, orderBy)
}
