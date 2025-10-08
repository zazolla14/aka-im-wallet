package imapi

import (
	"sync"
)

type CallerInterface interface{}

type Caller struct {
	imApi           string
	imSecret        string
	defaultIMUserID string
	lock            sync.RWMutex
}

func New(imApi, imSecret, defaultIMUserID string) (caller CallerInterface) {
	return &Caller{
		imApi:           imApi,
		imSecret:        imSecret,
		defaultIMUserID: defaultIMUserID,
		lock:            sync.RWMutex{},
	}
}
