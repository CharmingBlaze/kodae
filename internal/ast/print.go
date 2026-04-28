package ast

import (
	"fmt"
	"io"
	"strings"
)

// Fprint writes a simple outline of the program (for kodae parse / debugging).
func Fprint(w io.Writer, p *Program) {
	if p == nil {
		fmt.Fprintln(w, "<nil program>")
		return
	}
	for _, d := range p.Decls {
		fprintDecl(w, d, 0)
	}
}

func fprintDecl(w io.Writer, d Decl, ind int) {
	pad := strings.Repeat("  ", ind)
	switch x := d.(type) {
	case *EnumDecl:
		fmt.Fprintf(w, "%senum %s { %s }\n", pad, x.Name, strings.Join(x.Variants, ", "))
	case *LetDecl:
		ty := ""
		if x.T != nil {
			ty = ": " + x.T.String() + " "
		}
		fmt.Fprintf(w, "%slet %s%s= %s\n", pad, x.Name, ty, ExprString(x.Init))
	case *FnDecl:
		fmt.Fprintf(w, "%sfn %s (", pad, x.Name)
		for i, pr := range x.Params {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			ty := "_"
			if pr.T != nil {
				ty = pr.T.String()
			}
			fmt.Fprintf(w, "%s: %s", pr.Name, ty)
		}
		r := ""
		if x.Return != nil {
			r = " " + x.Return.String()
		}
		fmt.Fprintf(w, ")%s\n", r)
		fprintBlock(w, x.Body, ind+1)
	default:
		fmt.Fprintf(w, "%s(decl %T)\n", pad, d)
	}
}

func fprintBlock(w io.Writer, b *BlockStmt, ind int) {
	if b == nil {
		return
	}
	pad := strings.Repeat("  ", ind)
	fmt.Fprintf(w, "%s{\n", pad)
	for _, s := range b.Stmts {
		fprintStmt(w, s, ind+1)
	}
	fmt.Fprintf(w, "%s}\n", pad)
}

func fprintStmt(w io.Writer, s Stmt, ind int) {
	if s == nil {
		return
	}
	pad := strings.Repeat("  ", ind)
	switch x := s.(type) {
	case *BlockStmt:
		fprintBlock(w, x, ind)
	case *IfStmt:
		fmt.Fprintf(w, "%sif %s\n", pad, ExprString(x.Cond))
		fprintBlock(w, x.Thn, ind)
		if x.Els != nil {
			fmt.Fprintf(w, "%selse\n", pad)
			fprintStmt(w, x.Els, ind)
		}
	case *WhileStmt:
		fmt.Fprintf(w, "%swhile %s\n", pad, ExprString(x.Cond))
		fprintBlock(w, x.Body, ind)
	case *ForInStmt:
		fmt.Fprintf(w, "%sfor %s in %s\n", pad, x.Var, ExprString(x.In))
		fprintBlock(w, x.Body, ind)
	case *LoopStmt:
		fmt.Fprintf(w, "%sloop\n", pad)
		fprintBlock(w, x.Body, ind)
	case *ReturnStmt:
		if x.V == nil {
			fmt.Fprintf(w, "%sreturn\n", pad)
		} else {
			fmt.Fprintf(w, "%sreturn %s\n", pad, ExprString(x.V))
		}
	case *BreakStmt:
		fmt.Fprintf(w, "%sbreak\n", pad)
	case *ContinueStmt:
		fmt.Fprintf(w, "%scontinue\n", pad)
	case *DeferStmt:
		fmt.Fprintf(w, "%sdefer %s\n", pad, ExprString(x.E))
	case *LetStmt:
		ty := ""
		if x.T != nil {
			ty = ": " + x.T.String() + " "
		}
		kw := "let"
		if x.Const {
			kw = "const"
		}
		fmt.Fprintf(w, "%s%s %s%s= %s\n", pad, kw, x.Name, ty, ExprString(x.Init))
	case *ExprStmt:
		fmt.Fprintf(w, "%s%s\n", pad, ExprString(x.E))
	case *AssignStmt:
		fmt.Fprintf(w, "%s%s %s %s\n", pad, ExprString(x.Left), x.Op, ExprString(x.Right))
	case *MatchStmt:
		fmt.Fprintf(w, "%smatch %s\n", pad, ExprString(x.Scrutinee))
		fprintMatch(w, x, ind+1)
	default:
		fmt.Fprintf(w, "%s(stmt %T)\n", pad, s)
	}
}

func fprintMatch(w io.Writer, m *MatchStmt, ind int) {
	pad := strings.Repeat("  ", ind)
	for _, a := range m.Arms {
		fmt.Fprintf(w, "%s%s =>\n", pad, ExprString(a.Pat))
		fprintBlock(w, a.Body, ind+1)
	}
}
