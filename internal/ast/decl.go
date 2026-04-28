package ast

// --- Decl implementations ---

type EnumDecl struct {
	Name     string
	Variants []string
	// File is the absolute .kodae path; used for pub / cross-file rules.
	File     string
}

func (d *EnumDecl) decl() {}

// File is the absolute path to the .kodae source; used for cross-file visibility. Empty in tests / legacy.
type FnDecl struct {
	Name   string
	File   string
	Params []Param
	Return *TypeExpr
	Body   *BlockStmt
}

func (d *FnDecl) decl() {}

type LetDecl struct {
	Name string
	T    *TypeExpr
	Init Expr
	// File is the absolute path of the containing .kodae file (top-level let/const).
	File string
}

func (d *LetDecl) decl() {}

type ConstGroupDecl struct {
	Decls []*LetDecl
}

func (d *ConstGroupDecl) decl() {}

// StructField is a field in a struct { name: type }.
type StructField struct {
	Name string
	T    *TypeExpr
}

// StructDecl is at file scope: struct Name { a: int, b: int }.
type StructDecl struct {
	Name   string
	Fields []StructField
	// File is the absolute .kodae path; used for pub / cross-file rules.
	File string
}

func (d *StructDecl) decl() {}

// Param is a function parameter: name, type, optional.
// Dots is true for "..." (varargs) in extern fn.
type Param struct {
	Name string
	T    *TypeExpr
	Dots bool
	Init Expr
}

// ExternDecl is `extern fn name(...) T` with no body.
type ExternDecl struct {
	Name   string
	Params []Param
	Return *TypeExpr
	// File is the absolute .kodae path (externs are always linkable program-wide; field is for diagnostics).
	File string
}

func (d *ExternDecl) decl() {}

// ModuleDecl is `module name` (file-level).
type ModuleDecl struct{ Name string }

func (d *ModuleDecl) decl() {}

// UseDecl is `use name` (import module into scope; v1: accepted, resolution later).
type UseDecl struct{ Name string }

func (d *UseDecl) decl() {}

// LinkDecl is `# link "flags"` — extra argv for the C linker (e.g. "-lraylib" or a bare name "raylib" → -lraylib).
type LinkDecl struct{ Flags string }

func (d *LinkDecl) decl() {}

// LinkPathDecl is `# linkpath "dir"` — adds `-Ldir` so the linker finds `.a` / `.lib` files.
type LinkPathDecl struct{ Path string }

func (d *LinkPathDecl) decl() {}

// IncludeDecl is `# include "path"` — path is like `player`, `ui/hud`, or `raylib` (`.kodae` added when missing).
type IncludeDecl struct{ Path string }

func (d *IncludeDecl) decl() {}

// MetaDecl is a generic file directive: #mode/#library/#version/#author.
type MetaDecl struct {
	Key   string
	Value string
}

func (d *MetaDecl) decl() {}
