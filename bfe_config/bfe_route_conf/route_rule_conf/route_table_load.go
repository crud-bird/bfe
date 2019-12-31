package route_rule_conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crud-bird/bfe/bfe_basic/condition"
	"os"
)

type RouteRule struct {
	Cond        condition.Condition
	ClusterName string
}

type RouteRuleFile struct {
	Cond        *string
	ClusterName *string
}

type RouteRules []RouteRule
type RouteRuleFiles []RouteRuleFile

type ProductRouteRule map[string]RouteRules
type ProductRouteRuleFile map[string]RouteRuleFiles

type RouteTableFile struct {
	Version     *string
	ProductRule *ProductRouteRuleFile
}

type RouteTableConf struct {
	Version string
	RuleMap ProductRouteRule
}

func convert(fileConf *RouteTableFile) (*RouteTableConf, error) {
	conf := &RouteTableConf{
		RuleMap: make(ProductRouteRule),
	}

	if fileConf.Version == nil {
		return nil, errors.New("no Version")
	}

	if fileConf.ProductRule == nil {
		return nil, errors.New("no product rule")
	}

	conf.Version = *fileConf.Version

	for product, files := range *fileConf.ProductRule {
		rules := make(RouteRules, len(files))
		for i, file := range files {
			if file.ClusterName == nil {
				return nil, errors.New("no ClusterName")
			}

			if file.Cond == nil {
				return nil, errors.New("no cond")
			}

			rules[i].ClusterName = *file.ClusterName
			cond, err := condition.Build(*file.Cond)
			if err != nil {
				return nil, fmt.Errorf("error build [%s] [%s]", *file.Cond, err)
			}
			rules[i].Cond = cond
		}

		conf.RuleMap[product] = rules
	}

	return conf, nil
}

func (conf *RouteTableConf) LoadAndCheck(filename string) (string, error) {
	var fileConf RouteTableFile
	file, err := os.Open("filename")
	if err != nil {
		return "", err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fileConf); err != nil {
		return "", err
	}

	pConf, err := convert(&fileConf)
	if err != nil {
		return "", err
	}

	*conf = *pConf

	return conf.Version, nil
}

func RouteConfLoad(filename string) (*RouteTableConf, error) {
	var conf RouteTableConf
	if _, err := conf.LoadAndCheck(filename); err != nil {
		return nil, err
	}

	return &conf, nil
}
