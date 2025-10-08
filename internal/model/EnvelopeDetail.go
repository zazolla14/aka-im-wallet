package entity

import (
	"time"

	"github.com/1nterdigital/aka-im-wallet/pkg/eerrs"
)

const (
	UserSystem = "system"
)

type EnvelopeDetail struct {
	EnvelopeDetailID     int64                `json:"envelope_detail_id" gorm:"column:envelope_detail_id;primaryKey;autoIncrement"`
	EnvelopeID           int64                `json:"envelope_id" gorm:"column:envelope_id;not null"`
	Amount               float64              `json:"amount" gorm:"column:amount;not null"`
	UserID               string               `json:"user_id" gorm:"column:user_id;not null"`
	EnvelopeDetailStatus EnvelopeDetailStatus `json:"envelope_detail_status" gorm:"column:envelope_detail_status;type:enum('pending','claimed','refunded');not null"` //nolint:lll // long enum tag required by GORM
	ClaimedAt            *time.Time           `json:"claimed_at" gorm:"column:claimed_at"`
	IsActive             bool                 `json:"is_active" gorm:"column:is_active"`
	CreatedAt            time.Time            `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CreatedBy            string               `json:"created_by" gorm:"column:created_by"`
	UpdatedAt            time.Time            `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy            string               `json:"updated_by" gorm:"column:updated_by"`
	DeletedAt            *time.Time           `json:"deleted_at" gorm:"column:deleted_at;index"`
	DeletedBy            *string              `json:"deleted_by" gorm:"column:deleted_by"`
	Envelope             *Envelope            `json:"envelope,omitempty" gorm:"foreignKey:EnvelopeID;references:EnvelopeID"`
	Wallet               *Wallet              `json:"wallet,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}

func GetTransactionTypeByEnvelopeType(envType string) (TransactionType, error) {
	switch envType {
	case string(EnvelopeTypeFixed):
		return TransactionTypeEnvelopeFixed, nil
	case string(EnvelopeTypeLucky):
		return TransactionTypeEnvelopeLucky, nil
	case string(EnvelopeTypeSingle):
		return TransactionTypeEnvelopeSingle, nil
	default:
		return "", eerrs.ErrUnsupportedEnvelopeType
	}
}
