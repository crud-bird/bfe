package bfe_server

import (
	"github.com/crud-bird/bfe/bfe_conf"
	"github.com/crud-bird/bfe/bfe_util"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type BfeListener struct {
	Listener net.Listener

	BalanceType string

	ProxyHeaderTimeout time.Duration

	proxyHeaderLimit int64
}

func NewBfeListener(listener net.Listener, config bfe_conf.BfeConfig) *BfeListener {
	return &BfeListener{
		listener,
		config.Server.Layer4LoadBalancer,
		time.Duration(config.Server.ClientReadTimeout * time.Second),
		int64(config.Server.MaxProxyHeaderBytes),
	}
}

func (l *BfeListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		logrus.Debug("BfeListener: accept error: %s", err)
		return nil, err
	}

	switch l.BalanceType {
	case bfe_conf.BALANCER_BGW:
		conn = bfe_util.NewBgwConn(conn.(*net.TCPConn))
		logrus.Debug("BfeListener: accept connection via BGW")

	case bfe_conf.BALANCER_PROXY:
		conn = bfe_util.NewConn(conn, l.ProxyHeaderTimeout, l.proxyHeaderLimit)
		logrus.Debug("NewBfeListener: accept connection via PROXY")
	}

	return conn, nil
}
