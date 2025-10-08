package http

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/1nterdigital/aka-im-tools/apiresp"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/constant"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

// CreateTransfer creates a new transfer between users
//
// @Summary Create transfer
// @Description Create a new transfer from authenticated user to another user
// @Tags Transfer
// @Accept json
// @Produce json
// @Param request body domain.CreateTransferRequest true "Transfer creation request"
// @Success 200 {object} domain.CreateTransferResponse "Successfully created transfer"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid transfer data or insufficient balance"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Recipient user not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /transfer/create [post]
// @Security ApiKeyAuth
func (h *WalletHandler) CreateTransfer(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while CreateTransfer", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

	var request domain.CreateTransferRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	if utf8.RuneCountInString(request.Remark) > entity.MaxGreetingCharacters {
		apiresp.GinError(c, eerrs.ErrGreetingLength(entity.MaxGreetingCharacters))
		return
	}
	if request.Amount > entity.MaxAllowedSendTransferAmount {
		apiresp.GinError(c, eerrs.ErrTransferLimitExceeded)
		return
	}

	request.FromUserID = userID
	request.CreatedBy = fmt.Sprintf("%s-%s", userID, c.GetString(constant.RpcOpUserType))

	_, err = request.IsValid()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	var transferID *domain.CreateTransferResponse
	transferID, err = h.transferUsecase.CreateTransfer(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, transferID)
}

// ClaimTransfer claims a pending transfer
//
// @Summary Claim transfer
// @Description Claim a pending transfer that was sent to the authenticated user
// @Tags Transfer
// @Accept json
// @Produce json
// @Param request body domain.ClaimTransferRequest true "Transfer claim request"
// @Success 200 {object} apiresp.ApiResponse "Successfully claimed transfer"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid transfer ID or already claimed"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Transfer not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /transfer/claim [post]
// @Security ApiKeyAuth
func (h *WalletHandler) ClaimTransfer(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while ClaimTransfer", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

	var request domain.ClaimTransferRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	request.ClaimerUserID = userID
	request.OperateBy = fmt.Sprintf("%s-%s", userID, c.GetString(constant.RpcOpUserType))

	_, err = request.IsValid()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	err = h.transferUsecase.ClaimTransfer(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, nil)
}

// RefundTransfer Self refund transfer after expired
//
// @Summary Self Refund Transfer
// @Description Refund a transfer authentication user after expired
// @Tags Transfer
// @Accept json
// @Produce json
// @Param request body domain.RefundTransferReq true "Refund transfer request"
// @Success 200 {object} apiresp.ApiResponse "Successfully Refund transfer"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid transfer ID or should greater 0"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Transfer not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /transfer/refund [post]
// @Security ApiKeyAuth
func (h *WalletHandler) RefundTransfer(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while RefundTransfer", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

	var request domain.RefundTransferReq
	err = c.ShouldBindJSON(&request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	err = request.IsValid()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	request.UserID = userID
	request.OperatedBy = userID + "-" + c.GetString(constant.RpcOpUserType)

	err = h.transferUsecase.RefundTransfer(ctx, &request)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, nil)
}

// GetDetailTransfer Get detail transfer
//
// @Summary Get detail transfer
// @Description get detail transfer by transfer id and user id
// @Tags Transfer
// @Accept json
// @Produce json
// @Param transfer_id path string true "ID of transfer"
// @Success 200 {object} apiresp.ApiResponse "Successfully Get transfer detail"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid transfer ID or should greater 0"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Transfer not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /transfer/:transfer_id/detail [get]
// @Security ApiKeyAuth
func (h *WalletHandler) GetDetailTransfer(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while GetDetailTransfer", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

	var transferID int64
	transferID, err = getParamTransferID(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	var transfer *domain.Transfer
	transfer, err = h.transferUsecase.GetDetailTransfer(ctx, transferID, userID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, transfer)
}

func getParamTransferID(c *gin.Context) (transferID int64, err error) {
	rawTransferID := c.Param("transfer_id")
	if rawTransferID == "" {
		return 0, eerrs.ErrTransferIDRequired
	}

	transferID, err = strconv.ParseInt(rawTransferID, 10, 64)
	if err != nil {
		return 0, eerrs.ErrInvalidTransferID
	}

	return transferID, nil
}
