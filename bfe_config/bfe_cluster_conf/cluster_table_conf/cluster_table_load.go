package cluster_table_conf

import (
	"errors"
	"fmt"
	json "github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"sort"
)

type BackendConf struct {
	Name   *string
	Addr   *string
	Port   *int
	Weight *int
}

func (b *BackendConf) AddrInfo() string {
	return fmt.Sprintf("%s:%d", *b.Addr, b.Port)
}

type SubClusterBackend []*BackendConf
type ClusterBackend map[string]SubClusterBackend
type AllClusterBackend map[string]ClusterBackend

func (s SubClusterBackend) Len() int {
	return len(s)
}

func (s SubClusterBackend) Less(i, j int) bool {
	if *s[i].Addr != *s[j].Addr {
		return *s[i].Addr < *s[j].Addr
	}

	return *s[i].Port < *s[j].Port
}

func (s SubClusterBackend) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SubClusterBackend) Sort() {
	sort.Sort(s)
}

func (s SubClusterBackend) Shuffle() {
	for i := len(s) - 1; i > 1; i-- {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

func (allBackend AllClusterBackend) Sort() {
	for _, clusterBackend := range allBackend {
		for _, backends := range clusterBackend {
			backends.Sort()
		}
	}
}

func (allBackend AllClusterBackend) Shuffle() {
	for _, clusterBackend := range allBackend {
		for _, backends := range clusterBackend {
			backends.Shuffle()
		}
	}
}

func (allBackend AllClusterBackend) HasDiff(compared AllClusterBackend) bool {
	if len(allBackend) != len(compared) {
		return true
	}

	for k, v := range allBackend {
		val, ok := compared[k]
		if !ok {
			return true
		}

		if v.HasDiff(val) {
			return true
		}
	}

	return false
}

func (allBackend AllClusterBackend) IsSub(compared AllClusterBackend) bool {
	for k, v := range allBackend {
		val, ok := compared[k]
		if !ok {
			return false
		}

		if !v.IsSame(val) {
			return false
		}
	}

	return true
}

func (backend ClusterBackend) HasDiff(compared ClusterBackend) bool {
	return !reflect.DeepEqual(backend, compared)
}

func (backend ClusterBackend) IsSame(compared ClusterBackend) bool {
	return reflect.DeepEqual(backend, compared)
}

type ClusterTableConf struct {
	Version *string
	Config  *AllClusterBackend
}

func BackendConfCheck(conf *BackendConf) error {
	if conf.Name == nil {
		return errors.New("no name")
	}

	if conf.Addr == nil {
		return errors.New("no addr")
	}

	if conf.Port == nil {
		return errors.New("no port")
	}

	if conf.Weight == nil {
		return errors.New("no weight")
	}

	return nil
}

func (allBackend *AllClusterBackend) Check() error {
	return AllClusterBackendCheck(allBackend)
}

func (sub *SubClusterBackend) Check() error {
	avail := false
	for i, backendConf := range *sub {
		if err := BackendConfCheck(backendConf); err != nil {
			return fmt.Errorf("%d %s", i, err)
		}

		if *backendConf.Weight > 0 {
			avail = true
		}
	}

	if !avail {
		return fmt.Errorf("no avail backend")
	}

	return nil
}

func AllClusterBackendCheck(conf *AllClusterBackend) error {
	for name, backend := range *conf {
		for subName, subBackend := range backend {
			if err := subBackend.Check(); err != nil {
				return fmt.Errorf("%s %s %s", name, subName, err)
			}
		}
	}

	return nil
}

func ClusterTableConfCheck(conf ClusterTableConf) error {
	if conf.Version == nil {
		return errors.New("no version")
	}

	if conf.Config == nil {
		return errors.New("no config")
	}

	if err := AllClusterBackendCheck(conf.Config); err != nil {
		return fmt.Errorf("ClusterTableConf.Config: %s", err)
	}

	return nil
}

func CLusterTableLoad(filename string) (ClusterTableConf, error) {
	var config ClusterTableConf

	f, err := os.Open(filename)
	if err != nil {
		return config, err
	}

	decoder := json.NewDecoder()
	err = decoder.DecodeReader(f, &config)
	f.Close()
	if err != nil {
		return config, err
	}

	if err = ClusterTableConfCheck(config); err != nil {
		return config, err
	}

	return config, nil
}

func ClusterTableDump(conf ClusterTableConf, filename string) error {
	confJson, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(filename, confJson, 0644); err != nil {
		return err
	}

	return nil
}
