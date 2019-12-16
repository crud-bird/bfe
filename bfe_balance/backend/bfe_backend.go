package backend

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
	"sync"
)

type BfeBackend struct {
	Name       string
	Addr       string
	Port       int
	AddrInfo   string
	SubCluster string

	sync.RWMutex
	avail   bool
	connNum int
	failNum int
	succNum int

	closeChan chan bool
}

func NewBfeBackend() *BfeBackend {
	return &BfeBackend{
		avail:     true,
		closeChan: make(chan bool),
	}
}
func (back *BfeBackend) Init(subCluster string, conf *cluster_table_conf.BackendConf) {
	back.Name = *conf.Name
	back.Addr = *conf.Addr
	back.Port = *conf.Port
	back.AddrInfo = fmt.Sprintf("%s:%d", back.Addr, back.Port)
	back.SubCluster = subCluster
}

func (back *BfeBackend) GetAddr() string {
	return back.Addr
}

func (back *BfeBackend) GetAddrInfo() string {
	return back.AddrInfo
}

func (back *BfeBackend) Avail() bool {
	back.RLock()
	avail := back.avail
	back.RUnlock()

	return avail
}

func (back *BfeBackend) SetAvail(avail bool) {
	back.Lock()
	back.setAvail(avail)
	back.Unlock()
}

func (back *BfeBackend) setAvail(avail bool) {
	back.avail = avail
	if back.avail {
		back.connNum = 0
	}
}

func (back *BfeBackend) ConnNum() int {
	back.RLock()
	conns := back.connNum
	back.RUnlock()

	return conns
}

func (back *BfeBackend) AddConnNum() {
	back.Lock()
	back.connNum++
	back.Unlock()
}

func (back *BfeBackend) DecConnNum() {
	back.Lock()
	back.connNum--
	back.Unlock()
}

func (back *BfeBackend) AddFailNum() {
	back.Lock()
	back.failNum++
	back.Unlock()
}

func (back *BfeBackend) ResetFailNum() {
	back.Lock()
	back.failNum = 0
	back.Unlock()
}

func (back *BfeBackend) FailNum() int {
	back.RLock()
	failNum := back.failNum
	back.RUnlock()

	return failNum
}

func (back *BfeBackend) AddSuccNum() {
	back.Lock()
	back.succNum++
	back.Unlock()
}

func (back *BfeBackend) ResetSuccNum() {
	back.Lock()
	back.succNum = 0
	back.Unlock()
}

func (back *BfeBackend) SuccNum() int {
	back.RLock()
	succNum := back.succNum
	back.RUnlock()

	return succNum
}

func (back *BfeBackend) CheckAvail(succThreshould int) bool {
	back.Lock()
	defer back.Unlock()

	if back.succNum >= succThreshould {
		back.succNum = 0
		return true
	}

	return false
}

func (back *BfeBackend) UpdateStatus(failThreShould int) bool {
	back.Lock()
	defer back.Unlock()

	prevStatus := back.avail

	if back.failNum >= failThreShould {
		back.setAvail(false)
		if prevStatus {
			return true
		}
	}

	return false
}

func (back *BfeBackend) Release() {
	back.Close()
}

func (back *BfeBackend) Close() {
	close(back.closeChan)
}

func (back *BfeBackend) CloseChan() <-chan bool {
	return back.closeChan
}

func (back *BfeBackend) OnSuccess() {
	back.ResetFailNum()
}

func (back *BfeBackend) OnFail(cluster string) {
	back.AddFailNum()
	UpdateStatus(back, cluster)
}
