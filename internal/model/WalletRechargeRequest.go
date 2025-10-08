package entity

import (
	"time"

	"gorm.io/gorm"
)

type WalletRechargeRequest struct {
	WalletRechargeRequestID int64          `json:"wallet_recharge_request_id" gorm:"column:wallet_recharge_request_id;primaryKey;autoIncrement"` //nolint:lll // long enum tag required by GORM
	WalletID                int64          `json:"wallet_id" gorm:"column:wallet_id;not null"`
	Amount                  float64        `json:"amount" gorm:"column:amount;not null"`
	StatusRequest           StatusRequest  `json:"status_request" gorm:"column:status_request;type:enum('requested', 'approved', 'rejected', 'failed');default:'requested'"` //nolint:lll // long enum tag required by GORM
	Description             string         `json:"description" gorm:"column:description;type:text"`
	ApprovedAt              *time.Time     `json:"approved_at" gorm:"column:approved_at"`
	OperatedBy              *string        `json:"operated_by" gorm:"column:operated_by"`
	IsActive                bool           `json:"is_active" gorm:"column:is_active;not null"`
	CreatedAt               time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy               string         `json:"created_by" gorm:"column:created_by"`
	UpdatedAt               time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy               string         `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt               gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy               *string        `json:"deleted_by" gorm:"column:deleted_by"`
}
