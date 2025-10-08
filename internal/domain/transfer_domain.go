package domain

import (
	"errors"
	"time"
)

type TransferDomain struct {
	FromUserID     string
	ToUserID       string
	Amount         float64
	StatusTransfer string
	Remark         string
}

type (
	CreateTransferRequest struct {
		FromUserID string  `json:"fromUserID"`
		ToUserID   string  `json:"toUserID"`
		Amount     float64 `json:"amount"`
		Remark     string  `json:"remark"`
		CreatedBy  string
	}

	CreateTransferResponse struct {
		TransferID int64 `json:"transferID"`
	}

	MsgKafkaExpiredTransfer struct {
		Counter    int    `json:"counter"`
		TransferID int64  `json:"transferID"`
		OperatedBy string `json:"operatedBy"`
	}

	ClaimTransferRequest struct {
		TransferID    int64 `json:"transferID"`
		ClaimerUserID string
		OperateBy     string
	}

	RefundTransferReq struct {
		TransferID int64 `json:"transferID"`
		OperatedBy string
		UserID     string
	}

	Transfer struct {
		TransferID     int64      `json:"transferID"`
		FromUserID     string     `json:"fromUserID"`
		ToUserID       string     `json:"toUserID"`
		Amount         float64    `json:"amount"`
		StatusTransfer string     `json:"statusTransfer"`
		Remark         string     `json:"remark"`
		ExpiredAt      *time.Time `json:"expiredAt"`
		RefundedAt     *time.Time `json:"refundedAt"`
		ClaimedAt      *time.Time `json:"claimedAt"`
		CreatedAt      time.Time  `json:"createdAt"`
		CreatedBy      string     `json:"createdBy"`
		UpdatedAt      time.Time  `json:"updatedAt"`
		UpdatedBy      string     `json:"updatedBy"`
	}
)

func (c CreateTransferRequest) IsValid() (valid bool, err error) {
	if c.FromUserID == c.ToUserID {
		return false, errors.New("source user id and target user id must be different")
	}

	if c.ToUserID == "" {
		return false, errors.New("target user id is required")
	}

	if c.Amount <= 0 {
		return false, errors.New("amount must be greater than 0")
	}

	return true, nil
}

func (c ClaimTransferRequest) IsValid() (valid bool, err error) {
	if c.TransferID <= 0 {
		return false, errors.New("transfer id is required")
	}

	return true, nil
}

func (c RefundTransferReq) IsValid() error {
	if c.TransferID <= 0 {
		return errors.New("transfer id must be greater than 0")
	}

	return nil
}
