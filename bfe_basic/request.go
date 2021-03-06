package bfe_basic

import (
	"github.com/crud-bird/bfe/bfe_balance/backend"
	"github.com/crud-bird/bfe/bfe_http"
	"net"
	"net/url"
)

type BackendInfo struct {
	ClusterName    string
	SubclusterName string
	BackendAddr    string
	BackendPort    uint32
	BackendName    string
}

type RedirectInfo struct {
	Url  string
	Code int
}

type RequestRoute struct {
	Error       error
	HostTag     string
	Product     string
	ClusterName string
}

type RequestTags struct {
	Error    error
	TagTable map[string][]string
}

type RequestTransport struct {
	Backend   *backend.BfeBackend
	Transport bfe_http.RoundTripper
}

type Request struct {
	Connection net.Conn
	Session    *Session
	RemoteAddr *net.TCPAddr
	ClientAddr *net.TCPAddr

	HttpRequest  *bfe_http.Request
	OutRequest   *bfe_http.Request
	HttpResponse *bfe_http.Response

	CookieMap bfe_http.CookieMap
	Query     url.Values

	LogId         string
	ReqBody       []byte
	ReqBodyPeeked bool

	Route RequestRoute

	Tags RequestTags

	Trans RequestTransport

	BfeStatusCode int

	ErrCode error
	ErrMsg  string

	Stat *RequestStat

	RetryTime int
	Backend   BackendInfo

	Redirect RedirectInfo

	SvrDataConf ServerDataConfInterface

	Context map[interface{}]interface{}
}

func NewRequest(req *bfe_http.Request, conn net.Conn, stat *RequestStat, session *Session, svrDataConf ServerDataConfInterface) *Request {
	var addr *net.TCPAddr
	if session != nil {
		addr = session.RemoteAddr
	}

	return &Request{
		ErrCode:     nil,
		Connection:  conn,
		HttpRequest: req,
		Stat:        stat,
		Session:     session,
		Context:     make(map[interface{}]interface{}),
		Tags: RequestTags{
			TagTable: make(map[string][]string),
		},
		SvrDataConf: svrDataConf,
		RemoteAddr:  addr,
	}

}

func (req *Request) CachedQuery() url.Values {
	if req.Query == nil {
		req.Query = req.HttpRequest.URL.Query()
	}

	return req.Query
}

func (req *Request) CachedCookie() bfe_http.CookieMap {
	if req.CookieMap == nil {
		cookies := req.HttpRequest.Cookies()
		req.CookieMap = bfe_http.CookieMapGet(cookies)

	}

	return req.CookieMap
}

func (req *Request) Cookie(name string) (*bfe_http.Cookie, bool) {
	if req.CookieMap == nil {
		req.CachedCookie()
	}

	return req.CookieMap.Get(name)
}

func (req *Request) SetRequestTransport(backend *backend.BfeBackend, transport bfe_http.RoundTripper) {
	req.Trans.Backend = backend
	req.Trans.Transport = transport
}

func (req *Request) Protocali() string {
	if req.Session.IsSecure {
		return req.Session.Proto
	}

	return req.HttpRequest.Proto
}

func (req *Request) AddTags(name string, ntags []string) {
	if len(ntags) == 0 {
		return
	}

	tags := req.Tags.TagTable[name]
	tags = append(tags, ntags...)
	req.Tags.TagTable[name] = tags
}

func (req *Request) GetTags(name string) []string {
	return req.Tags.TagTable[name]
}

func (req *Request) SetContext(key, val interface{}) {
	req.Context[key] = val
}

func (req *Request) GetContext(key interface{}) interface{} {
	return req.Context[key]
}
