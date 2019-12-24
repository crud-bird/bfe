package action

import (
	"github.com/crud-bird/bfe/bfe_basic"
	"strings"
)

func ReqHostSet(req *bfe_basic.Request, hostName string) {
	req.HttpRequest.Host = hostName
}

func ReqHostSetFromPathPrefix(req *bfe_basic.Request) {
	path := req.HttpRequest.URL.Path

	segs := strings.SplitN(path, "/", 3)
	if len(segs) < 3 {
		return
	}

	req.HttpRequest.Host = segs[1]
	req.HttpRequest.URL.Path = "/" + segs[2]
}
