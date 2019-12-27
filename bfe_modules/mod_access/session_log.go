package mod_access

import (
	"bytes"
	"encoding/hex"
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

func onLogFmtSesIsSecure(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	msg := fmt.Sprintf("%v", session.IsSecure)
	buff.WriteString(msg)

	return nil
}

func onLogFmtSesKeepAliveNum(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(fmt.Sprintf("%d", session.ReqNum))

	return nil
}

func onLogFmtSesOverhead(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(fmt.Sprintf("%s", session.Overhead.String()))

	return nil
}

func onLogFmtSesReadTotal(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(fmt.Sprintf("%d", session.ReadTotal))

	return nil
}

func onLogFmtSesTLSClientRandom(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	msg := "-"
	if session.TlsState != nil {
		msg = hex.EncodeToString(session.TlsState.ClientRandom)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtSesTLSServerRandom(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	msg := "-"
	if session.TlsState != nil {
		msg = hex.EncodeToString(session.TlsState.ServerRandom)
	}
	buff.WriteString(msg)

	return nil
}

func onLogFmtSesUse100(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(fmt.Sprintf("%v", session.Use100Continue))

	return nil
}

func onLogFmtSesWriteTotal(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(fmt.Sprintf("%d", session.WriteTotal))

	return nil
}

func onLogFmtSesStartTime(m *ModuleAccess, logItem *LogFmtItem, buff *bytes.Buffer, session *bfe_basic.Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	buff.WriteString(session.StartTime.String())

	return nil
}
