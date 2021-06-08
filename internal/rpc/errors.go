package rpc

import "errors"

var (
	ErrorShuttingDown = errors.New("shutting down")
	ErrorBadPath  = errors.New("unknown path provided")
)