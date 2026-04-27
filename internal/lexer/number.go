package lex

import "clio/internal/token"

func (l *Lexer) readNumber(line, col int) token.Token {
	start := l.pos
	// Hex: 0xFF, 0X1a2b3c (not `0..2` range — handled below)
	if l.ch == '0' && (l.peek(1) == 'x' || l.peek(1) == 'X') {
		l.advance()
		l.advance()
		hexStart := l.pos
		for l.isHex() {
			l.advance()
		}
		if l.pos == hexStart {
			return l.illegal(line, col)
		}
		return token.Token{Type: token.INTLIT, Literal: l.input[start:l.pos], Line: line, Col: col}
	}
	typ := token.INTLIT
	for l.isDigit() {
		l.advance()
	}
	// Integer range: `0..n` or `1..9` must not be lexed as `0.` (float) + `..`.
	// If the next two runes are `..`, the integral prefix is a whole int literal.
	if l.ch == '.' && l.peek(1) == '.' {
		return token.Token{Type: token.INTLIT, Literal: l.input[start:l.pos], Line: line, Col: col}
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
