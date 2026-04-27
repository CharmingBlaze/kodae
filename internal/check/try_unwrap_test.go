package check

import (
	"strings"
	"testing"

	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestV1RejectsQuestionPropagation(t *testing.T) {
	const src = `fn f() { g()? }`
	p := parser.New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() == nil {
		t.Fatal("expected parse error for ? operator")
	}
	if !strings.Contains(p.Err().Error(), "not supported in Clio v1") {
		t.Fatalf("unexpected parse error: %v", p.Err())
	}
}
