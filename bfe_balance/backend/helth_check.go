package backend

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/crud-bird/bfe/bfe_debug"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"time"
)

func UpdateStatus(backend *BfeBackend, cluster string) bool {
	checkConf := getCheckConf(cluster)
	if checkConf == nil {
		return false
	}

	if backend.UpdateStatus(*checkConf.FailNum) {
		go check(backend, cluster)
		return true
	}

	return false
}

func check(backend *BfeBackend, cluster string) {
	logrus.Infof("start health check for %s", backend.Name)

	c := backend.CloseChan()

loop:
	for {
		select {
		case <-c:
			break loop
		default:
		}

		checkConf := getCheckConf(cluster)
		if checkConf == nil {
			time.Sleep(time.Second)
			continue
		}
		checkInterval := time.Duration(*checkConf.CheckInterval) * time.Millisecond

		if ok, err := CheckConnect(backend, checkConf); !ok {
			backend.ResetSuccNum()
			if bfe_debug.DebugHealthCheck {
				logrus.Debugf("backend %s still not avail (check failure: %s)", backend.Name, err)
			}
			time.Sleep(checkInterval)
			continue
		}

		backend.AddSuccNum()
		if !backend.CheckAvail(*checkConf.SuccNum) {
			if bfe_debug.DebugHealthCheck {
				logrus.Debugf("backend %s still not avail(check success, waiting for more checkes)", backend.Name)
			}
			time.Sleep(checkInterval)
			continue
		}

		logrus.Infof("backend %s back to Normal", backend.Name)
		backend.SetAvail(true)
		break loop
	}
}

func getHealthCheckAddrInfo(backend *BfeBackend, checkConf *cluster_conf.BackendCheck) string {
	if checkConf.Host != nil {
		hostInfo := strings.Split(*checkConf.Host, ":")
		if len(hostInfo) == 2 {
			return fmt.Sprintf("%s:%s", backend.GetAddr(), hostInfo[1])
		}

	}
	return backend.GetAddr()
}

func checkTCPConnect(backend *BfeBackend, checkConf *cluster_conf.BackendCheck) (bool, error) {
	addrInfo := getHealthCheckAddrInfo(backend, checkConf)

	var conn net.Conn
	var err error
	if checkConf.CheckTimeout != nil {
		conn, err = net.DialTimeout("tcp", addrInfo, time.Duration(*checkConf.CheckTimeout)*time.Millisecond)
	} else {
		conn, err = net.Dial("tcp", addrInfo)
	}

	if err != nil {
		return false, err
	}

	conn.Close()
	return true, nil
}

func doHTTPHealthCheck(req *http.Request, timeout time.Duration) (int, error) {
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Timeout:       timeout,
	}

	resp, err := client.Do((req))
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func checkHTTPConnect(backend *BfeBackend, checkConf *cluster_conf.BackendCheck) (bool, error) {
	addrInfo := getHealthCheckAddrInfo(backend, checkConf)
	urlStr := fmt.Sprintf("http://%s%s", addrInfo, *checkConf.Uri)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return false, err
	}

	if checkConf.Host != nil {
		req.Host = *checkConf.Host
	}

	req.Header.Set("Accept", "*/*")

	checkTimeout := time.Duration(0)
	if checkConf.CheckTimeout != nil {
		checkTimeout = time.Duration(*checkConf.CheckTimeout) * time.Millisecond
	}

	statusCode, err := doHTTPHealthCheck(req, checkTimeout)
	if err != nil {
		return false, err
	}

	return cluster_conf.MatchStatusCode(statusCode, *checkConf.StatusCode)
}

func CheckConnect(backend *BfeBackend, checkConf *cluster_conf.BackendCheck) (bool, error) {
	switch *checkConf.Schem {
	case "http":
		return checkHTTPConnect(backend, checkConf)
	case "tcp":
		return checkTCPConnect(backend, checkConf)
	default:
		return checkHTTPConnect(backend, checkConf)
	}
}

type CheckConfFetcher func(string) *cluster_conf.BackendCheck

var checkconfFetcher CheckConfFetcher

func getCheckConf(cluster string) *cluster_conf.BackendCheck {
	if checkconfFetcher == nil {
		return nil
	}

	return checkconfFetcher(cluster)
}

func SetCheckConfFetcher(fetcher CheckConfFetcher) {
	checkconfFetcher = fetcher
}
