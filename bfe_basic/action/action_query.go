package action

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_basic"
	"net/url"
	"strings"
)

func queryParse(req *bfe_basic.Request) url.Values {
	if req.Query == nil {
		req.Query = req.HttpRequest.URL.Query()
	}

	return req.Query
}

func queryDump(queries url.Values) string {
	strs := make([]string, 0)

	for key, values := range queries {
		for _, value := range values {
			str := fmt.Sprintf("%s=%s", key, value)
			strs = append(strs, str)
		}
	}

	return strings.Join(strs, "&")
}

func ReqQueryAdd(req *bfe_basic.Request, params []string) {
	var addQueryString string
	queries := queryParse(req)
	pairNum := len(params)

	for i := 0; i < pairNum; i++ {
		key := params[2*i]
		value := params[2*i+1]
		oldValue := queries.Get(key)

		if oldValue == "" {
			queries.Set(key, value)
		} else {
			queries.Add(key, value)
		}

		addQueryString += "&" + key + "=" + value
	}

	if req.HttpRequest.URL.RawQuery == "" {
		req.HttpRequest.URL.RawQuery = addQueryString[1:]
	} else {
		req.HttpRequest.URL.RawQuery = req.HttpRequest.URL.RawQuery + addQueryString
	}
}

func ReqQueryRename(req *bfe_basic.Request, oldName, newName string) {
	var values []string
	var ok bool
	rawQuery := "&" + req.HttpRequest.URL.RawQuery
	queries := queryParse(req)

	if values, ok = queries[oldName]; !ok {
		return
	}

	queries.Del(oldName)
	queries[newName] = values

	srcKey := "&" + oldName + "="
	dstKey := "&" + newName + "="
	rawQuery = strings.Replace(rawQuery, srcKey, dstKey, -1)

	req.HttpRequest.URL.RawQuery = rawQuery[1:]
}

func ReqQueryDel(req *bfe_basic.Request, keys []string) {
	rawQuery := "&" + req.HttpRequest.URL.RawQuery + "&"
	queries := queryParse(req)

	for _, key := range keys {
		queries.Del(key)
		for {
			start := strings.Index(rawQuery, "&"+key+"=")
			if start == -1 {
				break
			}

			end := strings.Index(rawQuery[start+1:], "&")
			if end == -1 {
				break
			}
			rawQuery = rawQuery[:start] + rawQuery[start+end+1:]
		}
	}

	if len(rawQuery) == 1 {
		req.HttpRequest.URL.RawQuery = ""
	} else {
		req.HttpRequest.URL.RawQuery = rawQuery[1 : len(rawQuery)-1]
	}
}

func ReqQueryDelAllExcect(req *bfe_basic.Request, keys []string) {
	rawQuery := "&" + req.HttpRequest.URL.RawQuery + "&"
	queries := queryParse(req)

	keysMap := make(map[string]bool)
	for _, key := range keys {
		keysMap[key] = true
	}

	for key := range queries {
		if _, ok := keysMap[key]; ok {
			continue
		}

		queries.Del(key)
		for {
			start := strings.Index(rawQuery, "&"+key+"=")
			if start == -1 {
				break
			}

			end := strings.Index(rawQuery[start+1:], "&")
			if end == -1 {
				break
			}

			rawQuery = rawQuery[:start] + rawQuery[start+end+1:]
		}
	}

	if len(rawQuery) == 1 {
		req.HttpRequest.URL.RawQuery = ""
	} else {
		req.HttpRequest.URL.RawQuery = rawQuery[1 : len(rawQuery)-1]
	}
}
