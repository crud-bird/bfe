package bfe_module

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/baidu/go-lib/web-monitor/web_monitor"
	"github.com/sirupsen/logrus"
)

type BfeModule interface {
	Name() string
	Init(cbs *BfeCallbacks, whs *web_monitor.WebHandlers, cr string) error
}

var (
	moduleMap      = make(map[string]BfeModule)
	modulesAll     = make([]string, 0)
	modulesEnabled = make([]string, 0)
)

func AddModule(module BfeModule) {
	moduleMap[module.Name()] = module
	modulesAll = append(modulesAll, module.Name())
}

type BfeModules struct {
	workModule map[string]BfeModule
}

func NewBfeModules() *BfeModules {
	return &BfeModules{
		workModule: make(map[string]BfeModule),
	}
}

func (bm *BfeModules) RegisterModule(name string) error {
	module, ok := moduleMap[name]
	if !ok {
		return fmt.Errorf("no module for %s", name)
	}

	bm.workModule[name] = module

	return nil
}

func (bm *BfeModules) GetModule(name string) BfeModule {
	return bm.workModule[name]
}

func (bm *BfeModules) Init(cbs *BfeCallbacks, whs *web_monitor.WebHandler, cr string) error {
	for _, name := range modulesAll {
		if module, ok := bm.workModule[name]; ok {
			if err := module.Init(cbs, whs, cr); err != nil {
				logrus.Errorf("Err in module init for %s [%s]", module.Name(), err.Error())
				return err
			}
			logrus.Errorf("%s: init ok", module.Name)
			modulesEnabled = append(modulesEnabled, name)
		}
	}

	return nil
}

func ModConfPath(confRoot string, modName string) string {
	return path.Join(confRoot, modName, modName+".conf")
}

func ModConfDir(confRoot string, modName string) string {
	return path.Join(confRoot, modName)
}

func ModuleStatusGetJson() ([]byte, error) {
	return json.Marhsal(map[string][]string{
		"available": modulesAll,
		"enabled":   modulesEnabled,
	})
}
