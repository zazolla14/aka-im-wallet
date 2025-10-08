package msgtransfer

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/IBM/sarama"

	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/mcontext"
	"github.com/1nterdigital/aka-im-tools/utils/stringutil"
	"github.com/1nterdigital/aka-im-wallet/internal/domain"
	"github.com/1nterdigital/aka-im-wallet/internal/usecase"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/db/kafka"
	"github.com/1nterdigital/aka-im-wallet/pkg/tools/batcher"
)

type CtxMsgEnvelope struct {
	msg *domain.MsgKafkaExpiredEnvelope
	ctx context.Context
}

type ExpiredEnvelopeConsumerHandler struct {
	consumerGroup       *kafka.MConsumerGroup
	producer            *kafka.Producer
	redisMessageBatches *batcher.Batcher[sarama.ConsumerMessage]
	envelopeUsecase     usecase.EnvelopeSvc
}

func NewExpiredEnvelopeConsumerHandler(
	_ context.Context,
	config *Config,
	producer *kafka.Producer,
	envelopeUsecase usecase.EnvelopeSvc,
) (*ExpiredEnvelopeConsumerHandler, error) {
	kafkaConf := config.KafkaConfig
	consumerGroup, err := kafka.NewMConsumerGroup(
		kafkaConf.Build(),
		kafkaConf.ToExpiredEnvelopeGroupID,
		[]string{kafkaConf.ToExpiredEnvelopeTopic},
		false,
	)
	if err != nil {
		return nil, err
	}

	och := ExpiredEnvelopeConsumerHandler{
		producer:        producer,
		envelopeUsecase: envelopeUsecase,
	}

	b := batcher.New[sarama.ConsumerMessage](
		batcher.WithSize(size),
		batcher.WithWorker(worker),
		batcher.WithInterval(interval),
		batcher.WithDataBuffer(mainDataBuffer),
		batcher.WithSyncWait(true),
		batcher.WithBuffer(subChanBuffer),
	)

	b.Sharding = func(key string) int {
		hashCode := stringutil.GetHashCode(key)
		return int(hashCode) % och.redisMessageBatches.Worker()
	}
	b.Key = func(consumerMessage *sarama.ConsumerMessage) string {
		return string(consumerMessage.Key)
	}
	b.Do = och.do
	och.redisMessageBatches = b
	och.consumerGroup = consumerGroup

	return &och, nil
}

func (och *ExpiredEnvelopeConsumerHandler) do(ctx context.Context, channelID int, val *batcher.Msg[sarama.ConsumerMessage]) {
	ctx = mcontext.WithTriggerIDContext(ctx, val.TriggerID())
	ctxMessages := parseConsumerEnvelopeMessages(ctx, val.Val())
	ctx = withAggregationCtxEnvelope(ctx, ctxMessages)
	log.ZInfo(ctx, "msg arrived channel", "channel id", channelID, "msgList length", len(ctxMessages), "key", val.Key())

	och.handleMsg(ctx, val.Key(), ctxMessages)
}

func parseConsumerEnvelopeMessages(ctx context.Context, consumerMessages []*sarama.ConsumerMessage) []*CtxMsgEnvelope {
	var ctxMessages []*CtxMsgEnvelope
	for i := range consumerMessages {
		ctxMsg := &CtxMsgEnvelope{}
		msgFromMQ := &domain.MsgKafkaExpiredEnvelope{}
		err := json.Unmarshal(consumerMessages[i].Value, msgFromMQ)
		if err != nil {
			log.ZWarn(ctx, "msg_transfer Unmarshal msg err", err, string(consumerMessages[i].Value))
			continue
		}

		var arr []string
		for i, header := range consumerMessages[i].Headers {
			arr = append(arr, strconv.Itoa(i), string(header.Key), string(header.Value))
		}
		log.ZDebug(ctx, "consumer.kafka.GetContextWithMQHeader", "len", len(consumerMessages[i].Headers),
			"header", strings.Join(arr, ", "))
		ctxMsg.ctx = kafka.GetContextWithMQHeader(consumerMessages[i].Headers)
		ctxMsg.msg = msgFromMQ
		log.ZDebug(ctx, "message parse finish", "message", msgFromMQ, "key",
			string(consumerMessages[i].Key))
		ctxMessages = append(ctxMessages, ctxMsg)
	}

	return ctxMessages
}

func (och *ExpiredEnvelopeConsumerHandler) handleMsg(ctx context.Context, key string, kafkaMsg []*CtxMsgEnvelope) {
	var err error
	defer func() {
		if err != nil {
			log.ZWarn(ctx, "an error occurs while handle kafka msg", err, "key", key)
		}
	}()
	log.ZInfo(ctx, "handle expired envelope")

	var expiredEnvelope []*domain.MsgKafkaExpiredEnvelope
	for idx := range kafkaMsg {
		expiredEnvelope = append(expiredEnvelope, kafkaMsg[idx].msg)
	}

	failedList, err := och.envelopeUsecase.ProcessExpiredEnvelopes(ctx, expiredEnvelope)
	if err != nil {
		log.ZError(ctx, "while ProcessExpiredEnvelopes", err, "key", key)
	}

	newCtx := context.WithoutCancel(ctx)
	for idx := range failedList {
		if _, _, errx := och.producer.SendMessage(newCtx, key, failedList[idx]); errx != nil {
			log.ZError(ctx, "while republish failed refund expired envelopes",
				errx, "key", key, "envelopeID", failedList[idx].EnvelopeID)
		}
	}
}

//nolint:revive // keep receiver for interface compliance, may be used in the future
func (och *ExpiredEnvelopeConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

//nolint:revive // keep receiver for interface compliance, may be used in the future
func (och *ExpiredEnvelopeConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

//nolint:dupl // similar code for different entities
func (och *ExpiredEnvelopeConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error { // a instance in the consumer group
	log.ZDebug(context.Background(), "new session expired envelope msg come", "highWaterMarkOffset",
		claim.HighWaterMarkOffset(), "topic", claim.Topic(), "partition", claim.Partition())
	och.redisMessageBatches.OnComplete = func(lastMessage *sarama.ConsumerMessage, _ int) {
		session.MarkMessage(lastMessage, "")
		session.Commit()
	}
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			if len(msg.Value) == 0 {
				continue
			}
			err := och.redisMessageBatches.Put(context.Background(), msg)
			if err != nil {
				log.ZWarn(context.Background(), "put msg to  error", err, "msg", msg)
			}
		case <-session.Context().Done():
			return nil
		}
	}
}

func withAggregationCtxEnvelope(ctx context.Context, values []*CtxMsgEnvelope) context.Context {
	var allMessageOperationID string
	for i, v := range values {
		if opid := mcontext.GetOperationID(v.ctx); opid != "" {
			if i == 0 {
				allMessageOperationID += opid
			} else {
				allMessageOperationID += "$" + opid
			}
		}
	}
	return mcontext.SetOperationID(ctx, allMessageOperationID)
}
