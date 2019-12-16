package bfe_server

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_config/bfe_conf"
	"github.com/crud-bird/bfe/bfe_modules"
	"github.com/sirupsen/logrus"
	"net"
)

func StartUp(cfg bfe_conf.BfeConfig, version string, confRoot string) error {
	lnMap, err := createListeners(cfg)
	if err != nil {
		logrus.Errorf("StartUp(): createListeners(): %s", err)
		return err
	}

	bfe_modules.SetModules()

	bfeServer := NewBfeServer(cfg, lnMap, version)

	if err = bfeServer.InitHttp(); err != nil {
		logrus.Errorf("StartUp(): InitHttp() %s", err)
		return err
	}

	if err = bfeServer.InitHttps(); err != nil {
		logrus.Errorf("StartUp(): InitHttps(): %s", err)
		return err
	}

	bfeServer.InitSignalTable()
	logrus.Info("StartUp(): bfeServer.InitSignalTable() OK")

	monitorPort := cfg.Server.MonitorPort
	if err = bfeServer.InitWebMonitor(monitorPort); err != nil {
		logrus.Errorf("StartUp(): InitWebMonitor(): %s", err)
		return err
	}

	if err = bfeServer.RegisterModules(cfg.Server.Modules); err != nil {
		logrus.Errorf("StartUp(): RegisterModules(): %s", err)
		return err
	}

	if err = bfeServer.InitModules(confRoot); err != nil {
		logrus.Errorf("StartUp(): BfeServer.InitModules(): %s", err)
		return err
	}

	if err = bfeServer.InitDataLoad(); err != nil {
		logrus.Errorf("StartUp(): bfeServer.InitDataLoad(): %s", err)
		return err
	}

	bfeServer.Monitor.Start()

	serveChan := make(chan error)

	go func() {
		httpErr := bfeServer.ServeHttp(bfeServer.HttpListener)
		serveChan <- httpErr
	}()

	go func() {
		httpsErr := bfeServer.ServeHttps(bfeServer.HttpsListener)
		serveChan <- httpsErr
	}()

	err = <-serveChan

	return nil
}

func createListeners(config bfe_conf.BfeConfig) (map[string]net.Listener, error) {
	lnMap := make(map[string]net.Listener)
	lnConf := map[string]int{
		"HTTP":  config.Server.HttpPort,
		"HTTPS": config.Server.HttpsPort,
	}

	for proto, port := range lnConf {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}

		listener = NewBfeListener(listener, config)
		lnMap[proto] = listener
		logrus.Info("Createlistener(): begin to listen port[%d]", port)
	}

	return lnMap, nil
}

func (p *BfeServer) closeListeners() {
	for _, ln := range p.listenerMap {
		if err := ln.Close(); err != nil {
			logrus.Errorf("closeListeners(): %s, %s", err, ln.Addr())
		}
	}
}
