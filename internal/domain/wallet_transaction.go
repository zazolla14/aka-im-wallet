package domain

import (
	"time"

	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

type (
	CreateTransactionReq struct {
		WalletID        int64   `json:"walletID"`
		TransactionType string  `json:"transactionType"`
		Entrytype       string  `json:"entryType"`
		Amount          float64 `json:"amount"`
		DescriptionEn   string  `json:"descriptionEn"`
		DescriptionZh   string  `json:"descriptionZh"`
		ReferenceCode   string  `json:"referenceCode"`
		ImpactedItem    int64   `json:"impactedItem"`
		CreatedBy       string  `json:"createdBy"`
	}

	WalletTransaction struct {
		WalletTransactionID int64     `json:"walletTransactionID"`
		WalletID            int64     `json:"walletID"`
		TransactionType     string    `json:"transactionType"`
		EntryType           string    `json:"entryType"`
		Amount              float64   `json:"amount"`
		BeforeBalance       float64   `json:"beforeBalance"`
		AfterBalance        float64   `json:"afterBalance"`
		DescriptionEn       string    `json:"descriptionEn"`
		DescriptionZh       string    `json:"descriptionZh"`
		ReferenceCode       string    `json:"referenceCode"`
		ImpactedItem        int64     `json:"impactedItem"`
		IsShown             bool      `json:"isShown"`
		TransactionDate     time.Time `json:"transactionDate"`
	}

	GetListTransactionRequest struct {
		UserID    string    `json:"userID"`
		WalletID  int64     `json:"walletID"`
		Page      int32     `json:"page"`
		Limit     int32     `json:"limit"`
		StartDate time.Time `json:"startDate"`
		EndDate   time.Time `json:"endDate"`
	}

	GetListTransactionResponse struct {
		Page         int32                `json:"page"`
		Limit        int32                `json:"limit"`
		TotalCount   int64                `json:"total"`
		Transactions []*WalletTransaction `json:"transactions"`
	}

	TransactionAction struct {
		Action          string  `json:"action"`
		EntryType       string  `json:"entryType"`
		TransactionType string  `json:"transactionType"`
		Amount          float64 `json:"amount"`
		WalletID        int64   `json:"walletID"`
		UserID          string  `json:"userId"`
	}
)

func (e *CreateTransactionReq) IsValid() (valid bool, err error) {
	if e.WalletID <= 0 {
		return false, eerrs.ErrWalletIDRequired
	}

	if !entity.EntryType(e.Entrytype).IsValid() {
		return false, eerrs.ErrEntryTypeInvalid
	}

	if !entity.TransactionType(e.TransactionType).IsValid() {
		return false, eerrs.ErrTransactionTypeInvalid
	}

	if e.DescriptionEn == "" {
		return false, eerrs.ErrDescriptionEnRequired
	}

	if e.DescriptionZh == "" {
		return false, eerrs.ErrDescriptionZhRequired
	}

	if e.ReferenceCode == "" {
		return false, eerrs.ErrReferenceCodeRequired
	}

	if e.ImpactedItem <= 0 {
		return false, eerrs.ErrImpactedItemRequired
	}

	return true, nil
}
