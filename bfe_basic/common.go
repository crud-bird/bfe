package bfe_basic

import (
	"github.com/crud-bird/bfe/bfe_route/bfe_cluster"
)

const (
	HeaderBfeIP         = "X-Bfe-Ip"
	HeaderBfeLogId      = "X-Bfe-Log-Id"
	HeaderForwardedFor  = "X-Forwarded-For"
	HeaderForWardedPort = "X-Forwarded-Port"
	HeaderRealIP        = "X-Real-Ip"
	HeaderRealPort      = "X-Real-Port"
)

type OperationStage int

const (
	StateStartConn OperationStage = iota
	StageReadReqHeader
	StageReadReqBody
	StageConnBackend
	StageWriteBackend
	StageReadResponseHeader
	StageReadResponseBody
	StageWriteClient
	StageEndRequest
)

type ServerDataConfInterface interface {
	ClusterTableLookup(string) (*bfe_cluster.BfeCluster, error)
	HostTableLookup(string) (string, error)
}
