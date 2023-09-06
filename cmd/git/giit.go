package main

import (
	"github.com/MR5356/syncer/cmd/git/app"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := app.NewGitCommand().Execute(); err != nil {
		logrus.Fatal(err)
	}
}
