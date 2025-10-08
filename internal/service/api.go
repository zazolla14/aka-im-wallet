package service

import (
	"github.com/1nterdigital/aka-im-wallet/internal/api/util"
	"github.com/1nterdigital/aka-im-wallet/internal/usecase"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/imapi"
)

func New(imApiCaller imapi.CallerInterface, api *util.Api, uc *usecase.UseCase) *Api {
	return &Api{
		Api:         api,
		imApiCaller: imApiCaller,
		uc:          uc,
	}
}

type Api struct {
	*util.Api
	imApiCaller imapi.CallerInterface
	uc          *usecase.UseCase
}

func (a *Api) WalletUseCase() *usecase.UseCase {
	return a.uc
}

func (a *Api) EnvelopeUseCase() *usecase.UseCase {
	return a.uc
}

func (a *Api) WalletRechargeRequestUseCase() *usecase.UseCase {
	return a.uc
}

func (a *Api) WalletTransactionUseCase() *usecase.UseCase {
	return a.uc
}

func (a *Api) TransferUseCase() *usecase.UseCase {
	return a.uc
}

func (a *Api) WalletMonitoringUseCase() *usecase.UseCase {
	return a.uc
}

func (a *Api) BalanceAdjustmentUsecase() *usecase.UseCase {
	return a.uc
}
