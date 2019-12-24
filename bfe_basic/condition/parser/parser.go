package parser

import (
	"fmt"
	"go/token"
	"strings"
	"sync"
)

type ErrorHandler func(token.Pos, string)

type Parser struct {
	fset    *token.FileSet
	scanner Scanner
	lexer   *condLex

	identList []*Ident
	errors    []Error
	ast       Node
}

type Error struct {
	pos token.Position
	msg string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s %s", e.pos, e.msg)
}

func (p *Parser) Init(src []byte) {
	p.fset = token.NewFileSet()
	p.errors = p.errors[0:0]
	p.identList = p.identList[0:0]

	file := p.fset.AddFile("", p.fset.Base(), len(src))
	p.scanner.Init(file, src, p.addError)
	p.lexer = &condLex{
		s:   &p.scanner,
		err: p.addError,
	}
}

func (p *Parser) addError(pos token.Pos, msg string) {
	p.errors = append(p.errors, Error{pos: p.fset.Position(pos), msg: msg})
}

func (p *Parser) Error() error {
	if len(p.errors) == 0 {
		return nil
	}

	return p.errors[0]
}

var parseLock sync.Mutex

func (p *Parser) Parse() {
	parseLock.Lock()
	defer parseLock.Unlock()

	condParse(p.lexer)
	p.ast = parseNode

	if len(p.errors) > 0 {
		return
	}

	Inspect(p.ast, p.collectVariable)
	Inspect(p.ast, p.primitiveCheck)
}

func (p Parser) String() string {
	var variables []string

	for _, ident := range p.identList {
		variables = append(variables, ident.Name)
	}

	var errors []string
	for _, err := range p.errors {
		errors = append(errors, err.Error())
	}

	return "names: " + strings.Join(variables, ",") + "\terrors: " + strings.Join(errors, ",")
}

func Parse(condStr string) (Node, []*Ident, error) {
	var p Parser

	p.Init([]byte(condStr))
	p.Parse()

	if err := p.Error(); err != nil {
		return nil, nil, err
	}

	return p.ast, p.identList, nil
}

//todo
