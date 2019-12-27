package bfe_http

import (
	"errors"
	"fmt"
	// 	"github.com/crud-bird/bfe/bfe_bufio"
	"github.com/crud-bird/bfe/bfe_net/textproto"
	"github.com/crud-bird/bfe/bfe_tls"
	"io"
	"mime"
	"mime/multipart"
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
	MultipartForm    *multipart.Form
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

func (r *Request) Cookies() []*Cookie {
	return readCookies(r.Header, "")
}

var ErrNoCookie = errors.New("http: named cookie not present")

func (r *Request) Cookie(name string) (*Cookie, error) {
	if cookies := readCookies(r.Header, name); len(cookies) > 0 {
		return cookies[0], nil
	}

	return nil, ErrNoCookie
}

func (r *Request) AddCookie(c *Cookie) {
	s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), sanitizeCookieValue(c.Value))
	if c := r.Header.Get("Cookie"); c != "" {
		r.Header.Set("Cookie", c+"; "+s)
	} else {
		r.Header.Set("Cookie", s)
	}
}

func (r *Request) Refer() string {
	return r.Header.Get("Referer")
}

var multipartByReader = &multipart.Form{
	Value: make(map[string][]string),
	File:  make(map[string][]*multipart.FileHeader),
}

func (r *Request) MultipartReader() (*multipart.Reader, error) {
	if r.MultipartForm == multipartByReader {
		return nil, errors.New("http: MultipartReader called twice")
	}

	if r.MultipartForm != nil {
		return nil, errors.New("http: multipart handled by ParseMultipartForm")
	}

	r.MultipartForm = multipartByReader
	return r.multipartReader()
}

func (r *Request) multipartReader() (*multipart.Reader, error) {
	v := r.Header.Get("Content-Type")
	if v == "" {
		return nil, ErrNotMultipart
	}

	d, params, err := mime.ParseMediaType(v)
	if err != nil || d != "multipart/form-data" {
		return nil, ErrNotMultipart
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, ErrMissingBoundary
	}

	return multipart.NewReader(r.Body, boundary), nil
}

func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}

	return def
}

const defaultUserAgent = "Go 1.1 package http"

// func (r *Request) Write(w io.Writer) error {
// 	return r.write(w, false, nil)
// }
//
// func (r *Request) WriteProxy(w io.Writer) error {
// 	return r.write(w, true, nil)
// }

// func (r *Request) write(w io.Writer, usingProxy bool, header Header) error {
// 	host := r.Host
// 	if host == "" {
// 		if r.URL == nil {
// 			return errors.New("http: Request.Write on Request without host or url")
// 		}
// 		host = r.URL.Host
// 	}
//
// 	ruri := r.URL.RequestURI()
// 	if usingProxy && r.URL.Scheme != "" && r.URL.Opaque == "" {
// 		ruri = r.URL.Scheme + "://" +host + ruri
// 	} else if r.Method == "CONNECT" && r.URL.Path == "" {
// 		ruri = host
// 	} else {
// 		rawurl, err := url.ParseRequestURI(r.RequestURI)
// 		if err == nil && rawurl.RequestURI() == ruri {
// 			if rawurl.Scheme == "" && rawurl.Host == "" &&rawurl.Opaque == "" {
// 				ruri = r.RequestURI
// 			}
// 		}
// 	}
//
// 	var bw *bfe_bufio.Writer
//
// 	switch w.(type) {
// 	case io.ByteWriter:
// 	case *MaxLatencyWriter:
// 	default:
// 		bw = bfe_bufio.NewWriter(w)
// 		w = bw
// 	}
//
// 	fmt.Fprintf(w, "%s %s HTTP/1.1\r\n", valueOrDefault(r.Method, "GET"), ruri)
// 	fmt.Fprintf(w, "Host: %s\r\n", host)
//
// 	tw, err := newTransferWriter(r)
// 	if err != nil {
// 		return err
// 	}
// 	err = tw.WriteHeader(w)
// 	if err != nil {
// 		return err
// 	}
//
// 	if err = r.Header.WriteSubset(w, ReqWriteExcludeHeader); err != nil {
// 		return err
// 	}
//
// 	if header != nil {
// 		if err = header.Write(w); err != nil {
// 			return err
// 		}
// 	}
//
// 	io.WriteString(w, "\r\n")
//
// 	if rbw, ok := w.(Flusher); ok {
// 		if err = rbw.Flush(); err !=nil {
// 			return err
// 		}
// 	}
//
// 	n, err := tw.WriteBody(w)
// 	if err != nil {
// 		return err
// 	}
// 	r.State.BodySize = uint32(n)
//
// 	if bw != nil {
// 		return bw.Flush()
// 	}
//
// 	return nil
// }
// todo
