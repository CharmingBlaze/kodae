// Package ast defines Kodae abstract syntax.
package ast

// Program is a translation unit: top-level declarations and statements.
type Program struct {
	Decls      []Decl
	Statements []Stmt
}

// Decl is a file-scope declaration.
type Decl interface{ decl() }

// Stmt is a statement in a function body.
type Stmt interface{ stmt() }

// Expr is an expression node.
type Expr interface{ expr() }
