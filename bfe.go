package main

import (
	"flag"
	"fmt"
	"github.com/crud-bird/bfe/bfe_config/bfe_conf"
	"github.com/crud-bird/bfe/bfe_util"
	"github.com/sirupsen/logrus"
	"path"
)

var (
	help     *bool   = flag.Bool("h", false, "to show help")
	confRoot *string = flag.String("c", "./conf", "root path of configuration")
	logPath  *string = flag.String("l", "./log", "dir path of log")
	stdOut   *bool   = flag.Bool("s", false, "to show log in stdout")
	showVer  *bool   = flag.Bool("v", false, "to show version of bfe")
	debugLog *bool   = flag.Bool("d", false, "to show debug log")
)

var version string

func main() {
	var (
		err       error
		config    bfe_conf.BfeConfig
		logSwitch string
	)

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	if *showVer {
		fmt.Printf("bfe: version %s\n", version)
		return
	}

	confPath := path.Join(*confRoot, "bfe.conf")
	if config, err = bfe_conf.BfeConfigLoad(confPath, *confRoot); err != nil {
		logrus.Errorf("main(): in BfeConfigLoad(): %s", err)
		bfe_util.AbnormalExit()
	}

}
