package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/1nterdigital/aka-im-tools/errs"
	"github.com/1nterdigital/aka-im-tools/log"
	"github.com/1nterdigital/aka-im-tools/utils/runtimeenv"
	"github.com/1nterdigital/aka-im-wallet/pkg/common/config"
	"github.com/1nterdigital/aka-im-wallet/version"
)

type RootCmd struct {
	Command     cobra.Command
	processName string
	log         config.Log
	index       int
	configPath  string
}

func (r *RootCmd) Index() int {
	return r.index
}

type CmdOpts struct {
	loggerPrefixName string
	configMap        map[string]any
}

func defaultCmdOpts() *CmdOpts {
	return &CmdOpts{
		loggerPrefixName: "im-wallet-log",
	}
}

func applyOptions(opts ...func(*CmdOpts)) *CmdOpts {
	cmdOpts := defaultCmdOpts()
	for _, opt := range opts {
		opt(cmdOpts)
	}

	return cmdOpts
}

func (r *RootCmd) getFlag(cmd *cobra.Command) (flag string, idx int, err error) {
	configDirectory, err := cmd.Flags().GetString(config.FlagConf)
	if err != nil {
		return "", 0, errs.Wrap(err)
	}
	index, err := cmd.Flags().GetInt(config.FlagTransferIndex)
	if err != nil {
		return "", 0, errs.Wrap(err)
	}
	r.index = index
	return configDirectory, index, nil
}

func (r *RootCmd) initializeConfiguration(cmd *cobra.Command, opts *CmdOpts) (err error) {
	configDirectory, _, err := r.getFlag(cmd)
	if err != nil {
		return err
	}
	r.configPath = configDirectory

	runtimeEnv := runtimeenv.PrintRuntimeEnvironment()

	for configFileName, configStruct := range opts.configMap {
		err := config.Load(configDirectory, configFileName,
			config.EnvPrefixMap[configFileName], runtimeEnv, configStruct)
		if err != nil {
			return err
		}
	}

	// Load common log configuration file
	return config.Load(configDirectory, config.LogConfigFileName,
		config.EnvPrefixMap[config.LogConfigFileName], runtimeEnv, &r.log)
}

func (r *RootCmd) initializeLogger(cmdOpts *CmdOpts) error {
	err := log.InitLoggerFromConfig(
		cmdOpts.loggerPrefixName,
		r.processName,
		"", "",
		r.log.RemainLogLevel,
		r.log.IsStdout,
		r.log.IsJson,
		r.log.StorageLocation,
		r.log.RemainRotationCount,
		r.log.RotationTime,
		version.Version,
		r.log.IsSimplify,
	)
	if err != nil {
		return errs.Wrap(err)
	}

	return errs.Wrap(log.InitConsoleLogger(r.processName, r.log.RemainLogLevel, r.log.IsJson, config.Version))
}

func (r *RootCmd) persistentPreRun(cmd *cobra.Command, opts ...func(*CmdOpts)) error {
	cmdOpts := applyOptions(opts...)
	if err := r.initializeConfiguration(cmd, cmdOpts); err != nil {
		return err
	}

	if err := r.initializeLogger(cmdOpts); err != nil {
		return errs.WrapMsg(err, "failed to initialize logger")
	}

	return nil
}

func WithConfigMap(configMap map[string]any) func(*CmdOpts) {
	return func(opts *CmdOpts) {
		opts.configMap = configMap
	}
}

func NewRootCmd(processName string, opts ...func(*CmdOpts)) *RootCmd {
	rootCmd := &RootCmd{processName: processName}
	cmd := cobra.Command{
		Use:  "Start IM wallet application",
		Long: fmt.Sprintf(`Start %s `, processName),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return rootCmd.persistentPreRun(cmd, opts...)
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	cmd.Flags().StringP(config.FlagConf, "c", "", "path of config directory")
	cmd.Flags().IntP(config.FlagTransferIndex, "i", 0, "process startup sequence number")

	rootCmd.Command = cmd
	return rootCmd
}

func (r *RootCmd) Execute() error {
	return r.Command.Execute()
}
