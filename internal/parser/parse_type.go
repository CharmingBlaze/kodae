package parser

import (
	"kodae/internal/ast"
	"kodae/internal/token"
)

// parseType reads a user-facing type in non-extern contexts.
func (p *Parser) parseType() *ast.TypeExpr {
	return p.parseTypeWithRules(false)
}

// parseExternType reads a type in extern fn signatures, where ptr[...] is allowed.
func (p *Parser) parseExternType() *ast.TypeExpr {
	return p.parseTypeWithRules(true)
}

func (p *Parser) parseTypeWithRules(allowPtr bool) *ast.TypeExpr {
	if p.tok.Type == token.IDENT {
		switch p.tok.Literal {
		case "f32", "i32", "u32", "u8":
			if !allowPtr {
				p.failf("type %s is only allowed in extern fn signatures or struct fields for C interop", p.tok.Literal)
				return nil
			}
			n := p.tok.Literal
			p.next()
			return &ast.TypeExpr{Name: n}
		}
	}
	if p.tok.Type == token.LPAREN {
		p.next()
		var elems []*ast.TypeExpr
		for p.tok.Type != token.RPAREN {
			if p.err != nil {
				return nil
			}
			if p.tok.Type == token.EOF {
				p.failf("tuple type: unclosed (")
				return nil
			}
			elems = append(elems, p.parseTypeWithRules(allowPtr))
			if p.tok.Type == token.RPAREN {
				break
			}
			if p.tok.Type == token.COMMA {
				p.next()
				continue
			}
			p.failf("tuple type: expected , or )")
			return nil
		}
		p.expect(token.RPAREN)
		return &ast.TypeExpr{TupleInner: elems}
	}
	if p.tok.Type != token.IDENT && p.tok.Type != token.RESULT {
		p.failf("type: need identifier, got %v", p.tok.Type)
		return nil
	}
	if p.tok.Type == token.RESULT || p.tok.Literal == "result" {
		p.failf("type: result[...] is not part of Kodae v1; use catch")
		return nil
	}
	if p.tok.Literal == "ptr" {
		if !allowPtr {
			p.failf("type: ptr[...] is only allowed in extern fn signatures")
			return nil
		}
		p.next()
		p.expect(token.LBRACK)
		inner := p.parseExternType()
		if inner == nil {
			return nil
		}
		p.expect(token.RBRACK)
		t := &ast.TypeExpr{PtrInner: inner}
		if p.tok.Type == token.QUEST {
			p.failf("type: T? is not part of Kodae v1; use plain none with implicit nullable values")
			return nil
		}
		return t
	}
	if p.tok.Literal == "list" {
		p.next()
		p.expect(token.LBRACK)
		inner := p.parseTypeWithRules(allowPtr)
		if inner == nil {
			return nil
		}
		p.expect(token.RBRACK)
		t := &ast.TypeExpr{ListInner: inner}
		if p.tok.Type == token.QUEST {
			p.failf("type: T? is not part of Kodae v1; use plain none with implicit nullable values")
			return nil
		}
		return t
	}
	t := &ast.TypeExpr{Name: p.tok.Literal}
	p.next()
	if p.tok.Type == token.QUEST {
		p.failf("type: T? is not part of Kodae v1; use plain none with implicit nullable values")
		return nil
	}
	return t
}
