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

// BalanceAdjustmentByAdmin Balance adjustment by admin
//
// @Summary Balance Adjustment processed by admin
// @Description added/deduct balance user by admin using balance adjustment
// @Tags BalanceAdjustment
// @Accept json
// @Produce json
// @Param request body domain.BalanceAdjustmentByAdminRequest true "Balance adjustment request"
// @Success 200 {object} apiresp.ApiResponse "Successfully adjust balance user"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Request invalid"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Wallet user not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /bo/balance_adjustment/process [post]
// @Security ApiKeyAuth
func (h *WalletHandler) BalanceAdjustmentByAdmin(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while BalanceAdjustmentByAdmin", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

	var request domain.BalanceAdjustmentByAdminRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	err = request.Validate()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	request.OperatedBy = fmt.Sprintf("%s-%s", userID, c.GetString(constant.RpcOpUserType))

	var resp *domain.BalanceAdjustmentByAdminResponse
	resp, err = h.balanceAdjustmentUsecase.BalanceAdjustmentByAdmin(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, resp)
}

// GetListBalanceAdjustment Get List Balance Adjustment
//
// @Summary Get List Balance Adjustment
// @Description Get list balance adjustment after processed by admin
// @Tags BalanceAdjustment
// @Accept json
// @Produce json
// @Param page query int false "page response" default(1)
// @Param limit query int false "limit response" default(10)
// @Param filterDateBy query string false "filter date by"
// @Param userID query string false "filter by userID"
// @Param sortBy query string false "sort list" default(created_at)
// @Param sortOrder query string false "sort order" Enums(asc, desc) default(desc)
// @Param startDate query string false "start date if filterdateby used" format(date)
// @Param endDate query string false "end date if filterdateby used" format(date)
// @Success 200 {object} apiresp.ApiResponse "Successfully get list balance adjustment"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Adjustment not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /bo/balance_adjustment/list [get]
// @Security ApiKeyAuth
func (h *WalletHandler) GetListBalanceAdjustment(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while GetListBalanceAdjustment", err)
		}
		span.End()
	}()

	var request domain.GetListbalanceAjustmentRequest
	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

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

	request.FilterDateBy = c.Query("filterDateBy")
	request.UserID = c.Query("userID")

	request.SortBy = c.Query("sortBy")
	if request.SortBy == "" {
		request.SortBy = "created_at"
	}

	request.SortOrder = c.Query("sortOrder")
	if request.SortOrder == "" {
		request.SortOrder = "desc"
	}

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

	var result *domain.GetListBalanceAdjustmentResponse
	result, err = h.balanceAdjustmentUsecase.GetListBalanceAdjustment(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}
