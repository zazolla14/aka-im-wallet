package config

import "strings"

var (
	ShareFileName              = "share.yml"
	DiscoveryConfigFileName    = "discovery.yml"
	WalletAPIWalletCfgFileName = "wallet-api-wallet.yml"
	LogConfigFileName          = "log.yml"
	AdminFileName              = "admin.yml"
	MongodbConfigFileName      = "mongodb.yml"
	MysqlConfigFileName        = "mysqldb.yml"
	RedisConfigFileName        = "redis.yml"
	KafkaConfigFileName        = "kafka.yml"
	MsgTransferCfgFileName     = "msgtransfer.yml"
	PublisherCfgFileName       = "publisher.yml"
	TracerCfgFileName          = "tracer.yml"
	Publisss                   = "her.yml"
)

var EnvPrefixMap map[string]string

func init() {
	EnvPrefixMap = make(map[string]string)
	fileNames := []string{
		ShareFileName,
		AdminFileName,
		DiscoveryConfigFileName,
		MongodbConfigFileName,
		LogConfigFileName,
		RedisConfigFileName,
		WalletAPIWalletCfgFileName,
		KafkaConfigFileName,
		MsgTransferCfgFileName,
		PublisherCfgFileName,
		TracerCfgFileName,
		Publisss,
	}

	for _, fileName := range fileNames {
		envKey := strings.TrimSuffix(strings.TrimSuffix(fileName, ".yml"), ".yaml")
		envKey = "WALLETENV_" + envKey
		envKey = strings.ToUpper(strings.ReplaceAll(envKey, "-", "_"))
		EnvPrefixMap[fileName] = envKey
	}
}

const (
	FlagConf          = "config_folder_path"
	FlagTransferIndex = "index"
)
