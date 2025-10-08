package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/1nterdigital/aka-im-tools/system/program"
	"github.com/1nterdigital/aka-im-wallet/internal/publisher"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/config"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/constant"
	"github.com/1nterdigital/aka-im-wallet/version"
)

type PublisherCmd struct {
	*RootCmd
	ctx             context.Context
	configMap       map[string]any
	publisherConfig *publisher.Config
}

func NewPublisherCmd() *PublisherCmd {
	var publisherConfig publisher.Config
	ret := &PublisherCmd{publisherConfig: &publisherConfig}
	ret.configMap = map[string]any{
		config.PublisherCfgFileName:    &publisherConfig.Publisher,
		config.KafkaConfigFileName:     &publisherConfig.KafkaConfig,
		config.MysqlConfigFileName:     &publisherConfig.MysqlConfig,
		config.ShareFileName:           &publisherConfig.Share,
		config.DiscoveryConfigFileName: &publisherConfig.Discovery,
	}
	ret.RootCmd = NewRootCmd(program.GetProcessName(), WithConfigMap(ret.configMap))
	ret.ctx = context.WithValue(context.Background(), constant.ContextKeyVersion, version.Version)
	ret.Command.Flags().StringVar(
		(*string)(&publisherConfig.Key),
		"key",
		"",
		"publisher key (e.g. refundTransferEnvelope)",
	)

	ret.Command.RunE = func(_ *cobra.Command, _ []string) error {
		return ret.runE()
	}
	return ret
}

func (m *PublisherCmd) Exec() error {
	return m.Execute()
}

func (m *PublisherCmd) runE() error {
	return publisher.Start(m.ctx, m.Index(), m.publisherConfig)
}
