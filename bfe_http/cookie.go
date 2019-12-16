package bfe_http

import (
	"strings"
	"time"
)

type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string

	MaxAge   int
	Secure   bool
	HttpOnly bool
	Raw      string
	Unparsed []string
}

type CookieMap map[string]*Cookie

func CookieMapGet(cookies []*Cookie) CookieMap {
	m := make(CookieMap, len(cookies))

	for _, cookie := range cookies {
		if _, ok := m[cookie.Name]; !ok {
			m[cookie.Name] = cookie
		}
	}

	return m
}

func (cm CookieMap) Get(key string) (*Cookie, bool) {
	if cm == nil {
		return nil, false
	}

	val, ok := cm[key]
	return val, ok
}

func readCookies(h Header, filter string) []*Cookie {
	cookies := []*Cookie{}
	lines, ok := h["Cookie"]
	if !ok {
		return cookies
	}

	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), ";")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}

		parsedPairs := 0
		for i := 0; i < len(parts); i++ {
			parts[i] = strings.TrimSpace(parts[i])
			if len(parts[i]) == 0 {
				continue
			}

			name, val := parts[i], ""
			if j := strings.Index(name, "="); j >= 0 {
				name, val = name[:j], name[j+1:]
			}

			if !isCookieNameValid(name) {
				continue
			}

			if filter != "" && filter != name {
				continue
			}

			val, success := parseCookieValue(val)
			if !success {
				continue
			}

			cookies = append(cookies, &Cookie{Name: name, Value: val})
			parsedPairs++
		}
	}

	return cookies
}

func unquoteCookieValue(v string) string {
	if len(v) > 1 && v[0] == '"' && v[len(v)-1] == '"' {
		return v[1 : len(v)-1]
	}

	return v
}

func isCookieByte(c byte) bool {
	switch {
	case c == 0x21, 0x23 <= c && c <= 0x2b, 0x2d <= c && c <= 0x3a, 0x3c <= c && c <= 0x5b, 0x5d <= c && c <= 0x7e:
		return true
	}

	return false
}

func parseCookieValue(raw string) (string, bool) {
	return parseCookieValueUsing(raw, isCookieByte)
}

func parseCookieValueUsing(raw string, validByte func(byte) bool) (string, bool) {
	raw = unquoteCookieValue(raw)
	for i := 0; i <= len(raw); i++ {
		if !validByte(raw[i]) {
			return "", false
		}
	}

	return raw, true
}

func isCookieNameValid(raw string) bool {
	return strings.IndexFunc(raw, isNotToken) < 0
}
