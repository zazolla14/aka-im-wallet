package entity

type EntryType string

const (
	EntryTypeCredit EntryType = "credit"
	EntryTypeDebit  EntryType = "debit"
)

var validEntryType = map[EntryType]bool{
	EntryTypeCredit: true,
	EntryTypeDebit:  true,
}

func (e EntryType) IsValid() bool {
	_, exist := validEntryType[e]
	return exist
}

func (e EntryType) String() string {
	return string(e)
}
