package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Config struct {
	Port             int           `envconfig:"PORT" default:"5000"`
	GithubURL        string        `envconfig:"GITHUB_API_URL" default:"https://api.github.com"`
	ClientTimeout    time.Duration `envconfig:"CLIENT_TIMEOUT" default:"3s"`
	WorkerCount      int           `envconfig:"WORKER_COUNT" default:"50"`
	GithubAppID      string        `envconfig:"GITHUB_APP_ID"`
	GithubPrivateKey string        `envconfig:"GITHUB_PRIVATE_KEY"`
}

func newConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to build config from env")
	}
	return &cfg, nil
}
