package lex

import (
	"testing"

	"clio/internal/token"
)

// readAll materializes the full token stream, including NEWLINE, until EOF.
func readAll(t *testing.T, src string) []token.Token {
	t.Helper()
	var out []token.Token
	l := New(src)
	for {
		tk := l.Next()
		if tk.Type == token.ILLEGAL {
			t.Fatalf("unexpected ILLEGAL: %q at %d:%d", tk.Literal, tk.Line, tk.Col)
		}
		out = append(out, tk)
		if tk.Type == token.EOF {
			return out
		}
	}
}

func readNoNL(t *testing.T, src string) []token.Token {
	t.Helper()
	var out []token.Token
	for _, tk := range readAll(t, src) {
		if tk.Type == token.NEWLINE {
			continue
		}
		out = append(out, tk)
	}
	return out
}

func typesOnly(toks []token.Token) []token.Type {
	tt := make([]token.Type, len(toks))
	for i := range toks {
		tt[i] = toks[i].Type
	}
	return tt
}
