package check

import (
	"strings"
	"testing"

	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestDiag_TypoName_Suggest(t *testing.T) {
	t.Parallel()
	const src = `fn main() {
  let score: int = 0
  print( str( scre ) )
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected error for typo scre")
	}
	if !strings.Contains(err.Error(), "did you mean") || !strings.Contains(err.Error(), "score") {
		t.Fatalf("want did-you-mean score, got: %v", err)
	}
}

func TestDiag_StructField_Suggest(t *testing.T) {
	t.Parallel()
	const src = `struct S { health: int }
fn main() {
  let s = S { health: 1 }
  print( str( s.heath ) )
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected error for wrong field heath")
	}
	if !strings.Contains(err.Error(), "did you mean") {
		t.Fatalf("expected suggestion, got: %v", err)
	}
}

func TestList_DotLen(t *testing.T) {
	t.Parallel()
	const src = `fn main() {
  let items: list[str] = ["a", "b"]
  print( str( items.len + len(items) ) )
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	if _, err := Check(pr); err != nil {
		t.Fatalf("check: %v", err)
	}
}
