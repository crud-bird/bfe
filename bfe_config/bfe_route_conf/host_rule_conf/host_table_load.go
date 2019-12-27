package host_rule_conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type HostnameList []string
type HostTagList []string

type HostTagToHost map[string]*HostnameList
type ProductToHostTag map[string]*HostTagList

type Host2HostTag map[string]string
type HostTag2Product map[string]string

type HostTableConf struct {
	Version        *string
	DefaultProduct *string
	Hosts          *HostTagToHost
	HostTags       *ProductToHostTag
}

type HostConf struct {
	Version        string
	DefaultProduct string
	HostMap        Host2HostTag
	HostTagMap     HostTag2Product
}

func (conf *HostTableConf) LoadAndCheck(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(*conf); err != nil {
		return "", err
	}

	if err := HostTableConfCheck(*conf); err != nil {
		return "", err
	}

	return *(conf.Version), nil
}

func HostTableConfCheck(conf HostTableConf) error {
	if conf.Version == nil {
		return errors.New("no version")
	}

	if conf.Hosts == nil {
		return errors.New("no Hosts")
	}

	if conf.HostTags == nil {
		return errors.New("no HostTags")
	}

	for product, tagList := range *conf.HostTags {
		if tagList == nil {
			return fmt.Errorf("no HostTagList for %s", product)
		}
	}

	for tag, list := range *conf.Hosts {
		if list == nil {
			return fmt.Errorf("no HostnameList for %s", tag)
		}

		find := false
	HOST_TAG_CHECK:
		for _, list := range *conf.HostTags {
			for _, ht := range *list {
				if ht == tag {
					find = true
					break HOST_TAG_CHECK
				}
			}
		}

		if !find {
			return fmt.Errorf("HostTag[%s] in Hosts should also exist in HostTags", tag)
		}
	}

	if conf.DefaultProduct != nil {
		tags := *conf.HostTags
		if _, ok := tags[*conf.DefaultProduct]; !ok {
			return fmt.Errorf("DefaultProduct[%s] must exist in HostTags", *conf.DefaultProduct)
		}
	}

	return nil
}

func HostRuleConfLoad(filename string) (HostConf, error) {
	var conf HostConf
	var config HostTableConf

	if _, err := config.LoadAndCheck(filename); err != nil {
		return conf, err
	}

	host2HostTag := make(Host2HostTag)
	for tag, list := range *config.Hosts {
		for _, name := range *list {
			if host2HostTag[name] != "" {
				return conf, fmt.Errorf("host duplicate for %s", name)
			}
			host2HostTag[name] = tag
		}
	}

	hostTag2Product := make(HostTag2Product)
	for product, list := range *config.HostTags {
		for _, tag := range *list {
			hostTag2Product[tag] = product
		}
	}

	if config.DefaultProduct != nil {
		conf.DefaultProduct = *config.DefaultProduct
	}

	conf.Version = *config.Version
	conf.HostMap = host2HostTag
	conf.HostTagMap = hostTag2Product

	return conf, nil
}
