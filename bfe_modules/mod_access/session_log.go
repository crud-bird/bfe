package mod_access

import (
	"bytes"
	"errors"
	"github.com/crud-bird/bfe/bfe_basic"
)

func onLogFmtSesClientIp(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(session.RemoteAddr.String())

	return nil
}

func onLogFmtSesEndTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(session.EndTime.String())

	return nil
}

func buildErrorMsg(err error, errMsg string) string {
	msg := "-"
	if err != nil {
		msg = err.Error()
		if len(errMsg) != 0 {
			msg += "," + errMsg
		}
	}

	return msg
}

func onLogFmtSesErrorCode(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	errCode, errMsg := session.GetError()
	msg := buildErrorMsg(errCode, errMsg)
	buff.WriteString(msg)

	return nil
}

// todo
