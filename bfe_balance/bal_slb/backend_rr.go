package bal_slb

import (
	"github.com/crud-bird/bfe/bfe_balance/backend"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
)

type BackendRR struct {
	weight  int
	current int
	backend *backend.BfeBackend
}

func NewBackendRR() *BackendRR {
	return &BackendRR{backend: backend.NewBfeBackend()}
}

func (backRR *BackendRR) Init(subClusterName string, conf *cluster_table_conf.BackendConf) {
	backRR.weight = *conf.Weight
	backRR.current = *conf.Weight
	backRR.backend.Init(subClusterName, conf)
}

func (backRR *BackendRR) UpdateWeight(weight int) {
	backRR.weight = weight
	if weight <= 0 {
		backRR.current = 0
	}
}

func (backRR *BackendRR) Release() {
	backRR.backend.Release()
}

func (backRR *BackendRR) MatchAddrPort(addr string, port int) bool {
	return backRR.backend.Addr == addr && backRR.backend.Port == port
}
