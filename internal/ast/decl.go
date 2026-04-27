package ast

// --- Decl implementations ---

type EnumDecl struct {
	Name     string
	Variants []string
}

func (d *EnumDecl) decl() {}

type FnDecl struct {
	Name   string
	Pub    bool
	Params []Param
	Return *TypeExpr
	Body   *BlockStmt
}

func (d *FnDecl) decl() {}

type LetDecl struct {
	Name string
	T    *TypeExpr
	Init Expr
}

func (d *LetDecl) decl() {}

// StructField is a field in a struct { name: type }.
type StructField struct {
	Name string
	T    *TypeExpr
}

// StructDecl is at file scope: struct Name { a: int, b: int }.
type StructDecl struct {
	Pub    bool
	Name   string
	Fields []StructField
}

func (d *StructDecl) decl() {}

// Param is a function parameter: name, type, optional.
// Dots is true for "..." (varargs) in extern fn.
type Param struct {
	Name string
	T    *TypeExpr
	Dots bool
}

// ExternDecl is `extern fn name(...) -> T` with no body.
type ExternDecl struct {
	Name   string
	Params []Param
	Return *TypeExpr
}

func (d *ExternDecl) decl() {}

// ModuleDecl is `module name` (file-level).
type ModuleDecl struct{ Name string }

func (d *ModuleDecl) decl() {}

// UseDecl is `use name` (import module into scope; v1: accepted, resolution later).
type UseDecl struct{ Name string }

func (d *UseDecl) decl() {}

// LinkDecl is `# link "flags"` — extra argv for the C linker (e.g. "-lraylib").
type LinkDecl struct{ Flags string }

func (d *LinkDecl) decl() {}

// MetaDecl is a generic file directive: #mode/#library/#version/#author.
type MetaDecl struct {
	Key   string
	Value string
}

func (d *MetaDecl) decl() {}
