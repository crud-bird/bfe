package bfe_util

import (
	"github.com/crud-bird/bfe/bfe_tls"
	"net"
)

type CloseWriter interface {
	CloseWriter() error
}

type AddrFetcher interface {
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	VirtualAddr() net.Addr
	BalanceAddr() net.Addr
}

type ConnFetcher interface {
	GetNetConn() net.Conn
}

func GetTCPConn(conn net.Conn) (*net.TCPConn, error) {
	switch conn.(type) {
	case *bfe_tls.Conn:
		c := (*bfe_tls.Conn).GetNetConn()
		return c.(*net.TCPConn), nil
	}
}
