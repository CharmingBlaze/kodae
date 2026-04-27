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

// examples/result_minimal.clio: result[T], ok/err, .ok/.value/.err, ? propagate, and catch/return-catch.
func TestResultMinimalCompiles(t *testing.T) {
	t.Parallel()
	path := findExample(t, "result_minimal.clio")
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
	if !strings.Contains(c, "clio_res_i64") {
		t.Fatal("expected clio_res_i64 in C for result[int]")
	}
	if !strings.Contains(c, "c_rc_") {
		t.Fatal("expected c_rc_ lowered catch temps")
	}
}
