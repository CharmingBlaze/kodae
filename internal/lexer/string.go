package lex

import (
	"strings"
	"unicode/utf8"

	"kodae/internal/token"
)

// readString reads a "..." string. Literal is the unescaped value for normal escapes.
func (l *Lexer) readString(line, col int) token.Token {
	if l.ch != '"' {
		return l.illegal(line, col)
	}
	l.advance()
	var b strings.Builder
	for l.ch != 0 && l.ch != '"' {
		if l.ch == '\\' {
			l.advance()
			if l.ch == 0 {
				return token.Token{Type: token.ILLEGAL, Literal: "unclosed string", Line: line, Col: col}
			}
			switch l.ch {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '0':
				b.WriteByte(0)
			case '\\', '"':
				b.WriteByte(l.ch)
			case 'x':
				l.advance()
				if !l.isHex() {
					return token.Token{Type: token.ILLEGAL, Literal: "bad \\x", Line: line, Col: col}
				}
				v := 0
				for i := 0; i < 2 && l.isHex(); i++ {
					v = v*16 + hexValue(l.ch)
					l.advance()
				}
				b.WriteByte(byte(v))
				continue
			case 'u':
				// \u + 4 hex
				l.advance()
				if !l.isHex() {
					return token.Token{Type: token.ILLEGAL, Literal: "bad \\u", Line: line, Col: col}
				}
				var runeVal uint32
				for i := 0; i < 4 && l.isHex(); i++ {
					runeVal = runeVal*16 + uint32(hexValue(l.ch))
					l.advance()
				}
				if runeVal > 0x10FFFF {
					return token.Token{Type: token.ILLEGAL, Literal: "invalid \\u", Line: line, Col: col}
				}
				var runeBuf [4]byte
				n := utf8.EncodeRune(runeBuf[:], rune(runeVal))
				b.Write(runeBuf[:n])
				continue
			default:
				b.WriteByte(l.ch)
			}
			l.advance()
			continue
		}
		if l.ch == '\n' {
			return token.Token{Type: token.ILLEGAL, Literal: "newline in string", Line: line, Col: col}
		}
		b.WriteByte(l.ch)
		l.advance()
	}
	if l.ch != '"' {
		return token.Token{Type: token.ILLEGAL, Literal: "unclosed string", Line: line, Col: col}
	}
	l.advance()
	return token.Token{Type: token.STRLIT, Literal: b.String(), Line: line, Col: col}
}

func hexValue(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(10 + c - 'a')
	case c >= 'A' && c <= 'F':
		return int(10 + c - 'A')
	default:
		return 0
	}
}

// readMultilineString reads a """...""" string. Literal is the unescaped value.
func (l *Lexer) readMultilineString(line, col int) token.Token {
	// Skip the three quotes
	l.advance()
	l.advance()
	l.advance()

	var b strings.Builder
	for l.ch != 0 {
		if l.ch == '"' && l.peek(1) == '"' && l.peek(2) == '"' {
			// End of multiline string
			l.advance()
			l.advance()
			l.advance()
			return token.Token{Type: token.STRLIT, Literal: b.String(), Line: line, Col: col}
		}
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		b.WriteByte(l.ch)
		l.advance()
	}
	return token.Token{Type: token.ILLEGAL, Literal: "unclosed multiline string", Line: line, Col: col}
}
