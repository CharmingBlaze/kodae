package lex

import "clio/internal/token"

func (l *Lexer) readNumber(line, col int) token.Token {
	start := l.pos
	typ := token.INTLIT
	for l.isDigit() {
		l.advance()
	}
	if l.ch == '.' {
		typ = token.FLOATLIT
		l.advance()
		for l.isDigit() {
			l.advance()
		}
	}
	// Exponent: optional 1.2e+3
	if l.ch == 'e' || l.ch == 'E' {
		typ = token.FLOATLIT
		l.advance()
		if l.ch == '+' || l.ch == '-' {
			l.advance()
		}
		for l.isDigit() {
			l.advance()
		}
	}
	return token.Token{Type: typ, Literal: l.input[start:l.pos], Line: line, Col: col}
}
