package bfe_conf

import (
	gcfg "gopkg.in/gcfg.v1"
)

type BfeConfig struct {
	Server ConfigBasic

	HttpsBasic ConfigHttpsBasic

	SessionCache ConfigSessionCache

	SessionTicket ConfigSessionTicket
}

func SetDefaultConf(conf *BfeConfig) {
	conf.Server.SetDeafaultConf()
	conf.HttpsBasic.SetDefaultConf()
	conf.SessionCache.SetDefaultConf()
	conf.SessionTicket.SetDefaultConf()
}

func BfeConfigLoad(filePath string, confRoot string) (BfeConfig, error) {
	var cfg BfeConfig

	SetDefaultConf(&cfg)

	if err := gcfg.ReadFileInto(&cfg, filePath); err != nil {
		return cfg, err
	}

	if err := cfg.Server.Check(confRoot); err != nil {
		return cfg, err
	}

	if err := cfg.HttpsBasic.Check(confRoot); err != nil {
		return cfg, err
	}

	if err := cfg.SessionCache.Check(confRoot); err != nil {
		return cfg, err
	}

	if err := cfg.SessionTicket.Check(confRoot); err != nil {
		return cfg, err
	}

	return cfg, nil
}
