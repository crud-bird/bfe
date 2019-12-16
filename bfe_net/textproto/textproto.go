package textproto

import (
	"fmt"
	"github.com/crud-bird/bfe_bufio"
	"io"
	"net"
)

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%03d %s", e.Code, e.Msg)
}

type ProtocolError string

func (p ProtocolError) Error() string {
	return string(p)
}

type Conn struct {
	Reader
	Writer
	Pipline
	conn io.ReadWriteCloser
}

func NewConn(conn io.ReadWriteCloser) *Conn {
	return &Conn{
		Reader: Reader{R: bfe_bufio.NewReader(conn)},
		Writer: Writer{W: bfe_bufio.NewWriter(conn)},
		conn:   conn,
	}
}
