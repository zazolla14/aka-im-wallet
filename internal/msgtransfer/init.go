package msgtransfer

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/db/mysqlutil"
	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/system/program"
	"github.com/1nterdigital/aka-im-tools/utils/datautil"
	"github.com/1nterdigital/aka-im-wallet/internal/repository"
	"github.com/1nterdigital/aka-im-wallet/internal/usecase"
	conf "github.com/1nterdigital/aka-im-wallet/pkg/common/config"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/kafka"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/kdisc"
)

type MsgTransfer struct {
	ctx    context.Context
	cancel context.CancelFunc

	expiredTransferCH *ExpiredTransferConsumerHandler
	expiredEnvelopeCH *ExpiredEnvelopeConsumerHandler
}

type Config struct {
	MsgTransfer conf.MsgTransfer
	RedisConfig conf.Redis
	KafkaConfig conf.Kafka
	MysqlConfig conf.Mysql
	Share       conf.Share
	Discovery   conf.Discovery
}

type mapKafkaProducer struct {
	expiredTransfer *kafka.Producer
	expiredEnvelope *kafka.Producer
}

func Start(ctx context.Context, index int, config *Config) error {
	log.CInfo(ctx, "Starting MSG-TRANSFER server instance1")
	log.CInfo(ctx, "MSG-TRANSFER server instance is initializing", "prometheusPorts",
		config.MsgTransfer.Prometheus.Ports, "index", index)

	db, err := mysqlutil.NewMysqlDB(ctx, config.MysqlConfig.Build())
	if err != nil {
		return err
	}

	dbGorm, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db.DB, // wrap existing *sql.DB
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	_, err = kdisc.NewDiscoveryRegister(&config.Discovery, "", nil)
	if err != nil {
		return err
	}

	// repository
	repos := repository.NewRepository(dbGorm)

	// usecase
	usecases, err := usecase.New(&usecase.Config{
		KafkaConfig: config.KafkaConfig,
	}, repos, dbGorm)
	if err != nil {
		return err
	}

	producers, err := initKafkaProducers(config)
	if err != nil {
		return err
	}

	expiredTransferCH, err := NewExpiredTransferConsumerHandler(
		ctx, config, producers.expiredTransfer,
		usecases.Transfer)
	if err != nil {
		return err
	}

	expiredEnvelopeCH, err := NewExpiredEnvelopeConsumerHandler(
		ctx, config, producers.expiredEnvelope,
		usecases.Envelope)
	if err != nil {
		return err
	}

	msgTransfer := &MsgTransfer{
		expiredTransferCH: expiredTransferCH,
		expiredEnvelopeCH: expiredEnvelopeCH,
	}
	return msgTransfer.Start(index, config)
}

func (m *MsgTransfer) Start(index int, cfg *Config) error {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	var (
		netDone = make(chan struct{}, 1)
		netErr  error
	)

	// consumer
	go m.expiredTransferCH.consumerGroup.RegisterHandleAndConsumer(m.ctx, m.expiredTransferCH)
	go m.expiredEnvelopeCH.consumerGroup.RegisterHandleAndConsumer(m.ctx, m.expiredEnvelopeCH)

	err := m.expiredTransferCH.redisMessageBatches.Start()
	if err != nil {
		return err
	}

	err = m.expiredEnvelopeCH.redisMessageBatches.Start()
	if err != nil {
		return err
	}

	_, err = kdisc.NewDiscoveryRegister(&cfg.Discovery, "", nil)
	if err != nil {
		return errs.WrapMsg(err, "failed to register discovery service")
	}

	if cfg.MsgTransfer.Prometheus.Enable {
		var (
			listener       net.Listener
			prometheusPort int
			err            error
		)

		prometheusPort, err = datautil.GetElemByIndex(cfg.MsgTransfer.Prometheus.Ports, index)
		if err != nil {
			return err
		}

		lc := &net.ListenConfig{}
		listener, err = lc.Listen(context.Background(), "tcp", fmt.Sprintf(":%d", prometheusPort))
		if err != nil {
			return errs.WrapMsg(err, "listen err", "addr", fmt.Sprintf(":%d", prometheusPort))
		}
		log.ZInfo(m.ctx, "listener", "listenerList", listener)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.ZPanic(m.ctx, "MsgTransfer Start Panic", errs.ErrPanic(r))
				}
			}()
		}()
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-sigs:
		program.SIGTERMExit()
		// graceful close kafka client.
		m.cancel()
		m.expiredTransferCH.redisMessageBatches.Close()
		m.expiredTransferCH.consumerGroup.Close()
		m.expiredEnvelopeCH.redisMessageBatches.Close()
		m.expiredEnvelopeCH.consumerGroup.Close()
		return nil
	case <-netDone:
		m.cancel()
		m.expiredTransferCH.redisMessageBatches.Close()
		m.expiredTransferCH.consumerGroup.Close()
		m.expiredEnvelopeCH.redisMessageBatches.Close()
		m.expiredEnvelopeCH.consumerGroup.Close()
		close(netDone)
		return netErr
	}
}

func initKafkaProducers(cfg *Config) (resp *mapKafkaProducer, err error) {
	kafkaConf := cfg.KafkaConfig
	configuration, err := kafka.BuildProducerConfig(kafkaConf.Build())
	if err != nil {
		return nil, err
	}

	var producers mapKafkaProducer
	producers.expiredTransfer, err = kafka.NewKafkaProducer(configuration, kafkaConf.Address, kafkaConf.ToExpiredTransferTopic)
	if err != nil {
		return nil, err
	}

	producers.expiredEnvelope, err = kafka.NewKafkaProducer(configuration, kafkaConf.Address, kafkaConf.ToExpiredEnvelopeTopic)
	if err != nil {
		return nil, err
	}

	return &producers, nil
}
