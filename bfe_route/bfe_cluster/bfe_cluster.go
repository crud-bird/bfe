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

func NewBfeCluster(name string) *BfeCluster {
	return &BfeCluster{
		Name: name,
	}
}

func (cluster *BfeCluster) BasicInit(conf cluster_conf.ClusterConf) {
	cluster.backendConf = conf.BackendConf
	cluster.CheckConf = conf.CheckConf
	cluster.GslbBasic = conf.GslbBasic
	cluster.timeoutReadClient = time.Duration(*conf.ClusterBasic.TimeoutReadClient)
	cluster.timeoutReadClientAgain = time.Duration(*conf.ClusterBasic.TimeoutReadClientAgain)
	cluster.timeoutWriteClient = time.Duration(*conf.ClusterBasic.TimeoutWriteClient)
	cluster.reqWriteBufferSize = *conf.ClusterBasic.ReqWriteBUfferSize
	cluster.reqFlushInternal = time.Duration(*conf.ClusterBasic.ReqFlushInterval)
	cluster.resFlushInternsl = time.Duration(*conf.ClusterBasic.ResFlushInterval)
	cluster.cancelOnClientClose = *conf.ClusterBasic.CancelOnClientClose
}

func (cluster *BfeCluster) BackendCheckConf() *cluster_conf.BackendCheck {
	cluster.RLock()
	res := cluster.CheckConf
	cluster.RUnlock()

	return res
}

func (cluster *BfeCluster) TimeoutConnSrv() int {
	cluster.RLock()
	t := *cluster.backendConf.TimeoutConnSrv
	cluster.RUnlock()

	return t
}

func (cluster *BfeCluster) BackendConf() *cluster_conf.BackendBasic {
	cluster.RLock()
	res := cluster.backendConf
	cluster.RUnlock()

	return res
}

func (cluster *BfeCluster) RetryLevel() int {
	cluster.RLock()
	res := cluster.backendConf.RetryLevel
	cluster.RUnlock()

	if res == nil {
		return cluster_conf.RetryConnect
	}

	return *res
}

func (cluster *BfeCluster) TimeoutReadClient() time.Duration {
	cluster.RLock()
	res := cluster.timeoutReadClient
	cluster.RUnlock()

	return res
}

func (cluster *BfeCluster) TimeoutReadClientAgain() time.Duration {
	cluster.RLock()
	res := cluster.timeoutReadClientAgain
	cluster.RUnlock()

	return res
}

func (cluster *BfeCluster) TimeoutWriteClient() time.Duration {
	cluster.RLock()
	res := cluster.timeoutWriteClient
	cluster.RUnlock()

	return res
}

func (cluster BfeCluster) ReqWriteBUfferSize() int {
	cluster.RLock()
	res := cluster.reqWriteBufferSize
	cluster.RUnlock()

	return res
}

func (cluster *BfeCluster) ReqFlushInterval() time.Duration {
	cluster.RLock()
	res := cluster.reqFlushInternal
	cluster.RUnlock()

	return res
}

func (cluster *BfeCluster) ResFlushInterval() time.Duration {
	cluster.RLock()
	res := cluster.resFlushInternsl
	cluster.RUnlock()

	return res
}

func (cluster *BfeCluster) CancelOnClientClose() bool {
	cluster.RLock()
	res := cluster.cancelOnClientClose
	cluster.RUnlock()

	return res
}
