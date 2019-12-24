package parser

type Token int

const IDENT = 57346
const LAND = 57347
const LOR = 57348
const LPAREN = 57349
const RPAREN = 57350
const NOT = 57351
const SEMICOLON = 57352
const BASICLIT = 57353
const COMMA = 57354
const BOOL = 57355
const STRING = 57356
const INT = 57357
const FLOAT = 57358
const IMAG = 57359
const COMMENT = 57360
const ILLEGAL = 57361

var keywords = []string{
	"break",
	"case",
	"chan",
	"const",
	"continue",

	"default",
	"defer",
	"else",
	"fallthrough",
	"for",

	"func",
	"go",
	"goto",
	"if",
	"import",

	"interface",
	"map",
	"package",
	"range",
	"return",

	"select",
	"struct",
	"switch",
	"type",
	"var",
}

var tokens = map[Token]string{
	IDENT:     "IDENT",
	LAND:      "LAND",
	LOR:       "LOR",
	LPAREN:    "LPAREN",
	RPAREN:    "RPAREN",
	NOT:       "NOT",
	SEMICOLON: "SEMICOLON",
	BASICLIT:  "BASICLIT",
	COMMA:     "COMMA",
	BOOL:      "BOOL",
	STRING:    "STRING",
	INT:       "INT",
	FLOAT:     "FLOAT",
	IMAG:      "IMAG",
	COMMENT:   "COMMENT",
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
}

var symbols = map[Token]string{
	IDENT:     "",
	LAND:      "&&",
	LOR:       "||",
	LPAREN:    "(",
	RPAREN:    ")",
	NOT:       "!",
	SEMICOLON: ";",
	BASICLIT:  "BASICLIT",
	COMMA:     ",",
	BOOL:      "BOOL",
	STRING:    "STRING",
	INT:       "INT",
	FLOAT:     "FLOAT",
	IMAG:      "IMAG",
	COMMENT:   "//",
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
}

func (t Token) Symbol() string {
	return symbols[t]
}

func (t Token) String() string {
	return tokens[t]
}

func Lookup(ident string) Token {
	for _, keyword := range keywords {
		if ident == keyword {
			return ILLEGAL
		}
	}

	return IDENT
}
