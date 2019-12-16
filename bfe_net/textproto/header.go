package textproto

type MIMEHeader map[string][]string

type MIMEKeys []string

func (h MIMEHeader) Add(key, value string) {
	key = CanonicalMIMEHeaderKey(key)
	h[key] = append(h[key], value)
}

func (h MIMEHeader) Set(key, value string) {
	h[CanonicalMIMEHeaderKey(key)] = []string{value}
}

func (h MIMEHeader) Get(key string) string {
	if h == nil {
		return ""
	}

	v := h[CanonicalMIMEHeaderKey(key)]
	if len(v) == 0 {
		return ""
	}

	return v[0]
}

func (h MIMEHeader) Del(key string) {
	delete(h, CanonicalMIMEHeaderKey(key))
}
