package lex

import "kodae/internal/token"

// Next returns the next non-comment token. Newlines are explicit NEWLINE tokens.
func (l *Lexer) Next() token.Token {
	for {
		l.skipHSpace()
		if l.ch == '\n' {
			line, col := l.line, l.col+1
			l.advance()
			return token.Token{Type: token.NEWLINE, Literal: "\n", Line: line, Col: col}
		}
		if l.isLineCommentStart() {
			l.skipLineComment()
			continue
		}
		if l.ch == 0 {
			return token.Token{Type: token.EOF, Line: l.line, Col: l.col + 1}
		}
		return l.nextAfterSpace()
	}
}

func (l *Lexer) nextAfterSpace() token.Token {
	line, col := l.line, l.col+1
	switch l.ch {
	case '(':
		l.advance()
		return token.Token{Type: token.LPAREN, Line: line, Col: col}
	case ')':
		l.advance()
		return token.Token{Type: token.RPAREN, Line: line, Col: col}
	case '{':
		l.advance()
		return token.Token{Type: token.LBRACE, Line: line, Col: col}
	case '}':
		l.advance()
		return token.Token{Type: token.RBRACE, Line: line, Col: col}
	case '[':
		l.advance()
		return token.Token{Type: token.LBRACK, Line: line, Col: col}
	case ']':
		l.advance()
		return token.Token{Type: token.RBRACK, Line: line, Col: col}
	case ',':
		l.advance()
		return token.Token{Type: token.COMMA, Line: line, Col: col}
	case ':':
		l.advance()
		return token.Token{Type: token.COLON, Line: line, Col: col}
	case ';':
		l.advance()
		return token.Token{Type: token.SEMI, Line: line, Col: col}
	case '+':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.PLUSEQ, Literal: "+=", Line: line, Col: col}
		}
		if l.peek(1) == '+' {
			l.advance()
			l.advance()
			return token.Token{Type: token.PLUSPLUS, Literal: "++", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.PLUS, Line: line, Col: col}
	case '-':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.MINUSEQ, Literal: "-=", Line: line, Col: col}
		}
		if l.peek(1) == '-' {
			l.advance()
			l.advance()
			return token.Token{Type: token.MINUSMINUS, Literal: "--", Line: line, Col: col}
		}
		if l.peek(1) == '>' {
			l.advance()
			l.advance()
			return token.Token{Type: token.ARROW, Literal: "->", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.MINUS, Line: line, Col: col}
	case '*':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.MULEQ, Literal: "*=", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.MUL, Line: line, Col: col}
	case '/':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.DIVEQ, Literal: "/=", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.DIV, Line: line, Col: col}
	case '%':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.MODEQ, Literal: "%=", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.MOD, Line: line, Col: col}
	case '=':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.EQ, Literal: "==", Line: line, Col: col}
		}
		if l.peek(1) == '>' {
			l.advance()
			l.advance()
			return token.Token{Type: token.FATARROW, Literal: "=>", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.ASSIGN, Line: line, Col: col}
	case '!':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.NEQ, Literal: "!=", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.NOT, Line: line, Col: col}
	case '<':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.LEQ, Literal: "<=", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.LT, Line: line, Col: col}
	case '>':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return token.Token{Type: token.GEQ, Literal: ">=", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.GT, Line: line, Col: col}
	case '&':
		if l.peek(1) == '&' {
			l.advance()
			l.advance()
			return token.Token{Type: token.AND, Literal: "&&", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.BITAND, Literal: "&", Line: line, Col: col}
	case '|':
		if l.peek(1) == '|' {
			l.advance()
			l.advance()
			return token.Token{Type: token.OR, Literal: "||", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.BITOR, Literal: "|", Line: line, Col: col}
	case '^':
		l.advance()
		return token.Token{Type: token.BITXOR, Literal: "^", Line: line, Col: col}
	case '~':
		l.advance()
		return token.Token{Type: token.BITNOT, Literal: "~", Line: line, Col: col}
	case '.':
		if l.peek(1) == '.' {
			l.advance()
			if l.peek(1) == '.' {
				l.advance()
				l.advance()
				return token.Token{Type: token.ELLIPSIS, Literal: "...", Line: line, Col: col}
			}
			l.advance()
			return token.Token{Type: token.DOTDOT, Literal: "..", Line: line, Col: col}
		}
		l.advance()
		return token.Token{Type: token.DOT, Line: line, Col: col}
	case '"':
		if l.peek(1) == '"' && l.peek(2) == '"' {
			return l.readMultilineString(line, col)
		}
		return l.readString(line, col)
	case '#':
		l.advance()
		return token.Token{Type: token.HASH, Line: line, Col: col}
	case '?':
		l.advance()
		return token.Token{Type: token.QUEST, Line: line, Col: col}
	default:
		if l.isIdentStart() {
			return l.readIdent(line, col)
		}
		if l.isDigit() {
			return l.readNumber(line, col)
		}
		return l.illegal(line, col)
	}
}
