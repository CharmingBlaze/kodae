package lex

import (
	"testing"

	"clio/internal/token"
)

func TestNext_HelloWorldish(t *testing.T) {
	src := `fn main() {
  let x = 1
  print("hi $x")
}`
	l := New(src)
	var toks []token.Type
	for {
		tk := l.Next()
		if tk.Type == token.EOF {
			toks = append(toks, tk.Type)
			break
		}
		if tk.Type == token.ILLEGAL {
			t.Fatalf("illegal: %q at %d:%d", tk.Literal, tk.Line, tk.Col)
		}
		// skip NEWLINE for this snapshot
		if tk.Type == token.NEWLINE {
			continue
		}
		toks = append(toks, tk.Type)
	}
	// fn main ( ) { let x = 1 print ( "hi $x" ) }
	want := []token.Type{
		token.FN, token.IDENT, token.LPAREN, token.RPAREN, token.LBRACE,
		token.LET, token.IDENT, token.ASSIGN, token.INTLIT,
		token.IDENT, token.LPAREN, token.STRLIT, token.RPAREN,
		token.RBRACE, token.EOF,
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(toks), len(want), toks)
	}
	for i := range want {
		if toks[i] != want[i] {
			t.Fatalf("idx %d: got %v want %v", i, toks[i], want[i])
		}
	}
}
