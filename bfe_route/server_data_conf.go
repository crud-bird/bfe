package bfe_route

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_config/bfe_route_conf/host_rule_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_route_conf/vip_rule_conf"
	"github.com/crud-bird/bfe/bfe_route/bfe_cluster"
)

type ServerDataConf struct {
	HostTable    *HostTable
	ClusterTable *ClusterTable
}

func newServerDataConf() *ServerDataConf {
	return &ServerDataConf{
		HostTable:    newHostTable(),
		ClusterTable: newClusterTable(),
	}
}

func LoadServerDataConf(hostFile, vipFile, routeFile, clusterConfFile string) (*ServerDataConf, error) {
	s := newServerDataConf()

	if err := s.hostTableLoad(hostFile, vipFile, routeFile); err != nil {
		return nil, fmt.Errorf("hostTableLoad error: %s", err)
	}

	if err := s.clusterTableLoad(clusterConfFile); err != nil {
		return nil, fmt.Errorf("clusterTableLoad error: %s", err)
	}

	if err := s.check(); err != nil {
		return nil, fmt.Errorf("ServerDataConf.check error: %s", err)
	}

	return s, nil
}

func (s *ServerDataConf) hostTableLoad(hostFile, vipFile, routeFile string) error {
	hostConf, err := host_rule_conf.HostRuleConfLoad(hostFile)
	if err != nil {
		return err
	}

	vipConf, err := vip_rule_conf.VipRuleConfLoad(vipFile)
	if err != nil {
		return err
	}

	routeConf, err := route_rule_conf.RouteConfLoad(routeFile)
	if err != nil {
		return err
	}

	s.HostTable.Update(hostConf, vipConf, routeConf)

	return nil
}

func (s *ServerDataConf) clusterTableLoad(clusterConf string) error {
	if err := s.ClusterTable.Init(clusterConf); err != nil {
		return err
	}

	return nil
}

func (s *ServerDataConf) check() error {
	for pro1 := range s.HostTable.productRouteTable {
		found := false
		for _, pro2 := range s.HostTable.hostTagTable {
			if pro1 == pro2 {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("product[%s] in route should exist in host", pro1)
		}
	}

	for _, rules := range s.HostTable.productRouteTable {
		for _, rule := range rules {
			if _, err := s.ClusterTable.Lookup(rule.ClusterName); err != nil {
				return fmt.Errorf("cluster[%s] in route should exist in cluster_conf", rule.ClusterName)
			}
		}
	}

	return nil
}

func (s *ServerDataConf) HostTableLookup(hostname string) (string, error) {
	return s.HostTable.LookupProduct(hostname)
}

func (s *ServerDataConf) CLusterTableLookup(clusterName string) (*bfe_cluster.BfeCluster, error) {
	return s.ClusterTable.Lookup(clusterName)
}
