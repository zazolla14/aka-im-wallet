package domain

import (
	"errors"
	"time"
)

type WalletRechargeRequestCreate struct {
	Amount float64 `json:"amount"`
	Notes  string  `json:"notes"`
}

type WalletRechargeRequest struct {
	WalletRechargeRequestID string  `json:"walletRechargeRequestID"`
	Amount                  float64 `json:"amount"`
	StatusRequest           string  `json:"statusRequest"`
	Description             string  `json:"description"`
	Status                  string  `json:"status"`
}
type WalletRechargeRequestProcess struct {
	Status    string `json:"status"`
	UpdatedBy string `json:"updatedBy"`
}

type (
	ProcessDepositByAdminRequest struct {
		Amount      float64 `json:"amount"`
		UserID      string  `json:"userID"`
		Description string  `json:"description"`
		OperatedBy  string
	}

	ProcessDepositByAdminResponse struct {
		WalletRechargeRequestID int64 `json:"walletRechargeRequestID"`
	}

	GetListDepositRequest struct {
		WalletID int64
		Page     int32 `json:"page"`
		Limit    int32 `json:"limit"`
		// StartDate: start date of FilterDateBy
		StartDate time.Time `json:"startDate"`
		// EndDate: end date of FilterDateBy
		EndDate time.Time `json:"endDate"`
		UserID  string    `json:"userID"`
		// StatusRequest: ["requested", "approved", "rejected", "failed"]
		StatusRequest string `json:"statusRequest"`
		// SortBy: sort by column
		SortBy string `json:"sortBy"`
		// SortOrder: ["asc", "desc"]
		SortOrder string `json:"sortOrder"`
		// FilterDateBy: ["created_at", "approved_at", "updated_at", "deleted_at"]
		FilterDateBy string `json:"filterDateBy"`
	}

	Deposit struct {
		DepositID     int64      `json:"depositID"`
		WalletID      int64      `json:"walletID"`
		Amount        float64    `json:"amount"`
		CreatedAt     time.Time  `json:"createdAt"`
		ApprovedAt    *time.Time `json:"approvedAt"`
		OperatedBy    *string    `json:"operatedBy"`
		StatusRequest string     `json:"statusRequest"`
		Description   string     `json:"description"`
		CreatedBy     string     `json:"createdBy"`
	}

	GetListDepositResponse struct {
		TotalCount int64      `json:"total"`
		Page       int32      `json:"page"`
		Limit      int32      `json:"limit"`
		Deposits   []*Deposit `json:"deposits"`
	}
)

func (p ProcessDepositByAdminRequest) IsValid() (valid bool, err error) {
	if p.Amount <= 0 {
		return false, errors.New("amount is required")
	}

	if p.UserID == "" {
		return false, errors.New("user id is required")
	}

	return true, nil
}
