// +build windows

package tipservice

import (
	"github.com/kardianos/service"
	"github.com/tktip/flyvo-rpc-client/internal/log"
)

var (
	logger service.Logger
	done   chan int
)

func New(args []string) service.Service {
	if len(args) < 2 {
		log.Logger.Fatal("Missing args! Please create service with 'configFile=[path]' set.")
	}

	svcConfig := &service.Config{
		Name:        "Tip Flyvo Service",
		DisplayName: "Tip Flyvo",
		Description: "Service that communicates with TIP via RPC, functioning as a 'proxy' between FLYVO and TIP",
	}

	prg := &program{
		args: args,
	}

	tipService, err := service.New(prg, svcConfig)
	if err != nil {
		log.Logger.Fatal(err)
	}

	//init loggers
	logger, err = tipService.Logger(nil)
	if err != nil {
		log.Logger.Fatal(err)
	}
	return tipService
}
