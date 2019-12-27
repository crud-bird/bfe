package bfe_http

import (
	"github.com/crud-bird/bfe/bfe_net/textproto"
	"io"
	"sort"
	"strings"
	"time"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

type Header map[string][]string

func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

func (h Header) GetDirect(key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}

	return ""
}

func (h Header) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

func (h Header) Write(w io.Writer) error {
	return h.WriteSubset(w, nil)
}

var timeFormats = []string{
	TimeFormat,
	time.RFC850,
	time.ANSIC,
}

func ParseTime(text string) (t time.Time, err error) {
	for _, layout := range timeFormats {
		t, err = time.Parse(layout, text)
		if err == nil {
			return
		}
	}

	return
}

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

type writeStringer interface {
	WriteString(string) (int, error)
}

type stringWriter struct {
	w io.Writer
}

func (w stringWriter) WriteString(s string) (n int, err error) {
	return w.w.Write([]byte(s))
}

type keyValues struct {
	key    string
	values []string
}

type headerSorter struct {
	kvs []keyValues
}

func (s *headerSorter) Len() int {
	return len(s.kvs)
}

func (s *headerSorter) Swap(i, j int) {
	s.kvs[i], s.kvs[j] = s.kvs[j], s.kvs[i]
}

func (s *headerSorter) Less(i, j int) bool {
	return s.kvs[i].key < s.kvs[j].key
}

var headerSorterCache = make(chan *headerSorter, 8)

func (h Header) sortedKeyValues(exclude map[string]bool) (kvs []keyValues, hs *headerSorter) {
	select {
	case hs = <-headerSorterCache:
	default:
		hs = new(headerSorter)
	}

	if cap(hs.kvs) < len(h) {
		hs.kvs = make([]keyValues, 0, len(h))
	}
	kvs = hs.kvs[:0]
	for k, vv := range h {
		if !exclude[k] {
			kvs = append(kvs, keyValues{k, vv})
		}
	}
	hs.kvs = kvs
	sort.Sort(hs)

	return kvs, hs
}

func (h Header) WriteSubset(w io.Writer, exclude map[string]bool) error {
	ws, ok := w.(writeStringer)
	if !ok {
		ws = stringWriter{w}
	}

	kvs, sorter := h.sortedKeyValues(exclude)
	for _, kv := range kvs {
		for _, v := range kv.values {
			v = headerNewlineToSpace.Replace(v)
			v = textproto.TrimString(v)
			for _, s := range []string{
				kv.key,
				": ",
				v,
				"\r\n",
			} {
				if _, err := ws.WriteString(s); err != nil {
					return err
				}
			}
		}
	}

	select {
	case headerSorterCache <- sorter:
	default:
	}
	return nil
}
