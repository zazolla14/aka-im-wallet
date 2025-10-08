package domain

import (
	"errors"
	"time"
)

type (
	BalanceAdjustmentByAdminRequest struct {
		Amount      float64 `json:"amount"`
		UserID      string  `json:"userID"`
		Reason      string  `json:"reason"`
		Description string  `json:"description"`
		OperatedBy  string
	}

	BalanceAdjustmentByAdminResponse struct {
		BalanceAdjustmentID int64 `json:"balanceAdjustmentID"`
	}

	GetListbalanceAjustmentRequest struct {
		WalletID int64
		Page     int32 `json:"page"`
		Limit    int32 `json:"limit"`
		// StartDate: start date of FilterDateBy
		StartDate time.Time `json:"startDate"`
		// EndDate: end date of FilterDateBy
		EndDate time.Time `json:"endDate"`
		UserID  string    `json:"userID"`
		// SortBy: sort by column
		SortBy string `json:"sortBy"`
		// SortOrder: ["asc", "desc"]
		SortOrder string `json:"sortOrder"`
		// FilterDateBy: ["created_at", "updated_at"]
		FilterDateBy string `json:"filterDateBy"`
	}

	BalanceAdjustment struct {
		BalanceAdjustmentID int64     `json:"balanceAdjustmentID"`
		WalletID            int64     `json:"walletID"`
		Amount              float64   `json:"amount"`
		Reason              string    `json:"reason"`
		Description         string    `json:"description"`
		CreatedAt           time.Time `json:"createdAt"`
		CreatedBy           string    `json:"createdBy"`
		UpdatedAt           time.Time `json:"updatedAt"`
		UpdatedBy           string    `json:"updatedBy"`
	}

	GetListBalanceAdjustmentResponse struct {
		TotalCount         int64                `json:"total"`
		Page               int32                `json:"page"`
		Limit              int32                `json:"limit"`
		BalanceAdjustments []*BalanceAdjustment `json:"balanceAdjustments"`
	}
)

func (r *BalanceAdjustmentByAdminRequest) Validate() error {
	if r.Amount == 0 {
		return errors.New("amount is required")
	}
	if r.UserID == "" {
		return errors.New("userID is required")
	}
	if r.Reason == "" {
		return errors.New("reason is required")
	}

	return nil
}
