package entity

import (
	"time"

	"gorm.io/gorm"
)

type WalletTransaction struct {
	WalletTransactionID int64           `json:"wallet_transaction_id" gorm:"column:wallet_transaction_id;primaryKey;autoIncrement"`
	WalletID            int64           `json:"wallet_id" gorm:"column:wallet_id;not null"`
	Amount              float64         `json:"amount" gorm:"column:amount;not null"`
	TransactionType     TransactionType `gorm:"column:transaction_type;type:enum('transfer','envelope_fixed','envelope_lucky','envelope_single','refund_envelope','refund_transfer','system_adjustment', 'deposit');not null"` //nolint:lll // long enum tag required by GORM
	EntryType           EntryType       `gorm:"column:entry_type;type:enum('credit', 'debit');not null"`
	BeforeBalance       float64         `json:"before_balance" gorm:"column:before_balance;not null"`
	AfterBalance        float64         `json:"after_balance" gorm:"column:after_balance;not null"`
	ReferenceCode       string          `json:"reference_code" gorm:"column:reference_code"`
	IsShown             bool            `json:"is_shown" gorm:"column:is_shown"`
	TransactionDate     time.Time       `json:"transaction_date" gorm:"column:transaction_date"`
	DescriptionEN       string          `json:"description_en" gorm:"column:description_en"`
	DescriptionZH       string          `json:"description_zh" gorm:"column:description_zh"`
	ImpactedItem        int64           `json:"impacted_item" gorm:"column:impacted_item;not null"`
	IsActive            bool            `json:"is_active" gorm:"column:is_active"`
	CreatedAt           time.Time       `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy           string          `json:"created_by" gorm:"column:created_by"`
	UpdatedAt           time.Time       `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy           string          `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt           gorm.DeletedAt  `gorm:"column:deleted_at;index"`
	DeletedBy           *string         `json:"deleted_by" gorm:"column:deleted_by"`
	Wallet              Wallet          `gorm:"foreignKey:WalletID"`
}

type TransactionCount struct {
	TransactionType string
	EntryType       string
	Count           int64
}

// Execute query based on scope
type UserStatResult struct {
	WalletID               int64   `gorm:"column:wallet_id"`
	TotalCredit            float64 `gorm:"column:total_credit"`
	TotalDebit             float64 `gorm:"column:total_debit"`
	CreditTransfer         int64   `gorm:"column:credit_transfer"`
	CreditEnvelopeFixed    int64   `gorm:"column:credit_envelope_fixed"`
	CreditEnvelopeLucky    int64   `gorm:"column:credit_envelope_lucky"`
	CreditEnvelopeSingle   int64   `gorm:"column:credit_envelope_single"`
	CreditDeposit          int64   `gorm:"column:credit_deposit"`
	CreditRefundEnvelope   int64   `gorm:"column:credit_refund_envelope"`
	CreditRefundTransfer   int64   `gorm:"column:credit_refund_transfer"`
	CreditSystemAdjustment int64   `gorm:"column:credit_system_adjustment"`
	DebitTransfer          int64   `gorm:"column:debit_transfer"`
	DebitEnvelopeFixed     int64   `gorm:"column:debit_envelope_fixed"`
	DebitEnvelopeLucky     int64   `gorm:"column:debit_envelope_lucky"`
	DebitEnvelopeSingle    int64   `gorm:"column:debit_envelope_single"`
	DebitDeposit           int64   `gorm:"column:debit_deposit"`
	DebitRefundEnvelope    int64   `gorm:"column:debit_refund_envelope"`
	DebitRefundTransfer    int64   `gorm:"column:debit_refund_transfer"`
	DebitSystemAdjustment  int64   `gorm:"column:debit_system_adjustment"`
	TransactionFrequency   int64   `gorm:"column:transaction_frequency"`
}
