package textproto

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_bufio"
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
	Pipeline
	conn io.ReadWriteCloser
}

func NewConn(conn io.ReadWriteCloser) *Conn {
	return &Conn{
		Reader: Reader{R: bfe_bufio.NewReader(conn)},
		Writer: Writer{W: bfe_bufio.NewWriter(conn)},
		conn:   conn,
	}
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func Dial(network, addr string) (*Conn, error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	return NewConn(c), nil
}

func (c *Conn) Cmd(format string, args ...interface{}) (id uint, err error) {
	id = c.Next()
	c.StartRequest(id)
	err = c.PrintfLine(format, args...)
	c.EndRequest(id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func TrimString(s string) string {
	for len(s) > 0 && isASCIISpace(s[0]) {
		s = s[1:]
	}

	for len(s) > 0 && isASCIISpace(s[len(s)-1]) {
		s = s[:len(s)-1]
	}

	return s
}

func TrimBytes(b []byte) []byte {
	for len(b) > 0 && isASCIISpace(b[0]) {
		b = b[1:]
	}

	for len(b) > 0 && isASCIISpace(b[len(b)-1]) {
		b = b[:len(b)-1]
	}

	return b
}

func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isASCIILetter(b byte) bool {
	b |= 0x20 //十进制的32, 小写字母 = 大写字母 + 32
	return 'a' <= b && b <= 'z'
}
