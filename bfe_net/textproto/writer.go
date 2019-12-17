package textproto

import (
	"fmt"
	"github.com/crud-bird/bfe/bfe_bufio"
	"io"
)

type Writer struct {
	W   *bfe_bufio.Writer
	dot *dotWriter
}

func NewWriter(w *bfe_bufio.Writer) *Writer {
	return &Writer{W: w}
}

var crnl = []byte{'\r', '\n'}
var dotcrnl = []byte{'.', '\r', '\n'}

func (w *Writer) PrintfLine(format string, args ...interface{}) error {
	w.closeDot()
	fmt.Fprintf(w.W, format, args...)
	w.W.Write(crnl)

	return w.W.Flush()
}

func (w *Writer) DotWriter() io.WriteCloser {
	w.closeDot()
	w.dot = &dotWriter{w: w}

	return w.dot
}

func (w *Writer) closeDot() {
	if w.dot != nil {
		w.dot.Close()
	}
}

type dotWriter struct {
	w     *Writer
	state int
}

const (
	wstateBeginLine = iota
	wstateCR
	wstateData
)

func (d *dotWriter) Write(b []byte) (n int, err error) {
	bw := d.w.W
	for n < len(b) {
		c := b[n]
		switch d.state {
		case wstateBeginLine:
			d.state = wstateData
			if c == '.' {
				bw.WriteByte('.')
			}
			fallthrough

		case wstateData:
			if c == '\r' {
				d.state = wstateCR
			}

			if c == '\n' {
				bw.WriteByte('\r')
				d.state = wstateBeginLine
			}

		case wstateCR:
			d.state = wstateData
			if c == '\n' {
				d.state = wstateBeginLine
			}
		}

		if err = bw.WriteByte(c); err != nil {
			break
		}
		n++
	}

	return
}

func (d *dotWriter) Close() error {
	if d.w.dot == d {
		d.w.dot = nil
	}

	bw := d.w.W
	switch d.state {
	default:
		bw.WriteByte('\r')
	case wstateCR:
		bw.WriteByte('\n')
	case wstateBeginLine:
		bw.Write(dotcrnl)
	}

	return bw.Flush()
}
