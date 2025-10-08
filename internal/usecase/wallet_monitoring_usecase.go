package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/internal/repository/wallet_monitoring"
)

type (
	WalletMonitoringSvcImpl struct {
		repo wallet_monitoring.WalletMonitoringRepository
	}

	WalletMonitoringSvc interface {
		GetDashboardTransactionVolume(
			ctx context.Context,
		) (resp *domain.DashboardTransactionVolumeResponse, err error)
		GetListTransactionMonitoring(
			ctx context.Context,
			req *domain.GetListTransactionMonitoringRequest,
		) (resp *domain.GetListTransactionMonitoringResponse, err error)
		GetListEnvelope(
			ctx context.Context,
			req *domain.GetListEnvelopeRequest,
		) (resp *domain.GetListEnvelopeResponse, err error)
		GetEnvelopeDetail(
			ctx context.Context,
			req *domain.GetEnvelopeDetailRequest,
		) (resp *domain.GetEnvelopeDetailResponse, err error)
		GetTransferHistory(
			ctx context.Context,
			req *domain.GetTransferHistoryRequest,
		) (resp *domain.GetTransferHistoryResponse, err error)
		GetTop10Users(
			ctx context.Context,
			req *domain.GetTop10UsersRequest,
		) (resp *domain.GetTop10UsersResponse, err error)
	}
)

func NewWalletMonitoringUseCase(repo wallet_monitoring.WalletMonitoringRepository) WalletMonitoringSvc {
	return &WalletMonitoringSvcImpl{
		repo: repo,
	}
}

func (u *WalletMonitoringSvcImpl) GetDashboardTransactionVolume(
	ctx context.Context,
) (resp *domain.DashboardTransactionVolumeResponse, err error) {
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

	var trxCountData []*entity.TransactionCount
	trxCountData, err = u.repo.GetDashboardTransactionVolume(ctx)
	if err != nil {
		return resp, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to get transaction volume data: %v", err)).Wrap()
	}
	span.SetAttributes(attribute.Int("entryTypeCount", len(trxCountData)))

	resp = &domain.DashboardTransactionVolumeResponse{
		Credit: domain.TransactionTypeCount{},
		Debit:  domain.TransactionTypeCount{},
	}

	entryTypeMap := map[string]*domain.TransactionTypeCount{
		entity.EntryTypeCredit.String(): &resp.Credit,
		entity.EntryTypeDebit.String():  &resp.Debit,
	}

	for _, data := range trxCountData {
		if ttc, exists := entryTypeMap[data.EntryType]; exists {
			setTransactionTypeCount(ttc, data.TransactionType, data.Count)
		}
	}

	return resp, nil
}

func (u *WalletMonitoringSvcImpl) GetListTransactionMonitoring(
	ctx context.Context,
	req *domain.GetListTransactionMonitoringRequest,
) (resp *domain.GetListTransactionMonitoringResponse, err error) {
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

	ctx, cancel := context.WithTimeout(ctx, time.Duration(domain.TimeoutContextIn30Seconds)*time.Second)
	defer cancel()

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

	var (
		transactions []*entity.WalletTransaction
		total        int64
	)
	transactions, total, err = u.repo.GetListTransactionMonitoring(ctx, req)
	if err != nil {
		return resp, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to get transaction monitoring data: %v", err)).Wrap()
	}

	result := make([]*domain.WalletTransactionDTO, 0, len(transactions))

	for _, tx := range transactions {
		dto := convertToDTO(tx)
		result = append(result, dto)
	}

	totalPages := int32(math.Ceil(float64(total) / float64(req.Limit)))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	resp = &domain.GetListTransactionMonitoringResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalCount: total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Data:       result,
	}
	span.SetAttributes(
		attribute.Int64("totalCount", total),
		attribute.Int("page", int(req.Page)),
		attribute.Int("limit", int(req.Limit)),
	)

	return resp, nil
}

func (u *WalletMonitoringSvcImpl) GetListEnvelope(
	ctx context.Context,
	req *domain.GetListEnvelopeRequest,
) (resp *domain.GetListEnvelopeResponse, err error) {
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

	ctx, cancel := context.WithTimeout(ctx, time.Duration(domain.TimeoutContextIn30Seconds)*time.Second)
	defer cancel()

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

	var (
		envelopes []*entity.Envelope
		total     int64
	)
	envelopes, total, err = u.repo.GetListEnvelope(ctx, req)
	if err != nil {
		return resp, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to get envelope list: %v", err)).Wrap()
	}

	result := make([]*domain.EnvelopeDTO, 0, len(envelopes))

	for _, env := range envelopes {
		dto := convertEnvelopeToDTO(env)
		result = append(result, dto)
	}

	totalPages := int32(math.Ceil(float64(total) / float64(req.Limit)))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	resp = &domain.GetListEnvelopeResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalCount: total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Data:       result,
	}

	span.SetAttributes(attribute.Int64("total", total))
	return resp, nil
}

func (u *WalletMonitoringSvcImpl) GetEnvelopeDetail(
	ctx context.Context,
	req *domain.GetEnvelopeDetailRequest,
) (resp *domain.GetEnvelopeDetailResponse, err error) {
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

	ctx, cancel := context.WithTimeout(ctx, time.Duration(domain.TimeoutContextIn30Seconds)*time.Second)
	defer cancel()

	span.SetAttributes(attribute.Int64("envelopeIDReq", req.EnvelopeID))

	var (
		envelope *entity.Envelope
		details  []*entity.EnvelopeDetail
	)
	envelope, details, err = u.repo.GetEnvelopeDetail(ctx, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, errs.ErrArgs.WithDetail("envelope ID not found").Wrap()
		}
		return resp, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to get envelope detail: %v", err)).Wrap()
	}

	detailsDTO := make([]*domain.EnvelopeDetailDTO, 0, len(details))

	// Convert details with enriched information
	for _, detail := range details {
		dto := convertToEnvelopeDetailDTO(detail)
		detailsDTO = append(detailsDTO, dto)
	}

	// Sort details by created_at DESC for consistent ordering
	sort.Slice(detailsDTO, func(i, j int) bool {
		return detailsDTO[i].CreatedAt.After(detailsDTO[j].CreatedAt)
	})

	response := &domain.GetEnvelopeDetailResponse{
		EnvelopeID:          envelope.EnvelopeID,
		UserID:              envelope.UserID,
		TotalAmount:         envelope.TotalAmount,
		TotalAmountClaimed:  envelope.TotalAmountClaimed,
		TotalAmountRefunded: envelope.TotalAmountRefunded,
		MaxNumReceived:      envelope.MaxNumReceived,
		EnvelopeType:        envelope.EnvelopeType,
		Remarks:             envelope.Remarks,
		ExpiredAt:           envelope.ExpiredAt,
		RefundedAt:          envelope.RefundedAt,
		CreatedAt:           envelope.CreatedAt,
		Details:             detailsDTO,
	}

	enrichEnvelopeDetailResponse(response)

	return response, nil
}

func (u *WalletMonitoringSvcImpl) GetTransferHistory(
	ctx context.Context,
	req *domain.GetTransferHistoryRequest,
) (resp *domain.GetTransferHistoryResponse, err error) {
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

	ctx, cancel := context.WithTimeout(ctx, domain.TimeoutContextIn30Seconds*time.Second)
	defer cancel()

	span.SetAttributes(
		attribute.Int("pageReq", int(req.Page)),
		attribute.Int("limitReq", int(req.Limit)),
		attribute.String("fromUserIDReq", req.FromUserID),
		attribute.String("toUserIDReq", req.ToUserID),
		attribute.String("statusTransferReq", req.StatusTransfer),
		attribute.Float64("fromAmountReq", req.FromAmount),
		attribute.Float64("toAmountReq", req.ToAmount),
		attribute.Int64("fromExpiredAtReq", req.FromExpiredAt),
		attribute.Int64("toExpiredAtReq", req.ToExpiredAt),
		attribute.Int64("fromClaimedAtReq", req.FromClaimedAt),
		attribute.Int64("toClaimedAtReq", req.ToClaimedAt),
	)

	var (
		transfers []*entity.Transfer
		total     int64
	)
	transfers, total, err = u.repo.GetTransferHistory(ctx, req)
	if err != nil {
		return resp, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to get transfer history: %v", err)).Wrap()
	}

	result := make([]*domain.TransferDTO, 0, len(transfers))

	for _, transfer := range transfers {
		dto := convertTransferToDTO(transfer)
		result = append(result, dto)
	}

	totalPages := int32(math.Ceil(float64(total) / float64(req.Limit)))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	statistics := calculateTransferStatistics(result)

	resp = &domain.GetTransferHistoryResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalCount: total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Statistics: statistics,
		Data:       result,
	}

	span.SetAttributes(
		attribute.Int64("totalCount", total),
		attribute.Int("page", int(req.Page)),
		attribute.Int("limit", int(req.Limit)),
	)

	return resp, nil
}

func (u *WalletMonitoringSvcImpl) GetTop10Users(
	ctx context.Context,
	req *domain.GetTop10UsersRequest,
) (resp *domain.GetTop10UsersResponse, err error) {
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
		attribute.Bool("sortByTotalAmount", req.SortByTotalAmount),
		attribute.String("scope", req.Scope),
	)

	var (
		results   []*entity.UserStatResult
		timeValue string
	)
	results, timeValue, err = u.repo.GetTop10Users(ctx, req)
	if err != nil {
		return nil, errs.ErrArgs.WithDetail(fmt.Sprintf("failed to get top 10 users: %v", err)).Wrap()
	}

	listTransaction := transformToHighFrequencyUserDTOs(results)

	return &domain.GetTop10UsersResponse{
		Time:            timeValue,
		ListTransaction: listTransaction,
	}, nil
}

// used by useCase GetDashboardTransactionVolume
func setTransactionTypeCount(ttc *domain.TransactionTypeCount, transactionType string, count int64) {
	// Use map for O(1) lookup instead of switch statements
	transactionTypeSetters := map[string]func(*domain.TransactionTypeCount, int64){
		entity.TransactionTypeTransfer.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.Transfer = c
		},
		entity.TransactionTypeEnvelopeFixed.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.EnvelopeFixed = c
		},
		entity.TransactionTypeEnvelopeLucky.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.EnvelopeLucky = c
		},
		entity.TransactionTypeEnvelopeSingle.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.EnvelopeSingle = c
		},
		entity.TransactionTypeRefundEnvelope.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.RefundEnvelope = c
		},
		entity.TransactionTypeRefundTransfer.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.RefundTransfer = c
		},
		entity.TransactionTypeSystemAdjustment.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.SystemAdjustment = c
		},
		entity.TransactionTypeDeposit.String(): func(t *domain.TransactionTypeCount, c int64) {
			t.Deposit = c
		},
	}

	if setter, exists := transactionTypeSetters[transactionType]; exists {
		setter(ttc, count)
	}
}

// used by useCase GetListTransactionMonitoring
// convertToDTO converts entity to DTO with proper error handling
func convertToDTO(tx *entity.WalletTransaction) (resp *domain.WalletTransactionDTO) {
	return &domain.WalletTransactionDTO{
		WalletTransactionID: tx.WalletTransactionID,
		WalletID:            tx.WalletID,
		Amount:              tx.Amount,
		TransactionType:     tx.TransactionType.String(),
		EntryType:           tx.EntryType.String(),
		BeforeBalance:       tx.BeforeBalance,
		AfterBalance:        tx.AfterBalance,
		ReferenceCode:       tx.ReferenceCode,
		TransactionDate:     tx.TransactionDate,
		DescriptionEN:       tx.DescriptionEN,
		DescriptionZH:       tx.DescriptionZH,
		ImpactedItem:        tx.ImpactedItem,
		CreatedAt:           tx.CreatedAt,
	}
}

// used by useCase GetListEnvelope
// convertEnvelopeToDTO converts entity to DTO with proper handling
func convertEnvelopeToDTO(env *entity.Envelope) (resp *domain.EnvelopeDTO) {
	dto := &domain.EnvelopeDTO{
		EnvelopeID:          env.EnvelopeID,
		UserID:              env.UserID,
		TotalAmount:         env.TotalAmount,
		TotalAmountClaimed:  env.TotalAmountClaimed,
		TotalAmountRefunded: env.TotalAmountRefunded,
		MaxNumReceived:      env.MaxNumReceived,
		EnvelopeType:        env.EnvelopeType,
		Remarks:             env.Remarks,
		ExpiredAt:           env.ExpiredAt,
		RefundedAt:          env.RefundedAt,
		CreatedAt:           env.CreatedAt,
	}

	// Calculate derived fields
	dto.RemainingAmount = env.TotalAmount - env.TotalAmountClaimed - env.TotalAmountRefunded
	dto.IsExpired = env.ExpiredAt != nil && env.ExpiredAt.Before(time.Now())
	dto.IsRefunded = env.RefundedAt != nil
	dto.Status = calculateEnvelopeStatus(env)

	return dto
}

const (
	EnvelopeStatusActive       = "active"
	EnvelopeStatusClaimed      = "claimed"
	EnvelopeStatusRefunded     = "refunded"
	EnvelopeStatusExpired      = "expired"
	EnvelopeDetailStatusActive = "active"
)

// calculateEnvelopeStatus determines envelope status based on its state
func calculateEnvelopeStatus(env *entity.Envelope) (status string) {
	now := time.Now()

	if env.RefundedAt != nil {
		return EnvelopeStatusRefunded
	}

	if env.ExpiredAt != nil && env.ExpiredAt.Before(now) {
		return EnvelopeStatusExpired
	}

	if env.TotalAmountClaimed >= env.TotalAmount {
		return EnvelopeStatusClaimed
	}

	return EnvelopeStatusActive
}

// used by useCase GetEnvelopeDetail
// convertToEnvelopeDetailDTO converts entity to DTO with enriched data
func convertToEnvelopeDetailDTO(detail *entity.EnvelopeDetail) (e *domain.EnvelopeDetailDTO) {
	dto := &domain.EnvelopeDetailDTO{
		EnvelopeDetailID:     detail.EnvelopeDetailID,
		EnvelopeID:           detail.EnvelopeID,
		Amount:               detail.Amount,
		UserID:               detail.UserID,
		EnvelopeDetailStatus: detail.EnvelopeDetailStatus.String(),
		ClaimedAt:            detail.ClaimedAt,
		CreatedAt:            detail.CreatedAt,
	}

	// Add calculated fields
	dto.IsClaimed = detail.ClaimedAt != nil
	if detail.ClaimedAt != nil {
		timeSinceClaimed := time.Since(*detail.ClaimedAt)
		dto.TimeSinceClaimed = &timeSinceClaimed
	}

	return dto
}

// enrichEnvelopeDetailResponse adds calculated fields and metadata to response
func enrichEnvelopeDetailResponse(response *domain.GetEnvelopeDetailResponse) {
	response.RemainingAmount = response.TotalAmount - response.TotalAmountClaimed - response.TotalAmountRefunded
	response.IsExpired = response.ExpiredAt != nil && response.ExpiredAt.Before(time.Now())
	response.IsRefunded = response.RefundedAt != nil
	response.Status = calculateDetailEnvelopeStatus(response)
	response.CompletionPercentage = calculateCompletionPercentage(response)
	response.Statistics = calculateEnvelopeStatistics(response.Details)

	// Add time-based information
	if response.ExpiredAt != nil {
		timeUntilExpiry := time.Until(*response.ExpiredAt)
		response.TimeUntilExpiry = &timeUntilExpiry
	}
}

// calculateEnvelopeStatus determines the current status of the envelope
func calculateDetailEnvelopeStatus(response *domain.GetEnvelopeDetailResponse) (status string) {
	now := time.Now()
	if response.TotalAmountClaimed == response.TotalAmount {
		return entity.EnvelopeClaimed.String()
	}
	if response.ExpiredAt != nil && response.ExpiredAt.Before(now) {
		return entity.EnvelopeExpired.String()
	}
	if response.TotalAmountClaimed == 0 {
		return EnvelopeDetailStatusActive
	}
	return entity.EnvelopePartial.String()
}

// calculateCompletionPercentage calculates the completion percentage
func calculateCompletionPercentage(response *domain.GetEnvelopeDetailResponse) (percentage float64) {
	var percentageFactor float64 = 100
	if response.TotalAmount == 0 {
		return 0
	}
	return (response.TotalAmountClaimed / response.TotalAmount) * percentageFactor
}

// calculateEnvelopeStatistics calculates statistics from envelope details
func calculateEnvelopeStatistics(details []*domain.EnvelopeDetailDTO) (stat *domain.EnvelopeStatistics) {
	stats := &domain.EnvelopeStatistics{}

	uniqueUsers := make(map[string]bool)

	for _, detail := range details {
		stats.TotalDetails++
		uniqueUsers[detail.UserID] = true

		switch strings.ToLower(detail.EnvelopeDetailStatus) {
		case "claimed":
			stats.ClaimedCount++
		case "pending":
			stats.PendingCount++
		case "expired":
			stats.ExpiredCount++
		case "refunded":
			stats.RefundedCount++
		}

		if detail.IsClaimed {
			stats.TotalClaimedAmount += detail.Amount
		}
	}

	stats.UniqueClaimantsCount = len(uniqueUsers)

	return stats
}

// used by useCase GetTransferHistory
// convertTransferToDTO converts entity to DTO with enriched data
func convertTransferToDTO(transfer *entity.Transfer) (resp *domain.TransferDTO) {
	dto := &domain.TransferDTO{
		TransferID:     transfer.TransferID,
		FromUserID:     transfer.FromUserID,
		ToUserID:       transfer.ToUserID,
		Amount:         transfer.Amount,
		StatusTransfer: transfer.StatusTransfer.String(),
		Remark:         transfer.Remark,
		ExpiredAt:      transfer.ExpiredAt,
		RefundedAt:     transfer.RefundedAt,
		ClaimedAt:      transfer.ClaimedAt,
		CreatedAt:      transfer.CreatedAt,
	}

	// Add calculated fields
	dto.IsClaimed = transfer.ClaimedAt != nil
	dto.IsExpired = transfer.ExpiredAt != nil && transfer.ExpiredAt.Before(time.Now())
	dto.IsRefunded = transfer.RefundedAt != nil
	dto.IsPending = dto.StatusTransfer == entity.StatusTransferPending.String()

	// Calculate time-based fields
	if transfer.ClaimedAt != nil {
		timeSinceClaimed := time.Since(*transfer.ClaimedAt)
		dto.TimeSinceClaimed = &timeSinceClaimed
	}

	if transfer.ExpiredAt != nil && !dto.IsExpired {
		timeUntilExpiry := time.Until(*transfer.ExpiredAt)
		dto.TimeUntilExpiry = &timeUntilExpiry
	}

	// Calculate processing time
	if transfer.ClaimedAt != nil {
		processingTime := transfer.ClaimedAt.Sub(transfer.CreatedAt)
		dto.ProcessingTime = &processingTime
	} else if transfer.RefundedAt != nil {
		processingTime := transfer.RefundedAt.Sub(transfer.CreatedAt)
		dto.ProcessingTime = &processingTime
	}

	return dto
}

// calculateTransferStatistics calculates statistics from transfers
func calculateTransferStatistics(transfers []*domain.TransferDTO) (resp *domain.TransferStatistics) {
	stats := &domain.TransferStatistics{}

	uniqueSenders := make(map[string]bool)
	uniqueReceivers := make(map[string]bool)

	for _, transfer := range transfers {
		stats.TotalTransfers++
		stats.TotalAmount += transfer.Amount

		uniqueSenders[transfer.FromUserID] = true
		uniqueReceivers[transfer.ToUserID] = true

		switch strings.ToLower(transfer.StatusTransfer) {
		case entity.StatusTransferPending.String():
			stats.PendingCount++
			stats.PendingAmount += transfer.Amount
		case entity.StatusTransferClaimed.String():
			stats.ClaimedCount++
			stats.ClaimedAmount += transfer.Amount
		case entity.StatusTransferRefunded.String():
			stats.RefundedCount++
			stats.RefundedAmount += transfer.Amount
		}
	}

	stats.UniqueSendersCount = len(uniqueSenders)
	stats.UniqueReceiversCount = len(uniqueReceivers)

	if stats.TotalTransfers > 0 {
		stats.AverageAmount = stats.TotalAmount / float64(stats.TotalTransfers)
	}

	return stats
}

// used by useCase GetTop10Users
// transformToHighFrequencyUserDTOs converts entity results to DTOs
func transformToHighFrequencyUserDTOs(
	results []*entity.UserStatResult,
) (resp []*domain.HighFrequencyUserDTO) {
	listTransaction := make([]*domain.HighFrequencyUserDTO, 0, len(results))

	for i, result := range results {
		dto := &domain.HighFrequencyUserDTO{
			WalletID:    result.WalletID,
			TotalCredit: result.TotalCredit,
			Credit: domain.TransactionTypeCount{
				Transfer:         result.CreditTransfer,
				EnvelopeFixed:    result.CreditEnvelopeFixed,
				EnvelopeLucky:    result.CreditEnvelopeLucky,
				EnvelopeSingle:   result.CreditEnvelopeSingle,
				Deposit:          result.DebitDeposit,
				RefundEnvelope:   result.CreditRefundEnvelope,
				RefundTransfer:   result.CreditRefundTransfer,
				SystemAdjustment: result.CreditSystemAdjustment,
			},
			TotalDebit: result.TotalDebit,
			Debit: domain.TransactionTypeCount{
				Transfer:         result.DebitTransfer,
				EnvelopeFixed:    result.DebitEnvelopeFixed,
				EnvelopeLucky:    result.DebitEnvelopeLucky,
				EnvelopeSingle:   result.DebitEnvelopeSingle,
				Deposit:          result.DebitDeposit,
				RefundEnvelope:   result.DebitRefundEnvelope,
				RefundTransfer:   result.DebitRefundTransfer,
				SystemAdjustment: result.DebitSystemAdjustment,
			},
			Position: i + 1,
		}
		listTransaction = append(listTransaction, dto)
	}

	return listTransaction
}
