package entity

type EnvelopeType string

const (
	EnvelopeTypeLucky  EnvelopeType = "lucky"
	EnvelopeTypeFixed  EnvelopeType = "fixed"
	EnvelopeTypeSingle EnvelopeType = "single"
)

var validEnvelopeType = map[EnvelopeType]bool{
	EnvelopeTypeLucky:  true,
	EnvelopeTypeFixed:  true,
	EnvelopeTypeSingle: true,
}

func (e EnvelopeType) IsValid() bool {
	_, exist := validEnvelopeType[e]
	return exist
}

func (e EnvelopeType) String() string {
	return string(e)
}
