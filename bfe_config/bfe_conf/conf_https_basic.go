package bfe_conf

import (
	"crypto/x509"
	"fmt"
	"github.com/crud-bird/bfe/bfe_tls"
	"github.com/crud-bird/bfe/bfe_util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

var TlsVersionMap = map[string]uint16{
	"VersionSSL30": bfe_tls.VersionSSL30,
	"VersionTLS10": bfe_tls.VersionTLS10,
	"VersionTLS11": bfe_tls.VersionTLS11,
	"VersionTLS12": bfe_tls.VersionTLS12,
}

var CurvesMap = map[string]bfe_tls.CurveID{
	"CurVeP256": bfe_tls.CurVeP256,
	"CurveP384": bfe_tls.CurveP384,
	"CurveP521": bfe_tls.CurveP521,
}

var CipherSuitesMap = map[string]uint16{
	"TLS_RSA_WITH_RC4_128_SHA":                      bfe_tls.TLS_RSA_WITH_RC4_128_SHA,
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA":                 bfe_tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA":                  bfe_tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"TLS_RSA_WITH_AES_256_CBC_SHA":                  bfe_tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":              bfe_tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":          bfe_tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":          bfe_tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_RC4_128_SHA":                bfe_tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":           bfe_tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":            bfe_tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":            bfe_tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":         bfe_tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":       bfe_tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":   bfe_tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": bfe_tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
}

const (
	EquivCipherSep = "|"
)

type ConfigHttpsBasic struct {
	ServerCertConf string
	TlsRuleConf    string

	CipherSuites     []string
	CurvePreferences []string

	MaxTlsVersion string
	MinTlsVersion string

	EnableSslv2ClientHello bool

	ClientCABaseDir string
}

func (cfg *ConfigHttpsBasic) SetDefaultConf() {
	cfg.ServerCertConf = "tls_conf/server_cert_conf.data"
	cfg.TlsRuleConf = "tls_conf/tls_rule_conf.data"

	cfg.CipherSuites = []string{
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256|TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256|TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256|TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		"TLS_ECDHE_RSA_WITH_RC4_128_SHA",
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		"TLS_RSA_WITH_RC4_128_SHA",
		"TLS_RSA_WITH_AES_128_CBC_SHA",
		"TLS_RSA_WITH_AES_256_CBC_SHA",
	}
	cfg.CurvePreferences = []string{
		"CurveP256",
	}

	cfg.EnableSslv2ClientHello = true

	cfg.ClientCABaseDir = "tls_conf/client_ca"
}

func (cfg *ConfigHttpsBasic) Check(confRoot string) error {
	if err := certConfCheck(cfg, confRoot); err != nil {
		return err
	}

	if err := certRuleCheck(cfg, confRoot); err != nil {
		return err
	}

	for _, cipherGroup := range cfg.CipherSuites {
		ciphers := strings.Split(cipherGroup, EquivCipherSep)
		for _, cipher := range ciphers {
			if _, ok := CipherSuitesMap[cipher]; !ok {
				return fmt.Errorf("cipher (%s) not support")
			}
		}
	}

	for _, curve := range cfg.CurvePreferences {
		if _, ok := CurvesMap[curve]; !ok {
			return fmt.Errorf("curve (%s) not support")
		}
	}

	if err := tlsVersionCheck(cfg); err != nil {
		return err
	}

	if len(cfg.ClientCABaseDir) == 0 {
		return fmt.Errorf("CLientCABaseDir empty")
	}

	return nil
}

func certConfCheck(cfg *ConfigHttpsBasic, confRoot string) error {
	if cfg.ServerCertConf == "" {
		logrus.Warn("ServerCertConf not set, use default value")
		cfg.ServerCertConf = "tls_conf/server_cert_conf.data"
	}
	cfg.ServerCertConf = bfe_util.ConfPathProc(cfg.ServerCertConf, confRoot)

	return nil
}

func certRuleCheck(cfg *ConfigHttpsBasic, confRoot string) error {
	if cfg.TlsRuleConf == "" {
		logrus.Warn("TlsRuleConf not set, use default value")
		cfg.TlsRuleConf = "tls_conf/tlsrule_conf.data"
	}
	cfg.TlsRuleConf = bfe_util.ConfPathProc(cfg.TlsRuleConf, confRoot)

	return nil
}

func tlsVersionCheck(cfg *ConfigHttpsBasic) error {
	if len(cfg.MaxTlsVersion) == 0 {
		cfg.MaxTlsVersion = "VersionTLS12"
	}

	if len(cfg.MinTlsVersion) == 0 {
		cfg.MinTlsVersion = "VersionSSl30"
	}

	minTlsVer, ok := TlsVersionMap[cfg.MinTlsVersion]
	if !ok {
		return fmt.Errorf("MinTlsVersion[%s] not support", cfg.MinTlsVersion)
	}

	maxTlsVer, ok := TlsVersionMap[cfg.MaxTlsVersion]
	if !ok {
		return fmt.Errorf("MaxTlsVersion[%s] not support", cfg.MaxTlsVersion)
	}

	if maxTlsVer < minTlsVer {
		return fmt.Errorf("MaxTlsVersion should be not less than MinTlsVersion")
	}

	return nil
}

//LoadClietCAFile ...
func LoadClientCAFile(path string) (*x509.CertPool, error) {
	roots := x509.NewCertPool()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	roots.AppendCertsFromPEM(data)
	return roots, nil
}

//GetCurvePreferences ...
func GetCurvePreferences(curveConf []string) ([]bfe_tls.CurveID, error) {
	curvePreferences := make([]bfe_tls.CurveID, 0, len(curveConf))
	for _, curveStr := range curveConf {
		curve, ok := CurvesMap[curveStr]
		if !ok {
			return nil, fmt.Errorf("ellptic curve (%s) not support", curveStr)
		}
		curvePreferences = append(curvePreferences, curve)
	}

	return curvePreferences, nil
}

//GetTlsVersion ...
func GetTlsVersion(cfg *ConfigHttpsBasic) (maxVer, minVer uint16) {
	maxTlsVersion, ok := TlsVersionMap[cfg.MaxTlsVersion]
	if !ok {
		maxTlsVersion = bfe_tls.VersionTLS12
	}

	minTLsVersion, ok := TlsVersionMap[cfg.MinTlsVersion]
	if !ok {
		minTLsVersion = bfe_tls.VersionSSL30
	}

	return maxTlsVersion, minTLsVersion
}
