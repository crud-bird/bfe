package bfe_balance

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_balance/backend"
	"github.com/crud-bird/bfe/bfe_balance/bal_gslb"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/gslb_conf"
	"github.com/crud-bird/bfe/bfe_route"
	"strings"
	"sync"
)

type BalMap map[string]*bal_gslb.BalanceGslb

type BalTable struct {
	lock     sync.Mutex
	balTable BalMap
	versions BalVersion
}

type BalVersion struct {
	ClusterTableConfVer string
	GslbCOnfTimeStamp   string
	GslbConfSrc         string
}

type BalTableState struct {
	Balancers  map[string]*bal_gslb.GslbState
	BackendNUm int
}

func NewBalTable(fetcher backend.CheckConfFetcher) *BalTable {
	backend.SetCheckConfFetcher(fetcher)
	return &BalTable{
		balTable: make(BalMap),
	}
}

func (t *BalTable) BalTableConfLoad(gFile, cFile string) (gslb_conf.GslbConf, cluster_table_conf.ClusterTableConf, error) {
	var gslbConf gslb_conf.GslbConf
	var backendConf cluster_table_conf.ClusterTableConf
	var err error

	gslbConf, err = gslb_conf.GslbConfLoad(gFile)
	if err != nil {
		return gslbConf, backendConf, err
	}

	backendConf, err = cluster_table_conf.CLusterTableLoad(cFile)

	return gslbConf, backendConf, err
}

func (t *BalTable) Init(gFile, cFile string) error {
	gConf, cConf, err := t.BalTableConfLoad(gFile, cFile)
	if err != nil {
		return err
	}

	if err := t.gslbInit(gConf); err != nil {
		return err
	}

	if err := t.backendInit(cConf); err != nil {
		return err
	}

	return nil
}

func (t *BalTable) gslbInit(gConfs gslb_conf.GslbConf) error {
	fails := make([]string, 0)
	for name, gConf := range *gConfs.Clusters {
		bal := bal_gslb.NewBalanceGslb(name)
		if err := bal.Init(gConf); err != nil {
			fails = append(fails, name)
			continue
		}
		t.balTable[name] = bal
	}

	t.versions.GslbCOnfTimeStamp = *gConfs.Ts
	t.versions.GslbConfSrc = *gConfs.Hostname

	if len(fails) != 0 {
		return fmt.Errorf("error in CLusterTable.gslbInit() for [%s]", strings.Join(fails, ","))
	}

	return nil
}

func (t *BalTable) backendInit(confs cluster_table_conf.ClusterTableConf) error {
	fails := make([]string, 0)
	for name, bal := range t.balTable {
		conf, ok := (*confs.Config)[name]
		if !ok {
			fails = append(fails, name)
			continue
		}

		if err := bal.BackendInit(conf); err != nil {
			fails = append(fails, name)
			continue
		}
	}

	t.versions.ClusterTableConfVer = *confs.Version

	if len(fails) > 0 {
		return fmt.Errorf("error in backendInit() for [%s]", strings.Join(fails, ","))
	}

	return nil
}

func (t *BalTable) SetGslbBasic(table *bfe_route.ClusterTable) {

}

// todo
