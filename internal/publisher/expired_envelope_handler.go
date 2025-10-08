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

type ExpiredEnvelopePublisherHandler struct {
	expiredEnvelopePublisher *kafka.Producer
	envelopeUsecase          usecase.EnvelopeSvc
}

func NewExpiredEnvelopePublisherHandler(
	_ context.Context,
	config *Config,
	envelopeUC usecase.EnvelopeSvc,
) (PublisherInterface, error) {
	kafkaConf := config.KafkaConfig
	conf, err := kafka.BuildProducerConfig(kafkaConf.Build())
	if err != nil {
		return nil, err
	}

	envelopePublisher, err := kafka.NewKafkaProducer(conf, kafkaConf.Address, kafkaConf.ToExpiredEnvelopeTopic)
	if err != nil {
		return nil, err
	}

	return &ExpiredEnvelopePublisherHandler{
		expiredEnvelopePublisher: envelopePublisher,
		envelopeUsecase:          envelopeUC,
	}, nil
}

func (p *ExpiredEnvelopePublisherHandler) Publish(ctx context.Context, key string) error {
	expEnvelopes, err := p.envelopeUsecase.FetchExpiredEnvelopes(ctx, []int64{})
	if err != nil {
		return err
	}

	if len(expEnvelopes) == 0 {
		log.ZInfo(ctx, "there are no expired envelopes is found")
		return nil
	}

	maxRetry := 3
	var lastErr error

	for _, envelope := range expEnvelopes {
		envelope.OperatedBy = domain.KafkaProducerOperator
		for retry := 1; retry <= maxRetry; retry++ {
			ctx = mcontext.SetOperationID(ctx, time.Now().String())
			if _, _, tempErr := p.expiredEnvelopePublisher.SendMessage(ctx, key, envelope); tempErr != nil {
				lastErr = tempErr
				if retry == maxRetry {
					log.ZError(ctx, "SendMessage expired envelope failed after max retries",
						tempErr,
						"envelopeID", envelope.EnvelopeID,
					)
				}

				continue
			}

			break
		}
	}

	return lastErr
}
