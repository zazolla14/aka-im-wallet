package http

import (
	"errors"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/1nterdigital/aka-im-tools/apiresp"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/tracer"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	model "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/constant"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

// CreateEnvelopeHandler creates a new envelope (red packet)
//
// @Summary Create envelope
// @Description Create a new envelope (red packet) that can be claimed by users
// @Tags Envelope
// @Accept json
// @Produce json
// @Param request body domain.EnvelopeCreateRequest true "Envelope creation request"
// @Success 200 {object} domain.EnvelopeCreateResponse "Successfully created envelope"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid envelope data or insufficient balance"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /envelope [post]
// @Security ApiKeyAuth
func (h *WalletHandler) CreateEnvelopeHandler(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while CreateEnvelopeHandler", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, errors.New("user ID not found in context"))
		return
	}
	var req domain.EnvelopeCreateRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	if req.TotalAmount < domain.EnvelopeMinimumAmount || req.TotalAmount > model.MaxAllowedSendEnvelopeAmount {
		apiresp.GinError(c, eerrs.ErrAmountRange(domain.EnvelopeMinimumAmount, model.MaxAllowedSendEnvelopeAmount))
		return
	}
	if req.TotalClaimer < 1 {
		apiresp.GinError(c, eerrs.ErrTotalClaimerMustGreaterThanZero)
		return
	}
	if utf8.RuneCountInString(req.Remarks) > model.MaxGreetingCharacters {
		apiresp.GinError(c, eerrs.ErrGreetingLength(model.MaxGreetingCharacters))
		return
	}

	if model.EnvelopeType(req.EnvelopeType) == model.EnvelopeTypeSingle && req.ToUserID == "" {
		apiresp.GinError(c, errors.New("toUserID is required for single envelope"))
		return
	}

	req.UserID = userID
	var envelopes *domain.EnvelopeCreateResponse
	envelopes, err = h.envelopeUsecase.CreateEnvelope(ctx, &req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, envelopes)
}

// AutoRefundEnvelopeHandler processes automatic refund for expired envelopes
//
// @Summary Auto refund envelopes
// @Description Process automatic refund for all expired envelopes
// @Tags Envelope
// @Accept json
// @Produce json
// @Success 200 {object} apiresp.ApiResponse "Successfully processed auto refund"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /envelope/autoRefund [get]
// @Security ApiKeyAuth
func (h *WalletHandler) AutoRefundEnvelopeHandler(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while AutoRefundEnvelopeHandler", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, errors.New("user ID not found in context"))
		return
	}
	err = h.envelopeUsecase.AutoRefund(ctx)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

// RefundEnvelopeByID processes automatic refund for expired envelopes
//
// @Summary refund envelopeByID
// @Description Process refund for all expired envelopeByID
// @Tags Envelope
// @Accept json
// @Produce json
// @Success 200 {object} apiresp.ApiResponse "Successfully processed auto refund"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /envelope/refund/:envelopeID [get]
// @Security ApiKeyAuth
func (h *WalletHandler) RefundEnvelopeByID(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while RefundEnvelopeByID", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, errors.New("user ID not found in context"))
		return
	}
	envIDString := c.Param("envelope_id")

	var envID int64
	envID, err = strconv.ParseInt(envIDString, 10, 64)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	err = h.envelopeUsecase.RefundByID(ctx, userID, envID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

// ClaimEnvelopeHandler claims an envelope
//
// @Summary Claim envelope
// @Description Claim an available envelope (red packet) to receive money
// @Tags Envelope
// @Accept json
// @Produce json
// @Param request body domain.EnvelopeClaimRequest true "Envelope claim request"
// @Success 200 {object} domain.EnvelopeClaimResponse "Successfully claimed envelope"
// @Failure 400 {object} apiresp.ApiResponse "Bad Request - Invalid envelope ID or wallet ID, or envelope already claimed"
// @Failure 401 {object} apiresp.ApiResponse "Unauthorized - User ID not found in context"
// @Failure 404 {object} apiresp.ApiResponse "Not Found - Envelope not found"
// @Failure 500 {object} apiresp.ApiResponse "Internal Server Error"
// @Router /envelope/claim [post]
// @Security ApiKeyAuth
func (h *WalletHandler) ClaimEnvelopeHandler(c *gin.Context) {
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
			log.ZError(ctx, "an error occurred while ClaimEnvelopeHandler", err)
		}
		span.End()
	}()

	userID := c.GetString(constant.RpcOpUserID)
	if userID == "" {
		apiresp.GinError(c, errors.New("user ID not found in context"))
		return
	}
	var req domain.EnvelopeClaimRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	if req.WalletID < 1 {
		apiresp.GinError(c, eerrs.ErrWalletNotFound)
		return
	}
	if req.EnvelopeID < 1 {
		apiresp.GinError(c, errors.New("envelopeID or walletID must be greater than 0"))
		return
	}
	req.UserID = userID
	var claimData *domain.EnvelopeClaimResponse
	claimData, err = h.envelopeUsecase.Claim(ctx, &req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, claimData)
}
