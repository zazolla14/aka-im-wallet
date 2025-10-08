package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/1nterdigital/aka-im-tools/errs"
	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type PaginationRequest struct {
	Page  int32 `json:"page" validate:"min=1"`
	Limit int32 `json:"limit" validate:"min=1"`
}

// DashboardTransactionVolumeResponse 1. Dashboard Transaction Volume
type DashboardTransactionVolumeResponse struct {
	Credit TransactionTypeCount `json:"credit" validate:"required"`
	Debit  TransactionTypeCount `json:"debit" validate:"required"`
}

type TransactionTypeCount struct {
	Transfer         int64 `json:"transfer" validate:"min=0"`
	EnvelopeFixed    int64 `json:"envelope_fixed" validate:"min=0"`
	EnvelopeLucky    int64 `json:"envelope_lucky" validate:"min=0"`
	EnvelopeSingle   int64 `json:"envelope_single" validate:"min=0"`
	Deposit          int64 `json:"deposit" validate:"min=0"`
	RefundEnvelope   int64 `json:"refund_envelope" validate:"min=0"`
	RefundTransfer   int64 `json:"refund_transfer" validate:"min=0"`
	SystemAdjustment int64 `json:"system_adjustment" validate:"min=0"`
}

// GetListTransactionMonitoringRequest 2. Get List Transaction
type GetListTransactionMonitoringRequest struct {
	PaginationRequest
	FromTransactionDate time.Time `json:"from_transaction_date"`
	ToTransactionDate   time.Time `json:"to_transaction_date"`
	TransactionType     string    `json:"transaction_type" validate:"omitempty,oneof=transfer deposit withdrawal"`
	EntryType           string    `json:"entry_type" validate:"omitempty,oneof=credit debit"`
	FromAmount          float64   `json:"from_amount" validate:"min=0"`
	ToAmount            float64   `json:"to_amount" validate:"min=0"`
	WalletID            int64     `json:"wallet_id" validate:"min=0"`
	ReferenceCode       string    `json:"reference_code"`
	WalletTransactionID string    `json:"wallet_transaction_id"`
}

type GetListTransactionMonitoringResponse struct {
	Page       int32                   `json:"page"`
	Limit      int32                   `json:"limit"`
	TotalCount int64                   `json:"total_count"`
	TotalPages int32                   `json:"total_pages"`
	HasNext    bool                    `json:"has_next"`
	HasPrev    bool                    `json:"has_prev"`
	Data       []*WalletTransactionDTO `json:"data"`
}

type WalletTransactionDTO struct {
	WalletTransactionID int64     `json:"wallet_transaction_id"`
	WalletID            int64     `json:"wallet_id"`
	Amount              float64   `json:"amount"`
	TransactionType     string    `json:"transaction_type"`
	EntryType           string    `json:"entry_type"`
	BeforeBalance       float64   `json:"before_balance"`
	AfterBalance        float64   `json:"after_balance"`
	ReferenceCode       string    `json:"reference_code"`
	TransactionDate     time.Time `json:"transaction_date"`
	DescriptionEN       string    `json:"description_en"`
	DescriptionZH       string    `json:"description_zh"`
	ImpactedItem        int64     `json:"impacted_item"`
	CreatedAt           time.Time `json:"created_at"`
}

// GetListEnvelopeRequest 3. Get List Envelope
type GetListEnvelopeRequest struct {
	PaginationRequest
	FromTotalAmount float64 `json:"from_total_amount" validate:"min=0"`
	ToTotalAmount   float64 `json:"to_total_amount" validate:"min=0"`
	UserID          string  `json:"user_id" validate:"omitempty,min=3,max=50"`
	MaxNumReceived  int32   `json:"max_num_received" validate:"min=0,max=10000"`
	EnvelopeType    string  `json:"envelope_type" validate:"omitempty,oneof=fixed lucky single"`
	FromExpiredAt   int64   `json:"from_expired_at" validate:"min=0"`
	ToExpiredAt     int64   `json:"to_expired_at" validate:"min=0"`
	FromRefundedAt  int64   `json:"from_refunded_at" validate:"min=0"`
	ToRefundedAt    int64   `json:"to_refunded_at" validate:"min=0"`
}

type GetListEnvelopeResponse struct {
	Page       int32          `json:"page"`
	Limit      int32          `json:"limit"`
	TotalCount int64          `json:"total_count"`
	TotalPages int32          `json:"total_pages"`
	HasNext    bool           `json:"has_next"`
	HasPrev    bool           `json:"has_prev"`
	Data       []*EnvelopeDTO `json:"data"`
}

type EnvelopeDTO struct {
	EnvelopeID          int64      `json:"envelope_id"`
	UserID              string     `json:"user_id"`
	TotalAmount         float64    `json:"total_amount"`
	TotalAmountClaimed  float64    `json:"total_amount_claimed"`
	TotalAmountRefunded float64    `json:"total_amount_refunded"`
	RemainingAmount     float64    `json:"remaining_amount"`
	MaxNumReceived      int        `json:"max_num_received"`
	EnvelopeType        string     `json:"envelope_type"`
	Remarks             string     `json:"remarks"`
	ExpiredAt           *time.Time `json:"expired_at"`
	RefundedAt          *time.Time `json:"refunded_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	Status              string     `json:"status"`
	IsExpired           bool       `json:"is_expired"`
	IsRefunded          bool       `json:"is_refunded"`
}

// GetEnvelopeDetailRequest 4. Get Envelope Detail
type GetEnvelopeDetailRequest struct {
	EnvelopeID   int64  `json:"envelope_id" validate:"required,min=1"`
	DetailStatus string `json:"detail_status,omitempty" validate:"omitempty,oneof=pending claimed expired refunded"`
	Claimed      *bool  `json:"claimed,omitempty"`
	UserID       string `json:"user_id,omitempty" validate:"omitempty,min=3,max=50"`
}

type GetEnvelopeDetailResponse struct {
	EnvelopeID           int64                `json:"envelope_id"`
	UserID               string               `json:"user_id"`
	TotalAmount          float64              `json:"total_amount"`
	TotalAmountClaimed   float64              `json:"total_amount_claimed"`
	TotalAmountRefunded  float64              `json:"total_amount_refunded"`
	RemainingAmount      float64              `json:"remaining_amount"`
	MaxNumReceived       int                  `json:"max_num_received"`
	EnvelopeType         string               `json:"envelope_type"`
	Remarks              string               `json:"remarks"`
	ExpiredAt            *time.Time           `json:"expired_at"`
	RefundedAt           *time.Time           `json:"refunded_at,omitempty"`
	CreatedAt            time.Time            `json:"created_at"`
	Status               string               `json:"status"`
	IsExpired            bool                 `json:"is_expired"`
	IsRefunded           bool                 `json:"is_refunded"`
	CompletionPercentage float64              `json:"completion_percentage"`
	TimeUntilExpiry      *time.Duration       `json:"time_until_expiry,omitempty"`
	Statistics           *EnvelopeStatistics  `json:"statistics"`
	Details              []*EnvelopeDetailDTO `json:"details"`
}

type EnvelopeDetailDTO struct {
	EnvelopeDetailID     int64          `json:"envelope_detail_id"`
	EnvelopeID           int64          `json:"envelope_id"`
	Amount               float64        `json:"amount"`
	UserID               string         `json:"user_id"`
	EnvelopeDetailStatus string         `json:"envelope_detail_status"`
	ClaimedAt            *time.Time     `json:"claimed_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	IsClaimed            bool           `json:"is_claimed"`
	TimeSinceClaimed     *time.Duration `json:"time_since_claimed,omitempty"`
}

type EnvelopeStatistics struct {
	TotalDetails         int     `json:"total_details"`
	ClaimedCount         int     `json:"claimed_count"`
	PendingCount         int     `json:"pending_count"`
	ExpiredCount         int     `json:"expired_count"`
	RefundedCount        int     `json:"refunded_count"`
	UniqueClaimantsCount int     `json:"unique_claimants_count"`
	TotalClaimedAmount   float64 `json:"total_claimed_amount"`
}

// GetTransferHistoryRequest 5. GetTransfer History Request
type GetTransferHistoryRequest struct {
	PaginationRequest
	FromUserID     string  `json:"from_user_id" validate:"omitempty,min=3,max=50"`
	ToUserID       string  `json:"to_user_id" validate:"omitempty,min=3,max=50"`
	StatusTransfer string  `json:"status_transfer" validate:"omitempty,oneof=pending claimed expired refunded canceled"`
	FromAmount     float64 `json:"from_amount" validate:"min=0"`
	ToAmount       float64 `json:"to_amount" validate:"min=0"`
	FromExpiredAt  int64   `json:"from_expired_at" validate:"min=0"`
	ToExpiredAt    int64   `json:"to_expired_at" validate:"min=0"`
	FromClaimedAt  int64   `json:"from_claimed_at" validate:"min=0"`
	ToClaimedAt    int64   `json:"to_claimed_at" validate:"min=0"`
}

type GetTransferHistoryResponse struct {
	Page       int32               `json:"page"`
	Limit      int32               `json:"limit"`
	TotalCount int64               `json:"total_count"`
	TotalPages int32               `json:"total_pages"`
	HasNext    bool                `json:"has_next"`
	HasPrev    bool                `json:"has_prev"`
	Statistics *TransferStatistics `json:"statistics"`
	Data       []*TransferDTO      `json:"data"`
}

type TransferDTO struct {
	TransferID       int64          `json:"transfer_id"`
	FromUserID       string         `json:"from_user_id"`
	ToUserID         string         `json:"to_user_id"`
	Amount           float64        `json:"amount"`
	StatusTransfer   string         `json:"status_transfer"`
	Remark           string         `json:"remark"`
	ExpiredAt        *time.Time     `json:"expired_at,omitempty"`
	RefundedAt       *time.Time     `json:"refunded_at,omitempty"`
	ClaimedAt        *time.Time     `json:"claimed_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	IsClaimed        bool           `json:"is_claimed"`
	IsExpired        bool           `json:"is_expired"`
	IsRefunded       bool           `json:"is_refunded"`
	IsPending        bool           `json:"is_pending"`
	TimeSinceClaimed *time.Duration `json:"time_since_claimed,omitempty"`
	TimeUntilExpiry  *time.Duration `json:"time_until_expiry,omitempty"`
	ProcessingTime   *time.Duration `json:"processing_time,omitempty"`
}

type TransferStatistics struct {
	TotalTransfers       int     `json:"total_transfers"`
	TotalAmount          float64 `json:"total_amount"`
	AverageAmount        float64 `json:"average_amount"`
	PendingCount         int     `json:"pending_count"`
	PendingAmount        float64 `json:"pending_amount"`
	ClaimedCount         int     `json:"claimed_count"`
	ClaimedAmount        float64 `json:"claimed_amount"`
	ExpiredCount         int     `json:"expired_count"`
	ExpiredAmount        float64 `json:"expired_amount"`
	RefundedCount        int     `json:"refunded_count"`
	RefundedAmount       float64 `json:"refunded_amount"`
	UniqueSendersCount   int     `json:"unique_senders_count"`
	UniqueReceiversCount int     `json:"unique_receivers_count"`
}

// GetTop10UsersRequest represents the request for getting top 10 users
type GetTop10UsersRequest struct {
	SortByTotalAmount bool   `json:"sort_by_total_amount" validate:""`
	Scope             string `json:"scope" validate:"required,oneof=all_time year month"`
}

// GetTop10UsersResponse represents the response for getting top 10 users
type GetTop10UsersResponse struct {
	Time            string                  `json:"time"`
	ListTransaction []*HighFrequencyUserDTO `json:"list_transaction"`
}

// HighFrequencyUserDTO represents a high frequency user data transfer object
type HighFrequencyUserDTO struct {
	WalletID    int64                `json:"wallet_id"`
	TotalCredit float64              `json:"total_credit"`
	Credit      TransactionTypeCount `json:"credit"`
	TotalDebit  float64              `json:"total_debit"`
	Debit       TransactionTypeCount `json:"debit"`
	Position    int                  `json:"position"`
}

// Validate validates the request parameters
func (r *GetListTransactionMonitoringRequest) Validate() error {
	var errors []string

	if r.Page < 1 {
		errors = append(errors, "page must be greater than 0")
	}

	if r.Limit < 1 {
		errors = append(errors, "limit must be greater than 0")
	}

	if !r.FromTransactionDate.IsZero() && !r.ToTransactionDate.IsZero() {
		if r.FromTransactionDate.After(r.ToTransactionDate) {
			errors = append(errors, "from_transaction_date cannot be after to_transaction_date")
		}

		if r.ToTransactionDate.Sub(r.FromTransactionDate) > 365*24*time.Hour {
			errors = append(errors, "date range cannot exceed 1 year")
		}
	}

	if r.FromAmount < 0 {
		errors = append(errors, "from_amount cannot be negative")
	}

	if r.ToAmount < 0 {
		errors = append(errors, "to_amount cannot be negative")
	}

	if r.FromAmount > 0 && r.ToAmount > 0 && r.FromAmount > r.ToAmount {
		errors = append(errors, "from_amount cannot be greater than to_amount")
	}

	if r.WalletID < 0 {
		errors = append(errors, "wallet_id cannot be negative")
	}

	if len(errors) > 0 {
		return errs.ErrArgs.WithDetail(fmt.Sprintf("validation errors: %s", strings.Join(errors, "; "))).Wrap()
	}

	return nil
}

// Validate validates the request parameters
func (r *GetListEnvelopeRequest) Validate() error {
	var errors []string

	errors = append(errors, r.validatePaginationRequest()...)
	errors = append(errors, r.validateMaxNumReceivedRequest()...)
	errors = append(errors, r.validateEnvelopeTypeRequest()...)
	errors = append(errors, r.validateAmountRequest()...)
	errors = append(errors, r.validateExpiredRequest()...)

	if len(errors) > 0 {
		return errs.ErrArgs.WithDetail(fmt.Sprintf("validation errors: %s", strings.Join(errors, "; "))).Wrap()
	}

	return nil
}

func (r *GetListEnvelopeRequest) validatePaginationRequest() (errors []string) {
	if r.Page < 1 {
		errors = append(errors, "page must be greater than 0")
	}

	if r.Limit < 1 {
		errors = append(errors, "limit must be greater than 0")
	}

	return errors
}

func (r *GetListEnvelopeRequest) validateAmountRequest() (errors []string) {
	if r.FromTotalAmount < 0 || r.ToTotalAmount < 0 {
		errors = append(errors, "amount values cannot be negative")
	}

	if r.FromTotalAmount > 0 && r.ToTotalAmount > 0 && r.FromTotalAmount > r.ToTotalAmount {
		errors = append(errors, "from_total_amount cannot be greater than to_total_amount")
	}

	return errors
}

func (r *GetListEnvelopeRequest) validateMaxNumReceivedRequest() (errors []string) {
	if r.MaxNumReceived < 0 {
		errors = append(errors, "max_num_received must be greater than 0")
	}

	return errors
}

func (r *GetListEnvelopeRequest) validateEnvelopeTypeRequest() (errors []string) {
	if r.EnvelopeType != "" && !entity.EnvelopeType(r.EnvelopeType).IsValid() {
		errors = append(errors, "invalid envelope_type")
	}

	return errors
}

func (r *GetListEnvelopeRequest) validateExpiredRequest() (errors []string) {
	if r.FromExpiredAt < 0 || r.ToExpiredAt < 0 || r.FromRefundedAt < 0 || r.ToRefundedAt < 0 {
		errors = append(errors, "timestamp values cannot be negative")
	}

	return errors
}

// Validate validates the request parameters
func (r *GetEnvelopeDetailRequest) Validate() error {
	var errors []string

	if r.EnvelopeID <= 0 {
		errors = append(errors, "envelope_id must be positive")
	}

	if r.DetailStatus != "" && !entity.EnvelopeDetailStatus(r.DetailStatus).IsValid() {
		errors = append(errors, "invalid detail_status")
	}

	if len(errors) > 0 {
		return errs.ErrArgs.WithDetail(fmt.Sprintf("validation errors: %s", strings.Join(errors, "; "))).Wrap()
	}

	return nil
}

// Validate validates the request parameters
func (r *GetTransferHistoryRequest) Validate() error {
	var errors []string

	if r.Page < 1 {
		errors = append(errors, "page must be greater than 0")
	}

	if r.Limit < 1 {
		errors = append(errors, "limit must be greater than 0")
	}

	if r.FromUserID != "" && r.ToUserID != "" && r.FromUserID == r.ToUserID {
		errors = append(errors, "from_user_id and to_user_id cannot be the same")
	}

	if r.StatusTransfer != "" && !entity.StatusTransfer(r.StatusTransfer).IsValid() {
		errors = append(errors, "invalid status_transfer")
	}

	if r.FromAmount < 0 || r.ToAmount < 0 {
		errors = append(errors, "amount values cannot be negative")
	}

	if r.FromAmount > 0 && r.ToAmount > 0 && r.FromAmount > r.ToAmount {
		errors = append(errors, "from_amount cannot be greater than to_amount")
	}

	if len(errors) > 0 {
		return errs.ErrArgs.WithDetail(fmt.Sprintf("validation errors: %s", strings.Join(errors, "; "))).Wrap()
	}

	return nil
}

// Top10ScopeType 6. Top 10 High-Frequency Users
type Top10ScopeType string

const (
	ScopeAllTime Top10ScopeType = "all_time"
	ScopeYear    Top10ScopeType = "year"
	ScopeMonth   Top10ScopeType = "month"
)

var validScopes = map[Top10ScopeType]bool{
	ScopeAllTime: true,
	ScopeYear:    true,
	ScopeMonth:   true,
}

func (r Top10ScopeType) String() string {
	return string(r)
}

func (r Top10ScopeType) IsValidate() bool {
	_, exist := validScopes[r]
	return exist
}

// Validate validates the request parameters
func (r GetTop10UsersRequest) Validate() error {
	var errors []string

	if r.Scope != "" && !Top10ScopeType(r.Scope).IsValidate() {
		errors = append(errors, "invalid scope")
	}

	if len(errors) > 0 {
		return errs.ErrArgs.WithDetail(fmt.Sprintf("validation errors: %s", strings.Join(errors, "; "))).Wrap()
	}

	return nil
}
