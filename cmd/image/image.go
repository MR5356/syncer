package main

import (
	"github.com/MR5356/syncer/cmd/image/app"
	_ "github.com/MR5356/syncer/pkg/log"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := app.NewImageCommand().Execute(); err != nil {
		logrus.Fatal(err)
	}
}
