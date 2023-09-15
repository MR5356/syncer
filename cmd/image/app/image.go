package app

import (
	"github.com/MR5356/syncer/pkg/domain/image/client"
	"github.com/MR5356/syncer/pkg/domain/image/config"
	"github.com/MR5356/syncer/pkg/utils/structutil"
	"github.com/MR5356/syncer/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"runtime"
)

const (
	defaultRetries = 3
)

var (
	configFile       string
	procNum, retries int

	debug bool

	defaultProcNum = runtime.NumCPU()
)

func NewImageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "A registry image sync tool",
		Long: `A registry image sync tool implement by Go.

Complete code is available at https://github.com/Mr5356/syncer`,
		Version: version.Version,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			cfg := config.NewConfig()
			if configFile != "" {
				cfg = config.NewConfigFromFile(configFile)
			} else {
				logrus.Fatalf("config file can not be empty")
			}
			if cfg.Proc == 0 || procNum != defaultProcNum {
				cfg.With(config.WithProc(procNum))
			}
			if cfg.Retries == 0 || retries != defaultRetries {
				cfg.With(config.WithRetries(retries))
			}
			logrus.Debugf("run with config: \n%s", structutil.Struct2String(cfg))
			cli := client.NewClient(cfg)
			if err := cli.Run(); err != nil {
				logrus.Fatalf("run image sync failed: %s", err)
			}
		},
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	cmd.PersistentFlags().IntVarP(&procNum, "proc", "p", defaultProcNum, "process num")
	cmd.PersistentFlags().IntVarP(&retries, "retries", "r", defaultRetries, "retries num")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	return cmd
}
