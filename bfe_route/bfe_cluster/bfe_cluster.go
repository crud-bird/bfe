package bfe_cluster

import (
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"sync"
	"time"
)

type BfeCluster struct {
	sync.RWMutex
	Name string

	backendConf *cluster_conf.BackendBasic
	CheckConf   *cluster_conf.BackendCheck
	GslbBasic   *cluster_conf.GslbBasicConf

	timeoutReadClient      time.Duration
	timeoutReadClientAgain time.Duration
	timeoutWriteClient     time.Duration

	reqWriteBufferSize  int
	reqFlushInternal    time.Duration
	resFlushInternsl    time.Duration
	cancelOnClientClose bool
}
