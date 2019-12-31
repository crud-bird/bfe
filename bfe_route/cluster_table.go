package bfe_route

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/crud-bird/bfe/bfe_route/bfe_cluster"
)

type ClusterMap map[string]*bfe_cluster.BfeCluster

type ClusterTable struct {
	clusterTable ClusterMap
	versions     ClusterVersion
}

type ClusterVersion struct {
	CLusterConfVer string
}

func newClusterTable() *ClusterTable {
	return new(ClusterTable)
}

func (t *ClusterTable) Init(filename string) error {
	t.clusterTable = make(ClusterMap)
	conf, err := cluster_conf.ClusterConfLoad(filename)
	if err != nil {
		return err
	}

	t.BasicInit(conf)

	return nil
}

func (t *ClusterTable) BasicInit(confs cluster_conf.BfeClusterConf) {
	t.clusterTable = make(ClusterMap)

	for name, conf := range *confs.Config {
		cluster := bfe_cluster.NewBfeCluster(name)
		cluster.BasicInit(conf)
		t.clusterTable[name] = cluster
	}

	t.versions.CLusterConfVer = *confs.Version
}

func (t *ClusterTable) Lookup(name string) (*bfe_cluster.BfeCluster, error) {
	cluster, ok := t.clusterTable[name]
	if !ok {
		return cluster, fmt.Errorf("no cluster found for %s", name)
	}

	return cluster, nil
}

func (t *ClusterTable) GetVersions() ClusterVersion {
	return t.versions
}

func (t *ClusterTable) ClusterMap() ClusterMap {
	return t.clusterTable
}
