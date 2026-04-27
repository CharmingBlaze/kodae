// Package lex tokenizes Clio source.
package lex

import "clio/internal/token"

// Lexer walks input byte-by-byte (UTF-8 string literals are handled as byte sequences).
type Lexer struct {
	input string
	pos   int
	line  int
	col   int
	ch    byte
}

// New creates a Lexer. Empty input yields immediate EOF.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	if len(input) == 0 {
		l.ch = 0
	} else {
		l.ch = input[0]
	}
	return l
}

// atEOF reports whether there is no current character.
func (l *Lexer) atEOF() bool { return l.ch == 0 }

// Char returns the current character (0 at EOF). Useful for tests and peek parsers.
func (l *Lexer) Char() byte { return l.ch }

// peek returns a byte at offset from the current l.pos. peek(0) is l.ch; peek(1) is the next byte in input.
func (l *Lexer) peek(ahead int) byte {
	n := l.pos + ahead
	if n >= 0 && n < len(l.input) {
		return l.input[n]
	}
	return 0
}

// advance consumes the current ch and moves to the next, or to EOF.
// Invariant: when not past end, l.ch == l.input[l.pos]; at EOF, l.ch == 0 and l.pos == len(input).
func (l *Lexer) advance() {
	if l.ch == 0 {
		return
	}
	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	l.pos++
	if l.pos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.pos]
	}
}

func (l *Lexer) illegal(line, col int) token.Token {
	c := l.ch
	l.advance()
	if c == 0 {
		return token.Token{Type: token.ILLEGAL, Literal: "<EOF>", Line: line, Col: col}
	}
	return token.Token{Type: token.ILLEGAL, Literal: string(c), Line: line, Col: col}
}
