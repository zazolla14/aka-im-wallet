package kafka

import (
	"bytes"
	"strings"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"

	"github.com/1nterdigital/aka-im-tools/errs"
)

//nolint:dupl // similar code for different entities
func BuildConsumerGroupConfig(conf *Config, initial int64, autoCommitEnable bool) (*sarama.Config, error) {
	kfk := sarama.NewConfig()
	kfk.Version = sarama.V2_0_0_0
	kfk.Consumer.Offsets.Initial = initial
	kfk.Consumer.Offsets.AutoCommit.Enable = autoCommitEnable
	kfk.Consumer.Return.Errors = false
	if conf.Username != "" || conf.Password != "" {
		kfk.Net.SASL.Enable = true
		kfk.Net.SASL.User = conf.Username
		kfk.Net.SASL.Password = conf.Password
	}
	if conf.TLS.EnableTLS {
		tls, err := newTLSConfig(
			conf.TLS.ClientCrt,
			conf.TLS.ClientKey,
			conf.TLS.CACrt,
			[]byte(conf.TLS.ClientKeyPwd),
			conf.TLS.InsecureSkipVerify,
		)
		if err != nil {
			return nil, err
		}
		kfk.Net.TLS.Config = tls
		kfk.Net.TLS.Enable = true
		kfk.Version = sarama.V2_8_0_0
		kfk.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		kfk.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: scram.SHA512}
		}
	}
	return kfk, nil
}

func NewConsumerGroup(conf *sarama.Config, addr []string, groupID string) (sarama.ConsumerGroup, error) {
	cg, err := sarama.NewConsumerGroup(addr, groupID, conf)
	if err != nil {
		return nil, errs.WrapMsg(err, "NewConsumerGroup failed", "addr", addr, "groupID", groupID, "conf", *conf)
	}
	return cg, nil
}

//nolint:dupl // similar code for different entities
func BuildProducerConfig(conf *Config) (config *sarama.Config, err error) {
	kfk := sarama.NewConfig()
	kfk.Producer.Return.Successes = true
	kfk.Producer.Return.Errors = true
	kfk.Producer.Partitioner = sarama.NewHashPartitioner
	if conf.Username != "" || conf.Password != "" {
		kfk.Net.SASL.Enable = true
		kfk.Net.SASL.User = conf.Username
		kfk.Net.SASL.Password = conf.Password
	}
	switch strings.ToLower(conf.ProducerAck) {
	case "no_response":
		kfk.Producer.RequiredAcks = sarama.NoResponse
	case "wait_for_local":
		kfk.Producer.RequiredAcks = sarama.WaitForLocal
	case "wait_for_all":
		kfk.Producer.RequiredAcks = sarama.WaitForAll
	default:
		kfk.Producer.RequiredAcks = sarama.WaitForAll
	}
	if conf.CompressType == "" {
		kfk.Producer.Compression = sarama.CompressionNone
	} else {
		err := kfk.Producer.Compression.UnmarshalText(bytes.ToLower([]byte(conf.CompressType)))
		if err != nil {
			return nil, errs.WrapMsg(err, "UnmarshalText failed", "compressType", conf.CompressType)
		}
	}
	if conf.TLS.EnableTLS {
		tls, err := newTLSConfig(
			conf.TLS.ClientCrt,
			conf.TLS.ClientKey,
			conf.TLS.CACrt,
			[]byte(conf.TLS.ClientKeyPwd),
			conf.TLS.InsecureSkipVerify,
		)
		if err != nil {
			return nil, err
		}
		kfk.Net.TLS.Config = tls
		kfk.Net.TLS.Enable = true
		kfk.Version = sarama.V2_8_0_0
		kfk.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		kfk.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: scram.SHA512}
		}
	}
	return kfk, nil
}

func NewProducer(conf *sarama.Config, addr []string) (sarama.SyncProducer, error) {
	producer, err := sarama.NewSyncProducer(addr, conf)
	if err != nil {
		return nil, errs.WrapMsg(err, "NewSyncProducer failed", "addr", addr, "conf", *conf)
	}
	return producer, nil
}

type XDGSCRAMClient struct {
	Client             *scram.Client
	ClientConversation *scram.ClientConversation
	HashGeneratorFcn   scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) error {
	var err error
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}
func (x *XDGSCRAMClient) Step(challenge string) (string, error) {
	return x.ClientConversation.Step(challenge)
}
func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}
