package entity

type TransactionType string

const (
	TransactionTypeTransfer         TransactionType = "transfer"
	TransactionTypeEnvelopeFixed    TransactionType = "envelope_fixed"
	TransactionTypeEnvelopeLucky    TransactionType = "envelope_lucky"
	TransactionTypeEnvelopeSingle   TransactionType = "envelope_single"
	TransactionTypeRefundEnvelope   TransactionType = "refund_envelope"
	TransactionTypeRefundTransfer   TransactionType = "refund_transfer"
	TransactionTypeSystemAdjustment TransactionType = "system_adjustment"
	TransactionTypeDeposit          TransactionType = "deposit"
)

var validTransactionType = map[TransactionType]bool{
	TransactionTypeTransfer:         true,
	TransactionTypeEnvelopeFixed:    true,
	TransactionTypeEnvelopeLucky:    true,
	TransactionTypeEnvelopeSingle:   true,
	TransactionTypeRefundEnvelope:   true,
	TransactionTypeRefundTransfer:   true,
	TransactionTypeSystemAdjustment: true,
	TransactionTypeDeposit:          true,
}

func (e TransactionType) IsValid() bool {
	_, exist := validTransactionType[e]
	return exist
}

func (e TransactionType) String() string {
	return string(e)
}
