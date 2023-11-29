package app

import (
	"github.com/MR5356/syncer/pkg/domain/git/client"
	"github.com/MR5356/syncer/pkg/domain/git/config"
	"github.com/MR5356/syncer/pkg/utils/structutil"
	"github.com/MR5356/syncer/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"runtime"
)

const defaultRetries = 3

var (
	debug                                          bool
	configFile, privateKeyFile, privateKeyPassword string
	retries, procNum                               int

	defaultProcNum = runtime.NumCPU()
)

func NewGitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "A git repo sync tool",
		Long: `A git repo sync tool implement by Go.

Complete code is available at https://github.com/Mr5356/syncer`,
		Version: version.Version,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			cfg := config.NewConfig()
			if configFile == "" {
				logrus.Fatalf("config file can not be empty")
			}
			cfg = config.NewConfigFromFile(configFile)

			if cfg.Proc == 0 || procNum != defaultProcNum {
				cfg.With(config.WithProc(procNum))
			}
			if cfg.Retries == 0 || retries != defaultRetries {
				cfg.With(config.WithRetries(retries))
			}
			if cfg.PrivateKeyFile == "" {
				cfg.With(config.WithPrivateKeyFile(privateKeyFile))
			}
			if privateKeyPassword != "" {
				cfg.With(config.WithPrivateKeyPassword(privateKeyPassword))
			}
			logrus.Debugf("run with config: \n%s", structutil.Struct2String(cfg))
			cli := client.NewClient(cfg)
			if err := cli.Run(); err != nil {
				logrus.Fatalf("run git sync failed: %+v", err)
			}
		},
	}

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	cmd.PersistentFlags().StringVar(&privateKeyFile, "privateKeyFile", "", "private key file")
	cmd.PersistentFlags().StringVar(&privateKeyPassword, "privateKeyPassword", "", "private key file password")
	cmd.PersistentFlags().IntVarP(&procNum, "proc", "p", defaultProcNum, "process num")
	cmd.PersistentFlags().IntVarP(&retries, "retries", "r", defaultRetries, "retries num")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	return cmd
}
