// Package parser builds an AST from Kodae source.
package parser

import (
	"fmt"
	"strconv"

	"kodae/internal/ast"
	lex "kodae/internal/lexer"
	"kodae/internal/token"
)

// Parser holds lexer state and the current token.
type Parser struct {
	lex *lex.Lexer
	tok token.Token
	err error
	// forInHeader: while parsing "for x in <expr>" without parens, <expr> must not
	// consume a following `{` as a struct literal (e.g. "for item in items { print(...) }").
	forInHeader bool
}

// New creates a parser. The lexer must be fresh.
func New(l *lex.Lexer) *Parser {
	p := &Parser{lex: l}
	p.next()
	return p
}

// Parse runs the full parse and returns a program and the first error (if any).
func Parse(l *lex.Lexer) (*ast.Program, error) {
	p := New(l)
	pr := p.ParseProgram()
	if e := p.Err(); e != nil {
		return pr, e
	}
	return pr, nil
}

func (p *Parser) next() {
	p.tok = p.lex.Next()
	if p.tok.Type == token.ILLEGAL {
		if p.err == nil {
			p.err = fmt.Errorf("illegal at %d:%d: %q", p.tok.Line, p.tok.Col, p.tok.Literal)
		}
	}
}

// Err returns a parse/lex error if any.
func (p *Parser) Err() error { return p.err }

func (p *Parser) failf(format string, args ...any) {
	if p.err != nil {
		return
	}
	loc := fmt.Sprintf("%d:%d", p.tok.Line, p.tok.Col)
	p.err = fmt.Errorf(loc+": "+format, args...)
}

func (p *Parser) skipNewlines() {
	for p.tok.Type == token.NEWLINE {
		p.next()
	}
}

func (p *Parser) expect(t token.Type) {
	if p.tok.Type != t {
		p.failf("expected %s, have %s (%q)", t.String(), p.tok.Type.String(), p.tok.Literal)
		return
	}
	p.next()
}

// Optional helper for numeric int literals
func (p *Parser) intFromTok() (int64, error) {
	s := p.tok.Literal
	if len(s) > 2 && s[0] == '0' && (s[1] == 'b' || s[1] == 'B') {
		return strconv.ParseInt(s[2:], 2, 64)
	}
	return strconv.ParseInt(s, 0, 64)
}
