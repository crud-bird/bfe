package bfe_bufio

import (
	"bytes"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"unicode/utf8"
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

var errNegativeRead = errors.New("bfe_bufio: reader returned negative count from Read")

func (b *Reader) fill() {
	if b.r > 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}

	n, err := b.rd.Read(b.buf[b.w:])
	if n < 0 {
		panic(errNegativeRead)
	}

	if (b.w + n) > len(b.buf) {
		logrus.Warnf("fill(), len(buf) = %d, b.r = %d, b.w = %d, n = %d", len(b.buf), b.r, b.w, n)
	}

	b.w += n
	if err != nil {
		b.err = err
	}
}

func (b *Reader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

func (b *Reader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrNegativeCount
	}
	if n > len(b.buf) {
		return nil, ErrBufferFull
	}

	for b.w-b.r < n && b.err == nil {
		b.fill()
	}
	m := b.w - b.r
	if m > n {
		m = n
	}
	var err error
	if m < n {
		err = b.readErr()
		if err == nil {
			err = ErrBufferFull
		}
	}
	return b.buf[b.r : b.r+m], err
}

func (b *Reader) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, b.readErr()
	}
	if b.w == b.r {
		if b.err != nil {
			return 0, b.readErr()
		}
		if len(p) >= len(b.buf) {
			n, b.err = b.rd.Read(p)
			if n > 0 {
				b.lastByte = int(p[n-1])
				b.lastRuneSize = -1
				b.TotalRead += n
			}

			return n, b, readErr()
		}
		b.fill()
		if b.w == b.r {
			return 0, b.readErr()
		}
	}

	if n > b.w-b.r {
		n = b.w - b.r
	}

	if b.r > len(b.buf) || (b.r+n) > len(b.buf) {
		logrus.Warnf("Read(): len(buf) = %d, b.r = %d, b.w = %d, n = %d", len(b.buf), b.r, b.w, n)
	}

	copy(p[0:n], b.buf[b.r:])
	b.r += n
	b.lastByte = int(b.buf[b.r-1])
	b.lastRuneSize = -1
	b.TotalRead += n

	return n, nil
}

func (b *Reader) ReadByte() (c byte, err error) {
	b.lastRuneSize = -1
	for b.w == b.r {
		if b.err != nil {
			return 0, b.readErr()
		}
		b.fill()
	}

	c = b.buf[b.r]
	b.r++
	b.lastByte = int(c)
	b.TotalRead++

	return c, nil
}

func (b *Reader) UnreadByte() error {
	b.lastRuneSize = -1
	if b.r == b.w && b.lastByte >= 0 {
		b.w = 1
		b.r = 0
		b.buf[0] = byte(b.lastByte)
		b.lastByte = -1

		if b.TotalRead > 0 {
			b.TotalRead--
		}

		return nil
	}

	if b.r <= 0 {
		return ErrInvalidUnreadByte
	}
	b.r--
	b.lastByte = -1

	if b.TotalRead > 0 {
		b.TotalRead--
	}

	return nil
}

func (b *Reader) ReadRune() (r rune, size int, err error) {
	for b.r+utf8.UTFMax > b.w && !utf8.FullRune(b.buf[b.r:b.w]) && b.err == nil {
		b.fill()
	}
	b.lastRuneSize = -1
	if b.r == b.w {
		return 0, 0, b.readErr()
	}

	r, size = rune(b.buf[b.r]), 1
	if r >= 0x80 {
		r, size = utf8.DecodeRune(b.buf[b.r:b.w])
	}
	b.r += size
	b.lastByte = int(b.buf[b.r-1])
	b.lastRuneSize = size
	b.TotalRead += size

	return r, size, nil
}

func (b *Reader) UnreadRune() error {
	if b.lastRuneSize < 0 || b.r == 0 {
		return ErrInvalidUnreadRune
	}
	b.r -= b.lastRuneSize

	if b.TotalRead >= b.lastRuneSize {
		b.TotalRead -= b.lastRuneSize
	}

	b.lastByte = -1
	b.lastRuneSize = -1

	return nil
}

func (b *Reader) Buffered() int {
	return b.w - b.r
}

func (b *Reader) ReadSlice(delim byte) (line []byte, err error) {
	if i := bytes.IndexByte(b.buf[b.r:b.w], delim); i >= 0 {
		line1 := b.buf[b.r : b.r+i+1]
		b.r += i + 1
		b.TotalRead += i + 1

		return line1, nil
	}

	for {
		if b.err != nil {
			line := b.buf[b.r:b.w]
			b.TotalRead += b.w - b.r
			b.r = b.w
			return line, b.readErr()
		}

		n := b.Buffered()
		b.fill()

		if i := bytes.IndexByte(b.buf[n:b.w], delim); i >= 0 {
			line := b.buf[0 : n+i+1]
			b.r = n + i + 1
			b.TotalRead += i + 1
			return line, nil
		}

		if b.Buffered() >= len(b.buf) {
			b.TotalRead += len(b.buf)
			b.r = b.w
			return b.buf, ErrBufferFull
		}
	}
}

func (b *Reader) ReadLine() (line []byte, isPrefix bool, err error) {
	line, err = b.ReadSlice('\n')
	if err == ErrBufferFull {
		if len(line) > 0 && line[len(line)-1] == '\r' {
			if b.r == 0 {
				panic("bfe_bufio: tried to rewind past start of buffer")
			}
			b.r--
			line = line[:len(line)-1]
		}

		return line, true, nil
	}

	if len(line) == 0 {
		if err != nil {
			line = nil
		}
		return
	}
	err = nil

	if line[len(line)-1] == '\n' {
		drop := 1
		if len(line) > 1 && line[len(line)-2] == '\r' {
			drop = 2
		}
		line = line[:len(line)-drop]
	}

	return
}

func (b *Reader) ReadBytes(delim byte) (line []byte, err error) {
	var frag []byte
	var full [][]byte

	for {
		var e error
		frag, e = b.ReadSlice(delim)
		if e == nil {
			break
		}
		if e != ErrBufferFull {
			err = e
			break
		}

		buf := make([]byte, len(frag))
		copy(buf, frag)
		full = append(full, buf)
	}

	n := 0
	for i := range full {
		n += len(full[i])
	}
	n += len(frag)

	buf := make([]byte, n)
	n = 0
	for i := range full {
		n += copy(buf[n:], full[i])
	}
	copy(buf[n:], frag)

	return buf, err
}

func (b *Reader) ReadString(delim byte) (line string, err error) {
	bytes, err := b.ReadBytes(delim)
	line = string(bytes)
	return line, err
}

func (b *Reader) WriteTO(w io.Writer) (n int64, err error) {
	n, err = b.writeBuf(w)
	if err != nil {
		return
	}

	if r, ok := b.rd.(io.WriterTo); ok {
		m, err := r.WriteTo(w)
		if m > 0 {
			b.TotalRead += int(n)
		}
		n += m
		return n, err
	}

	for b.fill(); b.r < b.w; b.fill() {
		m, err := b.writeBuf(w)
		n += m
		if err != nil {
			return n, err
		}
	}

	if b.err == io.EOF {
		b.err = nil
	}

	return n, b.readErr()
}

func (b *Reader) writeBuf(w io.Writer) (int64, error) {
	n, err := w.Write(b.buf[b.r:b.w])
	b.r += n

	if n > 0 {
		b.TotalRead += n
	}

	return int64(n), err
}

type Writer struct {
	err        error
	buf        []byte
	n          int
	wr         io.Writer
	TotalWrite int
}

func NewWriterSize(w io.Writer, size int) *Writer {
	b, ok := w.(*Writer)
	if ok && len(b.buf) >= size {
		return b
	}

	if size <= 0 {
		size = defaultBufSize
	}

	return &Writer{
		buf:        make([]byte, size),
		wr:         w,
		TotalWrite: 0,
	}
}

func NewWriter(w io.Writer) *Writer {
	return NewWriterSize(w, defaultBufSize)
}

func (b *Writer) Reset(w io.Writer) {
	b.err = nil
	b.n = 0
	b.wr = w
	b.TotalWrite = 0
}

func (b *Writer) Flush() error {
	return b.flush()
}

func (b *Writer) flush() error {
	if b.err != nil {
		return b.err
	}

	if b.n == 0 {
		return nil
	}

	n, err := b.wr.Write(b.buf[0:b.n])
	if n < b.n && err == nil {
		err = io.ErrShortWrite
	}

	if err != nil {
		if n > 0 && n < b.n {
			copy(b.buf[0:b.n-n], b.buf[n:b.n])
		}
		b.n -= n
		b.err = err
		return err
	}

	b.n = 0
	return nil
}

func (b *Writer) Available() int {
	return len(b.buf) - b.n
}

func (b *Writer) Buffered() int {
	return b.n
}

func (b *Writer) Write(p []byte) (nn int, err error) {
	for len(p) > b.Available() && b.err == nil {
		var n int
		if b.Buffered() == 0 {
			n, b.err = b.wr.Write(p)
		} else {
			n = copy(b.buf[b.n:], p)
			b.n += n
			b.flush()
		}
		nn += n
		p = p[n:]
	}

	if b.err != nil {
		b.TotalWrite += nn
		return nn, b.err
	}

	n := copy(b.buf[b.n:], p)
	b.n += n
	nn += n
	b.TotalWrite += nn

	return nn, nil
}

func (b *Writer) WriteByte(c byte) error {
	if b.err != nil {
		return b.err
	}

	if b.Available() <= 0 && b.flush() != nil {
		return b.err
	}

	b.buf[b.n] = c
	b.n++
	b.TotalWrite++

	return nil
}

func (b *Writer) WriteRune(r rune) (size int, err error) {
	if r < utf8.RuneSelf {
		err = b.WriteByte(byte(r))
		if err != nil {
			return 0, err
		}

		return 1, nil
	}

	if b.err != nil {
		return 0, b.err
	}

	n := b.Available()
	if n < utf8.UTFMax {
		if b.flush(); b.err != nil {
			return 0, b.err
		}
		n = b.Available()
		if n < utf8.UTFMax {
			return b.WriteString(string(r))
		}
	}

	size = utf8.EncodeRune(b.buf[b.n:], r)
	b.TotalWrite += size
	b.n += size

	return size, nil
}

func (b *Writer) WriteString(s string) (int, error) {
	nn := 0
	for len(s) > b.Available() && b.err == nil {
		n := copy(b.buf[b.n:], s)
		b.n += n
		nn += n
		s = s[n:]
		b.flush()
	}

	if b.err != nil {
		b.TotalWrite += nn
		return nn, nil
	}

	n := copy(b.buf[b.n:], s)
	b.n += n
	nn += n
	b.TotalWrite += nn

	return nn, nil
}

func (b *Writer) ReadForm(r io.Reader) (n int64, err error) {
	if b.Buffered() == 0 {
		if w, ok := b.wr.(io.ReaderFrom); ok {
			n, err = w.ReadFrom(r)
			b.TotalWrite += int(n)
			return n, err
		}
	}

	var m int
	for {
		if b.Available() == 0 {
			if err1 := b.flush(); err1 != nil {
				return n, err1
			}
		}

		m, err = r.Read(b.buf[b.n:])
		if m == 0 {
			break
		}

		b.n += m
		if b.n > len(b.buf) {
			logrus.Warnf("ReadForm(): len(buf) = %d, n = %d, m = %d", len(b.buf), b.n, m)
		}

		n += int64(m)
		if err != nil {
			break
		}
	}

	if err == io.EOF {
		if b.Available() == 0 {
			err = b.flush()
		} else {
			err = nil
		}
	}

	b.TotalWrite += int(n)

	return n, err
}

type ReadWriter struct {
	*Reader
	*Writer
}

func NewReadWriter(r *Reader, w *Writer) *ReadWriter {
	return &ReadWriter{r, w}
}
