package bfe_proxy

import (
	"github.com/baidu/go-lib/web-monitor/metrics"
)

type ProxyState struct {
	ProxyErrReadHeader      *metrics.Counter // connection with io err while read header
	ProxyErrNoProxyProtocol *metrics.Counter // connection with signature unmatched
	ProxyMatchedV1Signature *metrics.Counter // connection with signature v1 matched
	ProxyMatchedV2Signature *metrics.Counter // connection with signature v1 matched
	ProxyErrInvalidHeader   *metrics.Counter // connection with invalid header
	ProxyNormalV1Header     *metrics.Counter // connection with normal v1 header
	ProxyNormalV2Header     *metrics.Counter // connection with normal v2 header
}

var state ProxyState

func GetProxyState() *ProxyState {
	return &state
}
