package bfe_debug

import (
	"github.com/crud-bird/bfe/bfe_config/bfe_conf"
)

var (
	DebugIsOpen      = false
	DebugServHTTP    = false
	DebugBfeRoute    = false
	DebugBal         = false
	DebugHealthCheck = false
)

func SetDebugFlag(flag bfe_conf.ConfigBasic) {
	DebugServHTTP = flag.DebugServHttp
	DebugBfeRoute = flag.DebugBfeRoute
	DebugBal = flag.DebugBal
	DebugHealthCheck = flag.DebugHealthCheck
}
