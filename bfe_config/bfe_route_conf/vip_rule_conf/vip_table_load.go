package vip_rule_conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
)

type VipList []string
type Product2Vip map[string]VipList
type Vip2Product map[string]string

type VipTableConf struct {
	Version string
	Vips    Product2Vip
}

type VipConf struct {
	Version string
	VipMap  Vip2Product
}

func (conf *VipTableConf) LoadAndCheck(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(conf); err != nil {
		return "", nil
	}

	if err := VipTableConfCheck(*conf); err != nil {
		return "", err
	}

	return conf.Version, nil
}

func VipTableConfCheck(conf VipTableConf) error {
	if conf.Version == "" {
		return errors.New("no Version")
	}

	for product, vipList := range conf.Vips {
		var list VipList
		for _, vip := range vipList {
			ip, err := net.ResolveIPAddr("ip", vip)
			if err != nil {
				return fmt.Errorf("invalid vip %s for %s", vip, product)
			}

			list = append(list, ip.String())
		}
		conf.Vips[product] = list
	}
	return nil
}

func VipRuleConfLoad(filename string) (VipConf, error) {
	var vipConf VipConf

	var config VipTableConf
	if _, err := config.LoadAndCheck(filename); err != nil {
		return vipConf, err
	}

	vipConf.Version = config.Version
	vipConf.VipMap = make(Vip2Product)
	for product, vipList := range config.Vips {
		for _, vip := range vipList {
			vipConf.VipMap[vip] = product
		}
	}

	return vipConf, nil
}
