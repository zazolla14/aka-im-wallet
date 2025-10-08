//nolint:dupl // similar code for different entities
package publisher

import (
	"context"
	"time"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/mcontext"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	"github.com/1nterdigital/aka-im-wallet/internal/usecase"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/kafka"
)

type ExpiredTransferPublisherHandler struct {
	expiredTransferPublisher *kafka.Producer
	transferUsecase          usecase.TransferSvc
}

func NewExpiredTransferPublisherHandler(
	_ context.Context,
	config *Config,
	transferUC usecase.TransferSvc,
) (PublisherInterface, error) {
	kafkaConf := config.KafkaConfig
	conf, err := kafka.BuildProducerConfig(kafkaConf.Build())
	if err != nil {
		return nil, err
	}

	transferPublisher, err := kafka.NewKafkaProducer(conf, kafkaConf.Address, kafkaConf.ToExpiredTransferTopic)
	if err != nil {
		return nil, err
	}

	return &ExpiredTransferPublisherHandler{
		expiredTransferPublisher: transferPublisher,
		transferUsecase:          transferUC,
	}, nil
}

func (p *ExpiredTransferPublisherHandler) Publish(ctx context.Context, key string) error {
	expTransfers, err := p.transferUsecase.FetchExpiredTransfers(ctx, []int64{})
	if err != nil {
		return err
	}

	if len(expTransfers) == 0 {
		log.ZInfo(ctx, "there are no expired transfers is found")
		return nil
	}

	maxRetry := 3
	var lastErr error

	for _, transfer := range expTransfers {
		transfer.OperatedBy = domain.KafkaProducerOperator
		for retry := 1; retry <= maxRetry; retry++ {
			ctx = mcontext.SetOperationID(ctx, time.Now().String())
			if _, _, tempErr := p.expiredTransferPublisher.SendMessage(ctx, key, transfer); tempErr != nil {
				lastErr = tempErr
				if retry == maxRetry {
					log.ZError(ctx, "SendMessage expired transfer failed after max retries",
						tempErr,
						"transferID", transfer.TransferID,
					)
				}

				continue
			}

			break
		}
	}

	return lastErr
}
