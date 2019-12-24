package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/crud-bird/bfe/bfe_basic"
	"strings"
)

const (
	// connection actions
	ActionClose  = "CLOSE"  // close, close the connection directly
	ActionPass   = "PASS"   // pass, do nothing
	ActionFinish = "FINISH" // finish, close connection after reply

	// header actions
	ActionReqHeaderAdd = "REQ_HEADER_ADD" // add request header
	ActionReqHeaderSet = "REQ_HEADER_SET" // set request header
	ActionReqHeaderDel = "REQ_HEADER_DEL" // del request header

	// host actions
	ActionHostSetFromPathPrefix = "HOST_SET_FROM_PATH_PREFIX" // set host from path prefix
	ActionHostSet               = "HOST_SET"                  // set host

	// path actions
	ActionPathSet        = "PATH_SET"         // set path
	ActionPathPrefixAdd  = "PATH_PREFIX_ADD"  // add path prefix
	ActionPathPrefixTrim = "PATH_PREFIX_TRIM" // trim path prefix

	// query actions
	ActionQueryAdd          = "QUERY_ADD"            // add query
	ActionQueryDel          = "QUERY_DEL"            // del query
	ActionQueryRename       = "QUERY_RENAME"         // rename query
	ActionQueryDelAllExcept = "QUERY_DEL_ALL_EXCEPT" // del query except given query key
)

type ActionFile struct {
	Cmd    *string
	Params []string
}

type Action struct {
	Cmd    string
	Params []string
}

func (ac *Action) UnmarshalJSON(data []byte) error {
	var actionFile ActionFile

	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&actionFile); err != nil {
		return fmt.Errorf("decode Action err: %s", err)
	}

	if err := ActionFileCheck(actionFile); err != nil {
		return fmt.Errorf("ActionFileCheck error: %s", err)
	}

	ac.Cmd = *actionFile.Cmd
	ac.Params = actionFile.Params

	return nil
}

func (ac *Action) Check(allowAction map[string]bool) error {
	cmd := ac.Cmd
	if _, ok := allowAction[cmd]; !ok {
		return fmt.Errorf("action acm[%s], is not allowed", cmd)
	}

	return nil
}

func (ac *Action) Do(req *bfe_basic.Request) error {
	switch ac.Cmd {
	case ActionReqHeaderAdd:
		req.HttpRequest.Header.Add(ac.Params[0], ac.Params[1])
	case ActionReqHeaderSet:
		req.HttpRequest.Header.Set(ac.Params[0], ac.Params[1])
	case ActionReqHeaderDel:
		req.HttpRequest.Header.Del(ac.Params[0])

	case ActionHostSet:
		ReqHostSet(req, ac.Params[0])
	case ActionHostSetFromPathPrefix:
		ReqHostSetFromPathPrefix(req)

	case ActionPathSet:
		ReqPathSet(req, ac.Params[0])
	case ActionPathPrefixAdd:
		ReqPathPrefixAdd(req, ac.Params[0])
	case ActionPathPrefixTrim:
		ReqPathPrefixTrim(req, ac.Params[0])

	case ActionQueryAdd:
		ReqQueryAdd(req, ac.Params)
	case ActionQueryRename:
		ReqQueryRename(req, ac.Params[0], ac.Params[1])
	case ActionQueryDel:
		ReqQueryDel(req, ac.Params)
	case ActionQueryDelAllExcept:
		ReqQueryDelAllExcect(req, ac.Params)
	case ActionClose, ActionPass, ActionFinish:
		//pass
	default:
		return fmt.Errorf("unkown cmd[%s]", ac.Cmd)
	}

	return nil
}

const HeaderPrefix = "X-BFE-"

func ActionFileCheck(conf ActionFile) error {
	var paramsLenCheck int

	if conf.Cmd == nil {
		return fmt.Errorf("no cmd")
	}

	*conf.Cmd = strings.ToUpper(*conf.Cmd)
	switch *conf.Cmd {
	case ActionReqHeaderAdd, ActionReqHeaderSet:
		paramsLenCheck = 2
	case ActionReqHeaderDel:
		paramsLenCheck = 1
	case ActionClose, ActionPass, ActionFinish:
		paramsLenCheck = 0
	case ActionHostSetFromPathPrefix:
		paramsLenCheck = 0
	case ActionHostSet:
		paramsLenCheck = 1
	case ActionPathSet, ActionPathPrefixAdd, ActionPathPrefixTrim:
		paramsLenCheck = 1
	case ActionQueryAdd, ActionQueryRename:
		paramsLenCheck = 2
	case ActionQueryDel, ActionQueryDelAllExcept:
		paramsLenCheck = -1
	default:
		return fmt.Errorf("invalid cmd[%s]", *conf.Cmd)
	}

	if paramsLenCheck != -1 && len(conf.Params) != paramsLenCheck {
		return fmt.Errorf("num of params:[ok:%d, now:%d]", paramsLenCheck, len(conf.Params))
	}

	for _, p := range conf.Params {
		if len(p) == 0 {
			return fmt.Errorf("empty params")
		}
	}

	if *conf.Cmd == ActionReqHeaderSet || *conf.Cmd == ActionReqHeaderAdd {
		header := strings.ToUpper(conf.Params[0])
		if !strings.HasPrefix(header, HeaderPrefix) {
			return fmt.Errorf("add/set header key mast start with X-Bfe-, get %s", header)
		}
	}

	return nil
}
