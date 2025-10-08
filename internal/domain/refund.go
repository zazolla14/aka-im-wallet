package domain

type ManualRefundArgs struct {
	TransferIDs []int64 `json:"transferIDs"`
	EnvelopeIDs []int64 `json:"envelopeIDs"`
}
