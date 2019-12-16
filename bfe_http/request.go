package bfe_http

import (
	"errors"
	"fmt"
	"github.com/crud-bird/bfe/bfe_net/textproto"
	"github.com/crud-bird/bfe/bfe_tls"
	"io"
	"net"
	"net/url"
	"time"
)

const (
	maxValueLength   = 4096
	maxHeaderLines   = 1024
	chunkSize        = 4 << 10  // 4 KB chunks
	defaultMaxMemory = 32 << 20 // 32 MB
	MaxUriSize       = 1024 * 64
)

var ErrMissingFile = errors.New("http: no such file")

type ProtocolError struct {
	ErrorString string
}

func (err *ProtocolError) Error() string {
	return err.ErrorString
}

var (
	ErrHeaderTooLong        = &ProtocolError{"header too long"}
	ErrShortBody            = &ProtocolError{"entity body too short"}
	ErrNotSupported         = &ProtocolError{"feature not supported"}
	ErrUnexpectedTrailer    = &ProtocolError{"trailer header without chunked transfer encoding"}
	ErrMissingContentLength = &ProtocolError{"missing ContentLength in HEAD response"}
	ErrNotMultipart         = &ProtocolError{"request Content-Type isn't multipart/form-data"}
	ErrMissingBoundary      = &ProtocolError{"no multipart boundary param in Content-Type"}
)

type badStringError struct {
	what string
	str  string
}

func (e *badStringError) Error() string {
	return fmt.Sprintf("%s %q", e.what, e.str)
}

var reqWriteExcludeHeader = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

type Request struct {
	Method           string
	URL              *url.URL
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           Header
	HeaderKeys       textproto.MIMEKeys
	Body             io.ReadCloser
	ContentLength    int64
	TransferEncoding []string
	Host             string
	Form             url.Values
	PostForm         url.Values
	Trailer          Header
	RemoteAddr       string
	RequestURI       string
	TLS              *bfe_tls.ConnectionState
	State            *RequestState
}

type RequestState struct {
	SerialNumber        uint32
	Conn                net.Conn
	StartTime           time.Time
	ConnectBackendStart time.Time
	ConnectBackendEnd   time.Time
	HeaderSize          uint32
	BodySize            uint32
}

func (r *Request) ProtoAtLeast(major, minor int) bool {
	return r.ProtoMajor > major || r.ProtoMajor == major && r.ProtoMinor >= minor
}

func (r *Request) UserAgent() string {
	return r.Header.Get("User-Agent")
}
