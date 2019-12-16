package httputil

import (
	// "errors"
	// "io"
	// "net"
	// "net/textproto"
	// "sync"
	// bufio "github.com/crud-bird/bfe/bfe_bufio"
	http "github.com/crud-bird/bfe/bfe_http"
)

var (
	ErrPersistEOF = &http.ProtocolError{ErrorString: "persistent connection closed"}
	ErrClosed     = &http.ProtocolError{ErrorString: "connection closed by user"}
	ErrPipeline   = &http.ProtocolError{ErrorString: "pipeline error"}
)
