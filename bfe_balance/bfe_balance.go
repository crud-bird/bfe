package bfe_balance

import (
	"github.com/crud-bird/bfe/bfe_balance/backend"
	"github.com/crud-bird/bfe/bfe_basic"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/gslb_conf"
)

type BfeBalance interface {
	Init(cluster_table_conf.ClusterBackend, cluster_conf.GslbBasicConf, gslb_conf.GslbClusterConf) error
	Reload(cluster_table_conf.ClusterBackend, cluster_conf.GslbBasicConf, gslb_conf.GslbClusterConf) error
	Balance(*bfe_basic.Request) (*backend.BfeBackend, error)
	Release()
}
