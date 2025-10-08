package entity

import (
	"time"

	"gorm.io/gorm"
)

const (
	ActionCreate = "create"
	ActionClaim  = "claim"
	ActionRefund = "refund"

	MaxEnvelopeSendPerDay        = 500
	MaxEnvelopeClaimPerDay       = 200
	MaxGreetingCharacters        = 50
	MaxAllowedSendEnvelopeAmount = 1000
)

type Envelope struct {
	EnvelopeID          int64          `json:"envelope_id" gorm:"column:envelope_id;primaryKey;autoIncrement"`
	UserID              string         `json:"user_id" gorm:"column:user_id;not null"`
	WalletID            int64          `json:"wallet_id" gorm:"column:wallet_id;not null"`
	TotalAmount         float64        `json:"total_amount" gorm:"column:total_amount;not null"`
	TotalAmountClaimed  float64        `json:"total_amount_claimed" gorm:"column:total_amount_claimed;not null"`
	TotalAmountRefunded float64        `json:"total_amount_refunded" gorm:"column:total_amount_refunded;not null"`
	MaxNumReceived      int            `json:"max_num_received" gorm:"column:max_num_received;not null"`
	EnvelopeType        string         `gorm:"column:envelope_type;type:enum('lucky','fixed','single');not null"`
	Remarks             string         `json:"remarks" gorm:"column:remarks"`
	ExpiredAt           *time.Time     `json:"expired_at" gorm:"column:expired_at"`
	RefundedAt          *time.Time     `json:"refunded_at" gorm:"column:refunded_at"`
	IsActive            bool           `json:"is_active" gorm:"column:is_active"`
	CreatedAt           time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy           string         `json:"created_by" gorm:"column:created_by"`
	UpdatedAt           time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy           string         `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy           *string        `json:"deleted_by" gorm:"column:deleted_by"`
}

type ClaimStatus struct {
	TotalClaimed   int64
	UserHasClaimed int64
}

var EnvelopeDescFormat = map[string]struct {
	En        string
	Zh        string
	RefPrefix string
}{
	ActionCreate: {
		En:        "userID:%s created #%d envelope with amount %.0f",
		Zh:        "用户ID:%s 创建了 #%d 红包，金额为 %.0f",
		RefPrefix: "EN",
	},
	ActionClaim: {
		En:        "userID:%s claimed envelope #%d with amount %.0f",
		Zh:        "用户ID:%s 领取了红包 #%d，金额为 %.0f",
		RefPrefix: "CE",
	},
	ActionRefund: {
		En:        "userID:%s received refund from envelope #%d with amount %.0f",
		Zh:        "用户ID:%s 从红包 #%d 收到退款，金额为 %.0f",
		RefPrefix: "RE",
	},
}
