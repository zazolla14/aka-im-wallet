package domain

import "time"

type WalletDomain struct {
	WalletID  int64     `json:"walletID"`
	Balance   float64   `json:"balance"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type WalletCreateRequest struct {
	UserId string `json:"userID"`
}

type WalletActiveRequest struct {
	IsActive  bool   `json:"isActive"`
	UpdatedBy string `json:"updatedBy"`
}

type WalletUpdateBalanceRequest struct {
	Amount          float64 `json:"amount"`
	EntryType       string  `json:"entryType"`
	TransactionType string  `json:"transactionType"`
	ReferenceCode   string  `json:"referenceCode"`
	Description     string  `json:"description"`
	CreatedBy       string  `json:"createdBy"`
}

type (
	Wallet struct {
		ID        int64     `json:"walletID"`
		UserID    string    `json:"userID"`
		Balance   float64   `json:"balance"`
		CreatedAt time.Time `json:"createdAt"`
		CreatedBy string    `json:"createdBy"`
	}
)
