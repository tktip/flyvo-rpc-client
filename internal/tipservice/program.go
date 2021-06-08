package tipservice

// +build windows
import (
	"errors"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tktip/cfger"
	"github.com/tktip/flyvo-rpc-client/internal/api"
	"github.com/tktip/flyvo-rpc-client/internal/log"
	"github.com/tktip/flyvo-rpc-client/pkg/eventHook"
	filehook "github.com/tktip/flyvo-rpc-client/pkg/fileHook"

	"context"

	"github.com/kardianos/service"
)

type program struct {
	args       []string   `json:"-" yaml:"-"`
	LogFile    string     `json:"logFile" yaml:"logFile"`
	LogLevel   string     `json:"logLevel" yaml:"logLevel"`
	NoEventLog bool       `json:"noEventLog" yaml:"noEventLog"`
	Api        api.Server `json:"api" yaml:"api"`
}

func (p *program) readConfig(configFile string) (err error) {
	configKV := strings.Split(configFile, "=")
	if configKV[0] != "configFile" || len(configKV) < 2 || configKV[1] == "" {
		log.Logger.Debug("No config provided")
		return errors.New("missing configFile parameter")
	}

	configFile = configKV[1]
	_, err = cfger.ReadStructuredCfg("file::"+configFile, p)
	log.Logger.Infof("Contents: %+v", *p)
	return err
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) initialize() error {
	log.Logger.Info("initializing...")
	log.Logger.Info(`Config details:`)
	log.Logger.Infof("RPC client api port: %s", p.Api.Port)
	log.Logger.Infof("Rpc server address:  %s", p.Api.RpcClient.RpcServerAddress)
	log.Logger.Infof("Rpc cert file:       %s", p.Api.RpcClient.RpcCertFile)
	log.Logger.Infof("Flyvo api address:   %s", p.Api.RpcClient.FlyvoApiEndpoints.RootAddress)
	log.Logger.Infof("Log file:            %s", p.LogFile)
	log.Logger.Infof("Log level:           %s", p.LogLevel)

	if p.LogLevel != "" {
		level, err := logrus.ParseLevel(p.LogLevel)
		if err != nil {
			log.Logger.Warnf("Could not parse level '%s', falling back to default (info). Error: %s", p.LogLevel, err.Error())
		} else {
			log.Logger.SetLevel(level)
			log.Logger.Infof("Set log level to %s", level)
		}
	}

	if !p.NoEventLog {
		log.Logger.Info("Event log not disabled, adding event logger hook")
		hook := eventHook.NewHook(logger)
		log.Logger.AddHook(hook)
	}

	if p.LogFile != "" {
		log.Logger.Info()
		fHook := filehook.NewHook(p.LogFile)
		log.Logger.AddHook(fHook)
		log.Logger.Info("Logging to file enabled")
	} else {
		log.Logger.Warn("No log file location set - only logging to event log (if not disabled)")
	}
	return nil
}

func (p *program) run() {

	err := p.readConfig(os.Args[1])
	if err != nil {
		log.Logger.Fatal("Failed to read config: " + err.Error())
	}
	log.Logger.Debug("Config read")

	err = p.initialize()
	if err != nil {
		log.Logger.Fatal("Failed to initalize: " + err.Error())
	}

	err = p.Api.Run(context.Background())
	if err != nil {
		log.Logger.Fatalf("Failed to start api: %+v", err)
	}
}
func (p *program) Stop(s service.Service) error {
	log.Logger.Warn("Shutting down...")
	// Stop should not block. Return with a few seconds.
	return nil
}
