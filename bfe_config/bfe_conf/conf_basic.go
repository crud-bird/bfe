package bfe_conf

import (
	"fmt"
	"runtime"

	"github.com/crud-bird/bfe/bfe_util"
	"github.com/sirupsen/logrus"
)

const (
	BALANCE_BGW   = "BGW"
	BALANCE_PROXY = "PROXY"
	BALANCE_NONE  = "NONE"
)

type ConfigBasic struct {
	HttpPort    int
	HttpsPort   int
	MonitorPort int
	MaxCpus     int

	Layer4LoadBalancer      string
	TlsHandshakeTimeout     int
	ClientReadTimeout       int
	ClientWriteTimeout      int
	GracefulShutdownTimeout int
	MaxHeaderBytes          int
	MaxHeaderUriBytes       int
	MaxProxyHeaderBytes     int
	KeepAlivedEnabled       bool

	Modules []string

	HostRuleConf  string
	VipRuleConf   string
	RouteRuleConf string

	ClusterTableConf string
	GslbConf         string
	ClusterConf      string
	NameConf         string

	MonitorIterval int

	DebugServHttp    bool
	DebugBfeRoute    bool
	DebugBal         bool
	DebugHealthCheck bool
}

func (cfg *ConfigBasic) SetDeafaultConf() {
	cfg.HttpPort = 8080
	cfg.HttpsPort = 8443
	cfg.MonitorPort = 8421
	cfg.MaxCpus = 0

	cfg.TlsHandshakeTimeout = 30
	cfg.ClientReadTimeout = 60
	cfg.ClientWriteTimeout = 60
	cfg.GracefulShutdownTimeout = 10
	cfg.MaxHeaderBytes = 1048576
	cfg.MaxHeaderUriBytes = 8192
	cfg.KeepAlivedEnabled = true

	cfg.HostRuleConf = "server_data_conf/host_rule.data"
	cfg.VipRuleConf = "server_data_conf/vip_rule_data"
	cfg.RouteRuleConf = "server_data_conf/route_rule_data"

	cfg.ClusterTableConf = "cluster_conf/cluster_table.data"
	cfg.GslbConf = "cluster_conf/gslb.data"
	cfg.ClusterTableConf = "server_data_conf/cluster_conf.data"
	cfg.NameConf = "server_data_conf/name_conf.data"

	cfg.MonitorPort = 20
}

func (cfg *ConfigBasic) Check(confRoot string) error {
	return ConfBasicCheck(cfg, confRoot)
}

func ConfBasicCheck(cfg *ConfigBasic, confRoot string) error {
	if err := basicConfCheck(cfg); err != nil {
		return err
	}

	return nil
}

func basicConfCheck(cfg *ConfigBasic) error {
	if cfg.HttpPort < 1 || cfg.HttpPort > 65535 {
		return fmt.Errorf("HttpPort[%d] should be in [1, 65535]", cfg.HttpPort)
	}

	if cfg.HttpsPort < 1 || cfg.HttpsPort > 65536 {
		return fmt.Errorf("HttpsPort[%d] should be in [1, 65535]", cfg.HttpsPort)
	}

	if cfg.MonitorIterval < 1 || cfg.MonitorPort > 65536 {
		return fmt.Errorf("MonitorIterval[%d] should be in [1, 65535]", cfg.MonitorPort)
	}

	if cfg.MaxCpus < 0 {
		return fmt.Errorf("MaxCpus[%d] is too small", cfg.MaxCpus)
	} else if cfg.MaxCpus == 0 {
		cfg.MaxCpus = runtime.NumCPU()
	}

	if err := checkLayer4LoadBalancer(cfg); err != nil {
		return err
	}

	if cfg.TlsHandshakeTimeout > 1200 {
		return fmt.Errorf("TlsHandshakeTimeout[%d] should be < 1200", cfg.TlsHandshakeTimeout)
	}

	if cfg.TlsHandshakeTimeout <= 0 {
		return fmt.Errorf("TlsHandshakeTimeout[%d] should be > 0", cfg.TlsHandshakeTimeout)
	}

	if cfg.ClientReadTimeout <= 0 {
		return fmt.Errorf("ClientReadTimeout[%d] should be > 0", cfg.ClientReadTimeout)
	}

	if cfg.ClientWriteTimeout <= 0 {
		return fmt.Errorf("ClientWriteTimeout[%d] should be > 0", cfg.ClientWriteTimeout)
	}

	if cfg.GracefulShutdownTimeout <= 0 {
		return fmt.Errorf("GracefulShutdownTimeout[%d] should be > 0", cfg.GracefulShutdownTimeout)
	}

	if cfg.MonitorIterval <= 0 {
		logrus.Warn("cfg.MonitorIterval not set value, use default value 20")
		cfg.MonitorIterval = 20
	} else if cfg.MonitorIterval > 60 {
		logrus.Warn("MonitorIterval[%d] > 60, use 60", cfg.MonitorIterval)
	} else {
		if 60%cfg.MonitorIterval > 0 {
			return fmt.Errorf("MonitorIterval[%d] can not devide 60", cfg.MonitorIterval)
		}

		if cfg.MonitorIterval < 0 {
			return fmt.Errorf("MonitorIterval[%d] is too small", cfg.MonitorIterval)
		}
	}

	if cfg.MaxHeaderUriBytes <= 0 {
		return fmt.Errorf("MaxHeaderUriBytes[%d] should be > 0", cfg.MaxHeaderUriBytes)
	}

	if cfg.MaxHeaderBytes <= 0 {
		return fmt.Errorf("MaxHeaderBytes[%d] should be > 0", cfg.MaxHeaderBytes)
	}

	return nil
}

func checkLayer4LoadBalancer(cfg *ConfigBasic) error {
	if len(cfg.Layer4LoadBalancer) == 0 {
		cfg.Layer4LoadBalancer = BALANCE_BGW
	}

	if cfg.Layer4LoadBalancer == BALANCE_BGW || cfg.Layer4LoadBalancer == BALANCE_PROXY || cfg.Layer4LoadBalancer == BALANCE_NONE {
		return nil
	}

	return fmt.Errorf("Layer4LoadBalancer[%s] should be BGW/PROXY/NONE", cfg.Layer4LoadBalancer)
}

func dataFIleConfCheck(cfg *ConfigBasic, confRoot string) error {
	if cfg.HostRuleConf == "" {
		cfg.HostRuleConf = "server_data_conf/host_rule.data"
		logrus.Warn("HostRuleConf not set use defaault value[%s]", cfg.HostRuleConf)
	}
	cfg.HostRuleConf = bfe_util.ConfPathProc(cfg.HostRuleConf, confRoot)

	if cfg.VipRuleConf == "" {
		cfg.VipRuleConf = "server_data_conf/vip_rule.data"
		logrus.Warn("VipRuleConf not set, use default value[%s]", cfg.VipRuleConf)
	}
	cfg.VipRuleConf = bfe_util.ConfPathProc(cfg.VipRuleConf, confRoot)

	if cfg.RouteRuleConf == "" {
		cfg.RouteRuleConf = "server_data_conf/route_rule.data"
		logrus.Warn("RouteRuleConf not set, use default value[%s]", cfg.RouteRuleConf)
	}
	cfg.RouteRuleConf = bfe_util.ConfPathProc(cfg.RouteRuleConf, confRoot)

	if cfg.ClusterTableConf == "" {
		cfg.ClusterTableConf = "cluster_conf/cluster_table.data"
		logrus.Warn("ClusterTableConf not set, use default value[%s]", cfg.ClusterTableConf)
	}
	cfg.ClusterTableConf = bfe_util.ConfPathProc(cfg.ClusterTableConf, confRoot)

	if cfg.GslbConf == "" {
		cfg.GslbConf = "cluster_conf/gslb.data"
		logrus.Warn("GslbConf not set, use default value[%s]", cfg.GslbConf)
	}
	cfg.GslbConf = bfe_util.ConfPathProc(cfg.GslbConf, confRoot)

	if cfg.ClusterConf == "" {
		cfg.ClusterConf = "server_data_conf/cluster.data"
		logrus.Warn("ClusterConf not set, use default value[%s]", cfg.ClusterConf)
	}
	cfg.ClusterConf = bfe_util.ConfPathProc(cfg.ClusterConf, confRoot)

	if cfg.NameConf == "" {
		logrus.Warn("NameConf not set, ignore optionsl name conf")
	} else {
		cfg.NameConf = bfe_util.ConfPathProc(cfg.NameConf, confRoot)
	}

	return nil
}
