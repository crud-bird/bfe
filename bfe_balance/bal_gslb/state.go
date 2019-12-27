package bal_gslb

type SubClusterState struct {
	BackendNum int
}

type GslbState struct {
	SubClusters map[string]*SubClusterState
	BackendNum  int
}

func State(bal *BalanceGslb) *GslbState {
	gslbState := new(GslbState)
	gslbState.SubClusters = make(map[string]*SubClusterState)
	bal.lock.Lock()
	for _, sub := range bal.subClusters {
		subState := &SubClusterState{
			BackendNum: sub.Len(),
		}

		gslbState.SubClusters[sub.Name] = subState
		gslbState.BackendNum += subState.BackendNum
	}

	bal.lock.Unlock()

	return gslbState
}
