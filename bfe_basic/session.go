package bfe_basic

import (
	"github.com/crud-bird/bfe/bfe_tls"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Session struct {
	SessionID string
	StartTime time.Time
	EndTime   time.Time
	Overhead  time.Duration

	Connection net.Conn
	RemoteAddr *net.TCPAddr

	Use100Continue bool
	IsTrustIP      bool

	Proto    string
	IsSecure bool
	TlsState *bfe_tls.ConnectionState

	Vip     net.IP
	Vport   int
	Product string
	Rtt     uint32

	lock         sync.Mutex
	ReqNum       int64
	ReqNumActive int64
	ReadTotal    int64
	WriteTotal   int64
	ErrCode      error
	ErrMsg       string
	Context      map[interface{}]interface{}
}

func NewSession(conn net.Conn) *Session {
	var remoteAddr *net.TCPAddr
	if conn != nil {
		remoteAddr = conn.RemoteAddr().(*net.TCPAddr)
	}

	return &Session{
		StartTime:      time.Now(),
		Connection:     conn,
		Use100Continue: false,
		Context:        make(map[interface{}]interface{}),
		RemoteAddr:     remoteAddr,
	}
}

func (s *Session) GetVip() net.IP {
	return s.Vip
}

func (s *Session) Finish() {
	s.EndTime = time.Now()
	s.Overhead = s.EndTime.Sub(s.StartTime)
}

func (s *Session) IncReqNum(count int) int64 {
	return atomic.AddInt64(&s.ReqNum, int64(count))
}

func (s *Session) IncReqNumActive(count int) int64 {
	return atomic.AddInt64(&s.ReqNumActive, int64(count))
}

func (s *Session) UpdateReadTotal(total int) int {
	ntotal := int64(total)
	rtotal := atomic.SwapInt64(&s.ReadTotal, ntotal)
	if ntotal >= rtotal {
		return int(ntotal - rtotal)
	}

	return 0
}

func (s *Session) UpdateWriteTotal(total int) int {
	ntotal := int64(total)
	wtotal := atomic.SwapInt64(&s.WriteTotal, ntotal)

	if ntotal >= wtotal {
		return int(ntotal - wtotal)
	}

	return 0
}

func (s *Session) SetError(errCode error, errMsg string) {
	s.lock.Lock()
	s.ErrCode = errCode
	s.ErrMsg = errMsg
	s.lock.Unlock()
}

func (s *Session) GetError() (error, string) {
	s.lock.Lock()
	errCode := s.ErrCode
	errMsg := s.ErrMsg
	s.lock.Unlock()

	return errCode, errMsg
}

func (s *Session) SetContext(key, val interface{}) {
	s.lock.Lock()
	s.Context[key] = val
	s.lock.Unlock()
}

func (s *Session) GetContext(key interface{}) interface{} {
	s.lock.Lock()
	val := s.Context[key]
	s.lock.Unlock()

	return val
}
