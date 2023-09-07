package main

import (
	gitApp "github.com/MR5356/syncer/cmd/git/app"
	imageApp "github.com/MR5356/syncer/cmd/image/app"
	_ "github.com/MR5356/syncer/pkg/log"
	"github.com/MR5356/syncer/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug bool
)

func NewSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "syncer",
		Version: version.Version,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.AddCommand(
		imageApp.NewImageCommand(),
		gitApp.NewGitCommand(),
	)
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	return cmd
}

func main() {
	if err := NewSyncCommand().Execute(); err != nil {
		logrus.Fatal(err)
	}
}
