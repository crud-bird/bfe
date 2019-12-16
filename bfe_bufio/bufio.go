package bfe_bufio

import (
	"errors"
	"io"
)

const defaultBufSize = 4096

var (
	ErrInvalidUnreadByte = errors.New("bfe_bufio: invalid use of UnreadByte")
	ErrInvalidUnreadRune = errors.New("bfe_bufio: invalid use of UnreadRune")
	ErrBufferFull        = errors.New("bfe_bufio: buffer full")
	ErrNegativeCount     = errors.New("bfe_bufio: negative count")
)

type Reader struct {
	buf          []byte
	rd           io.Reader
	r, w         int
	err          error
	lastByte     int
	lastRuneSize int

	TotalRead int
}

const minReadBufferSize = 16

func NewReaderSize(rd io.Reader, size int) *Reader {
	b, ok := rd.(*Reader)
	if ok && len(b.buf) >= size {
		return b
	}

	if size < minReadBufferSize {
		size = minReadBufferSize
	}

	r := new(Reader)
	r.reset(make([]byte, size), rd)
	return r
}

func NewReader(rd io.Reader) *Reader {
	return NewReaderSize(rd, defaultBufSize)
}

func (b *Reader) Reset(r io.Reader) {
	b.reset(b.buf, r)
}

func (b *Reader) reset(buf []byte, r io.Reader) {
	*b = Reader{
		buf:          buf,
		rd:           r,
		lastByte:     -1,
		lastRuneSize: -1,
		TotalRead:    0,
	}
}
func (b *Reader) Read(p []byte) (n int, err error) {
	//todo
	return
}
