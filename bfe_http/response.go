package bfe_http

import (
	"github.com/crud-bird/bfe/bfe_tls"
	"io"
)

var respExcludeHeader = map[string]bool{
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

type SignCalculater interface {
	CalcSign(string) string
}

type Response struct {
	Status           string
	StatusCode       int
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           Header
	Body             io.ReadCloser
	ContentLength    int64
	TransferEncoding []string
	Signer           SignCalculater
	CLose            bool
	Trailer          Header
	Request          *Request
	TLS              *bfe_tls.ConnectionState
}
