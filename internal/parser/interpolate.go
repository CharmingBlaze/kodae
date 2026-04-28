package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"kodae/internal/ast"
	lex "kodae/internal/lexer"
	"kodae/internal/token"
)

// ParseExpressionFragment parses a single expression and requires EOF.
func ParseExpressionFragment(src string) (ast.Expr, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, fmt.Errorf("empty ${...} body")
	}
	p := New(lex.New(src))
	e := p.parseExpr()
	if p.err != nil {
		return e, p.err
	}
	p.skipNewlines()
	if p.tok.Type != token.EOF {
		return e, fmt.Errorf("extra text after expression in ${...}")
	}
	return e, p.Err()
}

// ExpandStringInterpolation decodes "Hello $bob" / "$${" code "}" and $$ → $
// into a left-associative a+b chain (or a single StringLit/expr when no + needed).
func ExpandStringInterpolation(s string) (ast.Expr, error) {
	if s == "" {
		return &ast.StringLit{Val: ""}, nil
	}
	if !strings.ContainsRune(s, '$') {
		return &ast.StringLit{Val: s}, nil
	}
	var chunks []ast.Expr
	var lit strings.Builder
	flush := func() {
		if lit.Len() == 0 {
			return
		}
		chunks = append(chunks, &ast.StringLit{Val: lit.String()})
		lit.Reset()
	}
	i := 0
	for i < len(s) {
		if s[i] != '$' {
			if s[i] < utf8.RuneSelf {
				lit.WriteByte(s[i])
				i++
				continue
			}
			r, w := utf8.DecodeRuneInString(s[i:])
			lit.WriteRune(r)
			i += w
			continue
		}
		// s[i] == '$' — $${ or $$ (literal $)
		if i+1 < len(s) && s[i+1] == '$' {
			flush()
			i += 2
			lit.WriteByte('$')
			continue
		}
		flush() // end literal run before a single $
		if i+1 < len(s) && s[i+1] == '{' {
			end := findMatchingRBrace(s, i+1)
			if end < 0 {
				return nil, fmt.Errorf("unclosed ${ in string")
			}
			inner := strings.TrimSpace(s[i+2 : end])
			sub, err := ParseExpressionFragment(inner)
			if err != nil {
				return nil, fmt.Errorf("${...}: %w", err)
			}
			chunks = append(chunks, sub)
			i = end + 1
			continue
		}
		pathEnd, pathExpr, err := parseDollarPath(s, i+1)
		if err != nil {
			return nil, err
		}
		if pathEnd <= i+1 {
			if i+1 >= len(s) {
				return nil, fmt.Errorf("trailing $ in string")
			}
			return nil, fmt.Errorf("expected name after $ in string")
		}
		chunks = append(chunks, pathExpr)
		i = pathEnd
	}
	flush()
	if len(chunks) == 0 {
		return &ast.StringLit{Val: ""}, nil
	}
	if len(chunks) == 1 {
		return chunks[0], nil
	}
	acc := chunks[0]
	for k := 1; k < len(chunks); k++ {
		acc = &ast.BinaryExpr{Op: "+", L: acc, R: chunks[k]}
	}
	return acc, nil
}

// parseDollarPath parses a single "$" that was already consumed: starting at s[from] with the first
// character of an identifier, reads ident (.ident)* and returns the member expression and end index.
func parseDollarPath(s string, from int) (end int, expr ast.Expr, err error) {
	if from >= len(s) {
		return from, nil, nil
	}
	seg, ok := scanDollarIdent(s, from)
	if !ok {
		return from, nil, nil
	}
	if seg == from {
		return from, nil, fmt.Errorf("incomplete $ name in string")
	}
	var left ast.Expr = &ast.IdentExpr{Name: s[from:seg]}
	end = seg
	for end < len(s) && s[end] == '.' {
		nx, nok := scanDollarIdent(s, end+1)
		if !nok || nx == end+1 {
			return 0, nil, fmt.Errorf("incomplete $ path after .")
		}
		field := s[end+1 : nx]
		left = &ast.MemberExpr{Left: left, Field: field}
		end = nx
	}
	return end, left, nil
}

// scanDollarIdent reads [a-zA-Z_][a-zA-Z0-9_]* starting at s[from]; returns one past name or ok=false.
func scanDollarIdent(s string, from int) (end int, ok bool) {
	if from >= len(s) {
		return from, false
	}
	b := s[from]
	if b != '_' && (b < 'A' || b > 'Z') && (b < 'a' || b > 'z') {
		return from, false
	}
	end = from + 1
	for end < len(s) {
		c := s[end]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			end++
			continue
		}
		break
	}
	return end, true
}

func findMatchingRBrace(s string, lbrace int) int {
	depth := 0
	for j := lbrace; j < len(s); j++ {
		switch s[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return -1
}
