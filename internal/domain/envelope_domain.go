package domain

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	entity "github.com/1nterdigital/aka-im-wallet/internal/model"
)

type EnvelopeDomain struct {
	TotalAmount float64 `json:"total_amount"`
}

type EnvelopeCreateRequest struct {
	UserID       string  `json:"userId"`
	WalletID     int64   `json:"walletId" binding:"required"`
	EnvelopeType string  `json:"envelopeType" binding:"required"`
	TotalAmount  float64 `json:"totalAmount" binding:"required"`
	TotalClaimer int     `json:"totalClaimer" binding:"required"`
	Remarks      string  `json:"remarks" binding:"required"`
	ToUserID     string  `json:"toUserId"`
}

type EnvelopeCreateResponse struct {
	EnvelopeID          int64          `json:"envelopeID"`
	UserID              string         `json:"userID"`
	WalletID            int64          `json:"walletID"`
	TotalAmount         float64        `json:"totalAmount"`
	TotalAmountClaimed  float64        `json:"totalAmountClaimed"`
	TotalAmountRefunded float64        `json:"totalAmountRefunded"`
	MaxNumReceived      int            `json:"maxNumReceived"`
	EnvelopeType        string         `json:"envelopeType"`
	Remarks             string         `json:"remarks"`
	ExpiredAt           *time.Time     `json:"expiredAt"`
	RefundedAt          *time.Time     `json:"refundedAt"`
	IsActive            bool           `json:"isActive"`
	CreatedAt           time.Time      `json:"createdAt"`
	CreatedBy           string         `json:"createdBy"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	UpdatedBy           string         `json:"updatedBy"`
	DeletedAt           gorm.DeletedAt `json:"deletedAt"`
	DeletedBy           *string        `json:"deletedBy"`
}

type EnvelopeDetailOption struct {
	Envelope   *entity.Envelope
	Amounts    []float64
	ReceiverID string
}

type EnvelopeCancelRequest struct {
	EnvelopeID int64 `json:"envelopeId" binding:"required"`
}

type EnvelopeClaimRequest struct {
	UserID     string `json:"userId"`
	EnvelopeID int64  `json:"envelopeId" binding:"required"`
	WalletID   int64  `json:"walletId"`
}

type EnvelopeClaimResponse struct {
	EnvelopeDetailID     int64            `json:"envelope_detail_id" gorm:"column:envelope_detail_id;primaryKey;autoIncrement"`
	EnvelopeID           int64            `json:"envelope_id" gorm:"column:envelope_id;not null"`
	Amount               float64          `json:"amount" gorm:"column:amount;not null"`
	UserID               string           `json:"user_id" gorm:"column:user_id;not null"`
	EnvelopeDetailStatus string           `json:"envelope_detail_status" gorm:"column:envelope_detail_status;type:enum('pending','claimed','refunded');not null"` //nolint:lll // long enum tag required by GORM
	ClaimedAt            *time.Time       `json:"claimed_at" gorm:"column:claimed_at"`
	IsActive             bool             `json:"is_active" gorm:"column:is_active"`
	CreatedAt            time.Time        `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy            string           `json:"created_by" gorm:"column:created_by"`
	UpdatedAt            time.Time        `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy            string           `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt            *time.Time       `json:"deleted_at" gorm:"column:deleted_at;index"`
	DeletedBy            *string          `json:"deleted_by" gorm:"column:deleted_by"`
	Envelope             *entity.Envelope `json:"envelope,omitempty" gorm:"foreignKey:EnvelopeID;references:EnvelopeID"`
	Wallet               *entity.Wallet   `json:"wallet,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}

type MsgKafkaExpiredEnvelope struct {
	Counter    int    `json:"counter"`
	EnvelopeID int64  `json:"envelopeID"`
	OperatedBy string `json:"operatedBy"`
}

func GenerateExpiredAt(duration time.Duration) *time.Time {
	expired := time.Now().Add(duration)
	return &expired
}

func IsEnvelopeExpired(expiredAt *time.Time) bool {
	if expiredAt == nil {
		return false
	}
	return time.Now().After(*expiredAt)
}

func ParseEnvelopeType(s string) (entity.EnvelopeType, error) {
	switch entity.EnvelopeType(s) {
	case entity.EnvelopeTypeLucky, entity.EnvelopeTypeFixed, entity.EnvelopeTypeSingle:
		return entity.EnvelopeType(s), nil
	default:
		return "", fmt.Errorf("invalid envelope type: %s", s)
	}
}
