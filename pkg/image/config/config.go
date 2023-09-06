package config

import (
	"github.com/MR5356/syncer/pkg/utils/configutil"
	"github.com/mcuadros/go-defaults"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Auth    map[string]*Auth `json:"auth" yaml:"auth"`
	Images  map[string]any   `json:"images" yaml:"images"`
	Proc    int              `json:"proc" yaml:"proc"`
	Retries int              `json:"retries" yaml:"retries"`
}

type Auth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Insecure bool   `json:"insecure" yaml:"insecure" default:"false"`
}

func NewConfig(cfg ...Cfg) *Config {
	config := &Config{
		Auth:   make(map[string]*Auth),
		Images: make(map[string]any),
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
		logrus.Fatalf("error parse config file: %s", err)
	}
	return config
}

func (c *Config) With(cs ...Cfg) {
	for _, cfg := range cs {
		cfg(c)
	}
}

func (c *Config) GetAuth(repo string) *Auth {
	if auth, ok := c.Auth[repo]; ok {
		return auth
	} else {
		return new(Auth)
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
