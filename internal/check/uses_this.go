package check

import (
	"kodae/internal/ast"
)

func stmtsUseThis(st []ast.Stmt) bool {
	for _, s := range st {
		if stmtUsesThis(s) {
			return true
		}
	}
	return false
}

func stmtUsesThis(s ast.Stmt) bool {
	if s == nil {
		return false
	}
	switch x := s.(type) {
	case *ast.BlockStmt:
		return stmtsUseThis(x.Stmts)
	case *ast.LetStmt:
		return exprUsesThis(x.Init)
	case *ast.ExprStmt:
		return exprUsesThis(x.E)
	case *ast.AssignStmt:
		return exprUsesThis(x.Left) || exprUsesThis(x.Right)
	case *ast.ReturnStmt:
		return exprUsesThis(x.V)
	case *ast.IfStmt:
		if stmtUsesThisBlock(x.Thn) {
			return true
		}
		if x.Els != nil {
			return stmtUsesThis(x.Els)
		}
		return false
	case *ast.WhileStmt:
		return stmtUsesThisBlock(x.Body)
	case *ast.LoopStmt:
		return stmtUsesThisBlock(x.Body)
	case *ast.ForInStmt:
		return stmtUsesThisBlock(x.Body)
	case *ast.RepeatStmt:
		return stmtUsesThisBlock(x.Body)
	case *ast.MatchStmt:
		for _, a := range x.Arms {
			if stmtUsesThisBlock(a.Body) {
				return true
			}
		}
		return false
	case *ast.DeferStmt:
		return exprUsesThis(x.E)
	default:
		return false
	}
}

func stmtUsesThisBlock(b *ast.BlockStmt) bool {
	if b == nil {
		return false
	}
	return stmtsUseThis(b.Stmts)
}

func exprUsesThis(e ast.Expr) bool {
	if e == nil {
		return false
	}
	switch x := e.(type) {
	case *ast.IdentExpr:
		return x.Name == "this"
	case *ast.MemberExpr:
		return exprUsesThis(x.Left)
	case *ast.CallExpr:
		if exprUsesThis(x.Fun) {
			return true
		}
		for _, a := range x.Args {
			if exprUsesThis(a) {
				return true
			}
		}
		return false
	case *ast.UnaryExpr:
		return exprUsesThis(x.X)
	case *ast.BinaryExpr:
		return exprUsesThis(x.L) || exprUsesThis(x.R)
	case *ast.ParenExpr:
		return exprUsesThis(x.Inner)
	case *ast.CastExpr:
		return exprUsesThis(x.Arg)
	case *ast.IndexExpr:
		return exprUsesThis(x.Left) || exprUsesThis(x.Index)
	case *ast.StructLit:
		for _, fi := range x.Inits {
			if exprUsesThis(fi.Init) {
				return true
			}
		}
		return false
	case *ast.StructUpdateExpr:
		if exprUsesThis(x.Base) {
			return true
		}
		for _, fi := range x.Inits {
			if exprUsesThis(fi.Init) {
				return true
			}
		}
		return false
	case *ast.ListLit:
		for _, el := range x.Elems {
			if exprUsesThis(el) {
				return true
			}
		}
		return false
	case *ast.TupleExpr:
		for _, el := range x.Exprs {
			if exprUsesThis(el) {
				return true
			}
		}
		return false
	case *ast.PostfixExpr:
		return exprUsesThis(x.X)
	case *ast.FuncLit:
		return stmtsUseThis(x.Body.Stmts)
	case *ast.ResultCatchExpr:
		if exprUsesThis(x.Subj) {
			return true
		}
		return stmtUsesThisBlock(x.Body)
	default:
		return false
	}
}
