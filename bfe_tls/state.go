package bfe_tls

import (
	"github.com/baidu/go-lib/web-monitor/metrics"
)

type TlsState struct {
	TlsHandshakeReadClientHelloErr        *metrics.Counter
	TlsHandshakeFullAll                   *metrics.Counter
	TlsHandshakeFullSucc                  *metrics.Counter
	TlsHandshakeResumeAll                 *metrics.Counter
	TlsHandshakeResumeSucc                *metrics.Counter
	TlsHandshakeCheckResumeSessionTicket  *metrics.Counter
	TlsHandshakeShouldResumeSessionTicket *metrics.Counter
	TlsHandshakeCheckResumeSessionCache   *metrics.Counter
	TlsHandshakeShouldResumeSessionCache  *metrics.Counter
	TlsHandshakeAcceptSslv2ClientHello    *metrics.Counter
	TlsHandshakeAcceptEcdheWithoutExt     *metrics.Counter
	TlsHandshakeNoSharedCipherSuite       *metrics.Counter
	TlsHandshakeSslv2NotSupport           *metrics.Counter
	TlsHandshakeOcspTimeErr               *metrics.Counter
	TlsStatusRequestExtCount              *metrics.Counter
	TlsHandshakeZeroData                  *metrics.Counter
}

var state TlsState

func GetTlsState() *TlsState {
	return &state
}
