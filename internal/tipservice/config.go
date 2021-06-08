package tipservice

import (
	"github.com/tktip/flyvo-rpc-client/internal/api"
)

type Config struct {
	LogFile  string     `json:"logFile" yaml:"logFile"`
	LogLevel string     `json:"logLevel" yaml:"logLevel"`
	Api      api.Server `json:"api" yaml:"api"`
}
