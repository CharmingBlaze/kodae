package lex

// skipHSpace advances past spaces, tabs, and \r. Newlines are not consumed here.
func (l *Lexer) skipHSpace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.advance()
	}
}

// isLineCommentStart is true for '--' or a single leading apostrophe.
func (l *Lexer) isLineCommentStart() bool {
	if l.ch == '\'' {
		return true
	}
	return l.ch == '-' && l.peek(1) == '-'
}

// skipLineComment consumes a line: either '--' or ' through the next newline (or EOF).
// ch is the first character of the comment token (' or first '-').
func (l *Lexer) skipLineComment() {
	switch l.ch {
	case '\'':
		l.advance()
	case '-':
		l.advance()
		if l.ch == '-' {
			l.advance()
		}
	}
	for l.ch != 0 && l.ch != '\n' {
		l.advance()
	}
}
