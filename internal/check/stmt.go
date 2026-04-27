package check

import (
	"fmt"

	"clio/internal/ast"
)

func (c *Checker) stmts(st []ast.Stmt) {
	for _, s := range st {
		if c.err != nil {
			return
		}
		c.stmt(s)
	}
}

func (c *Checker) stmt(s ast.Stmt) {
	if c.err != nil {
		return
	}
	switch x := s.(type) {
	case *ast.BlockStmt:
		c.deferNesting++
		c.push()
		c.stmts(x.Stmts)
		c.pop()
		c.deferNesting--
	case *ast.IfStmt:
		t, err := c.typeExpr(x.Cond)
		if err != nil {
			c.setErr(err)
			return
		}
		if t == nil || t.Kind != KBool {
			c.setErr(fmt.Errorf("if condition: expected bool, got %v", t))
			return
		}
		if x.Thn != nil {
			c.stmt(x.Thn)
		}
		if x.Els != nil {
			c.stmt(x.Els)
		}
	case *ast.WhileStmt:
		tt, err := c.typeExpr(x.Cond)
		if err != nil {
			c.setErr(err)
			return
		}
		if tt == nil || tt.Kind != KBool {
			c.setErr(fmt.Errorf("while: need bool, got %v", tt))
			return
		}
		c.loopDepth++
		if x.Body != nil {
			c.stmt(&ast.BlockStmt{Stmts: x.Body.Stmts})
		}
		c.loopDepth--
	case *ast.LoopStmt:
		c.loopDepth++
		if x.Body != nil {
			c.stmt(&ast.BlockStmt{Stmts: x.Body.Stmts})
		}
		c.loopDepth--
	case *ast.BreakStmt:
		if c.loopDepth == 0 {
			c.setErr(fmt.Errorf("break outside of loop"))
		}
	case *ast.ContinueStmt:
		if c.loopDepth == 0 {
			c.setErr(fmt.Errorf("continue outside of loop"))
		}
	case *ast.DeferStmt:
		if c.deferNesting != 0 {
			c.setErr(fmt.Errorf("defer is only allowed at the top of a function body (v1)"))
		} else {
			_, e := c.typeExpr(x.E) // e.g. void call
			if e != nil {
				c.setErr(e)
			}
		}
	case *ast.ForInStmt:
		inn, e := c.typeExpr(x.In)
		if e != nil {
			c.setErr(e)
			return
		}
		if inn.Kind == KRange {
			b, ok := x.In.(*ast.BinaryExpr)
			if !ok || b.Op != ".." {
				c.setErr(fmt.Errorf("for-in: range internal error"))
				return
			}
			c.push()
			c.set(x.Var, TpInt)
			c.loopDepth++
			if x.Body != nil {
				_, _ = b, ok
				c.stmt(&ast.BlockStmt{Stmts: x.Body.Stmts})
			}
			c.loopDepth--
			c.pop()
			return
		}
		if inn.Kind == KList {
			c.push()
			elem := inn.Elem
			if elem == nil {
				elem = TpInt
			}
			c.set(x.Var, elem)
			c.loopDepth++
			if x.Body != nil {
				c.stmt(&ast.BlockStmt{Stmts: x.Body.Stmts})
			}
			c.loopDepth--
			c.pop()
			return
		}
		c.setErr(fmt.Errorf("for-in: only for (i in a..b) (integer range) is supported; got %s", inn))
	case *ast.ReturnStmt:
		if c.returnWant == nil {
			// in global? shouldn't happen
			return
		}
		if x.V == nil {
			if c.returnWant.Kind != KVoid {
				c.setErr(fmt.Errorf("return: expected value of type %s", c.returnWant))
			}
			return
		}
		if x.V != nil && containsResultCatch(x.V) {
			if _, direct := unwrapExpr(x.V).(*ast.ResultCatchExpr); !direct {
				c.setErr(fmt.Errorf("result catch: must be the full return value, not nested in a larger expression"))
				return
			}
		}
		c.inReturn = true
		c.tryOK = true
		t, err := c.typeExpr(x.V)
		c.tryOK = false
		c.inReturn = false
		if err != nil {
			c.setErr(err)
			return
		}
		if c.assignable(c.returnWant, t) != nil {
			// int assign to float
			if c.returnWant.equal(TpFloat) && t.equal(TpInt) {
				return
			}
			c.setErr(c.assignable(c.returnWant, t))
		}
	case *ast.LetStmt:
		ty, err := c.inferLocal(x)
		if err != nil {
			c.setErr(err)
		} else {
			if x.Const {
				// reassign blocked in codegen, not here
			}
			_ = ty
		}
	case *ast.ExprStmt:
		if containsResultCatch(x.E) {
			if _, direct := unwrapExpr(x.E).(*ast.ResultCatchExpr); !direct {
				c.setErr(fmt.Errorf("result catch: must be the full expression, not nested in a larger expression"))
				return
			}
		}
		c.tryOK = true
		_, err := c.typeExpr(x.E) // e.g. call, f()? propagate, or f() catch
		c.tryOK = false
		if err != nil {
			c.setErr(err)
		}
	case *ast.AssignStmt:
		c.checkAssign(x)
	case *ast.MatchStmt:
		c.checkMatch(x)
	default:
		c.setErr(fmt.Errorf("statement: unsupported %T", s))
	}
}

// unwrapExpr strips outer ParenExpr wrappers.
func unwrapExpr(e ast.Expr) ast.Expr {
	for e != nil {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.Inner
	}
	return nil
}

// containsResultCatch is true if the tree has a `... catch (x) { }` node.
func containsResultCatch(e ast.Expr) bool {
	if e == nil {
		return false
	}
	switch t := e.(type) {
	case *ast.ResultCatchExpr:
		return true
	case *ast.BinaryExpr:
		return containsResultCatch(t.L) || containsResultCatch(t.R)
	case *ast.UnaryExpr:
		return containsResultCatch(t.X)
	case *ast.ParenExpr:
		return containsResultCatch(t.Inner)
	case *ast.CallExpr:
		if containsResultCatch(t.Fun) {
			return true
		}
		for _, a := range t.Args {
			if containsResultCatch(a) {
				return true
			}
		}
		return false
	case *ast.MemberExpr:
		return containsResultCatch(t.Left)
	case *ast.TryUnwrapExpr:
		return containsResultCatch(t.X)
	case *ast.CastExpr:
		return containsResultCatch(t.Arg)
	case *ast.PostfixExpr:
		return containsResultCatch(t.X)
	case *ast.StructLit:
		for _, fi := range t.Inits {
			if containsResultCatch(fi.Init) {
				return true
			}
		}
		return false
	case *ast.ListLit:
		for _, el := range t.Elems {
			if containsResultCatch(el) {
				return true
			}
		}
		return false
	case *ast.IndexExpr:
		return containsResultCatch(t.Left) || containsResultCatch(t.Index)
	default:
		return false
	}
}

func (c *Checker) inferLocal(x *ast.LetStmt) (*Type, error) {
	if x.Init == nil {
		if x.T == nil {
			return nil, fmt.Errorf("let: need a type and initializer, or = value")
		}
		want, e := c.resolveType(x.T)
		if e != nil {
			return nil, e
		}
		c.set(x.Name, want)
		return want, nil
	}
	if containsResultCatch(x.Init) {
		if _, direct := unwrapExpr(x.Init).(*ast.ResultCatchExpr); !direct {
			return nil, fmt.Errorf("result catch: must be the full initializer, not nested in a larger expression (e.g. f() catch (e) { ... } only)")
		}
	}
	if x.Const {
		if _, ok := unwrapExpr(x.Init).(*ast.ResultCatchExpr); ok {
			return nil, fmt.Errorf("const with result catch is not supported; use `let`")
		}
	}
	c.tryOK = true
	if ll, ok := x.Init.(*ast.ListLit); ok && len(ll.Elems) == 0 && x.T != nil {
		want, e := c.resolveType(x.T)
		c.tryOK = false
		if e != nil {
			return nil, e
		}
		if want == nil || want.Kind != KList {
			return nil, fmt.Errorf("list literal [] requires list[...] annotation")
		}
		c.setType(x.Init, want)
		c.set(x.Name, want)
		return want, nil
	}
	tt, err := c.typeExpr(x.Init)
	c.tryOK = false
	if err != nil {
		return nil, err
	}
	if x.T != nil {
		want, e := c.resolveType(x.T)
		if e != nil {
			return nil, e
		}
		if tt != nil && tt.Kind == KNil {
			want = optionalOf(want)
		}
		if e := c.assignable(want, tt); e != nil {
			return want, e
		}
		c.set(x.Name, want)
		return want, nil
	}
	c.set(x.Name, tt)
	return tt, nil
}

func (c *Checker) checkAssign(x *ast.AssignStmt) {
	lt, err := c.typeExpr(x.Left)
	if err != nil {
		c.setErr(err)
		return
	}
	// reassign: left must be ident
	id, isId := x.Left.(*ast.IdentExpr)
	if isId {
		_ = id
	} else {
		// *Member maybe later
	}
	if containsResultCatch(x.Right) {
		if _, direct := unwrapExpr(x.Right).(*ast.ResultCatchExpr); !direct {
			c.setErr(fmt.Errorf("result catch: must be the full right-hand side, not nested in a larger expression"))
			return
		}
	}
	c.tryOK = true
	rt, err2 := c.typeExpr(x.Right)
	c.tryOK = false
	if err2 != nil {
		c.setErr(err2)
		return
	}
	// compound op
	if x.Op != "=" {
		if x.Op == "+=" && lt != nil && lt.equal(TpStr) && rt != nil && rt.equal(TpStr) {
			// str += is lowered to concat in codegen
		} else if !lt.isNumeric() {
			c.setErr(fmt.Errorf("compound assign %s: need numeric, got %s (str allows += for concat)", x.Op, lt))
			return
		}
	}
	if e := c.assignable(lt, rt); e != nil {
		if x.Op == "+=" && lt != nil && lt.equal(TpStr) && rt != nil && rt.equal(TpStr) {
			return
		}
		// int += float? promote
		if lt != nil && rt != nil && lt.isNumeric() && rt.isNumeric() {
			if lt.equal(TpFloat) || (x.Op == "+=" && (lt.equal(TpInt) && rt.equal(TpFloat))) {
				return
			}
		}
		c.setErr(e)
	}
}
