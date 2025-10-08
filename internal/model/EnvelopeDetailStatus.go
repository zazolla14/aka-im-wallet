package entity

type EnvelopeDetailStatus string

const (
	EnvelopePending  EnvelopeDetailStatus = "pending"
	EnvelopeClaimed  EnvelopeDetailStatus = "claimed"
	EnvelopeRefunded EnvelopeDetailStatus = "refunded"
	EnvelopeExpired  EnvelopeDetailStatus = "expired"
	EnvelopePartial  EnvelopeDetailStatus = "partial-claimed"
)

var validEnvelopeDetailStatus = map[EnvelopeDetailStatus]bool{
	EnvelopePending:  true,
	EnvelopeClaimed:  true,
	EnvelopeRefunded: true,
	EnvelopeExpired:  true,
	EnvelopePartial:  true,
}

func (e EnvelopeDetailStatus) IsValid() bool {
	_, exist := validEnvelopeDetailStatus[e]
	return exist
}

func (e EnvelopeDetailStatus) String() string {
	return string(e)
}
