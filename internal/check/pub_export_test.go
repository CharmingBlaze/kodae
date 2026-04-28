package check

import (
	"strings"
	"testing"

	lex "kodae/internal/lexer"
	"kodae/internal/parser"
)

func TestPubFn_RejectsListInExport(t *testing.T) {
	const src = `pub fn bad(xs: list[int]) int { return len(xs) }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected exportability error for list in pub fn signature")
	}
	if !strings.Contains(err.Error(), "not exportable") {
		t.Fatalf("unexpected error: %v", err)
	}
}
