package http

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/1nterdigital/aka-im-tools/apiresp"
	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

// GetDashboardTransactionVolume retrieves dashboard transaction volume data
//
// @Summary Get dashboard transaction volume
// @Description Retrieve transaction volume statistics for dashboard display
// @Tags Wallet Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} domain.DashboardTransactionVolumeResponse "Successfully retrieved transaction volume data"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet-monitoring/transaction-volume [get]
// @Security ApiKeyAuth
func (h *WalletHandler) GetDashboardTransactionVolume(c *gin.Context) {
	var (
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while GetDashboardTransactionVolume", err)
		}
		span.End()
	}()

	var result *domain.DashboardTransactionVolumeResponse
	result, err = h.walletMonitoringUsecase.GetDashboardTransactionVolume(ctx)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}

// GetListTransactionMonitoring retrieves a paginated list of transaction monitoring data
//
// @Summary Get transaction monitoring list
// @Description Retrieve a filtered and paginated list of transactions for monitoring purposes
// @Tags Wallet Monitoring
// @Accept json
// @Produce json
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Number of items per page" minimum(1) default(10)
// @Param from_transaction_date query string false "Start date for transaction filter (RFC3339 format)" format(date-time)
// @Param to_transaction_date query string false "End date for transaction filter (RFC3339 format)" format(date-time)
// @Param transaction_type query string false "Type of transaction" Enums(transfer, deposit, withdrawal)
// @Param entry_type query string false "Entry type" Enums(credit, debit)
// @Param from_amount query number false "Minimum transaction amount" minimum(0)
// @Param to_amount query number false "Maximum transaction amount" minimum(0)
// @Param wallet_id query int false "Wallet ID to filter by" minimum(0)
// @Param reference_code query string false "Reference code to filter by"
// @Param wallet_transaction_id query string false "Wallet transaction ID to filter by"
// @Success 200 {object} domain.GetListTransactionMonitoringResponse "Successfully retrieved transaction list"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid parameters"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet-monitoring/transactions [get]
// @Security ApiKeyAuth
//
//nolint:dupl // similar code for different entities
func (h *WalletHandler) GetListTransactionMonitoring(c *gin.Context) {
	var (
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while GetListTransactionMonitoring", err)
		}
		span.End()
	}()

	var request *domain.GetListTransactionMonitoringRequest
	request, err = parseTransactionMonitoringRequest(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	err = request.Validate()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	var result *domain.GetListTransactionMonitoringResponse
	result, err = h.walletMonitoringUsecase.GetListTransactionMonitoring(ctx, request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}

// GetListEnvelope retrieves a paginated list of envelopes
//
// @Summary Get envelope list
// @Description Retrieve a filtered and paginated list of envelopes with various filtering options
// @Tags Wallet Monitoring
// @Accept json
// @Produce json
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Number of items per page" minimum(1) default(10)
// @Param from_total_amount query number false "Minimum total amount" minimum(0)
// @Param to_total_amount query number false "Maximum total amount" minimum(0)
// @Param user_id query string false "User ID to filter by" minlength(3) maxlength(50)
// @Param max_num_received query int false "Maximum number received" minimum(0) maximum(10000)
// @Param envelope_type query string false "Type of envelope" Enums(fixed, lucky, single)
// @Param from_expired_at query int false "Start timestamp for expiration filter" minimum(0)
// @Param to_expired_at query int false "End timestamp for expiration filter" minimum(0)
// @Param from_refunded_at query int false "Start timestamp for refund filter" minimum(0)
// @Param to_refunded_at query int false "End timestamp for refund filter" minimum(0)
// @Success 200 {object} domain.GetListEnvelopeResponse "Successfully retrieved envelope list"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid parameters"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet-monitoring/envelopes [get]
// @Security ApiKeyAuth
//
//nolint:dupl // similar code for different entities
func (h *WalletHandler) GetListEnvelope(c *gin.Context) {
	var (
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while GetListEnvelope", err)
		}
		span.End()
	}()

	var request *domain.GetListEnvelopeRequest
	request, err = parseEnvelopeListRequest(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	err = request.Validate()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	var result *domain.GetListEnvelopeResponse
	result, err = h.walletMonitoringUsecase.GetListEnvelope(ctx, request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}

// GetEnvelopeDetail retrieves detailed information about a specific envelope
//
// @Summary Get envelope details
// @Description Retrieve detailed information about a specific envelope including its status and recipient details
// @Tags Wallet Monitoring
// @Accept json
// @Produce json
// @Param envelope_id path int true "Envelope ID" minimum(1)
// @Param detail_status query string false "Detail status filter" Enums(pending, claimed, expired, refunded)
// @Param claimed query bool false "Filter by claimed status"
// @Param user_id query string false "User ID to filter by" minlength(3) maxlength(50)
// @Success 200 {object} domain.GetEnvelopeDetailResponse "Successfully retrieved envelope details"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid parameters"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Envelope not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet-monitoring/envelopes/{envelope_id}/details [get]
// @Security ApiKeyAuth
//
//nolint:dupl // similar code for different entities
func (h *WalletHandler) GetEnvelopeDetail(c *gin.Context) {
	var (
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while GetEnvelopeDetail", err)
		}
		span.End()
	}()

	var request *domain.GetEnvelopeDetailRequest
	request, err = parseEnvelopeDetailRequest(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	err = request.Validate()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	var result *domain.GetEnvelopeDetailResponse
	result, err = h.walletMonitoringUsecase.GetEnvelopeDetail(ctx, request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}

// GetTransferHistory retrieves a paginated list of transfer history
//
// @Summary Get transfer history
// @Description Retrieve a filtered and paginated list of transfer transactions with comprehensive filtering options
// @Tags Wallet Monitoring
// @Accept json
// @Produce json
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Number of items per page" minimum(1) default(10)
// @Param from_user_id query string false "Source user ID filter" minlength(3) maxlength(50)
// @Param to_user_id query string false "Destination user ID filter" minlength(3) maxlength(50)
// @Param status_transfer query string false "Transfer status filter" Enums(pending, claimed, expired, refunded, canceled)
// @Param from_amount query number false "Minimum transfer amount" minimum(0)
// @Param to_amount query number false "Maximum transfer amount" minimum(0)
// @Param from_expired_at query int false "Start timestamp for expiration filter" minimum(0)
// @Param to_expired_at query int false "End timestamp for expiration filter" minimum(0)
// @Param from_claimed_at query int false "Start timestamp for claimed filter" minimum(0)
// @Param to_claimed_at query int false "End timestamp for claimed filter" minimum(0)
// @Success 200 {object} domain.GetTransferHistoryResponse "Successfully retrieved transfer history"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid parameters"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet-monitoring/transfers [get]
// @Security ApiKeyAuth
//
//nolint:dupl // similar code for different entities
func (h *WalletHandler) GetTransferHistory(c *gin.Context) {
	var (
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while GetTransferHistory", err)
		}
		span.End()
	}()

	var request *domain.GetTransferHistoryRequest
	request, err = parseTransferHistoryRequest(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	err = request.Validate()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	var result *domain.GetTransferHistoryResponse
	result, err = h.walletMonitoringUsecase.GetTransferHistory(ctx, request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}

// GetTop10Users retrieves the top 10 users based on specified criteria
//
// @Summary Get top 10 users
// @Description Retrieve the top 10 users ranked by total amount within a specified time scope
// @Tags Wallet Monitoring
// @Accept json
// @Produce json
// @Param sort_by_total_amount query bool false "Sort by total amount" default(true)
// @Param scope query string true "Time scope for ranking" Enums(all_time, year, month)
// @Success 200 {object} domain.GetTop10UsersResponse "Successfully retrieved top 10 users"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid parameters"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet-monitoring/top-users [get]
// @Security ApiKeyAuth
func (h *WalletHandler) GetTop10Users(c *gin.Context) {
	var (
		request  domain.GetTop10UsersRequest
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while GetTop10Users", err)
		}
		span.End()
	}()

	err = parseGetTop10UsersRequest(c, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	err = request.Validate()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	var result *domain.GetTop10UsersResponse
	result, err = h.walletMonitoringUsecase.GetTop10Users(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}

// parsePagination
func parsePagination(c *gin.Context, request *domain.PaginationRequest) {
	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		page = domain.DefaultPage
	}
	request.Page = int32(page)

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 32)
	if err != nil {
		limit = domain.DefaultLimit
	}
	request.Limit = int32(limit)
}

// used by handler GetListEnvelope
// parseEnvelopeListRequest extracts and validates request parameters
func parseEnvelopeListRequest(c *gin.Context) (*domain.GetListEnvelopeRequest, error) {
	request := &domain.GetListEnvelopeRequest{}

	parsePagination(c, &request.PaginationRequest)

	if err := parseEnvelopeFilters(c, request); err != nil {
		return nil, err
	}

	return request, nil
}

func parseEnvelopeFilters(c *gin.Context, request *domain.GetListEnvelopeRequest) error {
	if userID := strings.TrimSpace(c.Query("user_id")); userID != "" {
		request.UserID = userID
	}

	if envelopeType := strings.TrimSpace(c.Query("envelope_type")); envelopeType != "" {
		if !entity.EnvelopeType(envelopeType).IsValid() {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid envelope_type: %s", envelopeType)).Wrap()
		}
		request.EnvelopeType = envelopeType
	}

	if maxNum := strings.TrimSpace(c.Query("max_num_received")); maxNum != "" {
		num, err := strconv.ParseInt(maxNum, 10, 32)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid max_num_received parameter: %v", err.Error())).Wrap()
		}
		if num <= 0 {
			return errs.ErrArgs.WithDetail("max_num_received must be positive").Wrap()
		}
		request.MaxNumReceived = int32(num)
	}

	// Parse date range filters (if uncommented in the future)
	if err := parseEnvelopeDateFilters(c, request); err != nil {
		return err
	}

	// Parse amount range filters (if uncommented in the future)
	if err := parseEnvelopeAmountFilters(c, request); err != nil {
		return err
	}

	return nil
}

// parseEnvelopeDateFilters handles date range parsing (for future use)
func parseEnvelopeDateFilters(c *gin.Context, request *domain.GetListEnvelopeRequest) error {
	// Parse FromExpiredAt
	if fromExpiredAt := strings.TrimSpace(c.Query("from_expired_at")); fromExpiredAt != "" {
		timestamp, err := strconv.ParseInt(fromExpiredAt, 10, 64)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid from_expired_at parameter: %v", err.Error())).Wrap()
		}
		if timestamp < 0 {
			return errs.ErrArgs.WithDetail("from_expired_at cannot be negative").Wrap()
		}
		request.FromExpiredAt = timestamp
	}

	// Parse ToExpiredAt
	if toExpiredAt := strings.TrimSpace(c.Query("to_expired_at")); toExpiredAt != "" {
		timestamp, err := strconv.ParseInt(toExpiredAt, 10, 64)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid to_expired_at parameter: %v", err.Error())).Wrap()
		}
		if timestamp < 0 {
			return errs.ErrArgs.WithDetail("to_expired_at cannot be negative").Wrap()
		}
		request.ToExpiredAt = timestamp
	}

	// Validate date range
	if request.FromExpiredAt > 0 && request.ToExpiredAt > 0 {
		if request.FromExpiredAt > request.ToExpiredAt {
			return errs.ErrArgs.WithDetail("from_expired_at cannot be after to_expired_at").Wrap()
		}
	}

	return nil
}

// parseEnvelopeAmountFilters handles amount range parsing (for future use)
func parseEnvelopeAmountFilters(c *gin.Context, request *domain.GetListEnvelopeRequest) error {
	var (
		fromTotalAmount, toTotalAmount string
		err                            error
	)

	if fromTotalAmount = strings.TrimSpace(c.Query("from_total_amount")); fromTotalAmount != "" {
		request.FromTotalAmount, err = validateAmountFilter(fromTotalAmount)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid from_total_amount parameter: %v", err)).Wrap()
		}
	}

	if toTotalAmount = strings.TrimSpace(c.Query("to_total_amount")); toTotalAmount != "" {
		request.ToTotalAmount, err = validateAmountFilter(toTotalAmount)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid to_total_amount parameter: %v", err.Error())).Wrap()
		}
	}

	if fromTotalAmount != "" && toTotalAmount != "" {
		if err := validateAmountRange(request.FromTotalAmount, request.ToTotalAmount); err != nil {
			return err
		}
	}

	return nil
}

// used by handler GetEnvelopeDetail
// parseEnvelopeDetailRequest extracts and validates envelope ID from path parameter
func parseEnvelopeDetailRequest(c *gin.Context) (*domain.GetEnvelopeDetailRequest, error) {
	envelopeIDStr := strings.TrimSpace(c.Param("envelope_id"))
	if envelopeIDStr == "" {
		return nil, errs.ErrArgs.WithDetail(eerrs.ErrEnvelopeIDNotFoundParam.Error()).Wrap()
	}

	envelopeID, err := strconv.ParseInt(envelopeIDStr, 10, 64)
	if err != nil {
		return nil, errs.ErrArgs.WithDetail("envelope_id must be a valid integer").Wrap()
	}

	if envelopeID <= 0 {
		return nil, errs.ErrArgs.WithDetail("envelope_id must be positive").Wrap()
	}

	request := &domain.GetEnvelopeDetailRequest{
		EnvelopeID: envelopeID,
	}

	if err := parseEnvelopeDetailFilters(c, request); err != nil {
		return nil, err
	}

	return request, nil
}

// parseEnvelopeDetailFilters parses optional filters for envelope details
func parseEnvelopeDetailFilters(c *gin.Context, request *domain.GetEnvelopeDetailRequest) error {
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		if !entity.EnvelopeDetailStatus(status).IsValid() {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid status parameter: %s", status)).Wrap()
		}
		request.DetailStatus = status
	}

	if claimed := strings.TrimSpace(c.Query("claimed")); claimed != "" {
		isClaimed, err := strconv.ParseBool(claimed)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid claimed parameter: %v", err.Error())).Wrap()
		}
		request.Claimed = &isClaimed
	}

	if userID := strings.TrimSpace(c.Query("user_id")); userID != "" {
		request.UserID = userID
	}

	return nil
}

// used by handler GetTransferHistory
// parseTransferHistoryRequest extracts and validates request parameters
func parseTransferHistoryRequest(c *gin.Context) (*domain.GetTransferHistoryRequest, error) {
	request := &domain.GetTransferHistoryRequest{}

	parsePagination(c, &request.PaginationRequest)

	if err := parseTransferFilters(c, request); err != nil {
		return nil, err
	}

	if err := parseTransferDateFilters(c, request); err != nil {
		return nil, err
	}

	return request, nil
}

func parseTransferFilters(c *gin.Context, request *domain.GetTransferHistoryRequest) error {
	if fromUserID := strings.TrimSpace(c.Query("from_user_id")); fromUserID != "" {
		request.FromUserID = fromUserID
	}

	if toUserID := strings.TrimSpace(c.Query("to_user_id")); toUserID != "" {
		request.ToUserID = toUserID
	}

	if request.FromUserID != "" && request.ToUserID != "" && request.FromUserID == request.ToUserID {
		return errs.ErrArgs.WithDetail("from_user_id and to_user_id cannot be the same").Wrap()
	}

	if statusTransfer := strings.TrimSpace(c.Query("status_transfer")); statusTransfer != "" {
		if !entity.StatusTransfer(statusTransfer).IsValid() {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid status_transfer: %s", statusTransfer)).Wrap()
		}
		request.StatusTransfer = statusTransfer
	}

	if err := parseTransferAmountFilters(c, request); err != nil {
		return err
	}

	return nil
}

func parseTransferAmountFilters(c *gin.Context, request *domain.GetTransferHistoryRequest) error {
	var (
		fromAmount, toAmount string
		err                  error
	)

	if fromAmount = strings.TrimSpace(c.Query("from_amount")); fromAmount != "" {
		request.FromAmount, err = validateAmountFilter(fromAmount)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid from_amount parameter: %v", err)).Wrap()
		}
	}

	if toAmount = strings.TrimSpace(c.Query("to_amount")); toAmount != "" {
		request.ToAmount, err = validateAmountFilter(toAmount)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid to_amount parameter: %v", err)).Wrap()
		}
	}

	if fromAmount != "" && toAmount != "" {
		if err := validateAmountRange(request.FromAmount, request.ToAmount); err != nil {
			return err
		}
	}

	return nil
}

func validateAmountFilter(amountRaw string) (amount float64, err error) {
	amount, err = strconv.ParseFloat(amountRaw, 64)
	if err != nil {
		return 0, err
	}

	if amount < 0 {
		return 0, fmt.Errorf("cannot be negative")
	}

	return amount, nil
}

func validateAmountRange(fromAmount, toAmount float64) error {
	if fromAmount > 0 && toAmount > 0 && fromAmount > toAmount {
		return errs.ErrArgs.WithDetail("from_amount cannot be greater than to_amount").Wrap()
	}

	return nil
}

// parseTransferDateFilters handles date range parsing (for future use)
func parseTransferDateFilters(c *gin.Context, request *domain.GetTransferHistoryRequest) (err error) {
	if request.FromExpiredAt, err = parseTimestampQuery(c, "from_expired_at"); err != nil {
		return err
	}

	if request.ToExpiredAt, err = parseTimestampQuery(c, "to_expired_at"); err != nil {
		return err
	}

	if request.FromClaimedAt, err = parseTimestampQuery(c, "from_claimed_at"); err != nil {
		return err
	}

	if request.ToClaimedAt, err = parseTimestampQuery(c, "to_claimed_at"); err != nil {
		return err
	}

	if request.FromExpiredAt > 0 && request.ToExpiredAt > 0 && request.FromExpiredAt > request.ToExpiredAt {
		return errs.ErrArgs.WithDetail("from_expired_at cannot be after to_expired_at").Wrap()
	}

	if request.FromClaimedAt > 0 && request.ToClaimedAt > 0 && request.FromClaimedAt > request.ToClaimedAt {
		return errs.ErrArgs.WithDetail("from_claimed_at cannot be after to_claimed_at").Wrap()
	}

	return nil
}

func parseTimestampQuery(c *gin.Context, param string) (int64, error) {
	value := strings.TrimSpace(c.Query(param))
	if value == "" {
		return 0, nil
	}

	timestamp, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, errs.ErrArgs.WithDetail(fmt.Sprintf("invalid %s parameter: %v", param, err.Error())).Wrap()
	}

	if timestamp < 0 {
		return 0, errs.ErrArgs.WithDetail(fmt.Sprintf("%s cannot be negative", param)).Wrap()
	}

	return timestamp, nil
}

// used by handler GetListTransactionMonitoring
// parseTransactionMonitoringRequest extracts and validates request parameters
func parseTransactionMonitoringRequest(c *gin.Context) (*domain.GetListTransactionMonitoringRequest, error) {
	request := &domain.GetListTransactionMonitoringRequest{}
	parsePagination(c, &request.PaginationRequest)

	if err := parseDateFilters(c, request); err != nil {
		return nil, err
	}

	if err := parseNumericFilters(c, request); err != nil {
		return nil, err
	}

	request.TransactionType = strings.TrimSpace(c.Query("transaction_type"))
	request.EntryType = strings.TrimSpace(c.Query("entry_type"))
	request.ReferenceCode = strings.TrimSpace(c.Query("reference_code"))
	request.WalletTransactionID = strings.TrimSpace(c.Query("wallet_transaction_id"))

	return request, nil
}

func parseDateFilters(c *gin.Context, request *domain.GetListTransactionMonitoringRequest) error {
	if rawFromDate := strings.TrimSpace(c.Query("from_transaction_date")); rawFromDate != "" {
		fromDate, err := time.Parse(domain.LayoutFilterDate, rawFromDate)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("expected format: " + domain.LayoutFilterDate)).Wrap()
		}

		request.FromTransactionDate = fromDate
	}

	if rawToDate := strings.TrimSpace(c.Query("to_transaction_date")); rawToDate != "" {
		toDate, err := time.Parse(domain.LayoutFilterDate, rawToDate)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("expected format: " + domain.LayoutFilterDate)).Wrap()
		}

		request.ToTransactionDate = toDate
	}

	if !request.FromTransactionDate.IsZero() && !request.ToTransactionDate.IsZero() {
		if request.FromTransactionDate.After(request.ToTransactionDate) {
			return errs.ErrArgs.WithDetail("from_date cannot be after to_date").Wrap()
		}
	}

	return nil
}

func parseNumericFilters(c *gin.Context, request *domain.GetListTransactionMonitoringRequest) error {
	if fromAmount := strings.TrimSpace(c.Query("from_amount")); fromAmount != "" {
		amount, err := strconv.ParseFloat(fromAmount, 64)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid from_amount parameter: %v", err.Error())).Wrap()
		}

		if amount < 0 {
			return errs.ErrArgs.WithDetail("from_amount cannot be negative").Wrap()
		}

		request.FromAmount = amount
	}

	if toAmount := strings.TrimSpace(c.Query("to_amount")); toAmount != "" {
		amount, err := strconv.ParseFloat(toAmount, 64)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid to_amount parameter: %v", err.Error())).Wrap()
		}

		if amount < 0 {
			return errs.ErrArgs.WithDetail("to_amount cannot be negative").Wrap()
		}

		request.ToAmount = amount
	}

	if request.FromAmount > 0 && request.ToAmount > 0 && request.FromAmount > request.ToAmount {
		return errs.ErrArgs.WithDetail("from_amount cannot be greater than to_amount").Wrap()
	}

	if walletID := strings.TrimSpace(c.Query("wallet_id")); walletID != "" {
		id, err := strconv.ParseInt(walletID, 10, 64)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid wallet_id parameter: %v", err.Error())).Wrap()
		}

		if id <= 0 {
			return errs.ErrArgs.WithDetail("wallet_id must be positive").Wrap()
		}

		request.WalletID = id
	}

	return nil
}

// used by handler GetTop10Users
// parseGetTop10UsersRequest parses and validates the request parameters
func parseGetTop10UsersRequest(c *gin.Context, request *domain.GetTop10UsersRequest) (err error) {
	// Parse sort_by_total_amount with default false
	sortByStr := c.Query("sort_by_total_amount")
	if sortByStr != "" {
		sortBy, err := strconv.ParseBool(sortByStr)
		if err != nil {
			return errs.ErrArgs.WithDetail(fmt.Sprintf("invalid sort_by_total_amount parameter: %v", err.Error())).Wrap()
		}

		request.SortByTotalAmount = sortBy
	}

	// Parse scope with default "month"
	scope := c.Query("scope")
	if scope == "" {
		scope = domain.ScopeMonth.String()
	}
	request.Scope = scope

	return nil
}
