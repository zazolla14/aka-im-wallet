package http

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/1nterdigital/aka-im-tools/apiresp"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/constant"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

// CreateTransaction creates a new wallet transaction
//
// @Summary Create wallet transaction
// @Description Create a new transaction for the specified wallet with transaction details
// @Tags Wallet Transactions
// @Accept json
// @Produce json
// @Param request body domain.CreateTransactionReq true "Transaction creation request"
// @Success 200 {object} apiresp.ApiResponse "Successfully created transaction"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid transaction data"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet_transactions/create [post]
// @Security ApiKeyAuth
func (h *WalletHandler) CreateTransaction(c *gin.Context) {
	var (
		req      = domain.CreateTransactionReq{}
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c, funcName)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while ProcessManualRefund", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}
	userType := c.GetString(constant.RpcOpUserType)

	err = c.ShouldBindJSON(&req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req.CreatedBy = fmt.Sprintf("%s-%s", userID, userType)

	var result int64
	result, err = h.walletTransactionUsecase.CreateTransaction(ctx, &req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}

// GetListTransaction retrieves a paginated list of wallet transactions
//
// @Summary Get transaction cash flow
// @Description Retrieve a filtered and paginated list of wallet transactions within a date range
// @Tags Wallet Transactions
// @Accept json
// @Produce json
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Number of items per page" minimum(1) default(10)
// @Param startDate query string false "Start date for filtering (YYYY-MM-DD format)" format(date)
// @Param endDate query string false "End date for filtering (YYYY-MM-DD format)" format(date)
// @Success 200 {object} domain.GetListTransactionResponse "Successfully retrieved transaction list"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid date format"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet_transactions/cash_flow [get]
// @Security ApiKeyAuth
func (h *WalletHandler) GetListTransaction(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while GetListTransaction", err)
		}
		span.End()
	}()

	var request domain.GetListTransactionRequest
	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}
	request.UserID = userID

	var defaultPage int64 = 1
	page, _ := strconv.ParseInt(c.Query("page"), 10, 32)
	if page <= 0 {
		page = defaultPage
	}
	request.Page = int32(page)

	var defaultLimit int64 = 10
	limit, _ := strconv.ParseInt(c.Query("limit"), 10, 32)
	if limit <= 0 {
		limit = defaultLimit
	}
	request.Limit = int32(limit)

	var (
		rawStartDate     = c.Query("startDate")
		rawEndDate       = c.Query("endDate")
		layoutFilterDate = "2006-01-02"
	)

	var startDate, endDate time.Time
	if rawStartDate != "" && rawEndDate != "" {
		startDate, err = time.Parse(layoutFilterDate, rawStartDate)
		if err != nil {
			apiresp.GinError(c, eerrs.ErrInvalidFormatStartDate)
			return
		}
		request.StartDate = startDate

		endDate, err = time.Parse(layoutFilterDate, rawEndDate)
		if err != nil {
			apiresp.GinError(c, eerrs.ErrInvalidFormatEndDate)
			return
		}
		request.EndDate = endDate
	}

	var result *domain.GetListTransactionResponse
	result, err = h.walletTransactionUsecase.GetListTransaction(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}
