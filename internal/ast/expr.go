package ast

import "strconv"

// --- Expr implementations (pointer receivers for a stable tree) ---

type IdentExpr struct{ Name string }

func (e *IdentExpr) expr() {}

type IntLit struct{ Val int64; Raw string }

func (e *IntLit) expr() {}

type FloatLit struct{ Raw string }

func (e *FloatLit) expr() {}

type StringLit struct{ Val string }

func (e *StringLit) expr() {}

type BoolLit struct{ Val bool }

func (e *BoolLit) expr() {}

type NoneExpr struct{}

func (e *NoneExpr) expr() {}

type ParenExpr struct{ Inner Expr }

func (e *ParenExpr) expr() {}

// PostfixExpr is lvalue++.
type PostfixExpr struct {
	X  Expr
	Op string // "++"
}

func (e *PostfixExpr) expr() {}

type UnaryExpr struct{ Op string; X Expr } // e.g. !, -

func (e *UnaryExpr) expr() {}

type BinaryExpr struct{ Op string; L, R Expr } // e.g. ==, +, !=

func (e *BinaryExpr) expr() {}

// CallExpr is a normal call. CastExpr below models int(x), float(x), str(x) after parsing.
type CallExpr struct {
	Fun  Expr
	Args []Expr
}

func (e *CallExpr) expr() {}

// CastExpr: explicit C-style casts: int(), float(), str() from the initial spec.
type CastExpr struct {
	To  string
	Arg Expr
}

func (e *CastExpr) expr() {}

type MemberExpr struct{ Left Expr; Field string }

func (e *MemberExpr) expr() {}

// StructFieldInit is a single `field: init` in a struct literal.
type StructFieldInit struct {
	Name string
	Init Expr
}

// StructLit: TypeName { a: 1, b: 2 }.
type StructLit struct {
	TypeName string
	Inits    []StructFieldInit
}

func (e *StructLit) expr() {}

// StructUpdateExpr: baseExpr with { field: value, ... } — functional update (copies base, overrides listed fields).
type StructUpdateExpr struct {
	Base  Expr
	Inits []StructFieldInit
}

func (e *StructUpdateExpr) expr() {}

// FuncLit: fn (params) -> T? { body } — expression-only closure (currently checked codegen supports fn() void only).
type FuncLit struct {
	Params []Param
	Return *TypeExpr
	Body   *BlockStmt
}

func (e *FuncLit) expr() {}

// ListLit: [a, b, c]
type ListLit struct {
	Elems []Expr
}

func (e *ListLit) expr() {}

// IndexExpr: listExpr[indexExpr]
type IndexExpr struct {
	Left  Expr
	Index Expr
}

func (e *IndexExpr) expr() {}

// TupleExpr: (a, b) or return a, b
type TupleExpr struct {
	Exprs []Expr
}

func (e *TupleExpr) expr() {}

// TryUnwrap: expr? — propagate error if result is not ok, otherwise unwrap to T.
type TryUnwrapExpr struct{ X Expr }

func (e *TryUnwrapExpr) expr() {}

// ResultCatch: subj catch (err) { block } on result[T] — success yields T, error runs block.
type ResultCatchExpr struct {
	Subj    Expr
	ErrName string
	Body    *BlockStmt
}

func (e *ResultCatchExpr) expr() {}

// ExprString is a compact, stable dump for the parse command.
func ExprString(e Expr) string {
	if e == nil {
		return "<nil>"
	}
	switch x := e.(type) {
	case *IdentExpr:
		return x.Name
	case *IntLit:
		return x.Raw
	case *FloatLit:
		return x.Raw
	case *StringLit:
		return strconv.Quote(x.Val)
	case *BoolLit:
		if x.Val {
			return "true"
		}
		return "false"
	case *NoneExpr:
		return "none"
	case *ParenExpr:
		return "(" + ExprString(x.Inner) + ")"
	case *UnaryExpr:
		return x.Op + ExprString(x.X)
	case *BinaryExpr:
		return "(" + ExprString(x.L) + " " + x.Op + " " + ExprString(x.R) + ")"
	case *CallExpr:
		s := ExprString(x.Fun) + "("
		for i, a := range x.Args {
			if i > 0 {
				s += ", "
			}
			s += ExprString(a)
		}
		return s + ")"
	case *CastExpr:
		return x.To + "(" + ExprString(x.Arg) + ")"
	case *MemberExpr:
		if id, ok := x.Left.(*IdentExpr); ok {
			return id.Name + "." + x.Field
		}
		return "(" + ExprString(x.Left) + ")." + x.Field
	case *PostfixExpr:
		return ExprString(x.X) + x.Op
	case *TryUnwrapExpr:
		return ExprString(x.X) + "?"
	case *ResultCatchExpr:
		return ExprString(x.Subj) + " catch (" + x.ErrName + ") { ... }"
	case *StructLit:
		s := x.TypeName + " {"
		for i, fi := range x.Inits {
			if i > 0 {
				s += ", "
			}
			s += fi.Name + ": " + ExprString(fi.Init)
		}
		return s + "}"
	case *StructUpdateExpr:
		s := "(" + ExprString(x.Base) + ") with {"
		for i, fi := range x.Inits {
			if i > 0 {
				s += ", "
			}
			s += fi.Name + ": " + ExprString(fi.Init)
		}
		return s + "}"
	case *FuncLit:
		s := "fn("
		for i, p := range x.Params {
			if i > 0 {
				s += ", "
			}
			s += p.Name + ": type"
		}
		s += ") "
		if x.Return != nil {
			s += "-> ... "
		}
		return s + "{ ... }"
	case *ListLit:
		s := "["
		for i, el := range x.Elems {
			if i > 0 {
				s += ", "
			}
			s += ExprString(el)
		}
		return s + "]"
	case *TupleExpr:
		s := "("
		for i, el := range x.Exprs {
			if i > 0 {
				s += ", "
			}
			s += ExprString(el)
		}
		return s + ")"
	case *IndexExpr:
		return ExprString(x.Left) + "[" + ExprString(x.Index) + "]"
	default:
		return "?"
	}
}
