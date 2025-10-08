package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/1nterdigital/aka-im-tools/system/program"
	"github.com/1nterdigital/aka-im-wallet/internal/msgtransfer"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/config"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/constant"
	"github.com/1nterdigital/aka-im-wallet/version"
)

type MsgTransferCmd struct {
	*RootCmd
	ctx               context.Context
	configMap         map[string]any
	msgTransferConfig *msgtransfer.Config
}

func NewMsgTransferCmd() *MsgTransferCmd {
	var msgTransferConfig msgtransfer.Config
	ret := &MsgTransferCmd{msgTransferConfig: &msgTransferConfig}
	ret.configMap = map[string]any{
		config.MsgTransferCfgFileName:  &msgTransferConfig.MsgTransfer,
		config.RedisConfigFileName:     &msgTransferConfig.RedisConfig,
		config.KafkaConfigFileName:     &msgTransferConfig.KafkaConfig,
		config.MysqlConfigFileName:     &msgTransferConfig.MysqlConfig,
		config.ShareFileName:           &msgTransferConfig.Share,
		config.DiscoveryConfigFileName: &msgTransferConfig.Discovery,
	}
	ret.RootCmd = NewRootCmd(program.GetProcessName(), WithConfigMap(ret.configMap))
	ret.ctx = context.WithValue(context.Background(), constant.ContextKeyVersion, version.Version)
	ret.Command.RunE = func(_ *cobra.Command, _ []string) error {
		return ret.runE()
	}
	return ret
}

func (m *MsgTransferCmd) Exec() error {
	return m.Execute()
}

func (m *MsgTransferCmd) runE() error {
	return msgtransfer.Start(m.ctx, m.Index(), m.msgTransferConfig)
}
