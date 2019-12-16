package bfe_util

import (
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

const (
	TCP_OPT_CIP_ANY = 230
	TCP_OPT_VIP_ANY = 229
)

var (
	ErrAddressFormat = errors.New("address format error")
)

func GetVipPort(conn net.Conn) (net.IP, int, error) {
	return net.IP{}, 0, nil
}

func GetVip(conn net.Conn) net.IP {
	return net.IP{}
}

func getVipPortViaBGW(conn net.Conn) (net.IP, int, error) {
	return net.IP{}, 0, nil
}
func getCipPortViaBGW(conn net.Conn) (net.IP, int, error) {
	return net.IP{}, 0, nil
}

func getCipPortBGW(conn net.Conn) (net.IP, int, error) {
	return net.IP{}, 0, nil
}

func parseSocketAddr(rawAddr []byte) (net.IP, int, error) {
	return net.IP{}, 0, nil
}

// var _ AddressFetcher = new(bgwConn)

type BgwConn struct {
	conn *net.TCPConn

	srcAddr *net.TCPAddr

	dstAddr *net.TCPAddr
	once    sync.Once
}

func NewBgwConn(conn *net.TCPConn) *BgwConn {
	return &BgwConn{
		conn: conn,
	}
}

func (c *BgwConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *BgwConn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *BgwConn) Close() error {
	return c.conn.Close()
}

func (c *BgwConn) CloseWrite() error {
	return c.conn.CloseWrite()
}

func (c *BgwConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *BgwConn) RemoteAddr() net.Addr {
	c.checkTtmInfoOnce()
	if c.srcAddr != nil {
		return c.srcAddr
	}

	return c.conn.RemoteAddr()
}

func (c *BgwConn) VirtualAddr() net.Addr {
	c.checkTtmInfoOnce()
	if c.dstAddr != nil {
		return c.dstAddr
	}

	return nil
}

func (c *BgwConn) BalancerAddr() net.Addr {
	return nil
}

func (c *BgwConn) GetNetConn() net.Conn {
	return c.conn
}

func (c *BgwConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *BgwConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *BgwConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *BgwConn) checkTtmInfoOnce() {
	c.once.Do(func() {
		c.checkTtmInfo()
	})
}

func (c *BgwConn) checkTtmInfo() {
	c.initSrcAddr()
	c.initDstAddr()
}

func (c *BgwConn) initSrcAddr() {
	cip, cport, err := getCipPortViaBGW(c)
	if err != nil {
		logrus.Debug("BgwConn getCipPortViaBGW failed, error: %s", err)
		return
	}

	c.srcAddr = &net.TCPAddr{
		IP:   cip,
		Port: cport,
	}
}

func (c *BgwConn) initDstAddr() {
	vip, vport, err := getVipPortViaBGW(c)
	if err != nil {
		logrus.Debug("BgwConn getVipPortViaBGW failed, error: %s", err)
		return
	}

	c.dstAddr = &net.TCPAddr{
		IP:   vip,
		Port: vport,
	}
}
