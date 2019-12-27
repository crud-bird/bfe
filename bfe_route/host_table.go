package bfe_route

import (
	"errors"
	"github.com/crud-bird/bfe/bfe_config/bfe_route_conf/host_rule_conf"
	"strings"
)

var (
	ErrNoProduct     = errors.New("no product found")
	ErrNoProductRule = errors.New("no route rule found for product")
	ErrNoMatchRule   = errors.New("no rule match for this req")
)

type HostTable struct {
	version   Versions
	hostTable host_rule_conf.Host2HostTag
}

// todo
