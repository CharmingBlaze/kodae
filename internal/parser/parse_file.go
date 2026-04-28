package parser

import (
	"fmt"
	"kodae/internal/ast"
	"kodae/internal/token"
)

// ParseProgram reads the file into an AST. On error, a partial tree may be returned; check Err().
func (p *Parser) ParseProgram() *ast.Program {
	if p.err != nil {
		return nil
	}
	pr := &ast.Program{}
	p.skipNewlines()
	for p.tok.Type != token.EOF {
		if p.tok.Type == token.NEWLINE {
			p.next()
			continue
		}
		d := p.parseTopDecl()
		if d == nil {
			// if no decl (error), try recover by skipping
			if p.err != nil {
				break
			}
			p.skipNewlines()
			continue
		}
		if group, ok := d.(*ast.ConstGroupDecl); ok {
			for _, c := range group.Decls {
				pr.Decls = append(pr.Decls, c)
			}
		} else {
			pr.Decls = append(pr.Decls, d)
		}
		p.skipNewlines()
	}
	return pr
}

// parseTopDecl: fn, enum, let, const, struct (unimplemented: skip line).
func (p *Parser) parseTopDecl() ast.Decl {
	if p.err != nil {
		return nil
	}
	switch p.tok.Type {
	case token.ENUM:
		return p.parseEnum()
	case token.FN:
		return p.parseFn()
	case token.LET:
		return p.parseTopLet()
	case token.CONST:
		return p.parseTopConst()
	case token.STRUCT:
		return p.parseStructDecl()
	case token.MODULE:
		return p.parseModule()
	case token.USE:
		return p.parseUse()
	case token.EXTERN:
		return p.parseExtern()
	case token.HASH:
		return p.parseLink()
	default:
		p.failf("unexpected at file scope: %s", p.tok.Type.String())
		return nil
	}
}

func (p *Parser) parseTopLet() *ast.LetDecl {
	if p.tok.Type != token.LET {
		return nil
	}
	p.next()
	if p.tok.Type != token.IDENT {
		p.failf("let: name")
		return nil
	}
	name := p.tok.Literal
	p.next()
	t := p.tryParseTypeWithColon() // if ":" parse type, else no type
	var init ast.Expr
	if p.tok.Type == token.ASSIGN {
		p.next()
		init = p.parseExpr()
	} else {
		p.failf("let: expected = or :")
	}
	return &ast.LetDecl{Name: name, T: t, Init: init}
}

func (p *Parser) parseTopConst() ast.Decl {
	// const NAME = value — treat like let without mutation
	if p.tok.Type != token.CONST {
		return nil
	}
	p.next()
	if p.tok.Type != token.IDENT {
		p.failf("const: name")
		return nil
	}
	name := p.tok.Literal
	p.next()

	if p.tok.Type == token.LBRACE {
		p.next()
		p.skipNewlines()
		var decls []*ast.LetDecl
		for p.tok.Type != token.RBRACE {
			if p.tok.Type == token.EOF {
				p.failf("const group: unclosed {")
				return nil
			}
			if p.tok.Type == token.NEWLINE || p.tok.Type == token.COMMA {
				p.next()
				continue
			}
			if p.tok.Type != token.IDENT {
				p.failf("const group: expected identifier")
				return nil
			}
			fName := p.tok.Literal
			p.next()
			
			// Optional type
			_ = p.tryParseTypeWithColon()
			
			if p.tok.Type != token.ASSIGN {
				p.failf("const group: expected =")
				return nil
			}
			p.next()
			v := p.parseExpr()
			decls = append(decls, &ast.LetDecl{
				Name: fmt.Sprintf("%s_%s", name, fName),
				T:    nil,
				Init: v,
			})
			
			if p.tok.Type == token.COMMA {
				p.next()
				p.skipNewlines()
				continue
			}
		}
		p.expect(token.RBRACE)
		return &ast.ConstGroupDecl{Decls: decls}
	}

	_ = p.tryParseTypeWithColon()
	if p.tok.Type != token.ASSIGN {
		p.failf("const: =")
		return nil
	}
	p.next()
	v := p.parseExpr()
	return &ast.LetDecl{Name: name, T: nil, Init: v}
}

// tryParseTypeWithColon: if : then parse type, else return nil
func (p *Parser) tryParseTypeWithColon() *ast.TypeExpr {
	if p.tok.Type != token.COLON {
		return nil
	}
	p.next()
	return p.parseType()
}

func (p *Parser) parseModule() *ast.ModuleDecl {
	p.expect(token.MODULE)
	if p.tok.Type != token.IDENT {
		p.failf("module: name")
		return nil
	}
	n := p.tok.Literal
	p.next()
	return &ast.ModuleDecl{Name: n}
}

func (p *Parser) parseUse() *ast.UseDecl {
	p.expect(token.USE)
	if p.tok.Type != token.IDENT {
		p.failf("use: module name")
		return nil
	}
	n := p.tok.Literal
	p.next()
	return &ast.UseDecl{Name: n}
}

// # link " -lfoo -L/path"
func (p *Parser) parseLink() ast.Decl {
	p.expect(token.HASH)
	if p.tok.Type != token.IDENT {
		p.failf("directive: expected name after #")
		return nil
	}
	key := p.tok.Literal
	p.next()
	if p.tok.Type != token.STRLIT {
		p.failf("# %s: expected string value", key)
		return nil
	}
	s := p.tok.Literal
	p.next()
	switch key {
	case "link":
		return &ast.LinkDecl{Flags: s}
	case "linkpath":
		return &ast.LinkPathDecl{Path: s}
	case "include":
		return &ast.IncludeDecl{Path: s}
	case "mode", "library", "version", "author":
		return &ast.MetaDecl{Key: key, Value: s}
	default:
		p.failf("directive: unsupported #%s", key)
		return nil
	}
}
