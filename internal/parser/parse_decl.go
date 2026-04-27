package parser

import (
	"fmt"
	"clio/internal/ast"
	"clio/internal/token"
)
func (p *Parser) parseEnum() *ast.EnumDecl { return p.parseEnumWithPub(false) }

func (p *Parser) parseEnumWithPub(pub bool) *ast.EnumDecl {
	p.expect(token.ENUM)
	if p.tok.Type != token.IDENT {
		p.failf("enum: name")
		return nil
	}
	name := p.tok.Literal
	p.next()
	p.expect(token.LBRACE)
	p.skipNewlines()
	var vars []string
	for p.tok.Type != token.RBRACE {
		if p.tok.Type == token.EOF {
			p.failf("enum: unclosed {")
			break
		}
		if p.tok.Type == token.NEWLINE || p.tok.Type == token.COMMA {
			p.next()
			continue
		}
		if p.tok.Type != token.IDENT {
			p.failf("enum: variant name")
			break
		}
		vars = append(vars, p.tok.Literal)
		p.next()
		if p.tok.Type == token.COMMA {
			p.next()
			p.skipNewlines()
			continue
		}
	}
	p.expect(token.RBRACE)
	return &ast.EnumDecl{Pub: pub, Name: name, Variants: vars}
}

// MangleMethod is the top-level function name for fn Receiver.method
func MangleMethod(recv, method string) string { return fmt.Sprintf("%s_%s", recv, method) }

func (p *Parser) parseFn() *ast.FnDecl { return p.parseFnWithPub(false) }

// parseFnWithPub is used for `fn` or `pub fn` (or `pub` already consumed).
func (p *Parser) parseFnWithPub(pub bool) *ast.FnDecl {
	p.expect(token.FN)
	if p.tok.Type != token.IDENT {
		p.failf("fn: name")
		return nil
	}
	first := p.tok.Literal
	p.next()
	var name, recv string
	if p.tok.Type == token.DOT {
		recv = first
		p.next()
		if p.tok.Type != token.IDENT {
			p.failf("fn: method name after .")
			return nil
		}
		meth := p.tok.Literal
		p.next()
		name = MangleMethod(recv, meth)
	} else {
		name = first
	}
	p.expect(token.LPAREN)
	params := p.parseParamList(recv, false, false)
	if recv != "" {
		params = p.fixMethodSelfType(recv, params)
	}
	var ret *ast.TypeExpr
	if p.tok.Type == token.ARROW {
		p.next()
		ret = p.parseType()
	}
	body := p.parseBlock()
	return &ast.FnDecl{Name: name, Pub: pub, Params: params, Return: ret, Body: body}
}

// parseExtern: `extern fn name(a: T, ...) -> R`  (no body)
func (p *Parser) parseExtern() *ast.ExternDecl {
	p.expect(token.EXTERN)
	p.expect(token.FN)
	if p.tok.Type != token.IDENT {
		p.failf("extern: name")
		return nil
	}
	name := p.tok.Literal
	p.next()
	p.expect(token.LPAREN)
	params := p.parseParamList("", true, true)
	var ret *ast.TypeExpr
	if p.tok.Type == token.ARROW {
		p.next()
		ret = p.parseExternType()
	} else {
		p.failf("extern: need -> return type (use void for none)")
		return nil
	}
	if p.tok.Type == token.LBRACE {
		p.failf("extern: must not have a body")
		return nil
	}
	return &ast.ExternDecl{Name: name, Params: params, Return: ret}
}

func (p *Parser) fixMethodSelfType(recv string, params []ast.Param) []ast.Param {
	if p.err != nil {
		return params
	}
	if len(params) == 0 || params[0].Name != "self" {
		self := ast.Param{Name: "self", T: &ast.TypeExpr{Name: recv, Optional: false}}
		return append([]ast.Param{self}, params...)
	}
	if params[0].T == nil {
		params[0].T = &ast.TypeExpr{Name: recv, Optional: false}
	}
	return params
}

// parseParamList reads parameters until ")".
// If allowVararg, a trailing "..." is allowed. If methodRecv is set, bare `self` may omit a type.
func (p *Parser) parseParamList(methodRecv string, allowVararg bool, allowPtr bool) []ast.Param {
	var ps []ast.Param
	if p.tok.Type == token.RPAREN {
		p.next()
		return nil
	}
	if allowVararg && p.tok.Type == token.ELLIPSIS {
		p.next()
		ps = append(ps, ast.Param{Dots: true})
		p.expect(token.RPAREN) // eat ')', advance to next
		return ps
	}
	for p.err == nil {
		if allowVararg && p.tok.Type == token.ELLIPSIS {
			p.next()
			ps = append(ps, ast.Param{Dots: true})
			p.expect(token.RPAREN)
			return ps
		}
		if p.tok.Type != token.IDENT {
			p.failf("param: name")
			return ps
		}
		pn := p.tok.Literal
		p.next()
		if p.tok.Type != token.COLON {
			if (p.tok.Type == token.COMMA || p.tok.Type == token.RPAREN) && methodRecv != "" && pn == "self" && len(ps) == 0 {
				ps = append(ps, ast.Param{Name: pn, T: nil})
				if p.tok.Type == token.RPAREN {
					p.next()
					return ps
				}
				p.next()
				p.skipNewlines()
				continue
			}
			p.failf("param: : type")
			return ps
		}
		p.next()
		var ty *ast.TypeExpr
		if allowPtr {
			ty = p.parseExternType()
		} else {
			ty = p.parseType()
		}
		ps = append(ps, ast.Param{Name: pn, T: ty})
		if p.tok.Type == token.RPAREN {
			p.next()
			return ps
		}
		if p.tok.Type == token.COMMA {
			p.next()
			p.skipNewlines()
			if allowVararg && p.tok.Type == token.ELLIPSIS {
				p.next()
				ps = append(ps, ast.Param{Dots: true})
				p.expect(token.RPAREN)
				return ps
			}
			continue
		}
		p.failf("param: , or )")
		return ps
	}
	return ps
}

func (p *Parser) parseStructDecl() *ast.StructDecl {
	return p.parseStructDeclWithPub(false)
}

func (p *Parser) parseStructDeclWithPub(pub bool) *ast.StructDecl {
	p.expect(token.STRUCT)
	if p.tok.Type != token.IDENT {
		p.failf("struct: name")
		return nil
	}
	name := p.tok.Literal
	p.next()
	p.expect(token.LBRACE)
	p.skipNewlines()
	var fields []ast.StructField
	for p.tok.Type != token.RBRACE {
		if p.tok.Type == token.EOF {
			p.failf("struct: unclosed {")
			return nil
		}
		if p.tok.Type == token.NEWLINE || p.tok.Type == token.COMMA {
			p.next()
			continue
		}
		if p.tok.Type != token.IDENT {
			p.failf("struct: field name")
			return nil
		}
		fn := p.tok.Literal
		p.next()
		if p.tok.Type != token.COLON {
			p.failf("struct: field: type (expected :)")
			return nil
		}
		p.next()
		ft := p.parseType()
		fields = append(fields, ast.StructField{Name: fn, T: ft})
		if p.tok.Type == token.COMMA {
			p.next()
			p.skipNewlines()
			continue
		}
	}
	p.expect(token.RBRACE)
	return &ast.StructDecl{Pub: pub, Name: name, Fields: fields}
}
