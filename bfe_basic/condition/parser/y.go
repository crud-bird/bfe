package parser

type condSymType struct {
	yys  int
	Node Node
	str  string
}

var condToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"IDENT",
	"LAND",
	"LOR",
	"LPAREN",
	"RPAREN",
	"NOT",
	"SEMICOLON",
	"BASICLIT",
	"COMMA",
	"BOOL",
	"STRING",
	"INT",
	"FLOAT",
	"IMAG",
	"COMMENT",
	"ILLEGAL",
}

var condStatenames = [...]string{}

const condEofCode = 1
const condErrCode = 2
const condInitialStackSize = 16

//line cond.y:96

// The parser expects the lexer to return 0 on EOF.  Give it a name
// for clarity.
const EOF = 0
