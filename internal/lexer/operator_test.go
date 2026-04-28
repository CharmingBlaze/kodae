package lex

import (
	"testing"

	"kodae/internal/token"
)

func TestTokenize_Operators(t *testing.T) {
	cases := []struct {
		src  string
		want []token.Type
	}{
		{"==", []token.Type{token.EQ, token.EOF}},
		{"!=", []token.Type{token.NEQ, token.EOF}},
		{"<=", []token.Type{token.LEQ, token.EOF}},
		{">=", []token.Type{token.GEQ, token.EOF}},
		{"&&", []token.Type{token.AND, token.EOF}},
		{"||", []token.Type{token.OR, token.EOF}},
		{"..", []token.Type{token.DOTDOT, token.EOF}},
		{"!", []token.Type{token.NOT, token.EOF}},
		{"=", []token.Type{token.ASSIGN, token.EOF}},
		{".", []token.Type{token.DOT, token.EOF}},
		{"0..2", []token.Type{token.INTLIT, token.DOTDOT, token.INTLIT, token.EOF}},
		{"10..20", []token.Type{token.INTLIT, token.DOTDOT, token.INTLIT, token.EOF}},
		{"+=", []token.Type{token.PLUSEQ, token.EOF}},
		{"=>", []token.Type{token.FATARROW, token.EOF}},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			toks := readNoNL(t, tc.src)
			got := typesOnly(toks)
			if len(got) != len(tc.want) {
				t.Fatalf("len got %d want %d: got=%v", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("i=%d: got %v want %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
