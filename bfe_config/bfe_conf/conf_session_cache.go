package bfe_conf

import (
	"fmt"
	"strings"
)

type ConfigSessionCache struct {
	SessionCacheDisable bool

	Servers string

	KeyPrefix string

	ConnectTimeout int
	ReadTimeout    int
	WriteTimeout   int

	MaxIdle int

	SessionExpire int
}

func (cfg *ConfigSessionCache) SetDefaultConf() {
	cfg.SessionCacheDisable = true
	cfg.KeyPrefix = "bfe"
	cfg.ConnectTimeout = 50
	cfg.WriteTimeout = 50
	cfg.MaxIdle = 20
	cfg.SessionExpire = 3600
}

func (cfg *ConfigSessionCache) Check(confRoot string) error {
	return ConfSessionCacheCheck(cfg, confRoot)
}

func ConfSessionCacheCheck(cfg *ConfigSessionCache, confRoot string) error {
	names := strings.Split(cfg.Servers, ",")
	if len(cfg.Servers) == 0 || len(names) < 1 {
		return fmt.Errorf("Servers[%s] invalid server names", cfg.Servers)
	}

	if cfg.ReadTimeout <= 0 {
		return fmt.Errorf("ReadTimeout[%d] should be > 0", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout <= 0 {
		return fmt.Errorf("WriteTimeout[%d] should be > 0", cfg.WriteTimeout)
	}

	if cfg.MaxIdle <= 0 {
		return fmt.Errorf("MaxIdle[%s] should be > 0", cfg.MaxIdle)
	}

	if cfg.SessionExpire <= 0 {
		return fmt.Errorf("SessionExpire[%d] should be > 0", cfg.SessionExpire)
	}

	return nil
}
