package bal_gslb

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_balance/backend"
	"github.com/crud-bird/bfe/bfe_balance/bal_slb"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
)

const (
	TypeGslbNormal    = 0
	TypeGslbBlackhole = 1
)

type SubCluster struct {
	Name     string
	sType    int
	backends *bal_slb.BalanceRR
	weight   int
}

func newSubCluster(name string) *SubCluster {
	sType := TypeGslbNormal
	if name == "GSLB_BLACKHOLE" {
		sType = TypeGslbBlackhole
	}

	return &SubCluster{
		Name:     name,
		sType:    sType,
		backends: bal_slb.NewBalanceRR(name),
	}
}

func (sub *SubCluster) init(backends cluster_table_conf.SubClusterBackend) {
	sub.backends.Init(backends)
}

func (sub *SubCluster) update(backends cluster_table_conf.SubClusterBackend) {
	sub.backends.Update(backends)
}

func (sub *SubCluster) release() {
	sub.backends.Release()
}

func (sub *SubCluster) Len() int {
	return sub.backends.Len()
}

func (sub *SubCluster) balance(algor int, key []byte) (*backend.BfeBackend, error) {
	if sub.backends.Len() == 0 {
		return nil, fmt.Errorf("no backend in sub cluster[%s]", sub.Name)
	}

	return sub.backends.Balance(algor, key)
}

type SubClusterList []*SubCluster

type SubClusterListSortor struct {
	l SubClusterList
}

func (s SubClusterListSortor) Len() int {
	return len(s.l)
}

func (s SubClusterListSortor) Swap(i, j int) {
	s.l[i], s.l[j] = s.l[j], s.l[i]
}

func (s SubClusterListSortor) Less(i, j int) bool {
	return s.l[i].Name < s.l[j].Name
}
