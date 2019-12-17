package textproto

import (
	"bytes"
	"github.com/crud-bird/bfe/bfe_bufio"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

type Reader struct {
	R   *bfe_bufio.Reader
	dot *dotReader
	buf []byte
}

func NewReader(r *bfe_bufio.Reader) *Reader {
	return &Reader{
		R: r,
	}
}

func (r *Reader) ReadLine() (string, error) {
	line, err := r.readLineSlice()
	return string(line), err
}

func (r *Reader) ReadLineBytes() ([]byte, error) {
	line, err := r.readLineSlice()
	if line != nil {
		buf := make([]byte, len(line))
		copy(buf, line)
		line = buf
	}

	return line, err
}

func (r *Reader) readLineSlice() ([]byte, error) {
	r.closeDot()
	var line []byte
	for {
		l, more, err := r.R.ReadLine()
		if err != nil {
			return nil, err
		}

		if line == nil && !more {
			return l, nil
		}

		line = append(line, l...)
		if !more {
			break
		}
	}

	return line, nil
}

func (r *Reader) ReadContinuedLine() (string, error) {
	line, err := r.readContinuedLineSlice()
	return string(line), err
}

func trim(s []byte) []byte {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	n := len(s)
	for n > i && (s[n-1] == ' ' || s[n-1] == '\t') {
		n--
	}

	return s[i:n]
}

func (r *Reader) ReadContinuedLineBytes() ([]byte, error) {
	line, err := r.readContinuedLineSlice()
	if line != nil {
		buf := make([]byte, len(line))
		copy(buf, line)
		line = buf
	}

	return line, err
}

func (r *Reader) readContinuedLineSlice() ([]byte, error) {
	line, err := r.readLineSlice()
	if err != nil {
		return nil, err
	}

	if len(line) == 0 {
		return line, nil
	}

	if r.R.Buffered() > 1 {
		peek, err := r.R.Peek(1)
		if err == nil && isASCIILetter(peek[0]) {
			return trim(line), nil
		}
	}

	r.buf = append(r.buf[:0], trim(line)...)

	for r.skipSpace() > 0 {
		line, err := r.readLineSlice()
		if err != nil {
			break
		}
		r.buf = append(r.buf, ' ')
		r.buf = append(r.buf, line...)
	}

	return r.buf, nil
}

func (r *Reader) skipSpace() int {
	n := 0
	for {
		c, err := r.R.ReadByte()
		if err != nil {
			break
		}

		if c != ' ' && c != '\t' {
			r.R.UnreadByte()
			break
		}
		n++
	}

	return n
}

func (r *Reader) readCodeLine(expectCode int) (code int, continued bool, message string, err error) {
	line, err := r.ReadLine()
	if err != nil {
		return
	}

	return parseCodeLine(line, expectCode)
}

func parseCodeLine(line string, expectCode int) (code int, continued bool, message string, err error) {
	if len(line) < 4 || line[3] != ' ' && line[3] != '-' {
		err = ProtocolError("short response: " + line)
		return
	}

	continued = line[3] == '-'
	code, err = strconv.Atoi(line[0:3])
	if err != nil || code < 100 {
		err = ProtocolError("invalid response code: " + line)
	}

	message = line[4:]
	if 1 <= expectCode && expectCode < 10 && code/100 != expectCode ||
		10 <= expectCode && expectCode < 100 && code/10 != expectCode ||
		100 <= expectCode && expectCode < 1000 && code != expectCode {
		err = &Error{code, message}
	}

	return
}

func (r *Reader) ReadCodeLine(expectCode int) (code int, message string, err error) {
	code, continued, message, err := r.readCodeLine(expectCode)
	if err == nil && continued {
		err = ProtocolError("unexpected multi-line response: " + message)
	}

	return
}

func (r *Reader) ReadResponse(expectCode int) (code int, message string, err error) {
	code, continued, message, err := r.readCodeLine(expectCode)
	for err == nil && continued {
		line, err := r.ReadLine()
		if err != nil {
			return 0, "", err
		}

		var code2 int
		var moreMessage string
		code2, continued, moreMessage, err = parseCodeLine(line, expectCode)
		if err != nil || code2 != code {
			message += "\n" + strings.TrimRight(line, "\r\n")
			continued = true
			continue
		}
		message += "\n" + moreMessage
	}

	return
}

func (r *Reader) DotReader() io.Reader {
	r.closeDot()
	r.dot = &dotReader{r: r}

	return r.dot
}

type dotReader struct {
	r     *Reader
	state int
}

func (d *dotReader) Read(b []byte) (n int, err error) {
	const (
		stateBeginLine = iota
		stateDot
		stateDotCR
		stateCR
		stateData
		stateEOF
	)

	br := d.r.R
	for n < len(b) && d.state != stateEOF {
		var c byte
		c, err = br.ReadByte()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			break
		}

		switch d.state {
		case stateBeginLine:
			if c == '.' {
				d.state = stateDot
				continue
			}

			if c == '\r' {
				d.state = stateCR
				continue
			}

			d.state = stateData

		case stateDot:
			if c == '\r' {
				d.state = stateDotCR
				continue
			}

			if c == '\n' {
				d.state = stateEOF
				continue
			}
			d.state = stateData

		case stateDotCR:
			if c == '\n' {
				d.state = stateEOF
				continue
			}

			br.UnreadByte()
			c = '\r'
			d.state = stateData

		case stateCR:
			if c == '\n' {
				d.state = stateBeginLine
				break
			}

			br.UnreadByte()
			c = '\r'
			d.state = stateData

		case stateData:
			if c == '\r' {
				d.state = stateCR
				continue
			}

			if c == '\n' {
				d.state = stateBeginLine
			}
		}

		b[n] = c
		n++
	}

	if err == nil && d.state == stateEOF {
		err = io.EOF
	}

	if err != nil && d.r.dot == d {
		d.r.dot = nil
	}

	return
}

func (r *Reader) closeDot() {
	if r.dot == nil {
		return
	}

	buf := make([]byte, 128)
	for r.dot != nil {
		r.dot.Read(buf)
	}
}

func (r *Reader) ReadDotBytes() ([]byte, error) {
	return ioutil.ReadAll(r.DotReader())
}

func (r *Reader) ReadDotLine() ([]string, error) {
	var v []string
	var err error
	for {
		var line string
		line, err = r.ReadLine()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			break
		}

		if len(line) > 0 && line[0] == '.' {
			if len(line) == 1 {
				break
			}
			line = line[1:]
		}

		v = append(v, line)
	}

	return v, err
}

func (r *Reader) ReadMIMEHeader() (MIMEHeader, error) {
	header, _, err := r.ReadMIMEHeaderAndKeys()
	return header, err
}

func (r *Reader) ReadMIMEHeaderAndKeys() (MIMEHeader, MIMEKeys, error) {
	var strs []string
	hint := r.upcomingHeaderNewlines()
	if hint > 0 {
		hint += 10
		strs = make([]string, hint)
	}

	m := make(MIMEHeader, hint)
	mkeys := make(MIMEKeys, 0, hint)
	for {
		kv, err := r.readContinuedLineSlice()
		if len(kv) == 0 {
			return m, mkeys, err
		}

		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			return m, mkeys, ProtocolError("malformed MIME header line: " + string(kv))
		}

		endKey := i
		for endKey > 0 && kv[endKey-1] == ' ' {
			endKey--
		}

		key := canonicalMIMEHeaderKey(kv[:endKey])

		i++
		for i < len(kv) && (kv[i] == ' ' || kv[i] == '\t') {
			i++
		}
		value := string(kv[i:])

		vv := m[key]
		if vv == nil && len(strs) > 0 {
			vv, strs = strs[:1], strs[1:]
			vv[0] = value
			m[key] = vv
		} else {
			m[key] = append(vv, value)
		}

		mkeys = append(mkeys, key)

		if err != nil {
			return m, mkeys, err
		}
	}
}

func (r *Reader) upcomingHeaderNewlines() (n int) {
	r.R.Peek(1)
	s := r.R.Buffered()
	if s == 0 {
		return
	}

	peek, _ := r.R.Peek(s)
	for len(peek) > 0 {
		i := bytes.IndexByte(peek, '\n')
		if i < 3 {
			return
		}
		n++
		peek = peek[i+1:]
	}

	return
}

func CanonicalMIMEHeaderKey(s string) string {
	upper := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if upper && 'a' <= c && c <= 'z' {
			return canonicalMIMEHeaderKey([]byte(s))
		}
		if !upper && 'A' <= c && c <= 'Z' {
			return canonicalMIMEHeaderKey([]byte(s))
		}
		upper = c == '-'
	}

	return s
}

func canonicalMIMEHeaderKey(a []byte) string {
	return canonicalMIMEHeaderKeyOriginal(a)
}

const toLower = 'a' - 'A'

func canonicalMIMEHeaderKeyOriginal(a []byte) string {
	upper := true
	lo := 0
	hi := len(commonHeaders)
	for i := 0; i < len(a); i++ {
		c := a[i]
		if c == ' ' {
			c = '-'
		} else if upper && 'a' <= c && c <= 'z' {
			c -= toLower
		} else if !upper && 'A' <= c && c <= 'Z' {
			c += toLower
		}

		a[i] = c
		upper = c == '-'

		if lo < hi {
			for lo < hi && (len(commonHeaders[lo]) <= i) || commonHeaders[lo][i] < c {
				lo++
			}
			for hi > lo && commonHeaders[hi-1][i] > c {
				hi--
			}
		}
	}

	if lo < hi && len(commonHeaders[lo]) == len(a) {
		return commonHeaders[lo]
	}

	return string(a)
}

var commonHeaders = []string{
	"Accept",
	"Accept-Charset",
	"Accept-Encoding",
	"Accept-Language",
	"Accept-Ranges",
	"Cache-Control",
	"Cc",
	"Connection",
	"Content-Id",
	"Content-Language",
	"Content-Length",
	"Content-Transfer-Encoding",
	"Content-Type",
	"Cookie",
	"Date",
	"Dkim-Signature",
	"Etag",
	"Expires",
	"From",
	"Host",
	"If-Modified-Since",
	"If-None-Match",
	"In-Reply-To",
	"Last-Modified",
	"Location",
	"Message-Id",
	"Mime-Version",
	"Pragma",
	"Received",
	"Return-Path",
	"Server",
	"Set-Cookie",
	"Subject",
	"To",
	"User-Agent",
	"Via",
	"X-Forwarded-For",
	"X-Imforwards",
	"X-Powered-By",
}
