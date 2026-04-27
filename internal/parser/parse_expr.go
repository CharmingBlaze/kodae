package parser

import (
	"clio/internal/ast"
	"clio/internal/token"
)

// parseExpr is the top-level entry for expression parsing (full precedence).
func (p *Parser) parseExpr() ast.Expr { return p.parseExprBP(0) }

// parseExprBP is Pratt-style binary parsing with the given minimum precedence.
func (p *Parser) parseExprBP(min int) ast.Expr {
	if p.err != nil {
		return nil
	}
	lhs := p.parsePrefix()
	if lhs == nil {
		return nil
	}
	for {
		if !isBinaryOp(p.tok.Type) {
			return lhs
		}
		prec := binPrec(p.tok.Type)
		if prec < min {
			return lhs
		}
		op := p.tok
		p.next()
		// left-associative: rhs at strictly higher binding power
		rhs := p.parseExprBP(prec + 1)
		lhs = &ast.BinaryExpr{Op: tokToOp(op.Type), L: lhs, R: rhs}
	}
}

func (p *Parser) parsePrefix() ast.Expr {
	if p.err != nil {
		return nil
	}
	switch p.tok.Type {
	case token.NOT, token.MINUS, token.PLUS:
		op := p.tok
		p.next()
		return &ast.UnaryExpr{Op: tokToUnaryOp(op.Type), X: p.parseExprBP(6)}
	case token.LPAREN:
		p.next()
		e := p.parseExpr()
		if p.tok.Type != token.RPAREN {
			p.failf("expected ) in group")
			return e
		}
		p.next()
		return &ast.ParenExpr{Inner: e}
	case token.LBRACK:
		return p.parseListLiteral()
	case token.INTLIT:
		i, e := p.intFromTok()
		if e != nil {
			p.failf("bad int: %v", e)
		}
		tok := p.tok
		p.next()
		return &ast.IntLit{Val: i, Raw: tok.Literal}
	case token.FLOATLIT:
		raw := p.tok.Literal
		p.next()
		return &ast.FloatLit{Raw: raw}
	case token.STRLIT:
		s := p.tok.Literal
		p.next()
		e, err := ExpandStringInterpolation(s)
		if err != nil {
			p.failf("%v", err)
			return &ast.StringLit{Val: s}
		}
		return e
	case token.TRUE:
		p.next()
		return &ast.BoolLit{Val: true}
	case token.FALSE:
		p.next()
		return &ast.BoolLit{Val: false}
	case token.NONE:
		p.next()
		return &ast.NoneExpr{}
	case token.IDENT:
		id := &ast.IdentExpr{Name: p.tok.Literal}
		p.next()
		return p.parsePostfix(id)
	case token.THIS:
		p.next()
		return p.parsePostfix(&ast.IdentExpr{Name: "this"})
	case token.OK, token.ERR:
		if p.tok.Type == token.OK {
			p.failf("ok(...) is not supported in Clio v1; use catch")
		} else {
			p.failf("err(...) is not supported in Clio v1; use catch")
		}
		return nil
	default:
		p.failf("unexpected token in expr: %v", p.tok.Type)
		return nil
	}
}

func (p *Parser) parsePostfix(lhs ast.Expr) ast.Expr {
	for p.err == nil {
		switch p.tok.Type {
		case token.DOT:
			p.next()
			var f string
			switch p.tok.Type {
			case token.IDENT:
				f = p.tok.Literal
				p.next()
			case token.OK, token.ERR:
				p.failf("result field access (.ok/.err) is not part of Clio v1; use catch")
				return lhs
			default:
				p.failf("field name after .")
				return lhs
			}
			lhs = &ast.MemberExpr{Left: lhs, Field: f}
		case token.LBRACE:
			// struct literal: TypeName { a: 1, b: 2 }
			// In "for v in e {", the `{` after e starts the for-body, not a struct literal.
			if p.forInHeader {
				return lhs
			}
			if id, ok := lhs.(*ast.IdentExpr); ok {
				p.next() // {
				p.skipNewlines()
				inits := p.parseStructFieldInits()
				if p.err != nil {
					return lhs
				}
				if p.tok.Type != token.RBRACE {
					p.failf("struct literal: expected }")
					return lhs
				}
				p.next()
				return &ast.StructLit{TypeName: id.Name, Inits: inits}
			}
			return lhs
		case token.LBRACK:
			p.next()
			p.skipNewlines()
			idx := p.parseExpr()
			p.skipNewlines()
			p.expect(token.RBRACK)
			lhs = &ast.IndexExpr{Left: lhs, Index: idx}
		case token.LPAREN:
			// call
			args := p.parseArgList()
			lhs = p.finishCall(lhs, args)
		case token.QUEST:
			p.failf("? is not supported in Clio v1; use catch")
			return lhs
		case token.CATCH:
			p.next()
			p.expect(token.LPAREN)
			if p.tok.Type != token.IDENT {
				p.failf("catch: need (name) for error value")
				return lhs
			}
			en := p.tok.Literal
			p.next()
			p.expect(token.RPAREN)
			p.skipNewlines()
			b := p.parseBlock()
			if b == nil {
				p.failf("catch: need { ... }")
				return lhs
			}
			lhs = &ast.ResultCatchExpr{Subj: lhs, ErrName: en, Body: b}
		case token.PLUSPLUS:
			p.next()
			lhs = &ast.PostfixExpr{X: lhs, Op: "++"}
		case token.MINUSMINUS:
			p.next()
			lhs = &ast.PostfixExpr{X: lhs, Op: "--"}
		default:
			return lhs
		}
	}
	return lhs
}

func (p *Parser) parseListLiteral() ast.Expr {
	p.expect(token.LBRACK)
	p.skipNewlines()
	if p.tok.Type == token.RBRACK {
		p.next()
		return &ast.ListLit{Elems: nil}
	}
	var elems []ast.Expr
	for p.err == nil {
		el := p.parseExpr()
		elems = append(elems, el)
		p.skipNewlines()
		if p.tok.Type == token.RBRACK {
			p.next()
			return &ast.ListLit{Elems: elems}
		}
		if p.tok.Type == token.COMMA {
			p.next()
			p.skipNewlines()
			continue
		}
		p.failf("list literal: expected , or ]")
		return &ast.ListLit{Elems: elems}
	}
	return &ast.ListLit{Elems: elems}
}

// parseStructFieldInits: field: expr, ...; current is first token after `{`.
func (p *Parser) parseStructFieldInits() []ast.StructFieldInit {
	if p.tok.Type == token.RBRACE {
		return nil
	}
	var inits []ast.StructFieldInit
	for p.err == nil {
		if p.tok.Type == token.EOF {
			p.failf("struct literal: unclosed {")
			return inits
		}
		if p.tok.Type == token.NEWLINE || p.tok.Type == token.COMMA {
			p.next()
			if p.tok.Type == token.RBRACE {
				return inits
			}
			continue
		}
		if p.tok.Type != token.IDENT {
			p.failf("struct literal: need field: value")
			return inits
		}
		fn := p.tok.Literal
		p.next()
		if p.tok.Type != token.COLON {
			p.failf("struct literal: expected : after field name %q", fn)
			return inits
		}
		p.next()
		p.skipNewlines()
		ie := p.parseExpr()
		p.skipNewlines()
		inits = append(inits, ast.StructFieldInit{Name: fn, Init: ie})
		if p.tok.Type == token.RBRACE {
			return inits
		}
		if p.tok.Type == token.COMMA {
			p.next()
			p.skipNewlines()
			if p.tok.Type == token.RBRACE {
				return inits
			}
			continue
		}
		p.failf("struct literal: , or }")
		return inits
	}
	return inits
}

// parseArgList: current tok is (
func (p *Parser) parseArgList() []ast.Expr {
	p.expect(token.LPAREN)
	if p.tok.Type == token.RPAREN {
		p.next()
		return nil
	}
	var a []ast.Expr
	for p.err == nil {
		e := p.parseExpr()
		a = append(a, e)
		if p.tok.Type == token.RPAREN {
			break
		}
		if p.tok.Type == token.COMMA {
			p.next()
			p.skipNewlines()
			continue
		}
		p.failf("in arg list, expected ) or comma")
		break
	}
	p.expect(token.RPAREN)
	return a
}

func (p *Parser) finishCall(f ast.Expr, args []ast.Expr) ast.Expr {
	if id, ok := f.(*ast.IdentExpr); ok {
		if (id.Name == "int" || id.Name == "float" || id.Name == "str" || id.Name == "bool") && len(args) == 1 {
			return &ast.CastExpr{To: id.Name, Arg: args[0]}
		}
	}
	return &ast.CallExpr{Fun: f, Args: args}
}

func isBinaryOp(t token.Type) bool {
	switch t {
	case token.PLUS, token.MINUS, token.MUL, token.DIV, token.MOD,
		token.EQ, token.NEQ, token.LT, token.GT, token.LEQ, token.GEQ,
		token.AND, token.OR, token.ASSIGN, token.PLUSEQ, token.MINUSEQ, token.MULEQ, token.DIVEQ, token.MODEQ, token.DOTDOT:
		return true
	default:
		return false
	}
}

func tokToOp(t token.Type) string {
	switch t {
	case token.PLUS:
		return "+"
	case token.MINUS:
		return "-"
	case token.MUL:
		return "*"
	case token.DIV:
		return "/"
	case token.MOD:
		return "%"
	case token.EQ:
		return "=="
	case token.NEQ:
		return "!="
	case token.LT:
		return "<"
	case token.GT:
		return ">"
	case token.LEQ:
		return "<="
	case token.GEQ:
		return ">="
	case token.AND:
		return "&&"
	case token.OR:
		return "||"
	case token.ASSIGN:
		return "="
	case token.PLUSEQ:
		return "+="
	case token.MINUSEQ:
		return "-="
	case token.MULEQ:
		return "*="
	case token.DIVEQ:
		return "/="
	case token.MODEQ:
		return "%="
	case token.DOTDOT:
		return ".."
	default:
		return "?"
	}
}

func tokToUnaryOp(t token.Type) string {
	switch t {
	case token.NOT:
		return "!"
	case token.MINUS:
		return "-"
	case token.PLUS:
		return "+"
	default:
		return "?"
	}
}

func binPrec(t token.Type) int {
	switch t {
	case token.OR:
		return 1
	case token.AND:
		return 2
	case token.EQ, token.NEQ, token.LT, token.GT, token.LEQ, token.GEQ:
		return 3
	case token.DOTDOT: // same tier as compares; `0..a+1` -> `0..(a+1)` by making + tighter
		return 3
	case token.PLUS, token.MINUS:
		return 4
	case token.MUL, token.DIV, token.MOD:
		return 5
	case token.ASSIGN, token.PLUSEQ, token.MINUSEQ, token.MULEQ, token.DIVEQ, token.MODEQ:
		return 0
	default:
		return -1
	}
}