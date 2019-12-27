package bal_gslb

import (
	"encoding/binary"
	"fmt"
	"github.com/baidu/go-lib/web-monitor/metrics"
	bal_backend "github.com/crud-bird/bfe/bfe_balance/backend"
	"github.com/crud-bird/bfe/bfe_balance/bal_slb"
	"github.com/crud-bird/bfe/bfe_basic"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_table_conf"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/gslb_conf"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	DefaultRetryMax      = 3
	defaultCrossRetrymax = 1
)

type BalanceGslb struct {
	lock sync.Mutex

	name        string
	subClusters SubClusterList

	totalWeight int
	single      bool
	avail       int

	retryMax    int
	crossRetry  int
	hashConf    cluster_conf.HashConf
	BalanceMode string
}

func NewBalanceGslb(name string) *BalanceGslb {
	strategy := cluster_conf.ClientIPOnly
	sticky := false
	return &BalanceGslb{
		name:       name,
		retryMax:   DefaultRetryMax,
		crossRetry: defaultCrossRetrymax,
		hashConf: cluster_conf.HashConf{
			HashStrategy:  &strategy,
			SessionSticky: &sticky,
		},
		BalanceMode: cluster_conf.BalanceModeWrr,
	}
}

func (bal *BalanceGslb) SetGslbBasic(basic cluster_conf.GslbBasicConf) {
	bal.lock.Lock()

	bal.crossRetry = *basic.CrossRetry
	bal.retryMax = *basic.RetryMax
	bal.hashConf = *basic.HashConf
	bal.BalanceMode = *basic.BalanceMode

	bal.lock.Unlock()
}

func (bal *BalanceGslb) Init(conf gslb_conf.GslbClusterConf) error {
	total := 0
	for subName, weight := range conf {
		subCluster := newSubCluster(subName)
		subCluster.weight = weight

		if weight > 0 {
			total += weight
		}

		bal.subClusters = append(bal.subClusters, subCluster)
	}

	if total == 0 {
		return fmt.Errorf("gslb[%s] total weight = 0", bal.name)
	}

	bal.totalWeight = total

	sort.Sort(SubClusterListSortor{bal.subClusters})
	availNum := 0
	for i, sub := range bal.subClusters {
		if sub.weight > 0 {
			bal.avail = i
			availNum++
		}
	}
	bal.single = (availNum == 1)

	return nil
}

func (bal *BalanceGslb) BackendInit(cBackend cluster_table_conf.ClusterBackend) error {
	bal.lock.Lock()
	for _, subCluster := range bal.subClusters {
		if backend, ok := cBackend[subCluster.Name]; ok {
			subCluster.init(backend)
		}
	}

	bal.lock.Unlock()
	return nil
}

func (bal *BalanceGslb) Reload(conf gslb_conf.GslbClusterConf) error {
	bal.lock.Lock()
	defer bal.lock.Unlock()

	var newList SubClusterList
	subExist := make(map[string]bool)

	for i := 0; i < len(bal.subClusters); i++ {
		sub := bal.subClusters[i]
		if w, ok := conf[sub.Name]; ok {
			sub.weight = w
			newList = append(newList, sub)
		} else {
			sub.release()
		}

		subExist[sub.Name] = true
	}

	for name, w := range conf {
		if _, ok := subExist[name]; !ok {
			sub := newSubCluster(name)
			sub.weight = w
			newList = append(newList, sub)
		}
	}

	sort.Sort(SubClusterListSortor{newList})

	totalWeight := 0
	availNum := 0
	lastAvail := 0

	for i, sub := range newList {
		if sub.weight > 0 {
			totalWeight += sub.weight
			availNum++
			lastAvail = i
		}
	}

	if totalWeight == 0 {
		return fmt.Errorf("gslb[%s] total weight = 0", bal.name)
	}

	bal.totalWeight = totalWeight

	if availNum == 1 {
		bal.single = true
		bal.avail = lastAvail
	} else {
		bal.single = false
	}

	bal.subClusters = newList

	return nil
}

func (bal *BalanceGslb) BackendReload(clusterBackend cluster_table_conf.ClusterBackend) error {
	bal.lock.Lock()
	for _, subCluster := range bal.subClusters {
		if backend, ok := clusterBackend[subCluster.Name]; ok {
			subCluster.update(backend)
		}
	}
	bal.lock.Unlock()

	return nil
}

func (bal *BalanceGslb) Release() {
	bal.lock.Lock()
	for i := 0; i < len(bal.subClusters); i++ {
		bal.subClusters[i].release()
	}

	bal.lock.Unlock()
}

func (bal *BalanceGslb) getHashKey(req *bfe_basic.Request) []byte {
	var clientIP net.IP
	var hashKey []byte

	if req.ClientAddr != nil {
		clientIP = req.ClientAddr.IP
	} else {
		clientIP = nil
	}

	switch *bal.hashConf.HashStrategy {
	case cluster_conf.ClientIDOnly:
		hashKey = getHashKeyByHeader(req, *bal.hashConf.HashHeader)
	case cluster_conf.ClientIPOnly:
		hashKey = clientIP
	case cluster_conf.ClientIDPreferred:
		hashKey = getHashKeyByHeader(req, *bal.hashConf.HashHeader)
		if hashKey == nil {
			hashKey = clientIP
		}
	}

	if len(hashKey) == 0 {
		hashKey = make([]byte, 8)
		binary.BigEndian.PutUint64(hashKey, rand.Uint64())
	}

	return hashKey
}

func getHashKeyByHeader(req *bfe_basic.Request, header string) []byte {
	if val := req.HttpRequest.Header.Get(header); len(val) > 0 {
		return []byte(val)
	}

	if key, ok := cluster_conf.GetCookieKey(header); ok {
		if cookie, ok := req.Cookie(key); ok {
			return []byte(cookie.Value)
		}
	}

	return nil
}

func (bal *BalanceGslb) Balance(req *bfe_basic.Request) (*bal_backend.BfeBackend, error) {
	var backend *bal_backend.BfeBackend
	var current *SubCluster
	var err error
	var balAlgor int

	bal.lock.Lock()
	defer bal.lock.Unlock()

	if req.RetryTime > (bal.retryMax + bal.crossRetry) {
		state.ErrBkRetryTooMany.Inc(1)
		return nil, bfe_basic.ErrBkRetryTooMany
	}

	switch bal.BalanceMode {
	case cluster_conf.BalanceModeWlc:
		balAlgor = bal_slb.WlcSmooth
	default:
		balAlgor = bal_slb.WrrSmooth
	}

	if *bal.hashConf.SessionSticky {
		balAlgor = bal_slb.WrrSticky
	}

	hashKey := bal.getHashKey(req)

	current, err = bal.subClustersBalance(hashKey)
	if err != nil {
		state.ErrBkNoSubCluster.Inc(1)
		req.ErrCode = bfe_basic.ErrBkNoSubCluster
		return nil, bfe_basic.ErrBkNoSubCluster
	}
	req.Backend.SubclusterName = current.Name

	if req.RetryTime <= bal.retryMax {
		backend, err = current.balance(balAlgor, hashKey)
		if err == nil {
			return backend, nil
		} else {
			state.ErrBkNoBackend.Inc(1)
			req.ErrMsg = fmt.Sprintf("cluster[%s], sub[%s], err[%s]", bal.name, current.Name, err.Error())
			req.RetryTime = bal.retryMax
		}
	}

	if bal.crossRetry <= 0 {
		req.ErrCode = bfe_basic.ErrBkNoBackend
		return nil, bfe_basic.ErrBkNoBackend
	}

	if req.Stat != nil {
		req.Stat.IsCrossCLuster = true
	}

	current, err = bal.randomSelectExclude(current)
	if err != nil {
		state.ErrBkNoSubClusterCross.Inc(1)
		req.ErrCode = bfe_basic.ErrBkNoSubClusterCross
		return nil, bfe_basic.ErrBkNoSubClusterCross
	}
	req.Backend.SubclusterName = current.Name

	backend, err = current.balance(balAlgor, hashKey)
	if err == nil {
		return backend, nil
	}

	state.ErrBkNoBackend.Inc(1)
	req.ErrCode = bfe_basic.ErrBkNoBackend

	return backend, bfe_basic.ErrBkCrossRetryBalance
}

func (bal *BalanceGslb) subClustersBalance(value []byte) (*SubCluster, error) {
	var subCluster *SubCluster
	var w int

	if bal == nil {
		return subCluster, fmt.Errorf("gslb is nil")
	}

	if bal.totalWeight == 0 {
		return subCluster, fmt.Errorf("totalWeight is 0")
	}

	if bal.single {
		return bal.subClusters[bal.avail], nil
	}

	w = bal_slb.GetHash(value, uint(bal.totalWeight))
	for i := 0; i < len(bal.subClusters); i++ {
		subCluster = bal.subClusters[i]
		if subCluster.weight <= 0 {
			continue
		}
		w -= subCluster.weight
		if w < 0 {
			break
		}
	}

	return subCluster, nil
}

func (bal *BalanceGslb) randomSelectExclude(excludeCluster *SubCluster) (*SubCluster, error) {
	var subCluster *SubCluster
	available := 0

	for i := 0; i < len(bal.subClusters); i++ {
		subCluster = bal.subClusters[i]
		if subCluster != excludeCluster && subCluster.weight >= 0 && subCluster.sType != TypeGslbBlackhole {
			available++
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := int(r.Int31()) % available
	for i := 0; i < len(bal.subClusters); i++ {
		subCluster = bal.subClusters[i]
		if subCluster != excludeCluster && subCluster.weight >= 0 && subCluster.sType != TypeGslbBlackhole {
			if n > 0 {
				n--
				continue
			}
			return subCluster, nil
		}
	}

	return subCluster, fmt.Errorf("randomSelectExclude(): should not reach here")
}

func (bal *BalanceGslb) SubClusterNum() int {
	return len(bal.subClusters)
}

type BalErrState struct {
	ErrBkNoSubCluster      *metrics.Counter
	ErrBkNoSubClusterCross *metrics.Counter
	ErrBkNoBackend         *metrics.Counter
	ErrBkRetryTooMany      *metrics.Counter
	ErrGslbBlackhole       *metrics.Counter
}

var state BalErrState

func GetBalErrState() *BalErrState {
	return &state
}
