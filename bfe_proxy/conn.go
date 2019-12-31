package bfe_proxy

import (
	"fmt"
	bufio "github.com/crud-bird/bfe/bfe_bufio"
	"io"
	"net"
	"sync"
	"time"
)

const (
	defaultProxyHeaderTimeout        = 30 * time.Second
	defaultMaxProxyHeaderBytes int64 = 2048
	noLimit                    int64 = (1 << 63) - 1
)

type Conn struct {
	conn          net.Conn
	lmtReader     *io.LimitedReader
	bufReader     *bufio.Reader
	headerTimeout time.Duration
	headerLimit   int64
	headerErr     error
	dstAddr       *net.TCPAddr
	srcAddr       *net.TCPAddr
	once          sync.Once
}

func NewConn(conn net.Conn, timeout time.Duration, max int64) *Conn {
	if timeout <= 0 {
		timeout = defaultProxyHeaderTimeout
	}

	if max <= 0 {
		max = defaultMaxProxyHeaderBytes
	}

	lmtReader := io.LimitReader(conn, max).(*io.LimitedReader)

	return &Conn{
		headerTimeout: timeout,
		headerLimit:   max,
		conn:          conn,
		lmtReader:     lmtReader,
		bufReader:     bufio.NewReader(lmtReader),
	}
}

func (p *Conn) Read(b []byte) (int, error) {
	p.checkProxyHeaderOnce()
	if p.headerErr != nil {
		return 0, p.headerErr
	}

	return p.bufReader.Read(b)
}

func (p *Conn) Write(b []byte) (int, error) {
	return p.conn.Write(b)
}

func (p *Conn) Close() error {
	return p.conn.Close()
}

func (p *Conn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *Conn) RemoteAddr() net.Addr {
	p.checkProxyHeaderOnce()
	if p.srcAddr != nil {
		return p.srcAddr
	}

	return p.conn.RemoteAddr()
}

func (p *Conn) VirtualAddr() net.Addr {
	p.checkProxyHeaderOnce()
	if p.dstAddr != nil {
		return p.dstAddr
	}

	return nil
}

func (p *Conn) BalancerAddr() net.Addr {
	p.checkProxyHeaderOnce()
	if p.dstAddr != nil {
		return p.conn.RemoteAddr()
	}

	return nil
}

func (p *Conn) GeNetConn() net.Conn {
	return p.conn
}

func (p *Conn) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}

func (p *Conn) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}

func (p *Conn) SetWriteDeadline(t time.Time) error {
	return p.conn.SetWriteDeadline(t)
}

func (p *Conn) checkProxyHeaderOnce() {
	p.once.Do(func() {
		if err := p.checkProxyHeader(); err != nil {
			p.Close()
		}
	})
}

func (p *Conn) checkProxyHeader() error {
	p.conn.SetReadDeadline(time.Now().Add(p.headerTimeout))

	defer func() {
		p.conn.SetReadDeadline(time.Time{})
		p.lmtReader.N = noLimit
	}()

	hdr, err := Read(p.bufReader)
	if err == ErrNoProxyProtocol {
		return nil
	}
	if err != nil {
		p.conn.Close()
		p.headerErr = err
		return err
	}

	srcAddr := net.JoinHostPort(hdr.SourceAddress.String(), fmt.Sprintf("%d", hdr.SourcePort))
	p.srcAddr, err = net.ResolveTCPAddr(hdr.TransportProtocol.String(), srcAddr)
	if err != nil {
		p.conn.Close()
		return err
	}

	dstAddr := net.JoinHostPort(hdr.DestinationAddress.String(), fmt.Sprintf("%d", hdr.DestinationPort))
	p.dstAddr, err = net.ResolveTCPAddr(hdr.TransportProtocol.String(), dstAddr)
	if err != nil {
		p.conn.Close()
		return err
	}

	return nil
}
