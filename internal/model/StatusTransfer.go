package entity

type StatusTransfer string

const (
	StatusTransferPending  StatusTransfer = "pending"
	StatusTransferClaimed  StatusTransfer = "claimed"
	StatusTransferRefunded StatusTransfer = "refunded"
)

var validStatusTransfer = map[StatusTransfer]bool{
	StatusTransferPending:  true,
	StatusTransferClaimed:  true,
	StatusTransferRefunded: true,
}

func (e StatusTransfer) IsValid() bool {
	_, exist := validStatusTransfer[e]
	return exist
}

func (e StatusTransfer) String() string {
	return string(e)
}
