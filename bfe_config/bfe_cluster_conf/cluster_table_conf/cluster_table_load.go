package cluster_table_conf

type BackendConf struct {
	Name   *string
	Addr   *string
	Port   *int
	Weight *int
}
