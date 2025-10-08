package entity

import (
	"time"

	"gorm.io/gorm"
)

type Wallet struct {
	WalletID  int64          `json:"wallet_id" gorm:"column:wallet_id;primaryKey;autoIncrement"`
	UserID    string         `json:"user_id" gorm:"column:user_id;unique;not null"`
	Balance   float64        `json:"balance" gorm:"column:balance;not null"`
	IsActive  bool           `json:"is_active" gorm:"column:is_active;not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy string         `json:"created_by" gorm:"column:created_by"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy string         `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy *string        `json:"deleted_by" gorm:"column:deleted_by"`
}
