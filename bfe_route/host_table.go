package bfe_route

import (
	"errors"
	"github.com/crud-bird/bfe/bfe_basic"
	"github.com/crud-bird/bfe/bfe_config/bfe_route_conf/host_rule_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_route_conf/route_rule_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_route_conf/vip_rule_conf"
	"github.com/crud-bird/bfe/bfe_route/trie"
	"strings"
)

var (
	ErrNoProduct     = errors.New("no product found")
	ErrNoProductRule = errors.New("no route rule found for product")
	ErrNoMatchRule   = errors.New("no rule match for this req")
)

type HostTable struct {
	versions       Versions
	hostTable      host_rule_conf.Host2HostTag
	hostTagTable   host_rule_conf.HostTag2Product
	vipTable       vip_rule_conf.Vip2Product
	defaultProduct string

	hostTrie          *trie.Trie
	productRouteTable route_rule_conf.ProductRouteRule
}

type Versions struct {
	HostTag      string
	Vip          string
	ProductRoute string
}

type Status struct {
	HostTableSize         int
	HostTagTableSize      int
	VipTableSize          int
	ProductRouteTableSize int
}

type route struct {
	product string
	tag     string
}

func newHostTable() *HostTable {
	return new(HostTable)
}

func (t *HostTable) updateHostTable(conf host_rule_conf.HostConf) {
	t.versions.HostTag = conf.Version
	t.hostTable = conf.HostMap
	t.hostTagTable = conf.HostTagMap
	t.defaultProduct = conf.DefaultProduct
	t.hostTrie = buildHostRoute(conf)
}

func (t *HostTable) updateVipTable(conf vip_rule_conf.VipConf) {
	t.versions.Vip = conf.Version
	t.vipTable = conf.VipMap
}

func (t *HostTable) updateRouteTable(conf *route_rule_conf.RouteTableConf) {
	t.versions.ProductRoute = conf.Version
	t.productRouteTable = conf.RuleMap
}

func (t *HostTable) Update(hostConf host_rule_conf.HostConf, vipConf vip_rule_conf.VipConf, routeConf *route_rule_conf.RouteTableConf) {
	t.updateHostTable(hostConf)
	t.updateVipTable(vipConf)
	t.updateRouteTable(routeConf)
}

func (t *HostTable) LookupHostTagAndProduct(req *bfe_basic.Request) error {
	hostName := req.HttpRequest.Host
	hostRoute, err := t.findHostRoute(hostName)
	if err != nil {
		if vip := req.Session.Vip; vip != nil {
			hostRoute, err = t.findVipRoute(vip.String())
		}
	}

	if err != nil && t.defaultProduct != "" {
		hostRoute, err = route{product: t.defaultProduct}, nil
	}

	req.Route.HostTag = hostRoute.tag
	req.Route.Product = hostRoute.product
	req.Route.Error = err

	return err
}

func (t *HostTable) LookupCluster(req *bfe_basic.Request) error {
	var clusterName string

	rules, ok := t.productRouteTable[req.Route.Product]
	if !ok {
		req.Route.ClusterName = ""
		req.Route.Error = ErrNoProductRule
		return req.Route.Error
	}

	for _, rule := range rules {
		if rule.Cond.Match(req) {
			clusterName = rule.ClusterName
			break
		}
	}

	if clusterName == "" {
		req.Route.ClusterName = ""
		req.Route.Error = ErrNoMatchRule
		return req.Route.Error
	}

	req.Route.ClusterName = clusterName

	return nil
}

func (t *HostTable) Lookup(req *bfe_basic.Request) bfe_basic.RequestRoute {
	route := bfe_basic.RequestRoute{}

	if err := t.LookupHostTagAndProduct(req); err != nil {
		route.Error = err
		return route
	}

	route.Product = req.Route.Product
	route.HostTag = req.Route.HostTag

	if err := t.LookupCluster(req); err != nil {
		route.Error = err
		return route
	}

	route.ClusterName = req.Route.ClusterName

	return route
}

func (t *HostTable) LookupPoductByVip(vip string) (string, error) {
	route, err := t.findVipRoute(vip)
	if err != nil {
		return "", err
	}

	return route.product, nil
}

func (t *HostTable) LookupProduct(hostname string) (string, error) {
	route, err := t.findHostRoute(hostname)
	if err != nil {
		return "", err
	}

	return route.product, nil
}

func (t *HostTable) GetVersions() Versions {
	return t.versions
}

func (t *HostTable) GetStatus() Status {
	return Status{
		ProductRouteTableSize: len(t.productRouteTable),
		HostTableSize:         len(t.hostTable),
		HostTagTableSize:      len(t.hostTagTable),
		VipTableSize:          len(t.vipTable),
	}
}

func (t *HostTable) findHostRoute(host string) (route, error) {
	if t.hostTrie == nil {
		return route{}, ErrNoProduct
	}

	host = strings.ToLower(host)
	match, ok := t.hostTrie.Get(strings.Split(reverseFqdnHost(hostnameStrip(host)), "."))
	if ok {
		return match.(route), nil
	}

	return route{}, ErrNoProduct
}

func (t *HostTable) findVipRoute(vip string) (route, error) {
	if len(t.vipTable) == 0 {
		return route{}, ErrNoProduct
	}

	if product, ok := t.vipTable[vip]; ok {
		return route{product: product}, nil
	}

	return route{}, ErrNoProduct
}

func hostnameStrip(hostname string) string {
	return strings.Split(hostname, ":")[0]
}

func reverseFqdnHost(host string) string {
	r := []rune(host)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	if len(r) > 0 && r[0] == '.' {
		r = r[1:]
	}

	return string(r)
}

func buildHostRoute(conf host_rule_conf.HostConf) *trie.Trie {
	hostTrie := trie.NewTrie()

	for host, tag := range conf.HostMap {
		host = strings.ToLower(host)
		product := conf.HostTagMap[tag]
		hostTrie.Set(strings.Split(reverseFqdnHost(host), "."), route{product: product, tag: tag})
	}

	return hostTrie
}
