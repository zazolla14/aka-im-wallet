package entity

type StatusRequest string

const (
	StatusRequestRequested StatusRequest = "requested"
	StatusRequestApproved  StatusRequest = "approved"
	StatusRequestRejected  StatusRequest = "rejected"
	StatusRequestFailed    StatusRequest = "failed"
)

var validStatusRequest = map[StatusRequest]bool{
	StatusRequestRequested: true,
	StatusRequestApproved:  true,
	StatusRequestRejected:  true,
	StatusRequestFailed:    true,
}

func (e StatusRequest) IsValid() bool {
	_, exist := validStatusRequest[e]
	return exist
}

func (e StatusRequest) String() string {
	return string(e)
}
