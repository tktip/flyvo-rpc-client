// +build linux

package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tktip/cfger"
	"github.com/tktip/flyvo-rpc-client/internal/api"
	"github.com/tktip/flyvo-rpc-client/internal/log"
)

type config struct {
	Api      api.Server `yaml:"api"`
	LogFile  string     `yaml:"logFile"`
	LogLevel string     `yaml:"logLevel"`
}

func main() {
	conf := config{}
	_, err := cfger.ReadStructuredCfg(os.Getenv("CONFIG"), &conf)
	if err != nil {
		log.Logger.Fatal(err.Error())
	}

	logLevel, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		log.Logger.Warnf("Bad log level '%s', defaulting to info.", conf.LogLevel)
	}
	log.Logger.SetLevel(logLevel)

	srv := conf.Api
	srv.Run(context.Background())
}
