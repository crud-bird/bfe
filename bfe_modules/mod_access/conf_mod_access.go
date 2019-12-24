package mod_access

import (
	"fmt"

	gcfg "gopkg.in/gcfg.v1"
)

type ConfModAccess struct {
	Log struct {
		LogPrefix   string
		LogDir      string
		RotateWhen  string
		BackupCount int
	}

	Template struct {
		RequestTemplate string
		SessionTemplate string
	}
}

func ConfLoad(filePath string) (*ConfModAccess, error) {
	var err error
	var cfg ConfModAccess

	if err = gcfg.ReadFileInto(&cfg, filePath); err != nil {
		return &cfg, err
	}

	if err = cfg.Check(); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func (cfg *ConfModAccess) Check() error {
	if cfg.Log.LogPrefix == "" {
		return fmt.Errorf("ConfModAccess.LogPrefix is empty")
	}

	if cfg.Log.LogDir == "" {
		return fmt.Errorf("ConfModAccess.LogDir is empty")
	}

	if cfg.Log.BackupCount <= 0 {
		return fmt.Errorf("ConfModAccess.BackupCount[%d] should > 0", cfg.Log.BackupCount)
	}

	if cfg.Template.RequestTemplate == "" {
		return fmt.Errorf("ConfModAccess.RequestTemplate not set")
	}

	if cfg.Template.SessionTemplate == "" {
		return fmt.Errorf("ConfModAccess.SessionTemplate not set")
	}

	return nil
}

func checkLogFmt(item LogFmtItem, logFmtType string) error {
	if logFmtType != Request && logFmtType != Session {
		return fmt.Errorf("logFmtType should be Request or Session")
	}

	domain, found := fmtItemDomainTable[item.Type]
	if !found {
		return fmt.Errorf("type: (%d, %s) not configured in domain table", item.Type, item.Key)
	}

	if domain != DomainAll && domain != logFmtType {
		return fmt.Errorf("type: (%d, %s) should not in request finish log", item.Type, item.Key)
	}

	return nil
}

func tokenTypeGet(templatePtr *string, offset int) (int, int, error) {
	templateLen := len(*templatePtr)

	for key, logItemType := range fmtTable {
		n := len(key)
		if offset+n > templateLen {
			continue
		}

		if key == (*templatePtr)[offset:(offset+n)] {
			return logItemType, offset + n - 1, nil
		}
	}

	return -1, -1, fmt.Errorf("no such log item format type: %s", *templatePtr)
}

func parseBracketToken(templatePtr *string, offset int) (LogFmtItem, int, error) {
	length := len(*templatePtr)
	var end int
	for end = offset + 1; end < length; end++ {
		if (*templatePtr)[end] == '}' {
			break
		}
	}

	if end >= length {
		return LogFmtItem{}, -1, fmt.Errorf("log format: { must terminated by a }")
	}

	if end == (length - 1) {
		return LogFmtItem{}, -1, fmt.Errorf("log format: { must followed by a charactor")
	}

	key := (*templatePtr)[offset+1 : end]
	logItemType, last, err := tokenTypeGet(templatePtr, end+1)
	if err != nil {
		return LogFmtItem{}, -1, err
	}

	return LogFmtItem{key, logItemType}, last, nil
}

func parseLogTemplate(template string) ([]LogFmtItem, error) {
	reqFmts := make([]LogFmtItem, 0)
	start := 0
	length := len(template)
	var token string

	for i := 0; i < length; i++ {
		if template[i] != '$' {
			continue
		}

		if (i + 1) == length {
			return nil, fmt.Errorf("log format: $ must followed with a charactor")
		}

		if start <= (i - 1) {
			token = template[start:i]
			item := LogFmtItem{token, FormatString}
			reqFmts = append(reqFmts, item)
		}

		if template[i+1] == '{' {
			item, end, err := parseBracketToken(&template, i+1)
			if err != nil {
				return nil, err
			}
			reqFmts = append(reqFmts, item)
			i = end
			start = end + 1
		} else {
			logItemType, end, err := tokenTypeGet(&template, i+1)
			if err != nil {
				return nil, err
			}

			token = template[(i + 1) : end+1]
			item := LogFmtItem{token, logItemType}
			reqFmts = append(reqFmts, item)

			i = end
			start = end + 1
		}
	}

	if start < length {
		token = template
		item := LogFmtItem{token, FormatString}
		reqFmts = append(reqFmts, item)
	}

	return reqFmts, nil
}
