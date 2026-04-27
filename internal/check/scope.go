package check

import (
	"fmt"

	"clio/internal/ast"
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

func (c *Checker) isBuiltin(name string) bool {
	switch name {
	case "print", "int", "float", "str", "bool", "min", "max", "abs",
		"input", "random", "clear_screen", "ok", "err":
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
	case "float", "f64", "float64", "double":
		base = TpFloat
	case "str", "string":
		base = TpStr
	case "bool":
		base = TpBool
	case "byte":
		base = TpByte
	default:
		return nil, fmt.Errorf("unknown type %q", tx.Name)
	}
	if tx.Optional {
		return optionalOf(base), nil
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
	if !want.equal(got) {
		if (want.equal(TpInt) && got.equal(TpFloat)) || (want.equal(TpFloat) && got.equal(TpInt)) {
			return nil
		}
		return errAssign(want, got)
	}
	return nil
}
func errAssign(a, b *Type) error { return fmt.Errorf("type mismatch: expected %s, have %s", a, b) }
