package lex

import (
	"testing"

	"kodae/internal/token"
)

func TestNumber_Int(t *testing.T) {
	toks := readNoNL(t, "0 42 9")
	lits := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		if toks[i].Type != token.INTLIT {
			t.Fatalf("tok %d: want int got %v", i, toks[i])
		}
		lits = append(lits, toks[i].Literal)
	}
	if lits[0] != "0" || lits[1] != "42" {
		t.Fatalf("lits %v", lits)
	}
}

func TestNumber_Hex(t *testing.T) {
	toks := readNoNL(t, "0xFF 0X10 0x181818FF")
	for i := 0; i < 3; i++ {
		if toks[i].Type != token.INTLIT {
			t.Fatalf("i=%d: want int, got %v", i, toks[i].Type)
		}
	}
}

func TestNumber_Float(t *testing.T) {
	toks := readNoNL(t, "3.14 1.2e3 5E-2")
	for _, tk := range toks[:3] {
		if tk.Type != token.FLOATLIT {
			t.Fatalf("want float got %v lit=%q", tk.Type, tk.Literal)
		}
	}
}

func TestString_Escapes(t *testing.T) {
	// Kodae source: a string containing newline, tab, backslash, quote, and \x41 -> A
	src := "\x22" + // "
		"\\" + "n" +
		"\\" + "t" +
		"\\" + "\\" +
		"\\" + "\"" +
		"\\" + "x" + "41" +
		"\x22"
	toks := readNoNL(t, src)
	if toks[0].Type != token.STRLIT {
		t.Fatalf("got %v", toks[0].Type)
	}
	// Kodae: \n, \t, one backslash, escaped quote, \x41 -> A
	const want = "\n\t\\\"" + "A"
	if toks[0].Literal != want {
		t.Fatalf("got %q want %q", toks[0].Literal, want)
	}
}

func TestString_Unclosed(t *testing.T) {
	l := New(`"nope`) // no closing "
	for {
		tk := l.Next()
		if tk.Type == token.ILLEGAL {
			if tk.Literal == "" {
				t.Error("ILLEGAL with empty literal")
			}
			return
		}
		if tk.Type == token.EOF {
			t.Fatalf("expected ILLEGAL for unclosed string, got EOF")
		}
	}
}
