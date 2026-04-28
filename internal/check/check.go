package check

import (
	"fmt"
	"sort"
	"strings"

	"kodae/internal/ast"
)

func methodReceiverType(fn *ast.FnDecl, structs map[string]*Struct) *Type {
	if fn == nil || len(fn.Params) == 0 || fn.Params[0].Name != "self" || fn.Params[0].T == nil {
		return nil
	}
	recvName := fn.Params[0].T.Name
	if recvName == "" {
		return nil
	}
	if _, ok := structs[recvName]; !ok {
		return nil
	}
	prefix := recvName + "_"
	if !strings.HasPrefix(fn.Name, prefix) {
		return nil
	}
	return StructType(structs[recvName])
}

func typeExprHasPtr(tx *ast.TypeExpr) bool {
	if tx == nil {
		return false
	}
	if tx.PtrInner != nil {
		return true
	}
	if tx.ResultInner != nil {
		return typeExprHasPtr(tx.ResultInner)
	}
	if tx.ListInner != nil {
		return typeExprHasPtr(tx.ListInner)
	}
	return false
}

// Info holds type info for the whole program, used by codegen
type Info struct {
	Types   map[uintptr]*Type
	Enums   map[string]*Enum
	Struct  map[string]*Struct
	Fns     map[string]*ast.FnDecl
	Externs map[string]*ast.ExternDecl
	// LinkFlags: raw tokens from # link "..." (split in driver when compiling).
	LinkFlags []string
	// UsesConsole: true if the program uses print, input, or clear_screen.
	UsesConsole bool
	Module      string
	Meta        map[string]string
	Tuples      []*Type // collected tuple types
}

// Checker is the semantic/type checker
type Checker struct {
	pr    *ast.Program
	inf   *Info
	err     error
	enums   map[string]*Enum
	structs map[string]*Struct
	fns     map[string]*ast.FnDecl
	externs map[string]*ast.ExternDecl
	// file-scope let (values)
	globals map[string]*Type

	scopes     []map[string]*Type
	loopDepth  int
	deferNesting int // 0 = only direct fn body statements may use defer
	returnWant   *Type
	inReturn     bool // type-checking a return's expression (for err() → result[T])
	tryOK        bool // set while type-checking let init, return value, assign RHS, or expr-stmt
	// curFile: absolute .kodae path of the function / let being checked (for pub / cross-file rules)
	curFile string
	// externTypeCtx > 0 while resolving types inside extern fn signatures (allows f32)
	externTypeCtx int
	// sizedTypeCtx > 0 while resolving struct field types (allows f32/i32/u32/u8 for C layout)
	sizedTypeCtx int
}

// Check type-checks a program
func Check(pr *ast.Program) (*Info, error) {
	if pr == nil {
		return nil, fmt.Errorf("nil program")
	}
	c := &Checker{
		pr: pr,
		inf: &Info{
			Types:   make(map[uintptr]*Type),
			Enums:   make(map[string]*Enum),
			Struct:  make(map[string]*Struct),
			Fns:     make(map[string]*ast.FnDecl),
			Externs: make(map[string]*ast.ExternDecl),
			Meta:    make(map[string]string),
		},
		enums:   make(map[string]*Enum),
		structs: make(map[string]*Struct),
		fns:     make(map[string]*ast.FnDecl),
		externs: make(map[string]*ast.ExternDecl),
		globals: make(map[string]*Type),
	}
	// module, use, # link (v1: record; use does not resolve)
	for _, d := range pr.Decls {
		switch t := d.(type) {
		case *ast.ModuleDecl:
			if c.inf.Module != "" {
				c.setErr(fmt.Errorf("only one module declaration allowed"))
			} else {
				c.inf.Module = t.Name
			}
		case *ast.UseDecl:
			_ = t
		case *ast.LinkDecl:
			c.inf.LinkFlags = append(c.inf.LinkFlags, expandLinkArgs(splitLinkFlagString(t.Flags))...)
		case *ast.LinkPathDecl:
			if strings.TrimSpace(t.Path) == "" {
				c.setErr(fmt.Errorf("# linkpath: path is empty"))
			} else {
				c.inf.LinkFlags = append(c.inf.LinkFlags, "-L"+t.Path)
			}
		case *ast.MetaDecl:
			c.inf.Meta[t.Key] = t.Value
		}
	}
	for _, d := range pr.Decls {
		switch t := d.(type) {
		case *ast.EnumDecl:
			idx := make(map[string]int, len(t.Variants))
			for i, n := range t.Variants {
				if _, du := idx[n]; du {
					c.failf("enum %s: duplicate variant %q", t.Name, n)
					continue
				}
				idx[n] = i
			}
			if c.err != nil {
				break
			}
			e := &Enum{Name: t.Name, Index: idx, Pub: t.Pub, File: t.File}
			c.enums[t.Name] = e
			c.inf.Enums[t.Name] = e
		}
	}
	if c.err != nil {
		return c.inf, c.err
	}
	// struct declarations (order in file must satisfy field types; e.g. inner structs first)
	for _, d := range pr.Decls {
		sdecl, ok := d.(*ast.StructDecl)
		if !ok {
			continue
		}
		if c.enums[sdecl.Name] != nil {
			c.setErr(fmt.Errorf("name %q already used as an enum", sdecl.Name))
			continue
		}
		if c.structs[sdecl.Name] != nil {
			c.setErr(fmt.Errorf("duplicate struct name %q", sdecl.Name))
			continue
		}
		c.curFile = sdecl.File
		c.sizedTypeCtx++
		m := make(map[string]*Type, len(sdecl.Fields))
		var order []string
		for _, f := range sdecl.Fields {
			if f.Name == "" {
				continue
			}
			if m[f.Name] != nil {
				c.setErr(fmt.Errorf("struct %s: duplicate field %q", sdecl.Name, f.Name))
				continue
			}
			if typeExprHasPtr(f.T) {
				c.setErr(fmt.Errorf("type: ptr[...] is only allowed in extern fn signatures"))
				break
			}
			ft, err := c.resolveType(f.T)
			if err != nil {
				c.setErr(err)
				break
			}
			if ft != nil && ft.Kind == KOptional {
				if ft.Opt != nil && ft.Opt.Kind == KStruct {
					c.setErr(fmt.Errorf("struct %s: field %q: optional of struct is not supported yet", sdecl.Name, f.Name))
					break
				}
			}
			order = append(order, f.Name)
			m[f.Name] = ft
		}
		c.sizedTypeCtx--
		if c.err != nil {
			c.curFile = ""
			break
		}
		c.curFile = ""
		sd := &Struct{Name: sdecl.Name, Order: order, Fields: m, Pub: sdecl.Pub, SrcFile: sdecl.File}
		c.structs[sdecl.Name] = sd
		c.inf.Struct[sdecl.Name] = sd
	}
	if c.err != nil {
		return c.inf, c.err
	}
	c.curFile = ""
	for _, d := range pr.Decls {
		if ex, ok := d.(*ast.ExternDecl); ok {
			if c.externs[ex.Name] != nil {
				c.setErr(fmt.Errorf("duplicate extern %q", ex.Name))
				continue
			}
			if c.fns[ex.Name] != nil {
				c.setErr(fmt.Errorf("name %q already a function; cannot extern", ex.Name))
				continue
			}
			if c.structs[ex.Name] != nil {
				c.setErr(fmt.Errorf("name %q already a struct; cannot extern", ex.Name))
				continue
			}
			c.externs[ex.Name] = ex
			c.inf.Externs[ex.Name] = ex
			continue
		}
		if fn, ok := d.(*ast.FnDecl); ok {
			for _, p := range fn.Params {
				if typeExprHasPtr(p.T) {
					c.setErr(fmt.Errorf("type: ptr[...] is only allowed in extern fn signatures"))
				}
			}
			if typeExprHasPtr(fn.Return) {
				c.setErr(fmt.Errorf("type: ptr[...] is only allowed in extern fn signatures"))
			}
			if c.fns[fn.Name] != nil {
				c.setErr(fmt.Errorf("duplicate function %q", fn.Name))
				continue
			}
			if c.externs[fn.Name] != nil {
				c.setErr(fmt.Errorf("name %q already extern; cannot declare fn", fn.Name))
				continue
			}
			if c.structs[fn.Name] != nil {
				c.setErr(fmt.Errorf("name %q already used as a struct; cannot declare fn with same name", fn.Name))
				continue
			}
			c.fns[fn.Name] = fn
			c.inf.Fns[fn.Name] = fn
		}
	}
	if c.err != nil {
		return c.inf, c.err
	}
	for _, d := range pr.Decls {
		if l, ok := d.(*ast.LetDecl); ok {
			if l.Name == "main" {
				c.setErr(fmt.Errorf("cannot declare 'main' as a variable; use fn main()"))
				continue
			}
			if typeExprHasPtr(l.T) {
				c.setErr(fmt.Errorf("type: ptr[...] is only allowed in extern fn signatures"))
				continue
			}
			if c.structs[l.Name] != nil {
				c.setErr(fmt.Errorf("name %q already used as a struct", l.Name))
				continue
			}
			if c.fns[l.Name] != nil {
				c.setErr(fmt.Errorf("name %q already used as a function", l.Name))
				continue
			}
			if c.globals[l.Name] != nil {
				c.setErr(fmt.Errorf("duplicate top-level name %q", l.Name))
				continue
			}
			ty, err := c.checkGlobalInit(l)
			if err != nil {
				c.setErr(err)
			} else {
				c.globals[l.Name] = ty
			}
		}
	}
	if c.err != nil {
		return c.inf, c.err
	}
	for _, d := range pr.Decls {
		fn, ok := d.(*ast.FnDecl)
		if !ok {
			continue
		}
		c.curFile = fn.File
		c.scopes = nil
		c.push()
		c.loopDepth = 0
		for k, t := range c.globals {
			c.set(k, t)
		}
		if fn.Return == nil {
			c.returnWant = TpVoid
		} else {
			tt, e := c.resolveType(fn.Return)
			if e != nil {
				c.setErr(e)
				tt = TpVoid
			}
			c.returnWant = tt
		}
		for _, p := range fn.Params {
			tt, e := c.resolveType(p.T)
			if e != nil {
				c.setErr(e)
				tt = TpInt
			}
			c.set(p.Name, tt)
		}
		if recvT := methodReceiverType(fn, c.structs); recvT != nil {
			c.set("this", recvT)
		}
		if fn.Body != nil {
			c.stmts(fn.Body.Stmts)
		}
		c.curFile = ""
	}
	if c.err != nil {
		return c.inf, c.err
	}
	c.curFile = ""
	for _, d := range pr.Decls {
		switch t := d.(type) {
		case *ast.StructDecl:
			if !t.Pub {
				continue
			}
			c.curFile = t.File
			func() {
				c.sizedTypeCtx++
				defer func() { c.sizedTypeCtx-- }()
				for _, f := range t.Fields {
					ft, err := c.resolveType(f.T)
					if err != nil {
						c.setErr(err)
						break
					}
					if err := c.validatePubStructFieldType(ft); err != nil {
						c.setErr(fmt.Errorf("pub struct %s field %s: %v", t.Name, f.Name, err))
						break
					}
				}
			}()
			c.curFile = ""
		case *ast.FnDecl:
			if !t.Pub {
				continue
			}
			c.curFile = t.File
			for _, p := range t.Params {
				pt, err := c.resolveType(p.T)
				if err != nil {
					c.setErr(err)
					break
				}
				if err := c.validateExportType(pt); err != nil {
					c.setErr(fmt.Errorf("pub fn %s param %s: %v", t.Name, p.Name, err))
					break
				}
			}
			if t.Return != nil {
				rt, err := c.resolveType(t.Return)
				if err != nil {
					c.setErr(err)
				} else if err := c.validateExportType(rt); err != nil {
					c.setErr(fmt.Errorf("pub fn %s return: %v", t.Name, err))
				}
			}
			c.curFile = ""
		}
	}
	if c.err != nil {
		return c.inf, c.err
	}
	return c.inf, nil
}

// validatePubStructFieldType allows C layout types (f32, i32, …) for pub structs used in bindings.
func (c *Checker) validatePubStructFieldType(t *Type) error {
	if t == nil {
		return nil
	}
	switch t.Kind {
	case KInt, KFloat, KBool, KStr, KVoid, KF32, KI32, KU32, KU8:
		return nil
	case KStruct:
		return nil
	case KList:
		return fmt.Errorf("list[T] is not exportable in pub API")
	case KPtr:
		return fmt.Errorf("ptr[...] is not exportable in pub API")
	case KOptional:
		return fmt.Errorf("optional/none type is not exportable in pub API")
	case KResult:
		return fmt.Errorf("result[...] is not exportable in pub API")
	default:
		return fmt.Errorf("type %s is not exportable", t)
	}
}

func (c *Checker) validateExportType(t *Type) error {
	if t == nil {
		return nil
	}
	switch t.Kind {
	case KInt, KFloat, KBool, KStr, KVoid:
		return nil
	case KF32, KI32, KU32, KU8:
		return fmt.Errorf("C-sized type %s is not for pub export; use int/float only", t)
	case KStruct:
		return nil
	case KList:
		return fmt.Errorf("list[T] is not exportable in pub API")
	case KPtr:
		return fmt.Errorf("ptr[...] is not exportable in pub API")
	case KOptional:
		return fmt.Errorf("optional/none type is not exportable in pub API")
	case KResult:
		return fmt.Errorf("result[...] is not exportable in pub API")
	default:
		return fmt.Errorf("type %s is not exportable", t)
	}
}

func (c *Checker) setErr(e error) {
	if c.err == nil && e != nil {
		c.err = e
	}
}
func (c *Checker) failf(format string, a ...any) { c.setErr(fmt.Errorf(format, a...)) }

func (c *Checker) setType(e ast.Expr, t *Type) { c.inf.Types[exprKey(e)] = t }

func (c *Checker) checkGlobalInit(l *ast.LetDecl) (*Type, error) {
	c.curFile = l.File
	defer func() { c.curFile = "" }()
	c.push()
	for k, t := range c.globals {
		c.set(k, t)
	}
	if ll, ok := l.Init.(*ast.ListLit); ok && len(ll.Elems) == 0 && l.T != nil {
		want, e := c.resolveType(l.T)
		if e != nil {
			c.pop()
			return nil, e
		}
		if want == nil || want.Kind != KList {
			c.pop()
			return nil, fmt.Errorf("list literal [] requires list[...] annotation")
		}
		c.setType(l.Init, want)
		c.pop()
		return want, nil
	}
	ti, err := c.typeExpr(l.Init)
	if err != nil {
		c.pop()
		return nil, err
	}
	if l.T != nil {
		want, e := c.resolveType(l.T)
		if e != nil {
			c.pop()
			return nil, e
		}
		if ti != nil && ti.Kind == KNil {
			want = optionalOf(want)
		}
		if e := c.assignable(want, ti); e != nil {
			c.pop()
			return nil, e
		}
		c.pop()
		return want, nil
	}
	c.pop()
	return ti, nil
}

// stmts is in stmt.go

func (c *Checker) checkMatch(m *ast.MatchStmt) {
	t, e := c.typeExpr(m.Scrutinee)
	if e != nil {
		c.setErr(e)
		return
	}
	if t.Kind != KEnum || t.EnumRef == nil {
		c.setErr(fmt.Errorf("match requires an enum, got %s", t))
		return
	}
	en := t.EnumRef
	want := make(map[string]struct{}, len(en.Index))
	for s := range en.Index {
		want[s] = struct{}{}
	}
	seen := make(map[string]struct{})
	for i, a := range m.Arms {
		pme, mOk := a.Pat.(*ast.MemberExpr)
		if !mOk {
			c.setErr(fmt.Errorf("match arm %d: use %s.<Variant>", i, en.Name))
			return
		}
		id, idOk := pme.Left.(*ast.IdentExpr)
		if !idOk || id.Name != en.Name {
			c.setErr(fmt.Errorf("match pattern must be %s.<Variant>", en.Name))
			return
		}
		if _, o := en.Index[pme.Field]; !o {
			c.setErr(fmt.Errorf("no variant %q in enum %s", pme.Field, en.Name))
			return
		}
		seen[pme.Field] = struct{}{}
		_ = i
	}
	var missing []string
	for s := range want {
		if _, o := seen[s]; !o {
			missing = append(missing, s)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		c.setErr(fmt.Errorf("non-exhaustive match on %s: missing %s", en.Name, strings.Join(missing, ", ")))
	}
	for _, a := range m.Arms {
		c.push()
		if a.Body != nil {
			c.stmts(a.Body.Stmts)
		}
		c.pop()
	}
}

// splitLinkFlagString splits a # link "..." string into argv tokens (whitespace; no quoting yet).
func splitLinkFlagString(s string) []string { return strings.Fields(s) }

// expandLinkArgs turns bare names (e.g. "raylib" from # link "raylib") into "-lraylib";
// tokens that already start with '-' are left unchanged.
func expandLinkArgs(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		if p[0] == '-' {
			out = append(out, p)
			continue
		}
		out = append(out, "-l"+p)
	}
	return out
}
