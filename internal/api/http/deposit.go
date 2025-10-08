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

// ProcessDepositByAdmin processes a deposit transaction by admin
//
// @Summary Process deposit by admin
// @Description Create and process a deposit transaction on behalf of a user by admin
// @Tags Deposit
// @Accept json
// @Produce json
// @Param request body domain.ProcessDepositByAdminRequest true "Deposit processing request"
// @Success 200 {object} domain.ProcessDepositByAdminResponse "Successfully processed deposit"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid parameters"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /deposit/process [post]
// @Security ApiKeyAuth
func (h *WalletHandler) ProcessDepositByAdmin(c *gin.Context) {
	var (
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c, funcName)

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while ProcessDepositByAdmin", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

	var request domain.ProcessDepositByAdminRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	request.OperatedBy = fmt.Sprintf("%s-%s", userID, c.GetString(constant.RpcOpUserType))

	var resp *domain.ProcessDepositByAdminResponse
	resp, err = h.depositUsecase.ProcessDepositByAdmin(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, resp)
}

// GetListDeposit retrieves a filtered and paginated list of deposit transactions
//
// @Summary Get deposit list
// @Description Retrieve a filtered and paginated list of deposit transactions with various filtering and sorting options
// @Tags Deposit
// @Accept json
// @Produce json
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Number of items per page" minimum(1) default(10)
// @Param startDate query string false "Start date for filtering (YYYY-MM-DD format)" format(date)
// @Param endDate query string false "End date for filtering (YYYY-MM-DD format)" format(date)
// @Param statusRequest query string false "Status of the deposit request" Enums(requested, approved, rejected, failed)
// @Param sortBy query string false "Column to sort by" default(created_at)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
// @Param filterDateBy query string false "Date column to filter by" Enums(created_at, approved_at, updated_at, deleted_at)
// @Success 200 {object} domain.GetListDepositResponse "Successfully retrieved deposit list"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid date format or parameters"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /deposit/list [get]
// @Security ApiKeyAuth
func (h *WalletHandler) GetListDeposit(c *gin.Context) {
	var (
		request  domain.GetListDepositRequest
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			log.ZError(ctx, "an error occurred while GetListDeposit", err)
		}
		span.End()
	}()

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
	request.FilterDateBy = c.Query("statusRequest")
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

	var result *domain.GetListDepositResponse
	result, err = h.depositUsecase.GetListDeposit(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, result)
}
