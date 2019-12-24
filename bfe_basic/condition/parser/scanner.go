package parser

import (
	"fmt"
	"go/token"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

type Scanner struct {
	file *token.File
	dir  string
	src  []byte
	err  ErrorHandler

	ch         rune
	offset     int
	rdOffset   int
	lineOffset int

	ErrCount int
}

const bom = 0xFEFF

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler) {
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size(%d) does not match src len(%d)", file.Size(), len(src)))
	}

	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.ErrCount = 0

	s.next()
	if s.ch == bom {
		s.next()
	}
}

func (s *Scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Pos(offs), msg)
	}
	s.ErrCount++
}

func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal charactor NUL")
		case r >= 0x80:
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal utf-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}

		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch -= 1
	}
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\r' || s.ch == '\n' {
		s.next()
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '-' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}

	return string(s.src[offs:s.offset])
}

func stripCR(b []byte) []byte {
	c := make([]byte, len(b))
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			c[i] = ch
			i++
		}
	}

	return c[:i]
}

func (s *Scanner) scanString() string {
	offs := s.offset - 1
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "strings literal not terminated")
			break
		}
		s.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			s.scanEspace('"')
		}
	}

	return string(s.src[offs+1 : s.offset-1])
}

func (s *Scanner) scanEspace(quote rune) bool {
	offs := s.offset
	var n int
	var base, max uint32

	switch s.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		s.next()
		return true
	case '1', '2', '3', '4', '5', '6', '7', '0':
		n, base, max = 3, 8, 255
	case 'x':
		s.next()
		n, base, max = 2, 16, 255
	case 'u':
		s.next()
		n, base, max = 4, 6, unicode.MaxRune
	case 'U':
		s.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "unkown escape sequence"
		if s.ch < 0 {
			msg = "escape sequence not terminated"
		}
		s.error(offs, msg)
		return false
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(s.ch))
		if d >= base {
			msg := fmt.Sprintf("illegal charactor %#U in escape sequence", s.ch)
			if s.ch < 0 {
				msg = "escape sequence not terminated"
			}
			s.error(s.offset, msg)
			return false
		}
		x = x*base + d
		s.next()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		s.error(offs, "espace sequence os invalid Unicode point")
		return false
	}

	return true
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}

	return 16
}

func (s *Scanner) scanRawString() string {
	offs := s.offset - 1
	hasCR := false
	for {
		ch := s.ch
		if ch < 0 {
			s.error(offs, "raw string literal not terminated")
			break
		}
		s.next()
		if ch == '`' {
			break
		}
		if ch == '\r' {
			hasCR = true
		}
	}

	lit := s.src[offs+1 : s.offset-1]
	if hasCR {
		lit = stripCR(lit)
	}

	return string(lit)
}

func (s *Scanner) scanCommet() string {
	offs := s.offset - 1
	hasCR := false

	s.next()
	for s.ch != '\n' && s.ch >= 0 {
		if s.ch == '\r' {
			hasCR = true
		}
		s.next()
	}

	lit := s.src[offs:s.offset]
	if hasCR {
		lit = stripCR(lit)
	}

	return string(lit)
}

func (s *Scanner) scanMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func (s *Scanner) scanNumber(point bool) (Token, string) {
	offs := s.offset
	tok := INT

	if point {
		offs--
		tok = FLOAT
		s.scanMantissa(10)
		goto exponent
	}

	if s.ch == '0' {
		offs := s.offset
		s.next()
		if s.ch == 'x' || s.ch == 'X' {
			s.next()
			s.scanMantissa(16)
			if s.offset-offs <= 2 {
				s.error(offs, "illegal hexadecimal number")
			}
		} else {
			digit := false
			s.scanMantissa(8)
			if s.ch == '8' || s.ch == '9' {
				digit = true
				s.scanMantissa(10)
			}
			if s.ch == '.' || s.ch == 'e' || s.ch == 'E' || s.ch == 'i' {
				goto fraction
			}
			if digit {
				s.error(offs, "illegal octal number")
			}
		}

		goto exit
	}
	s.scanMantissa(10)

fraction:
	if s.ch == '.' {
		tok = FLOAT
		s.scanMantissa(10)
	}

exponent:
	if s.ch == 'e' || s.ch == 'E' {
		tok = FLOAT
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		s.scanMantissa(10)
	}

	if s.ch == 'i' {
		tok = IMAG
		s.next()
	}

exit:
	return Token(tok), string(s.src[offs:s.offset])
}

func (s *Scanner) Scan() (pos token.Pos, tok Token, lit string) {
acanAgain:
	s.skipWhitespace()
	pos = s.file.Pos(s.offset)

	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanIdentifier()
		if len(lit) > 1 {
			tok = Lookup(lit)
			if tok != IDENT {
				tok = ILLEGAL
				s.error(s.offset, "keyword can't be used")
			}

			if lit == "true" || lit == "false" {
				tok = BOOL
			}
		} else {
			tok = IDENT
		}

	case '0' <= ch && ch <= '9':
		tok, lit = s.scanNumber(false)

	default:
		s.next()
		switch ch {
		case -1:
			tok = EOF
		case ';':
			tok = SEMICOLON
			lit = ";"
		case '"':
			tok = STRING
			lit = s.scanString()
		case '`':
			tok = STRING
			lit = s.scanRawString()
		case '(':
			tok = LPAREN
			lit = "("
		case ')':
			tok = RPAREN
			lit = ")"
		case '!':
			tok = NOT
			lit = "!"
		case ',':
			tok = COMMA
			lit = ","
		case '&':
			if s.ch == '&' {
				s.next()
				tok = LAND
				lit = "&&"
			} else {
				tok = ILLEGAL
				lit = string(ch)
			}
		case '|':
			if s.ch == '|' {
				s.next()
				tok = LOR
				lit = "||"
			} else {
				tok = ILLEGAL
				lit = string(ch)
			}
		case '/':
			if s.ch == '/' {
				tok = COMMENT
				lit = s.scanCommet()
				goto acanAgain
			} else {
				tok = ILLEGAL
				lit = string(ch)
			}
		default:
			tok = ILLEGAL
			lit = string(ch)
		}
	}

	return
}
