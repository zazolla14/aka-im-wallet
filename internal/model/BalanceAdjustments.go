package entity

import (
	"time"

	"gorm.io/gorm"
)

type BalanceAdjustments struct {
	BalanceAdjustmentID int64          `json:"balance_adjustment_id" gorm:"column:balance_adjustment_id;primaryKey;autoIncrement"`
	WalletID            int64          `json:"wallet_id" gorm:"column:wallet_id;not null"`
	Amount              float64        `json:"amount" gorm:"column:amount;not null"`
	Reason              string         `json:"reason" gorm:"column:reason;not null"`
	Description         string         `json:"description" gorm:"column:description;not null"`
	IsActive            bool           `json:"is_active" gorm:"column:is_active;index"`
	CreatedAt           time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy           string         `json:"created_by" gorm:"column:created_by"`
	UpdatedAt           time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy           string         `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy           *string        `json:"deleted_by" gorm:"column:deleted_by"`
}
