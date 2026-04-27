package lex

import (
	"testing"

	"clio/internal/token"
)

func TestLineComment_ConsumesToEOL(t *testing.T) {
	// "let" after newline must tokenize, not be eaten by comment
	toks := readNoNL(t, "let  -- a comment with ) } weird stuff "+ "\n" + "x")
	want := []token.Type{token.LET, token.IDENT, token.EOF}
	if got := typesOnly(toks); !sliceEq(got, want) {
		t.Fatalf("types: got %v want %v (literals: %#v)", got, want, toks)
	}
	if toks[1].Literal != "x" {
		t.Errorf("ident literal: got %q", toks[1].Literal)
	}
}

func TestLineComment_OnlyLineThenCode(t *testing.T) {
	// Comment swallows the rest of the first line; '+' is on a new line.
	toks := readNoNL(t, "-- to EOL, nothing here\n+")
	if got := typesOnly(toks); !sliceEq(got, []token.Type{token.PLUS, token.EOF}) {
		t.Fatalf("got %v, want + EOF", got)
	}
}

func TestLineComment_SingleQuote(t *testing.T) {
	toks := readNoNL(t, "1 ' ignore ) everything here\n+")
	if got := typesOnly(toks); !sliceEq(got, []token.Type{token.INTLIT, token.PLUS, token.EOF}) {
		t.Fatalf("got %v, want 1 + EOF (types)", got)
	}
}

func TestLineComment_SingleQuote_Midline(t *testing.T) {
	toks := readNoNL(t, "a ' rest is gone\nb")
	if got := typesOnly(toks); !sliceEq(got, []token.Type{token.IDENT, token.IDENT, token.EOF}) {
		t.Fatalf("got %v, want a b (types)", got)
	}
	if toks[0].Literal != "a" || toks[1].Literal != "b" {
		t.Fatalf("lits %q %q", toks[0].Literal, toks[1].Literal)
	}
}

func sliceEq(a, b []token.Type) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
