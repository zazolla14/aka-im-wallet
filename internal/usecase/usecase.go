package usecase

import (
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-wallet/internal/repository"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/config"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/kafka"
)

type Config struct {
	KafkaConfig config.Kafka
}

type mapKafkaProducer struct {
	expiredTransfer *kafka.Producer
	expiredEnvelope *kafka.Producer
}

type UseCase struct {
	Wallet                WalletSvc
	Envelope              EnvelopeSvc
	WalletRechargeRequest WalletRechargeRequestSvc
	WalletTransaction     WalletTransactionSvc
	Transfer              TransferSvc
	WalletMonitoring      WalletMonitoringSvc
	BalanceAdjustment     BalanceAdjustmentSvc
}

func New(cfg *Config, repo repository.Repository, trx *gorm.DB) (*UseCase, error) {
	producers, err := initKafkaProducers(cfg)
	if err != nil {
		return nil, err
	}

	walletUsecase := NewWalletUseCase(
		repo.Wallet(),
		trx,
		repo.WalletTransaction(),
	)

	walletTransactionUsecase := NewWalletTransactionUseCase(
		repo.WalletTransaction(),
		repo.Wallet(),
		repo.TxRepo(),
	)

	walletRechargeRequestUsecase := NewWalletRechargeRequestUseCase(
		trx,
		walletTransactionUsecase,
		repo.WalletRechargeRequest(),
		repo.WalletTransaction(),
		repo.Wallet(),
		repo.TxRepo(),
	)

	envelopeUsecase := NewEnvelopeUseCase(
		producers.expiredEnvelope,
		walletUsecase,
		walletTransactionUsecase,
		repo.TxRepo(),
		repo.Envelope(),
		repo.Wallet(),
	)

	transferUsecase := NewTransferUseCase(
		producers.expiredTransfer,
		walletTransactionUsecase,
		repo.Transfer(),
		repo.Wallet(),
		repo.TxRepo(),
	)

	walletMonitoringUsecase := NewWalletMonitoringUseCase(
		repo.WalletMonitoring(),
	)

	adjustmentUsecase := NewBalanceAdjustmentUseCase(
		walletTransactionUsecase,
		repo.BalanceAdjustment(),
		repo.Wallet(),
		repo.TxRepo(),
	)

	return &UseCase{
		Wallet:                walletUsecase,
		Envelope:              envelopeUsecase,
		WalletRechargeRequest: walletRechargeRequestUsecase,
		WalletTransaction:     walletTransactionUsecase,
		Transfer:              transferUsecase,
		WalletMonitoring:      walletMonitoringUsecase,
		BalanceAdjustment:     adjustmentUsecase,
	}, nil
}

func initKafkaProducers(cfg *Config) (*mapKafkaProducer, error) {
	kafkaConf := cfg.KafkaConfig
	conf, err := kafka.BuildProducerConfig(kafkaConf.Build())
	if err != nil {
		return nil, err
	}

	var producers mapKafkaProducer
	producers.expiredTransfer, err = kafka.NewKafkaProducer(conf, kafkaConf.Address, kafkaConf.ToExpiredTransferTopic)
	if err != nil {
		return nil, err
	}

	producers.expiredEnvelope, err = kafka.NewKafkaProducer(conf, kafkaConf.Address, kafkaConf.ToExpiredEnvelopeTopic)
	if err != nil {
		return nil, err
	}

	return &producers, nil
}
