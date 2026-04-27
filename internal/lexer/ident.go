package lex

import "clio/internal/token"

func (l *Lexer) isIdentStart() bool {
	return l.ch == '_' || (l.ch >= 'a' && l.ch <= 'z') || (l.ch >= 'A' && l.ch <= 'Z')
}

func (l *Lexer) isIdentCont() bool {
	return l.isIdentStart() || l.isDigit()
}

func (l *Lexer) readIdent(line, col int) token.Token {
	start := l.pos
	for l.isIdentCont() {
		l.advance()
	}
	lit := l.input[start:l.pos]
	return token.Token{Type: token.Lookup(lit), Literal: lit, Line: line, Col: col}
}
