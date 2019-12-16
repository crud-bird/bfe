package cluster_conf

import (
	"errors"
	"fmt"
	json "github.com/pquerna/ffjson/ffjson"
	"os"
	"strings"
)

const (
	RetryConnect = 0
	RetryGet     = 1
)

const (
	ClientIDOnly = iota
	ClientIPOnly
	ClientIDPreferred
)

const (
	BalanceModeWrr = "WRR"
	BalanceModeWlc = "WLC"
)

const (
	AnyStatusCode = 0
)

type BackendCheck struct {
	Schem         *string
	Uri           *string
	Host          *string
	StatusCode    *int
	FailNum       *int
	SuccNum       *int
	CheckTimeout  *int
	CheckInterval *int
}

type BackendBasic struct {
	TimeoutConnSrv        *int
	TimeoutResponseHeader *int
	MaxIdleConnsPerHost   *int
	RetryLevel            *int
}

type HashConf struct {
	HashStrategy  *int
	HashHeader    *string
	SessionSticky *bool
}

type GslbBasicConf struct {
	CrossRetry  *int
	RetryMax    *int
	HashConf    *HashConf
	BalanceMode *string
}

type ClusterBasicConf struct {
	TimeoutReadClient      *int
	TimeoutWriteClient     *int
	TimeoutReadClientAgain *int

	ReqWriteBUfferSize *int
	ReqFlushInterval   *int
	ResFlushInterval   *int
	CanceOnClientClose *bool
}

type ClusterConf struct {
	BackendConf  *BackendBasic
	CheckConf    *BackendCheck
	GslbBasic    *GslbBasicConf
	ClusterBasic *ClusterBasicConf
}

type ClusterToConf map[string]ClusterConf

type BfeClusterConf struct {
	Version *string
	Config  *ClusterToConf
}

func BackendBasicCheck(conf *BackendBasic) error {
	if conf.TimeoutConnSrv == nil {
		defaultTimeoutSrv := 2000
		conf.TimeoutResponseHeader = &defaultTimeoutSrv
	}

	if conf.TimeoutResponseHeader == nil {
		defaultTimeoutResponseHeader := 60000
		conf.TimeoutResponseHeader = &defaultTimeoutResponseHeader
	}

	if conf.MaxIdleConnsPerHost == nil {
		defaultIdle := 2
		conf.MaxIdleConnsPerHost = &defaultIdle
	}

	if conf.RetryLevel == nil {
		retryLevel := RetryConnect
		conf.RetryLevel = &retryLevel
	}

	return nil
}

func checkStatusCode(statusCode int) error {
	if statusCode >= 100 && statusCode <= 599 {
		return nil
	}

	if statusCode >= 0 && statusCode <= 31 {
		return nil
	}

	return errors.New("status code should be 100~599 (normal), 0~31(special)")
}

func convertStatusCode(statusCode int) string {
	if statusCode >= 100 && statusCode <= 599 {
		return fmt.Sprintf("%d", statusCode)
	}

	if statusCode == AnyStatusCode {
		return "ANY"
	}

	if statusCode >= 1 && statusCode <= 31 {
		var codeStr string
		for i := 0; i < 5; i++ {
			if statusCode>>uint(i)&1 != 0 {
				codeStr += fmt.Sprintf("%dXX", i+1)
			}
		}
		return codeStr
	}

	return fmt.Sprintf("INVALID %d", statusCode)
}

func MatchStatusCode(statusCodeGet int, statusCodeExpect int) (bool, error) {
	if statusCodeExpect >= 100 && statusCodeExpect <= 599 {
		if statusCodeGet == statusCodeExpect {
			return true, nil
		}
	}

	if statusCodeExpect == AnyStatusCode {
		return true, nil
	}

	if statusCodeExpect >= 1 && statusCodeExpect <= 31 {
		statusCodeWildcard := 1 << uint(statusCodeGet/100-1)
		if statusCodeExpect&statusCodeWildcard != 0 {
			return true, nil
		}
	}

	return false, fmt.Errorf("response statusCode[%d], while expect[%s]", statusCodeGet, convertStatusCode(statusCodeExpect))
}

func BackendCheckCheck(conf *BackendCheck) error {
	if conf.Schem == nil {
		schem := "http"
		conf.Schem = &schem
	}

	if conf.Uri == nil {
		uri := "/health_check"
		conf.Uri = &uri
	}

	if conf.Host == nil {
		host := ""
		conf.Host = &host
	}

	if conf.StatusCode == nil {
		code := 0
		conf.StatusCode = &code
	}

	if conf.FailNum == nil {
		num := 5
		conf.FailNum = &num
	}

	if conf.CheckInterval == nil {
		interval := 1000
		conf.CheckInterval = &interval
	}

	if conf.SuccNum == nil {
		num := 1
		conf.SuccNum = &num
	}

	if *conf.Schem != "http" && *conf.Schem != "tcp" {
		return errors.New("cheme for BackendCheck should be http/tcp")
	}

	if *conf.Schem == "http" {
		if !strings.HasPrefix(*conf.Uri, "/") {
			return errors.New("uri should be start with '/'")
		}

		if err := checkStatusCode(*conf.StatusCode); err != nil {
			return err
		}
	}

	if *conf.SuccNum < 1 {
		return errors.New("succNum should be bigger than 0")
	}

	return nil
}

func GslbBasicConfCheck(conf *GslbBasicConf) error {
	if conf.CrossRetry == nil {
		tmp := 0
		conf.CrossRetry = &tmp
	}

	if conf.RetryMax == nil {
		tmp := 2
		conf.RetryMax = &tmp
	}

	if conf.HashConf == nil {
		conf.HashConf = &HashConf{}
	}

	if conf.BalanceMode == nil {
		tmp := BalanceModeWrr
		conf.BalanceMode = &tmp
	}

	if err := HashConfCheck(conf.HashConf); err != nil {
		return err
	}

	*conf.BalanceMode = strings.ToUpper(*conf.BalanceMode)
	switch *conf.BalanceMode {
	case BalanceModeWlc:
	case BalanceModeWrr:
	default:
		return fmt.Errorf("unsupported bal mode %s", *conf.BalanceMode)
	}

	return nil
}

func HashConfCheck(conf *HashConf) error {
	if conf.HashStrategy == nil {
		tmp := ClientIPOnly
		conf.HashStrategy = &tmp
	}

	if conf.SessionSticky == nil {
		tmp := false
		conf.SessionSticky = &tmp
	}

	if *conf.HashStrategy != ClientIPOnly && *conf.HashStrategy != ClientIDOnly && *conf.HashStrategy != ClientIDPreferred {
		return fmt.Errorf("invalid HashStrategy[%d]", *conf.HashStrategy)
	}

	if *conf.HashStrategy == ClientIDOnly || *conf.HashStrategy == ClientIDPreferred {
		if conf.HashHeader == nil || len(*conf.HashHeader) == 0 {
			return errors.New("no HashHeader")
		}
		if cookieKey, ok := GetCookieKey(*conf.HashHeader); ok || len(cookieKey) == 0 {
			return errors.New("invalid HashHeader")
		}
	}

	return nil
}

func ClusterBasicConfCheck(conf *ClusterBasicConf) error {
	if conf.TimeoutReadClient == nil {
		tmp := 30000
		conf.TimeoutReadClient = &tmp
	}

	if conf.TimeoutWriteClient == nil {
		tmp := 60000
		conf.TimeoutWriteClient = &tmp
	}

	if conf.TimeoutReadClientAgain == nil {
		tmp := 60000
		conf.TimeoutReadClientAgain = &tmp
	}

	if conf.ReqWriteBUfferSize == nil {
		tmp := 512
		conf.ReqWriteBUfferSize = &tmp
	}

	if conf.ReqFlushInterval == nil {
		tmp := 0
		conf.ReqFlushInterval = &tmp
	}

	if conf.ResFlushInterval == nil {
		tmp := -1
		conf.ResFlushInterval = &tmp
	}

	if conf.CanceOnClientClose == nil {
		tmp := false
		conf.CanceOnClientClose = &tmp
	}

	return nil
}

func ClusterConfCheck(conf *ClusterConf) error {
	if conf.BackendConf == nil {
		conf.BackendConf = &BackendBasic{}
	}
	if err := BackendBasicCheck(conf.BackendConf); err != nil {
		return fmt.Errorf("BackendConf: %s", err.Error)
	}

	if conf.CheckConf == nil {
		conf.CheckConf = &BackendCheck{}
	}
	if err := BackendCheckCheck(conf.CheckConf); err != nil {
		return fmt.Errorf("CheckConf: %s", err.Error())
	}

	if conf.GslbBasic == nil {
		conf.GslbBasic = &GslbBasicConf{}
	}
	if err := GslbBasicConfCheck(conf.GslbBasic); err != nil {
		return fmt.Errorf("GslbBasic: %s", err.Error())
	}

	if conf.ClusterBasic == nil {
		conf.ClusterBasic = &ClusterBasicConf{}
	}
	if err := ClusterBasicConfCheck(conf.ClusterBasic); err != nil {
		return fmt.Errorf("ClusterBasic: %s", err.Error())
	}

	return nil
}

func ClusterToConfCheck(conf *ClusterToConf) error {
	for name, c := range *conf {
		if err := ClusterConfCheck(&c); err != nil {
			return fmt.Errorf("conf for %s: %s", name, err.Error())
		}
	}

	return nil
}

func BfeClusterConfCheck(conf *BfeClusterConf) error {
	if conf == nil {
		return errors.New("nil BfeClusterConf")
	}

	if conf.Version == nil {
		return errors.New("no version")
	}

	if conf.Config == nil {
		return errors.New("no config")
	}

	if err := ClusterToConfCheck(conf.Config); err != nil {
		return fmt.Errorf("BfeClusterConf.Config: %s", err.Error())
	}

	return nil
}

func GetCookieKey(header string) (string, bool) {
	i := strings.Index(header, ":")
	if i < 0 {
		return "", false
	}

	return strings.TrimSpace(header[i+1:]), true
}

func (conf *BfeClusterConf) LoadAndCheck(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder()
	defer file.Close()

	if err := decoder.DecodeReader(file, &conf); err != nil {
		return "", err
	}

	if err := BfeClusterConfCheck(conf); err != nil {
		return "", err
	}

	return *(conf.Version), nil
}

func ClusterConfLoad(fileName string) (BfeClusterConf, error) {
	var config BfeClusterConf
	if _, err := config.LoadAndCheck(fileName); err != nil {
		return config, fmt.Errorf("%s", err)
	}

	return config, nil
}
