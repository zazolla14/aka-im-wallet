package entity

import (
	"time"

	"gorm.io/gorm"
)

const (
	MaxTransferSendPerDay        = 200
	MaxTransferClaimPerDay       = 200
	MaxAllowedSendTransferAmount = 5000
)

type Transfer struct {
	TransferID     int64          `json:"transfer_id" gorm:"column:transfer_id;primaryKey;autoIncrement"`
	FromUserID     string         `json:"from_user_id" gorm:"column:from_user_id;type:varchar(20);not null"`
	ToUserID       string         `json:"to_user_id" gorm:"column:to_user_id;type:varchar(20);not null"`
	Amount         float64        `json:"amount" gorm:"column:amount;type:decimal(15,2)"`
	StatusTransfer StatusTransfer `json:"status_transfer" gorm:"column:status_transfer;type:enum('pending', 'claimed', 'refunded');default:'pending'"` //nolint:lll // long enum tag required by GORM
	Remark         string         `json:"remark" gorm:"column:remark;type:varchar(255)"`
	ExpiredAt      *time.Time     `json:"expired_at" gorm:"column:expired_at"`
	RefundedAt     *time.Time     `json:"refunded_at" gorm:"column:refunded_at"`
	ClaimedAt      *time.Time     `json:"claimed_at" gorm:"column:claimed_at"`
	IsActive       bool           `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt      time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy      string         `json:"created_by" gorm:"column:created_by;default:system"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy      string         `json:"updated_by" gorm:"column:updated_by;default:system"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
	DeletedBy      *string        `json:"deleted_by" gorm:"column:deleted_by"`
}
