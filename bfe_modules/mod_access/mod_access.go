package mod_access

import (
	"bytes"
	"fmt"

	"github.com/baidu/go-lib/web-monitor/web_monitor"
	"github.com/crud-bird/bfe/bfe_basic"
	"github.com/crud-bird/bfe/bfe_http"
	"github.com/crud-bird/bfe/bfe_module"
	"github.com/sirupsen/logrus"
)

type ModuleAccess struct {
	name   string
	logger *logrus.Logger
	conf   *ConfModAccess

	reqFmts     []LogFmtItem
	sessionFmts []LogFmtItem
}

func NewModuleAccess() *ModuleAccess {
	return &ModuleAccess{
		name: "mod_access",
	}
}

func (m *ModuleAccess) Name() string {
	return m.name
}

func (m *ModuleAccess) ParseConfig(conf *ConfModAccess) error {
	var err error

	m.reqFmts, err = parseLogTemplate(conf.Template.RequestTemplate)
	if err != nil {
		return fmt.Errorf("%s.Init(): RequestTemplate %s", m.name, err.Error())
	}

	m.sessionFmts, err = parseLogTemplate(conf.Template.SessionTemplate)
	if err != nil {
		return fmt.Errorf("%s.Init(): SessionTemplate %s", m.name, err.Error())
	}

	return nil
}

func (m *ModuleAccess) Init(cbs *bfe_module.BfeCallbacks, whs *web_monitor.WebHandlers, cr string) error {
	var err error
	var conf *ConfModAccess

	confPath := bfe_module.ModConfPath(cr, m.name)
	if conf, err = ConfLoad(confPath); err != nil {
		return fmt.Errorf("%s: cond load err %s", m.name, err.Error())
	}

	return m.init(conf, cbs, whs)

}

func (m *ModuleAccess) init(conf *ConfModAccess, cbs *bfe_module.BfeCallbacks, whs *web_monitor.WebHandlers) error {
	var err error
	if err = m.ParseConfig(conf); err != nil {
		return fmt.Errorf("%s.Init(): ParseConfig %s", m.name, err.Error())
	}

	m.conf = conf

	if err = m.CheckLogFormat(); err != nil {
		return fmt.Errorf("%s.Init(): CheckLogFormat %s", m.name, err.Error())
	}

	m.logger = logrus.New()

	if err = cbs.AddFilter(bfe_module.HANDLE_REQUEST_FINISH, m.requestLogHandler); err != nil {
		return fmt.Errorf("%s.Init(): AddFilter(m.requestLogHandler): %s", m.name, err.Error())
	}

	if err = cbs.AddFilter(bfe_module.HANDLE_FINISH, m.sessionLogHandler); err != nil {
		return fmt.Errorf("%s.Init(): AddFilter(m.sessionLogHandler) %s", m.name, err.Error())
	}

	return nil
}

func (m *ModuleAccess) CheckLogFormat() error {
	for _, item := range m.reqFmts {
		err := checkLogFmt(item, Request)
		if err != nil {
			return err
		}
	}

	for _, item := range m.sessionFmts {
		err := checkLogFmt(item, Session)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ModuleAccess) requestLogHandler(req *bfe_basic.Request, res *bfe_http.Response) int {
	byteStr := bytes.NewBuffer(nil)

	for _, item := range m.reqFmts {
		switch item.Type {
		case FormatString:
			byteStr.WriteString(item.Key)
		case FormatTime:
			onLogFmtTime(m, byteStr)
		default:
			if handler, ok := fmtHandlerTable[item.Type]; ok {
				h := handler.(func(*ModuleAccess, *LogFmtItem, *bytes.Buffer, *bfe_basic.Request, *bfe_http.Response) error)
				h(m, &item, byteStr, req, res)
			}

		}
	}

	byteStr.WriteString("\n")
	m.logger.Info(byteStr.Bytes())

	return bfe_module.BFE_HANDLER_GOON
}

func (m *ModuleAccess) sessionLogHandler(session *bfe_basic.Session) int {
	byteStr := bytes.NewBuffer(nil)
	for _, item := range m.sessionFmts {
		switch item.Type {
		case FormatString:
			byteStr.WriteString(item.Key)
		case FormatTime:
			onLogFmtTime(m, byteStr)
		default:
			if handler, ok := fmtHandlerTable[item.Type]; ok {
				h := handler.(func(*ModuleAccess, *LogFmtItem, *bytes.Buffer, *bfe_basic.Session) error)
				h(m, &item, byteStr, session)
			}
		}
	}

	byteStr.WriteString("\n")
	m.logger.Info(byteStr.Bytes())

	return bfe_module.BFE_HANDLER_GOON
}
