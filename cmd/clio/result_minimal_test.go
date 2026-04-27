package main

import (
	"os"
	"strings"
	"testing"

	"clio/internal/check"
	"clio/internal/codegen"
	lexapi "clio/internal/lexer"
	"clio/internal/parser"
)

// examples/features.clio is the v1 language conformance sample.
func TestFeaturesExampleCompiles(t *testing.T) {
	t.Parallel()
	path := findExample(t, "features.clio")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.Parse(lexapi.New(string(b)))
	if err != nil {
		t.Fatal("parse:", err)
	}
	inf, err := check.Check(p)
	if err != nil {
		t.Fatal("check:", err)
	}
	c, err := codegen.EmitC(p, inf)
	if err != nil {
		t.Fatal("codegen:", err)
	}
	if !strings.Contains(c, "match") && !strings.Contains(c, "switch") {
		t.Fatal("expected generated C to include enum/match lowering")
	}
}
