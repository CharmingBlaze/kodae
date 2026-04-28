package check

import (
	"fmt"

	"kodae/internal/ast"
)

func (c *Checker) push() {
	c.scopes = append(c.scopes, make(map[string]*Type))
}
func (c *Checker) pop() {
	if len(c.scopes) == 0 {
		return
	}
	c.scopes = c.scopes[:len(c.scopes)-1]
}
func (c *Checker) set(n string, t *Type) {
	if len(c.scopes) == 0 {
		c.push()
	}
	c.scopes[len(c.scopes)-1][n] = t
}
func (c *Checker) get(n string) *Type {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][n]; ok {
			return t
		}
	}
	if t, ok := c.globals[n]; ok {
		return t
	}
	return nil
}

// typeNameCandidates is used for "did you mean" on unknown type names.
func (c *Checker) typeNameCandidates() []string {
	out := make([]string, 0, len(c.enums)+len(c.structs)+12)
	for k := range c.enums {
		out = append(out, k)
	}
	for k := range c.structs {
		out = append(out, k)
	}
	out = append(out, "int", "float", "str", "string", "bool", "byte", "void", "f64", "float64", "double")
	return out
}

// visibleNameCandidates collects names in scope for typo suggestions on unknown identifiers.
func (c *Checker) visibleNameCandidates() []string {
	var out []string
	for i := 0; i < len(c.scopes); i++ {
		for k := range c.scopes[i] {
			out = append(out, k)
		}
	}
	for k := range c.globals {
		out = append(out, k)
	}
	for k := range c.enums {
		out = append(out, k)
	}
	for k := range c.structs {
		out = append(out, k)
	}
	for k := range c.fns {
		out = append(out, k)
	}
	// callables and builtins (usually written as a call, but help typos in expressions)
	for k := range c.externs {
		out = append(out, k)
	}
	out = append(out,
		"print", "int", "float", "str", "bool", "min", "max", "abs",
		"input", "random", "clear_screen", "len", "this",
	)
	return out
}

// callableNameCandidates for typo suggestions on unknown function calls.
func (c *Checker) callableNameCandidates() []string {
	var out []string
	for k := range c.fns {
		out = append(out, k)
	}
	for k := range c.externs {
		out = append(out, k)
	}
	out = append(out,
		"print", "int", "float", "str", "bool", "min", "max", "abs",
		"input", "random", "clear_screen", "len",
	)
	return out
}



func (c *Checker) isBuiltin(name string) bool {
	switch name {
	case "print", "int", "float", "str", "bool", "min", "max", "abs",
		"input", "random", "clear_screen", "len", "ok", "err":
		return true
	default:
		return false
	}
}

// resolveType: ast.TypeExpr -> *Type
func (c *Checker) resolveType(tx *ast.TypeExpr) (*Type, error) {
	if tx == nil {
		return TpInt, nil
	}
	if tx.PtrInner != nil {
		inner, err := c.resolveType(tx.PtrInner)
		if err != nil {
			return nil, err
		}
		pt := &Type{Kind: KPtr, Pointee: inner}
		if tx.Optional {
			return optionalOf(pt), nil
		}
		return pt, nil
	}
	if tx.ListInner != nil {
		inner, err := c.resolveType(tx.ListInner)
		if err != nil {
			return nil, err
		}
		lt := &Type{Kind: KList, Elem: inner}
		if tx.Optional {
			return optionalOf(lt), nil
		}
		return lt, nil
	}
	if tx.ResultInner != nil {
		inner, err := c.resolveType(tx.ResultInner)
		if err != nil {
			return nil, err
		}
		rt := &Type{Kind: KResult, Res: inner}
		if tx.Optional {
			return optionalOf(rt), nil
		}
		return rt, nil
	}
	if tx.TupleInner != nil {
		var elems []*Type
		for _, x := range tx.TupleInner {
			inner, err := c.resolveType(x)
			if err != nil {
				return nil, err
			}
			elems = append(elems, inner)
		}
		return &Type{Kind: KTuple, TupleElems: elems}, nil
	}
	if tx.Name == "void" {
		if tx.Optional {
			return nil, fmt.Errorf("void? is invalid")
		}
		return TpVoid, nil
	}
	en, hasEnum := c.enums[tx.Name]
	if hasEnum {
		t := c.enumTypeFor(en)
		if tx.Optional {
			return optionalOf(t), nil
		}
		return t, nil
	}
	if sdef, ok := c.structs[tx.Name]; ok {
		t := StructType(sdef)
		if t == nil {
			return nil, fmt.Errorf("struct %q", tx.Name)
		}
		if tx.Optional {
			if t.Kind == KStruct {
				return nil, fmt.Errorf("optional of struct %q is not supported yet", tx.Name)
			}
			return optionalOf(t), nil
		}
		return t, nil
	}
	// not enum or struct — primitives
	var base *Type
	switch tx.Name {
	case "int":
		base = TpInt
	case "f32":
		if c.externTypeCtx <= 0 && c.sizedTypeCtx <= 0 {
			return nil, fmt.Errorf("type f32 (C float) is only allowed in extern fn signatures or struct fields for C layout")
		}
		base = TpF32
	case "i32":
		if c.externTypeCtx <= 0 && c.sizedTypeCtx <= 0 {
			return nil, fmt.Errorf("type i32 (C int32_t) is only allowed in extern fn signatures or struct fields for C layout")
		}
		base = TpI32
	case "u32":
		if c.externTypeCtx <= 0 && c.sizedTypeCtx <= 0 {
			return nil, fmt.Errorf("type u32 (C uint32_t) is only allowed in extern fn signatures or struct fields for C layout")
		}
		base = TpU32
	case "u8":
		if c.externTypeCtx <= 0 && c.sizedTypeCtx <= 0 {
			return nil, fmt.Errorf("type u8 (C uint8_t) is only allowed in extern fn signatures or struct fields for C layout")
		}
		base = TpU8
	case "float", "f64", "float64", "double":
		base = TpFloat
	case "str", "string":
		base = TpStr
	case "bool":
		base = TpBool
	case "byte":
		base = TpByte
	case "Any":
		base = TpAny
	default:
		if sug, ok := suggestName(tx.Name, c.typeNameCandidates()); ok {
			return nil, fmt.Errorf("SCOPE unknown type %q — did you mean %q?", tx.Name, sug)
		}
		return nil, fmt.Errorf("SCOPE unknown type %q", tx.Name)
	}
	if tx.Optional {
		return nil, fmt.Errorf("type: T? is not part of Kodae v1; use plain none with implicit nullable values")
	}
	return base, nil
}

func (c *Checker) enumTypeFor(e *Enum) *Type { return &Type{Kind: KEnum, EnumName: e.Name, EnumRef: e} }

// assignable: want, got; coerce int/float
func (c *Checker) assignable(want, got *Type) error {
	if want == nil {
		return nil
	}
	if got == nil {
		return nil
	}
	// optionals
	if want.Kind == KOptional {
		if got.Kind == KNil {
			return nil
		}
		if got.Kind == KOptional {
			return c.assignable(want.Opt, got.Opt)
		}
		return errAssign(want, got)
	}
	if got.Kind == KNil {
		// need optional on left
		if want.Kind != KOptional {
			return fmt.Errorf("cannot use none (none) for non-optional type %s", want)
		}
		return nil
	}
	// int <= float
	if want.Kind == KFloat && got.Kind == KInt {
		return nil
	}
	// f32 in extern: accept int or Kodae float (double) as the argument
	if want.Kind == KF32 {
		if got.Kind == KInt || got.Kind == KFloat || got.Kind == KF32 {
			return nil
		}
		return errAssign(want, got)
	}
	if want.Kind == KI32 || want.Kind == KU32 || want.Kind == KU8 {
		if got.Kind == KInt || got.Kind == KFloat || got.Kind == want.Kind {
			return nil
		}
		return errAssign(want, got)
	}
	if !want.equal(got) {
		if (want.equal(TpInt) && got.equal(TpFloat)) || (want.equal(TpFloat) && got.equal(TpInt)) {
			return nil
		}
		return errAssign(want, got)
	}
	return nil
}
func errAssign(a, b *Type) error { return fmt.Errorf("type mismatch: expected %s, have %s", a, b) }
