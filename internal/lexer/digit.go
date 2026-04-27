package lex

func (l *Lexer) isDigit() bool { return l.ch >= '0' && l.ch <= '9' }
func (l *Lexer) isHex() bool {
	return l.isDigit() || (l.ch >= 'a' && l.ch <= 'f') || (l.ch >= 'A' && l.ch <= 'F')
}
