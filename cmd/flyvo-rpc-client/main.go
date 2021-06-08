// +build windows

package main

import (
	"os"

	"github.com/tktip/flyvo-rpc-client/internal/log"
	"github.com/tktip/flyvo-rpc-client/internal/tipservice"
)

func main() {
	s := tipservice.New(os.Args)
	err := s.Run()
	if err != nil {
		log.Logger.Error(err)
	}
}
