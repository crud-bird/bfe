package gslb_conf

import (
	"errors"
	"fmt"
	json "github.com/pquerna/ffjson/ffjson"
	"os"
	"reflect"
)

var (
	ErrGslbNoHostname = errors.New("no HostName")
	ErrGslbNoTs       = errors.New("no ts")
)

type GslbClusterConf map[string]int

type GslbClustersConf map[string]GslbClusterConf

type GslbConf struct {
	Clusters *GslbClustersConf
	Hostname *string
	Ts       *string
}

func (gslb GslbClustersConf) HasDiff(compared GslbClustersConf) bool {
	if len(gslb) != len(compared) {
		return true
	}

	for cluster, conf := range gslb {
		comparedConf, ok := compared[cluster]
		if !ok {
			return true
		}

		if conf.HasDiff(comparedConf) {
			return true
		}
	}

	return false
}

func (gslbConf GslbConf) IsSub(compared GslbConf) bool {
	for cluster, conf := range *gslbConf.Clusters {
		comparedConf, ok := (*compared.Clusters)[cluster]
		if !ok {
			return false
		}

		if !conf.IsSame(comparedConf) {
			return false
		}
	}

	return true
}

func (conf GslbClusterConf) Check() error {
	total := 0
	for _, weight := range conf {
		if weight > 0 {
			total += weight
		}
	}

	if total <= 0 {
		return errors.New("GslbClusterConf check: total weight <= 0")
	}

	return nil
}

func (conf GslbClusterConf) IsSame(compared GslbClusterConf) bool {
	return reflect.DeepEqual(conf, compared)
}

func (conf GslbClusterConf) HasDiff(compared GslbClusterConf) bool {
	return !reflect.DeepEqual(conf, compared)
}

func (conf GslbClustersConf) Check() error {
	for cluster, clusterConf := range conf {
		if err := clusterConf.Check(); err != nil {
			return fmt.Errorf("[%s] check conf err [%s]", cluster, err)
		}
	}

	return nil
}

func (conf *GslbConf) Check() error {
	return GslbConfNilCheck(*conf)
}

func GslbConfNilCheck(conf GslbConf) error {
	if conf.Clusters == nil {
		return errors.New("no Clusters")
	}

	if conf.Hostname == nil {
		return ErrGslbNoHostname
	}

	if conf.Ts == nil {
		return ErrGslbNoTs
	}

	return nil
}

func GslbConfCheck(conf GslbConf) error {
	if err := GslbConfNilCheck(conf); err != nil {
		return fmt.Errorf("Check nul: %s", err)
	}

	if err := conf.Clusters.Check(); err != nil {
		return fmt.Errorf("Clusters check err：%s", err)
	}

	return nil
}

func GslbConfLoad(filename string) (GslbConf, error) {
	var config GslbConf

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}

	decoder := json.NewDecoder()
	err = decoder.DecodeReader(file, &config)
	file.Close()
	if err != nil {
		return config, err
	}

	if err = GslbConfCheck(config); err != nil {
		return config, err
	}

	return config, nil
}
