package domain

var (
	TransactionRefundTransferEN = "refund transfer from %d\n"
	TransactionRefundTransferZN = "refund transfer from %d\n"

	TransactionRefundEnvelopeEN = "refund envelope from %d\n"
	TransactionRefundEnvelopeZN = "refund envelope from %d\n"
)

var (
	FormatReferenceCodeRefundTransfer = "R-TRF:%d.WLT:%d\n"
	FormatReferenceCodeRefundEnvelope = "R-EVL:%d.WLT:%d\n"
)

const (
	KafkaProducerOperator = "SYSTEM:PRODUCER"
)

type ctxKey string

const (
	KeyOperatedBy ctxKey = "operatedBy"
)

const (
	LayoutFilterDate       = "2006-01-02"
	DefaultPage      int64 = 1
	DefaultLimit     int64 = 10
)

type PublisherKey string

const (
	PublisherKeyRefundTransferEnvelope PublisherKey = "refundTransferEnvelope"
)

const (
	BatchSizeCreateEnvelope   = 100
	TimeoutContextIn30Seconds = 30
)

const (
	EnvelopeMinimumAmount = 0.01
)
