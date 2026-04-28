package codegen

import (
	"kodae/internal/ast"
)

// flattenFuncLitsPostOrder returns nested fn(){ } literals in post-order (inner lambdas first).
func flattenFuncLitsPostOrder(st []ast.Stmt) []*ast.FuncLit {
	var out []*ast.FuncLit
	for _, s := range st {
		out = append(out, flattenStmtFuncLits(s)...)
	}
	return out
}

func flattenStmtFuncLits(s ast.Stmt) []*ast.FuncLit {
	if s == nil {
		return nil
	}
	switch x := s.(type) {
	case *ast.BlockStmt:
		return flattenFuncLitsPostOrder(x.Stmts)
	case *ast.LetStmt:
		if fl, ok := stripExprToFuncLit(x.Init); ok {
			nested := flattenFuncLitsPostOrder(fl.Body.Stmts)
			return append(nested, fl)
		}
	case *ast.IfStmt:
		var out []*ast.FuncLit
		if x.Thn != nil {
			out = append(out, flattenStmtFuncLits(x.Thn)...)
		}
		if x.Els != nil {
			out = append(out, flattenStmtFuncLits(x.Els)...)
		}
		return out
	case *ast.WhileStmt:
		if x.Body != nil {
			return flattenStmtFuncLits(x.Body)
		}
	case *ast.LoopStmt:
		if x.Body != nil {
			return flattenStmtFuncLits(x.Body)
		}
	case *ast.ForInStmt:
		if x.Body != nil {
			return flattenStmtFuncLits(x.Body)
		}
	case *ast.RepeatStmt:
		if x.Body != nil {
			return flattenStmtFuncLits(x.Body)
		}
	case *ast.MatchStmt:
		var out []*ast.FuncLit
		for _, a := range x.Arms {
			if a.Body != nil {
				out = append(out, flattenStmtFuncLits(a.Body)...)
			}
		}
		return out
	case *ast.ExprStmt:
		return flattenExprFuncLits(x.E)
	case *ast.AssignStmt:
		return append(flattenExprFuncLits(x.Left), flattenExprFuncLits(x.Right)...)
	case *ast.ReturnStmt:
		return flattenExprFuncLits(x.V)
	case *ast.DeferStmt:
		return flattenExprFuncLits(x.E)
	}
	return nil
}

func flattenExprFuncLits(e ast.Expr) []*ast.FuncLit {
	e = stripExprParens(e)
	if e == nil {
		return nil
	}
	switch x := e.(type) {
	case *ast.FuncLit:
		return flattenFuncLitsPostOrder(x.Body.Stmts)
	case *ast.CallExpr:
		var out []*ast.FuncLit
		out = append(out, flattenExprFuncLits(x.Fun)...)
		for _, a := range x.Args {
			out = append(out, flattenExprFuncLits(a)...)
		}
		return out
	default:
		return nil
	}
}

func stripExprToFuncLit(e ast.Expr) (*ast.FuncLit, bool) {
	e = stripExprParens(e)
	fl, ok := e.(*ast.FuncLit)
	return fl, ok
}
