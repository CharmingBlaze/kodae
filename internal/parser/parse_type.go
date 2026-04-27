package parser

import (
	"clio/internal/ast"
	"clio/internal/token"
)

// parseType reads a type: void, T?, ptr[U], result[V], or name?.
func (p *Parser) parseType() *ast.TypeExpr {
	if p.tok.Type != token.IDENT && p.tok.Type != token.RESULT {
		p.failf("type: need identifier, got %v", p.tok.Type)
		return nil
	}
	// Lexer maps the word "result" to token.RESULT, not IDENT.
	if p.tok.Type == token.RESULT {
		p.next()
		p.expect(token.LBRACK)
		inner := p.parseType()
		if inner == nil {
			return nil
		}
		p.expect(token.RBRACK)
		t := &ast.TypeExpr{Name: "result", ResultInner: inner}
		if p.tok.Type == token.QUEST {
			p.failf("type: result[...]? is not supported")
			return nil
		}
		return t
	}
	if p.tok.Literal == "ptr" {
		p.next()
		p.expect(token.LBRACK)
		inner := p.parseType()
		if inner == nil {
			return nil
		}
		p.expect(token.RBRACK)
		t := &ast.TypeExpr{PtrInner: inner}
		if p.tok.Type == token.QUEST {
			t.Optional = true
			p.next()
		}
		return t
	}
	t := &ast.TypeExpr{Name: p.tok.Literal}
	p.next()
	if p.tok.Type == token.QUEST {
		t.Optional = true
		p.next()
	}
	return t
}
