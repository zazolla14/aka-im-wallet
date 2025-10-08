package publisher

import (
	"context"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/1nterdigital/aka-im-tools/db/mysqlutil"
	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	"github.com/1nterdigital/aka-im-wallet/internal/repository"
	"github.com/1nterdigital/aka-im-wallet/internal/usecase"
	conf "github.com/1nterdigital/aka-im-wallet/pkg/common/config"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/kdisc"
)

type PublisherInterface interface {
	Publish(ctx context.Context, key string) error
}

type Publisher struct {
	ctx    context.Context
	cancel context.CancelFunc

	mapPublisher map[domain.PublisherKey][]PublisherInterface
}

type Config struct {
	Publisher   conf.Publisher
	MysqlConfig conf.Mysql
	KafkaConfig conf.Kafka
	Share       conf.Share
	Discovery   conf.Discovery

	Key domain.PublisherKey
}

func Start(ctx context.Context, index int, config *Config) error {
	log.CInfo(ctx, "Starting PUBLISHER server")
	log.CInfo(ctx, "PUBLISHER server is initializing", "prometheusPorts",
		config.Publisher.Prometheus.Ports, "index", index)

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

	refundTransfer, err := NewExpiredTransferPublisherHandler(ctx, config, usecases.Transfer)
	if err != nil {
		return err
	}
	refundEnvelope, err := NewExpiredEnvelopePublisherHandler(ctx, config, usecases.Envelope)
	if err != nil {
		return err
	}

	publisher := &Publisher{
		ctx: ctx,
		mapPublisher: map[domain.PublisherKey][]PublisherInterface{
			domain.PublisherKeyRefundTransferEnvelope: {
				refundEnvelope,
				refundTransfer,
			},
		},
	}

	return publisher.Start(config)
}

func (m *Publisher) Start(cfg *Config) error {
	m.ctx, m.cancel = context.WithCancel(context.Background())

	if pubs, ok := m.mapPublisher[cfg.Key]; ok {
		for idx := range pubs {
			err := pubs[idx].Publish(m.ctx, string(cfg.Key))
			if err != nil {
				return err
			}
		}
	} else {
		err := errs.New("key handler is not registered yet", "key", cfg.Key).Wrap()
		log.ZError(m.ctx, "while check publisher key", err, "key", cfg.Key)
		return err
	}

	return nil
}
