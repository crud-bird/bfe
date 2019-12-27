package bal_slb

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_balance/backend"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
	"github.com/spaolacci/murmur3"
	"math/rand"
	"sort"
	"sync"
)

const (
	WrrSimple = iota
	WrrSmooth
	WrrSticky
	WlcSimple
	WlcSmooth
)

type BackendList []*BackendRR

func (bl *BackendList) ResetWeight() {
	for _, rr := range *bl {
		rr.current = rr.weight
	}
}

type BackendListSortor struct {
	l BackendList
}

func (s BackendListSortor) Len() int {
	return len(s.l)
}

func (s BackendListSortor) Swap(i, j int) {
	s.l[i], s.l[j] = s.l[j], s.l[i]
}

func (s BackendListSortor) Less(i, j int) bool {
	return s.l[i].backend.AddrInfo < s.l[j].backend.AddrInfo
}

type BalanceRR struct {
	sync.Mutex
	Name     string
	backends BackendList
	sorted   bool
	next     int
}

func NewBalanceRR(name string) *BalanceRR {
	return &BalanceRR{Name: name}
}

func (brr *BalanceRR) Init(conf cluster_table_conf.SubClusterBackend) {
	for _, c := range conf {
		backendRR := NewBackendRR()
		backendRR.Init(brr.Name, c)
		brr.backends = append(brr.backends, backendRR)
	}

	brr.sorted = false
	brr.next = 0
}

func (brr *BalanceRR) Release() {
	for _, back := range brr.backends {
		back.Release()
	}
}

func confMapMake(conf cluster_table_conf.SubClusterBackend) map[string]*cluster_table_conf.BackendConf {
	retVal := make(map[string]*cluster_table_conf.BackendConf)

	for _, backend := range conf {
		retVal[backend.AddrInfo()] = backend
	}

	return retVal
}

func (brr *BalanceRR) Update(conf cluster_table_conf.SubClusterBackend) {
	var backendsNew BackendList
	confMap := confMapMake(conf)
	brr.Lock()
	defer brr.Unlock()

	for idx := 0; idx < len(brr.backends); idx++ {
		backendRR := brr.backends[idx]
		backendKey := backendRR.backend.GetAddrInfo()
		if bkConf, ok := confMap[backendKey]; ok && backendRR.MatchAddrPort(*bkConf.Addr, *bkConf.Port) {
			backendRR.UpdateWeight(*bkConf.Weight)
			backendsNew = append(backendsNew, backendRR)
			delete(confMap, backendKey)
		} else {
			backendRR.Release()
		}
	}

	for _, bkConf := range confMap {
		backendRR := NewBackendRR()
		backendRR.Init(brr.Name, bkConf)
		backendsNew = append(backendsNew, backendRR)
	}

	brr.backends = backendsNew
	brr.sorted = false
	brr.next = 0
}

func (brr *BalanceRR) initWeight() {
	brr.backends.ResetWeight()
}

func moveTONext(next int, backends BackendList) int {
	next += 1
	if next >= len(backends) {
		next = 0
	}

	return next
}

func (brr *BalanceRR) ensureSortedUnlocked() {
	if !brr.sorted {
		sort.Sort(BackendListSortor{brr.backends})
		brr.sorted = true
	}
}

func (brr *BalanceRR) Balance(algor int, key []byte) (*backend.BfeBackend, error) {
	switch algor {
	case WrrSimple:
		return brr.simpleBalance()
	case WrrSmooth:
		return brr.smoothBalance()
	case WrrSticky:
		return brr.stickyBalance(key)
	case WlcSimple:
		return brr.leastConnsSimpleBalance()
	case WlcSmooth:
		return brr.leastConnsSmoothBalance()
	default:
		return brr.smoothBalance()
	}
}

func (brr *BalanceRR) smoothBalance() (*backend.BfeBackend, error) {
	brr.Lock()
	defer brr.Unlock()

	return smoothBalance(brr.backends)
}

func smoothBalance(backs BackendList) (*backend.BfeBackend, error) {
	var best *BackendRR
	total, max := 0, 0
	for _, backendRR := range backs {
		backend := backendRR.backend
		if !backend.Avail() || backendRR.weight <= 0 {
			continue
		}

		if best == nil || backendRR.current > max {
			best = backendRR
			max = backendRR.current
		}
		total += backendRR.current

		backendRR.current += backendRR.weight
	}

	if best == nil {
		return nil, fmt.Errorf("rr_bal: all backend is down")
	}

	best.current -= total

	return best.backend, nil
}

func (brr *BalanceRR) leastConnsSmoothBalance() (*backend.BfeBackend, error) {
	brr.Lock()
	defer brr.Unlock()

	candidates, err := leastConnsBalance(brr.backends)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 1 {
		return candidates[0].backend, nil
	}

	return smoothBalance(candidates)
}

func (brr *BalanceRR) leastConnsSimpleBalance() (*backend.BfeBackend, error) {
	brr.Lock()
	defer brr.Unlock()

	candidates, err := leastConnsBalance(brr.backends)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 1 {
		return candidates[0].backend, nil
	}

	return randomBalance(candidates)
}

func leastConnsBalance(backs BackendList) (BackendList, error) {
	var best *BackendRR
	candidates := make(BackendList, 0, len(backs))

	single := true
	for _, backendRR := range backs {
		if backendRR.backend.Avail() || backendRR.weight <= 0 {
			continue
		}

		if best == nil {
			best = backendRR
			single = true
			continue
		}

		if ret := compLCWeight(best, backendRR); ret > 0 {
			best = backendRR
			single = true
		} else if ret == 0 {
			single = false
		}
	}

	if best == nil {
		return nil, fmt.Errorf("rr_bal: all backend is down")
	}

	if single {
		candidates = append(candidates, best)
		return candidates, nil
	}

	for _, backendRR := range backs {
		if !backendRR.backend.Avail() || backendRR.weight <= 0 {
			continue
		}

		if ret := compLCWeight(best, backendRR); ret == 0 {
			candidates = append(candidates, backendRR)
		}
	}

	return candidates, nil
}

func randomBalance(backs BackendList) (*backend.BfeBackend, error) {
	return backs[rand.Int()%len(backs)].backend, nil
}

func (brr *BalanceRR) simpleBalance() (*backend.BfeBackend, error) {
	var backend *backend.BfeBackend
	var backendRR *BackendRR

	brr.Lock()
	defer brr.Unlock()

	backends := brr.backends
	allDown := true

	next := brr.next
	for {
		backendRR = backends[next]
		backend = backendRR.backend

		avail := backend.Avail()
		if avail && backendRR.current > 0 {
			break
		}

		if avail && backendRR.weight != 0 {
			allDown = false
		}

		next = moveTONext(next, backends)

		if next == brr.next {
			if allDown {
				return backend, fmt.Errorf("rr_bal: all backend are down")
			}

			brr.initWeight()
			brr.next = 0
			next = 0
		}
	}

	backendRR.current--
	brr.next = moveTONext(next, backends)

	return backend, nil
}

func (brr *BalanceRR) stickyBalance(key []byte) (*backend.BfeBackend, error) {
	candidates := make(BackendList, 0, brr.Len())
	total := 0

	brr.Lock()
	defer brr.Unlock()

	brr.ensureSortedUnlocked()
	for _, backendRR := range brr.backends {
		if backendRR.backend.Avail() && backendRR.weight > 0 {
			candidates = append(candidates, backendRR)
			total += backendRR.weight
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("rr_bal: all backends are down")
	}

	value := GetHash(key, uint(total))
	for _, backendRR := range candidates {
		value -= backendRR.weight
		if value < 0 {
			return backendRR.backend, nil
		}
	}

	return nil, fmt.Errorf("rr_bal: stickyBalance fail")
}

func compLCWeight(a, b *BackendRR) int {
	ret := a.backend.ConnNum()*b.weight - b.backend.ConnNum()*a.weight

	if ret > 0 {
		return 1
	}

	if ret == 0 {
		return 0
	}

	return -1
}

func (brr *BalanceRR) Len() int {
	return len(brr.backends)
}

func GetHash(value []byte, base uint) int {
	var hash uint64
	if value == nil {
		hash = uint64(rand.Uint32())
	} else {
		hash = murmur3.Sum64(value)
	}

	return int(hash % uint64(base))
}
