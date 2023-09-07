package config

import (
	"github.com/MR5356/syncer/pkg/utils/configutil"
	"github.com/mcuadros/go-defaults"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Proc               int    `json:"proc" yaml:"proc"`
	Retries            int    `json:"retries" yaml:"retries"`
	PrivateKeyFile     string `json:"privateKeyFile" yaml:"privateKeyFile"`
	PrivateKeyPassword string `json:"privateKeyPassword" yaml:"privateKeyPassword"`

	Repos map[string]any `json:"repos" yaml:"repos"`
}

func NewConfig(cfg ...Cfg) *Config {
	config := &Config{
		Repos: make(map[string]any),
	}

	defaults.SetDefaults(config)

	for _, c := range cfg {
		c(config)
	}

	return config
}

func NewConfigFromFile(cf string) *Config {
	config := NewConfig()
	err := configutil.NewConfigFromFile(cf, config)
	if err != nil {
		logrus.Fatalf("error parse config file: %+v", err)
	}
	return config
}

func (c *Config) With(cs ...Cfg) {
	for _, cfg := range cs {
		cfg(c)
	}
}

type Cfg func(config *Config)

func WithProc(proc int) Cfg {
	return func(config *Config) {
		config.Proc = proc
	}
}

func WithRetries(retries int) Cfg {
	return func(config *Config) {
		config.Retries = retries
	}
}

func WithPrivateKeyFile(privateKeyFile string) Cfg {
	return func(config *Config) {
		config.PrivateKeyFile = privateKeyFile
	}
}

func WithPrivateKeyPassword(privateKeyPassword string) Cfg {
	return func(config *Config) {
		config.PrivateKeyPassword = privateKeyPassword
	}
}
