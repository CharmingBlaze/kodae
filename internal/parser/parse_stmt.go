package parser

import (
	"clio/internal/ast"
	"clio/internal/token"
)

func (p *Parser) parseBlock() *ast.BlockStmt {
	p.expect(token.LBRACE)
	p.skipNewlines()
	b := &ast.BlockStmt{}
	for p.tok.Type != token.RBRACE {
		if p.tok.Type == token.EOF {
			p.failf("block: unclosed {")
			break
		}
		if p.tok.Type == token.NEWLINE {
			p.next()
			continue
		}
		s := p.parseStmt()
		if s != nil {
			b.Stmts = append(b.Stmts, s)
		}
		if p.err != nil {
			break
		}
	}
	p.expect(token.RBRACE)
	return b
}

func (p *Parser) parseStmt() ast.Stmt {
	if p.err != nil {
		return nil
	}
	p.skipNewlines()
	switch p.tok.Type {
	case token.LET, token.CONST:
		return p.parseLocalLet()
	case token.IF:
		return p.parseIf()
	case token.WHILE:
		return p.parseWhile()
	case token.FOR:
		return p.parseFor()
	case token.LOOP:
		return p.parseLoop()
	case token.MATCH:
		return p.parseMatch()
	case token.RETURN:
		return p.parseReturn()
	case token.BREAK:
		p.next()
		return &ast.BreakStmt{}
	case token.CONTINUE:
		p.next()
		return &ast.ContinueStmt{}
	case token.DEFER:
		p.next()
		e := p.parseExpr()
		if e == nil {
			return nil
		}
		return &ast.DeferStmt{E: e}
	case token.LBRACE:
		return p.parseBlock()
	}
	return p.parseExprOrAssignStmt()
}

// parseExprOrAssignStmt: expression statement or assignment
func (p *Parser) parseExprOrAssignStmt() ast.Stmt {
	e := p.parseExpr()
	if p.err != nil {
		return nil
	}
	if b, ok := e.(*ast.BinaryExpr); ok {
		switch b.Op {
		case "=", "+=", "-=", "*=", "/=", "%=":
			return &ast.AssignStmt{Left: b.L, Op: b.Op, Right: b.R}
		}
	}
	return &ast.ExprStmt{E: e}
}

func (p *Parser) parseLocalLet() ast.Stmt {
	cons := p.tok.Type == token.CONST
	p.next()
	if p.tok.Type != token.IDENT {
		p.failf("let/const: name")
		return nil
	}
	name := p.tok.Literal
	p.next()
	ty := p.tryParseTypeWithColon()
	if p.tok.Type != token.ASSIGN {
		if ty == nil {
			p.failf("let: need a type and = value, or = expr")
			return nil
		}
		// `let v: S` without initializer — zero-initialized
		return &ast.LetStmt{Const: cons, Name: name, T: ty, Init: nil}
	}
	p.next()
	init := p.parseExpr()
	return &ast.LetStmt{Const: cons, Name: name, T: ty, Init: init}
}

func (p *Parser) parseIf() *ast.IfStmt {
	p.expect(token.IF)
	p.expect(token.LPAREN)
	cond := p.parseExpr()
	p.expect(token.RPAREN)
	thn := p.parseBlock()
	var els ast.Stmt
	if p.tok.Type == token.ELSE {
		p.next()
		if p.tok.Type == token.IF {
			els = p.parseIf()
		} else {
			els = p.parseBlock()
		}
	}
	return &ast.IfStmt{Cond: cond, Thn: thn, Els: els}
}

func (p *Parser) parseWhile() *ast.WhileStmt {
	p.expect(token.WHILE)
	p.expect(token.LPAREN)
	cond := p.parseExpr()
	p.expect(token.RPAREN)
	b := p.parseBlock()
	return &ast.WhileStmt{Cond: cond, Body: b}
}

func (p *Parser) parseFor() *ast.ForInStmt {
	p.expect(token.FOR)
	// for (i in 0..10) { }  or  for i in 0..10 { }
	if p.tok.Type == token.LPAREN {
		p.next()
		if p.tok.Type != token.IDENT {
			p.failf("for: var")
			return nil
		}
		v := p.tok.Literal
		p.next()
		p.expect(token.IN)
		inn := p.parseExpr()
		if p.err != nil {
			return nil
		}
		p.expect(token.RPAREN)
		p.skipNewlines()
		b := p.parseBlock()
		if p.err != nil {
			return nil
		}
		return &ast.ForInStmt{Var: v, In: inn, Body: b}
	}
	if p.tok.Type != token.IDENT {
		p.failf("for: need (name in range) or name in range before {")
		return nil
	}
	v := p.tok.Literal
	p.next()
	p.expect(token.IN)
	p.forInHeader = true
	inn := p.parseExpr()
	p.forInHeader = false
	if p.err != nil {
		return nil
	}
	p.skipNewlines()
	b := p.parseBlock()
	if p.err != nil {
		return nil
	}
	return &ast.ForInStmt{Var: v, In: inn, Body: b}
}

func (p *Parser) parseLoop() *ast.LoopStmt {
	p.expect(token.LOOP)
	b := p.parseBlock()
	return &ast.LoopStmt{Body: b}
}

func (p *Parser) parseReturn() *ast.ReturnStmt {
	p.expect(token.RETURN)
	if p.tok.Type == token.RBRACE || p.tok.Type == token.NEWLINE {
		// no value
		if p.tok.Type == token.NEWLINE {
			p.next()
		}
		return &ast.ReturnStmt{V: nil}
	}
	v := p.parseExpr()
	return &ast.ReturnStmt{V: v}
}

func (p *Parser) parseMatch() *ast.MatchStmt {
	p.expect(token.MATCH)
	p.expect(token.LPAREN)
	s := p.parseExpr()
	p.expect(token.RPAREN)
	p.expect(token.LBRACE)
	p.skipNewlines()
	var arms []ast.MatchArm
	for p.tok.Type != token.RBRACE {
		if p.tok.Type == token.EOF {
			p.failf("match: unclosed }")
			break
		}
		if p.tok.Type == token.NEWLINE || p.tok.Type == token.COMMA {
			p.next()
			continue
		}
		pat := p.parseExpr()
		if p.tok.Type != token.FATARROW {
			p.failf("match arm: =>")
			break
		}
		p.next()
		p.skipNewlines()
		body := p.parseBlock()
		arms = append(arms, ast.MatchArm{Pat: pat, Body: body})
	}
	p.expect(token.RBRACE)
	return &ast.MatchStmt{Scrutinee: s, Arms: arms}
}
