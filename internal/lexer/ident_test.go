package lex

import "testing"

// Regresses advance() leaving l.pos on the last byte when the source is a single id character.
func TestReadIdent_SingleCharAtEndOfFile(t *testing.T) {
	toks := readNoNL(t, "x")
	if len(toks) != 2 {
		t.Fatalf("toks %#v", toks)
	}
	if toks[0].Literal != "x" {
		t.Fatalf("literal: got %q want x", toks[0].Literal)
	}
}

func TestReadIdent_OneLineProgram(t *testing.T) {
	toks := readNoNL(t, "a")
	if toks[0].Literal != "a" {
		t.Fatalf("got %q", toks[0].Literal)
	}
}
