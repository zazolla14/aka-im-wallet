package http

import (
	"github.com/1nterdigital/aka-im-wallet/internal/service"
	"github.com/1nterdigital/aka-im-wallet/internal/usecase"
)

type WalletHandler struct {
	walletUsecase            usecase.WalletSvc
	depositUsecase           usecase.WalletRechargeRequestSvc
	walletTransactionUsecase usecase.WalletTransactionSvc
	envelopeUsecase          usecase.EnvelopeSvc
	transferUsecase          usecase.TransferSvc
	walletMonitoringUsecase  usecase.WalletMonitoringSvc
	balanceAdjustmentUsecase usecase.BalanceAdjustmentSvc
}

func NewWalletHandler(u *service.Api) *WalletHandler {
	return &WalletHandler{
		walletUsecase:            u.WalletUseCase().Wallet,
		depositUsecase:           u.WalletRechargeRequestUseCase().WalletRechargeRequest,
		walletTransactionUsecase: u.WalletTransactionUseCase().WalletTransaction,
		envelopeUsecase:          u.EnvelopeUseCase().Envelope,
		transferUsecase:          u.TransferUseCase().Transfer,
		walletMonitoringUsecase:  u.WalletMonitoringUseCase().WalletMonitoring,
		balanceAdjustmentUsecase: u.BalanceAdjustmentUsecase().BalanceAdjustment,
	}
}
