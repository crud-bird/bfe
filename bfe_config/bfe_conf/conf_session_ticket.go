package bfe_conf

import (
	"github.com/crud-bird/bfe/bfe_util"
	"github.com/sirupsen/logrus"
)

type ConfigSessionTicket struct {
	SessionTIcketsDisable bool
	SessionTicketsKeyFile string
}

func (cfg *ConfigSessionTicket) SetDefaultConf() {
	cfg.SessionTIcketsDisable = true
	cfg.SessionTicketsKeyFile = "tls_conf/session_ticket_key.data"
}

func (cfg *ConfigSessionTicket) Check(confRoot string) error {
	return ConfSessionTicketCheck(cfg, confRoot)
}

func ConfSessionTicketCheck(cfg *ConfigSessionTicket, confRoot string) error {
	if cfg.SessionTicketsKeyFile == "" {
		logrus.Warn("SessionTicketsKeyFile not set, use default value")
		cfg.SessionTicketsKeyFile = "tls_conf/server_ticket_key.data"
	}
	cfg.SessionTicketsKeyFile = bfe_util.ConfPathProc(cfg.SessionTicketsKeyFile, confRoot)

	return nil
}
