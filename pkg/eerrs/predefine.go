package eerrs

import (
	"fmt"

	"github.com/1nterdigital/aka-im-tools/errs"
)

var (
	ErrPassword                 = errs.NewCodeError(ErrorCodePasswordError, "PasswordError")
	ErrAccountNotFound          = errs.NewCodeError(ErrorCodeAccountNotFound, "AccountNotFound")
	ErrPhoneAlreadyRegister     = errs.NewCodeError(ErrorCodePhoneAlreadyRegister, "PhoneAlreadyRegister")
	ErrAccountAlreadyRegister   = errs.NewCodeError(ErrorCodeAccountAlreadyRegister, "AccountAlreadyRegister")
	ErrVerifyCodeSendFrequently = errs.NewCodeError(ErrorCodeVerifyCodeSendFrequently, "VerifyCodeSendFrequently")
	ErrVerifyCodeNotMatch       = errs.NewCodeError(ErrorCodeVerifyCodeNotMatch, "VerifyCodeNotMatch")
	ErrVerifyCodeExpired        = errs.NewCodeError(ErrorCodeVerifyCodeExpired, "VerifyCodeExpired")
	ErrVerifyCodeMaxCount       = errs.NewCodeError(ErrorCodeVerifyCodeMaxCount, "VerifyCodeMaxCount")
	ErrVerifyCodeUsed           = errs.NewCodeError(ErrorCodeVerifyCodeUsed, "VerifyCodeUsed")
	ErrInvitationCodeUsed       = errs.NewCodeError(ErrorCodeInvitationCodeUsed, "InvitationCodeUsed")
	ErrInvitationNotFound       = errs.NewCodeError(ErrorCodeInvitationNotFound, "InvitationNotFound")
	ErrForbidden                = errs.NewCodeError(ErrorCodeForbidden, "Forbidden")
	ErrRefuseFriend             = errs.NewCodeError(ErrorCodeRefuseFriend, "RefuseFriend")
	ErrEmailAlreadyRegister     = errs.NewCodeError(ErrorCodeEmailAlreadyRegister, "EmailAlreadyRegister")

	ErrTokenNotExist = errs.NewCodeError(ErrorTokenNotExist, "ErrTokenNotExist")

	ErrEntryTypeInvalid       = errs.NewCodeError(ErrorCodeEntryTypeInvalid, "invalid type of entry type")
	ErrTransactionTypeInvalid = errs.NewCodeError(ErrorCodeTransactionTypeInvalid, "invalid type of transaction type")
	ErrWalletIDRequired       = errs.NewCodeError(ErrorCodeWalletIDRequired, "wallet id is required")
	ErrDescriptionEnRequired  = errs.NewCodeError(ErrorCodeDescriptionEnRequired, "description of transaction english is required")
	ErrDescriptionZhRequired  = errs.NewCodeError(ErrorCodeDescriptionZhRequired, "description of transaction zh is required")
	ErrReferenceCodeRequired  = errs.NewCodeError(ErrorCodeReferenceCodeRequired, "reference code of transaction is required")
	ErrImpactedItemRequired   = errs.NewCodeError(ErrorCodeImpactedItemRequired, "impacted item is required")
	ErrInsufficientBalance    = errs.NewCodeError(ErrorCodeInsufficientBalance, "insufficient balance")
	ErrWalletNotFound         = errs.NewCodeError(ErrorCodeWalletNotFound, "wallet not found")
	ErrInvalidFormatStartDate = errs.NewCodeError(ErrorCodeInvalidFormatStartDate, "invalid format start date (format: YYYY-MM-DD)")
	ErrInvalidFormatEndDate   = errs.NewCodeError(ErrorCodeInvalidFormatEndDate, "invalid format end date (format: YYYY-MM-DD)")
	ErrUserIDNotFoundCtx      = errs.NewCodeError(ErrorCodeUserIDNotFoundCtx, "user id not found in context")
	ErrWalletExisted          = errs.NewCodeError(ErrorCodeWalletExisted, "wallet already existed")
	ErrInvalidStatusRequest   = errs.NewCodeError(ErrorCodeInvalidStatusRequest, "invalid status request deposit")
	ErrCurrencyMismatch       = errs.NewCodeError(ErrorCodeCurrencyMismatch, "currency mismatch")

	// Envelope
	ErrEnvelopeNotFound                = errs.NewCodeError(ErrorCodeEnvelopeNotFound, "envelope not found")
	ErrUnsupportedEnvelopeType         = errs.NewCodeError(ErrorCodeUnsupportedEnvelopeType, "unsupported envelope type")
	ErrExpiredEnvelope                 = errs.NewCodeError(ErrorCodeExpiredEnvelope, "envelope already expired")
	ErrEnvelopeNotActive               = errs.NewCodeError(ErrorCodeEnvelopeNotActive, "envelope is no longer active")
	ErrUnauthorizedUserID              = errs.NewCodeError(ErrorCodeUnauthorizedUserID, "unauthorized user id")
	ErrUnauthorizedClaimer             = errs.NewCodeError(ErrorCodeUnauthorizedClaimer, "unauthorized to claim")
	ErrAmountExceedsWalletBalance      = errs.NewCodeError(ErrorCodeAmountExceedsWalletBalance, "amount exceeds wallet balance")
	ErrNoMoreSharesToClaim             = errs.NewCodeError(ErrorCodeNoMoreSharesToClaim, "no more shares to claim")
	ErrTotalClaimerMustGreaterThanZero = errs.NewCodeError(ErrorCodeTotalClaimerMustGreaterThanZero, "total claimer must be greater than 0")
	ErrSendingDailyLimit               = errs.NewCodeError(ErrorCodeSendingDailyLimit, "you have reached the daily envelope sending limit")
	ErrClaimingDailyLimit              = errs.NewCodeError(ErrorCodeClaimingDailyLimit, "you have reached the daily envelope claiming limit")
	ErrAllEnvelopeHasBeenClaimed       = errs.NewCodeError(ErrorCodeAllEnvelopeHasBeenClaimed, "all envelope has been claimed")
	ErrUserAlreadyClaimedThisEnvelope  = errs.NewCodeError(ErrorCodeUserAlreadyClaimedThisEnvelope, "user has already claimed this envelope")
	ErrNoRemainingAmountToRefund       = errs.NewCodeError(ErrorCodeNoRemainingAmountToRefund, "no remaining amount to refund")
	ErrEnvelopeIDNotFoundParam         = errs.NewCodeError(ErrorCodeEnvelopeIDNotFoundParam, "envelope ID not found in path param")

	// transfer
	ErrNoEligibleTransferRefund  = errs.NewCodeError(ErrorCodeNoEligibleTransferRefund, "transfer not eligible to refund")
	ErrTransferIDRequired        = errs.NewCodeError(ErrorCodeTransferIDRequired, "transfer id is required")
	ErrInvalidTransferID         = errs.NewCodeError(ErrorCodeInvalidTransferID, "transfer id is invalid")
	ErrTransferNotFound          = errs.NewCodeError(ErrorCodeTransferNotFound, "transfer not found")
	ErrTransferFailed            = errs.NewCodeError(ErrorCodeTransferFailed, "transfer failed")
	ErrReceiverWalletNotFound    = errs.NewCodeError(ErrorCodeReceiverWalletNotFound, "receiver wallet not found")
	ErrCannotTransferToSelf      = errs.NewCodeError(ErrorCodeCannotTransferToSelf, "cannot transfer to self")
	ErrTransferLimitExceeded     = errs.NewCodeError(ErrorCodeTransferLimitExceeded, "transfer limit exceeded")
	ErrInvalidTransferAmount     = errs.NewCodeError(ErrorCodeInvalidTransferAmount, "invalid transfer amount")
	ErrReceiverWalletLocked      = errs.NewCodeError(ErrorCodeReceiverWalletLocked, "receiver wallet locked")
	ErrTransferAlreadyClaimed    = errs.NewCodeError(ErrorCodeTransferAlreadyClaimed, "transfer has already been claimed")
	ErrTransferExpired           = errs.NewCodeError(ErrorCodeTransferExpired, "transfer has expired")
	ErrTransferInactive          = errs.NewCodeError(ErrorCodeTransferInactive, "transfer is inactive")
	ErrNoEligibleTransfer        = errs.NewCodeError(ErrorCodeNoEligibleTransfer, "transfer not eligible")
	ErrSendingTransferDailyLimit = errs.NewCodeError(ErrorCodeSendingTransferDailyLimit, "you have reached the daily transfer sending limit")
	ErrClaimTransferDailyLimit   = errs.NewCodeError(ErrorCodeClaimTransferDailyLimit, "you have reached the daily transfer claiming limit")
)

func ErrUnsupportedAction(action string) (err error) {
	return fmt.Errorf("unsupported description action: %s", action)
}

func ErrAmountRange(minAmount, maxAmount float64) (err error) {
	return errs.NewCodeError(
		ErrCodeRedEnvelopeAmountRange,
		fmt.Sprintf("amount must be between %.2f and %.2f", minAmount, maxAmount),
	)
}

func ErrGreetingLength(maxLength int) (err error) {
	return errs.NewCodeError(
		ErrCodeExceedGreetingLength,
		fmt.Sprintf("exceed max greeting length, max: %d character", maxLength),
	)
}

func ErrDeactivateDetails(inErr error) (err error) {
	return fmt.Errorf("failed to deactivate envelope details: %w", inErr)
}

func ErrRefund(envID int64, userID string, inErr error) (err error) {
	return fmt.Errorf("[Refund] Failed refund. EnvelopeID=%d UserID=%s Error=%w", envID, userID, inErr)
}
