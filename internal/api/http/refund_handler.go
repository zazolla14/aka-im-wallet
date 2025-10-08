package http

import (
	"context"
	"fmt"

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

// ProcessManualRefund Process manual refund by admin
//
// @Summary Process manual refund processed by admin
// @Description manual refund processed by admin when auto refund is failed
// @Tags ManualRefund
// @Accept json
// @Produce json
// @Param request body domain.ManualRefundArgs true "Mnual refund request"
// @Success 200 {object} apiresp.ApiResponse "Successfully refund by admin"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Request invalid"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Wallet user not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /bo/refund/manual_process [post]
// @Security ApiKeyAuth
func (h *WalletHandler) ProcessManualRefund(c *gin.Context) {
	var (
		req      = domain.ManualRefundArgs{}
		err      error
		funcName = tracer.GetFullFunctionPath()
		t        = otel.Tracer(tracer.LevelHandler)
	)

	ctx, span := t.Start(c.Request.Context(), funcName)
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

	ctx = context.WithoutCancel(ctx)
	ctx = context.WithValue(ctx, domain.KeyOperatedBy, fmt.Sprintf("%s-%s", userID, userType))
	go func() {
		err = h.envelopeUsecase.ProcessManualRefund(ctx, req.EnvelopeIDs)
		if err != nil {
			log.ZError(ctx, "failed to trigger manual refund envelopes", err)
		}

		err = h.transferUsecase.ProcessManualRefund(ctx, req.TransferIDs)
		if err != nil {
			log.ZError(ctx, "failed to trigger manual refund transfers", err)
		}
	}()

	apiresp.GinSuccess(c, nil)
}
