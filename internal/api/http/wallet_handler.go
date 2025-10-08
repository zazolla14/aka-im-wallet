package http

import (
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

// GetWalletDetail retrieves detailed information about user's wallet
//
// @Summary Get wallet details
// @Description Retrieve detailed information about the authenticated user's wallet including balance and status
// @Tags Wallet
// @Accept json
// @Produce json
// @Success 200 {object} apiresp.ApiResponse "Successfully retrieved wallet details"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Wallet not found for user"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet/detail [get]
// @Security ApiKeyAuth
func (h *WalletHandler) GetWalletDetail(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while GetWalletDetail", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}

	var wallet *domain.Wallet
	wallet, err = h.walletUsecase.GetWalletDetail(ctx, userID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, wallet)
}

// CreateWallet creates a new wallet for the authenticated user
//
// @Summary Create new wallet
// @Description Create a new wallet for the authenticated user if they don't have one
// @Tags Wallet
// @Accept json
// @Produce json
// @Success 200 {object} apiresp.ApiResponse "Successfully created wallet"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Wallet already exists"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /wallet/create [post]
// @Security ApiKeyAuth
func (h *WalletHandler) CreateWallet(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while CreateWallet", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, eerrs.ErrUserIDNotFoundCtx)
		return
	}
	createdBy := fmt.Sprintf("%s-%s", userID, c.GetString(constant.RpcOpUserType))

	var wallet *domain.Wallet
	wallet, err = h.walletUsecase.CreateWallet(ctx, userID, createdBy)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	apiresp.GinSuccess(c, wallet)
}
