package bfe_util

import (
	"path"
	"strings"
)

func ConfPathProc(confPath string, confRoot string) string {
	if !strings.HasPrefix(confPath, "/") {
		confPath = path.Join(confRoot, confPath)
	}

	return confPath
}
