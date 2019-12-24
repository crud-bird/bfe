package mod_access

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/crud-bird/bfe/bfe_basic"
	"github.com/crud-bird/bfe/bfe_http"
)

func onLogFmtAllServeTime(m *ConfModAccess, logItem *LogFmtItem, buff bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	now := time.Now()
	ms := now.Sub(req.Stat.ReadReqStart).Nanoseconds() / 1000000
	msg := fmt.Sprintf("%d", ms)
	buff.WriteString(msg)

	return nil
}

func onLogFmtBackend(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := fmt.Sprintf("%s,%s,%s %s", req.Backend.ClusterName, req.Backend.SubclusterName, req.Backend.BackendAddr, req.Backend.BackendName)
	buff.WriteString(msg)
	return nil
}

func onLogFmtBodyLenIn(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	msg := fmt.Sprintf("%d", req.Stat.BodyLenIn)
	buff.WriteString(msg)

	return nil
}

func onLogFmtClientReadTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	ms := req.Stat.ReadReqEnd.Sub(req.Stat.ReadReqStart).Nanoseconds() / 1000000
	msg := fmt.Sprintf("%d", ms)
	buff.WriteString(msg)

	return nil
}

func onLogFmtClusterDuration(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.HttpRequest == nil {
		return errors.New("req.HttpRequest is nil")
	}

	msg := fmt.Sprintf("%d", req.HttpRequest.ContentLength)
	buff.WriteString(msg)

	return nil
}

func onLogFmtClusterName(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	msg := "-"
	if req.Backend.ClusterName != "" {
		msg = req.Backend.ClusterName
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtClusterServeTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req != nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	ms := req.Stat.CLusterEnd.Sub(req.Stat.ClusterStart).Nanoseconds() / 1000000
	msg := fmt.Sprintf("%d", ms)
	buff.WriteString(msg)

	return nil
}

func onLogFmtConnectBackendTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil || req.OutRequest == nil {
		return errors.New("req is nil")
	}

	if req.OutRequest.State == nil {
		return errors.New("req.OutRequest.Stat is nil")
	}

	msg := "-"
	stat := req.OutRequest.State
	ms := stat.ConnectBackendEnd.Sub(stat.ConnectBackendStart).Nanoseconds() / 1000000
	if ms >= 0 {
		msg = fmt.Sprintf("%d", ms)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtReqContentLen(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.HttpRequest == nil {
		return errors.New("req.HttpRequest is nil")
	}

	msg := fmt.Sprintf("%d", req.HttpRequest.ContentLength)
	buff.WriteString(msg)

	return nil
}

func onLogFmtHost(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.HttpRequest == nil {
		return errors.New("req.HttpRequest is nil")
	}

	buff.WriteString(req.HttpRequest.Host)

	return nil
}

func onLogFmtHttp(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.HttpRequest == nil {
		return errors.New("req.HttpRequest is nil")
	}

	msg := "-"
	if data, ok := req.HttpRequest.Header[logItem.Key]; ok {
		msg = strings.Join(data, ",")
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtIsTrustip(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := fmt.Sprintf("%v", req.Session.IsTrustIP)
	buff.WriteString(msg)

	return nil
}

func onLogFmtLastBackendDuration(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	ms := req.Stat.BackendEnd.Sub(req.Stat.BackendStart).Nanoseconds() / 1000000
	msg := fmt.Sprintf("%d", ms)
	buff.WriteString(msg)

	return nil
}

func onLogFmtLogId(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := fmt.Sprintf("%s", req.LogId)
	buff.WriteString(msg)

	return nil
}

func onLogFmtNthReqInSession(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if req.HttpRequest != nil && req.HttpRequest.State != nil {
		msg = fmt.Sprintf("%d", req.HttpRequest.State.SerialNumber)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtResContentLen(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if req.Stat != nil {
		msg = fmt.Sprintf("%d", req.Stat.BodyLenOut)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtStatusCode(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	msg := "-"
	if res != nil {
		msg = fmt.Sprintf("%d", res.StatusCode)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtProduct(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if req.Route.Product != "" {
		msg = req.Route.Product
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtProxyDelayTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	msg := "-"
	if !req.Stat.BackendFirst.IsZero() {
		ms := req.Stat.BackendFirst.Sub(req.Stat.ReadReqEnd).Nanoseconds() / 1000000
		msg = fm.Sprintf("%d", ms)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtReadReqDuration(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	ms := req.Stat.ReadReqEnd.Sub(req.Stat.ReadReqStart).Nanoseconds() / 1000000
	msg := fmt.Sprintf("%d", ms)
	buff.WriteString(msg)

	return nil
}

func onLogFmtReadWriteSrvTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	msg := "-"
	if !req.Stat.BackendStart.IsZero() {
		now := time.Now()
		ms := now.Sub(req.Stat.BackendStart).Nanoseconds() / 1000000
		msg = fmt.Sprintf("%d", ms)
	}

	buff.WriteString(msg)

	return nil
}

func onLogFmtRedirect(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if req.Redirect.Url != "" {
		msg = fmt.Sprintf("%s,%d", req.Redirect.Url, req.Redirect.Code)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtClientIp(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if req.RemoteAddr != nil {
		msg = req.RemoteAddr.IP.String()
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtReqCookie(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if co, ok := req.Cookie(logItem.Key); ok {
		msg = co.Value
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtErrorCode(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := buildErrorMsg(req.ErrCode, req.ErrMsg)
	buff.WriteString(msg)

	return nil
}

func onLogFmtReqHeaderLen(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

msg:
	fmt.Sprintf("%d", req.Stat.HeaderLenIn)
	buff.WriteString(msg)

	return nil
}

func onLogFmtRequestHeader(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.HttpRequest == nil {
		return errors.New("req.HttpRequest is nil")
	}

	msg := "-"
	if data := req.HttpRequest.Header.Get(logItem.Key); data != "" {
		msg = data
	}
	buff.WriteString("msg")

	return nil
}

func onLogFmtRequestUri(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.HttpRequest == nil {
		return errors.New("req.HttpRequest is nil")
	}

	buff.WriteString(req.HttpRequest.RequestURI)

	return nil
}

func onLogFmtResCookie(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	if res == nil {
		buff.WriteString("-")
		return nil
	}

	msg := "-"
	cookies := res.Cookies()
	for _, co := range cookies {
		if co.Name == logItem.Key {
			msg = co.Value
			break
		}
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtResDuration(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	ms := req.Stat.ResponseEnd.Sub(req.Stat.ResponseStart).Nanoseconds() / 1000000
	msg := fmt.Sprintf("%d", ms)
	buff.WriteString(msg)

	return nil
}

func onLogFmtResProto(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if res != nil {
		msg = fmt.Sprintf("%s", res.Proto)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtResponseHeader(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if res == nil {
		buff.WriteString("-")
		return nil
	}

	msg := "-"
	if data, ok := res.Header[logItem.Key]; ok {
		msg = strings.Join(data, ",")
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtResHeaderLen(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("res is nil")
	}

	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	msg := "-"
	if res != nil {
		msg = fmt.Sprintf("%d", req.Stat.HeaderLenOut)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtResStatus(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if res != nil {
		msg = fmt.Sprintf("%s", res.Status)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtRetryNum(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := fmt.Sprintf("%d", req.RetryTime)
	buff.WriteString(msg)

	return nil
}

func onLogFmtServerAddr(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("rea is nil")
	}

	msg := "-"
	if req.Connection != nil {
		msg = req.Connection.LocalAddr().String()
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtSinceSessionTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Session == nil {
		return errors.New("req.Session is nil")
	}

	ms := time.Since(req.Session.StartTime).Nanoseconds() / 1000000
	msg := fmt.Sprintf("%d", ms)
	buff.WriteString(msg)

	return nil
}

func onLogFmtSubclusterName(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	buff.WriteString(req.Backend.SubclusterName)

	return nil
}

func onLogFmtTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	now := time.Now()
	buff.WriteString(fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()))

	return nil
}

func onLogFmtVip(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}

	msg := "-"
	if req.Session.Vip != nil {
		if vip := req.Session.Vip.String(); len(vip) > 0 {
			msg = vip
		}
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtUri(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.HttpRequest == nil {
		return errors.New("req.HttpRequest is nil")
	}

	buff.WriteString(req.HttpRequest.URL.String())

	return nil
}

func onLogFmtWriteSrvTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, req *bfe_basic.Request, res *bfe_http.Response) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if req.Stat == nil {
		return errors.New("req.Stat is nil")
	}

	ms := req.Stat.BackendEnd.Sub(req.Stat.BackendStart).Nanoseconds() / 1000000
	buff.WriteString(fmt.Sprintf("%d", ms))

	return nil
}
