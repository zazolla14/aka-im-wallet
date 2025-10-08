package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"

	"github.com/1nterdigital/aka-im-tools/errs"
)

// Producer represents a Kafka producer.
type Producer struct {
	addr     []string
	topic    string
	config   *sarama.Config
	producer sarama.SyncProducer
}

func NewKafkaProducer(config *sarama.Config, addr []string, topic string) (*Producer, error) {
	producer, err := NewProducer(config, addr)
	if err != nil {
		return nil, err
	}
	return &Producer{
		addr:     addr,
		topic:    topic,
		config:   config,
		producer: producer,
	}, nil
}

// SendMessage sends a message to the Kafka topic configured in the Producer.
func (p *Producer) SendMessage(
	ctx context.Context, key string, msg interface{},
) (partition int32, offset int64, err error) {
	// Marshal the json message
	bMsg, err := json.Marshal(msg)
	if err != nil {
		return 0, 0, errs.WrapMsg(err, "kafka json Marshal err")
	}
	if len(bMsg) == 0 {
		return 0, 0, errs.WrapMsg(errEmptyMsg, "kafka json Marshal err")
	}

	// Prepare Kafka message
	kMsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(bMsg),
	}

	// Validate message key and value
	if kMsg.Key.Length() == 0 || kMsg.Value.Length() == 0 {
		return 0, 0, errs.Wrap(errEmptyMsg)
	}

	// Attach context metadata as headers
	header, err := GetMQHeaderWithContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	kMsg.Headers = header

	// Send the message
	partition, offset, err = p.producer.SendMessage(kMsg)
	if err != nil {
		return 0, 0, errs.WrapMsg(err, "p.producer.SendMessage error")
	}

	return partition, offset, nil
}
