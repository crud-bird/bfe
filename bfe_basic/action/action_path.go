package action

import (
	"github.com/crud-bird/bfe/bfe_basic"
	"strings"
)

func ReqPathSet(req *bfe_basic.Request, path string) {
	req.HttpRequest.URL.Path = path
}

func ReqPathPrefixAdd(req *bfe_basic.Request, prefix string) {
	httpReq := req.HttpRequest
	pathStr := httpReq.URL.Path
	pathStr = strings.TrimPrefix(pathStr, "/")
	pathStr = prefix + pathStr

	if !strings.HasPrefix(pathStr, "/") {
		pathStr = "/" + pathStr
	}

	httpReq.URL.Path = pathStr
}

func ReqPathPrefixTrim(req *bfe_basic.Request, prefix string) {
	httpReq := req.HttpRequest
	pathStr := httpReq.URL.Path
	pathStr = strings.TrimPrefix(pathStr, prefix)
	if !strings.HasPrefix(pathStr, "/") {
		pathStr = "/" + pathStr
	}
	httpReq.URL.Path = pathStr
}
