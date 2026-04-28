package check

import (
	"fmt"

	"kodae/internal/ast"
)

func (c *Checker) typeStructUpdateExpr(x *ast.StructUpdateExpr, e ast.Expr) (*Type, error) {
	bt, err := c.typeExpr(x.Base)
	if err != nil {
		return nil, err
	}
	if bt == nil || bt.Kind != KStruct || bt.StructDef == nil {
		return nil, fmt.Errorf("`with` update: need a struct value on the left, got %v", bt)
	}
	sdef := bt.StructDef
	seen := make(map[string]bool)
	for _, in := range x.Inits {
		if in.Name == "" {
			return nil, fmt.Errorf("`with`: empty field name")
		}
		if seen[in.Name] {
			return nil, fmt.Errorf("`with`: duplicate field %q", in.Name)
		}
		seen[in.Name] = true
		want := sdef.Fields[in.Name]
		if want == nil {
			if sug, ok := suggestName(in.Name, sdef.Order); ok {
				return nil, fmt.Errorf("struct %s has no field %q — did you mean %q?", sdef.Name, in.Name, sug)
			}
			return nil, fmt.Errorf("struct %s has no field %q", sdef.Name, in.Name)
		}
		et, err := c.typeExpr(in.Init)
		if err != nil {
			return nil, err
		}
		if err := c.assignable(want, et); err != nil {
			return nil, fmt.Errorf("`with` field %q: %v", in.Name, err)
		}
	}
	out := StructType(sdef)
	c.setType(e, out)
	return out, nil
}

func (c *Checker) typeFuncLit(fl *ast.FuncLit, e ast.Expr) (*Type, error) {
	if len(fl.Params) > 0 {
		return nil, fmt.Errorf("lambda: only fn() {{ }} with no parameters is supported for now")
	}
	if fl.Return != nil {
		rt, err := c.resolveType(fl.Return)
		if err != nil {
			return nil, err
		}
		if rt == nil || rt.Kind != KVoid {
			return nil, fmt.Errorf("lambda: only -> void is supported for now")
		}
	}
	if fl.Body == nil {
		return nil, fmt.Errorf("lambda: need a body")
	}
	was := c.returnWant
	c.returnWant = TpVoid
	c.stmt(fl.Body)
	c.returnWant = was

	capThis := stmtsUseThis(fl.Body.Stmts)
	if capThis && c.curFn == nil {
		return nil, fmt.Errorf("lambda: internal error (no enclosing function)")
	}
	if capThis && methodReceiverType(c.curFn, c.structs) == nil {
		return nil, fmt.Errorf("`this` is only valid inside struct methods (lambda uses `this` outside a method)")
	}

	var recv string
	if rt := methodReceiverType(c.curFn, c.structs); rt != nil && capThis {
		recv = rt.StructName
	}
	c.closureSeq++
	mangled := fmt.Sprintf("f_%s_lam%d", c.curFn.Name, c.closureSeq)
	if c.inf.Closures == nil {
		c.inf.Closures = make(map[uintptr]*ClosureInfo)
	}
	c.inf.Closures[exprKey(fl)] = &ClosureInfo{
		Mangled:      mangled,
		CapturesThis: capThis,
		RecvStruct:   recv,
	}
	c.setType(e, TpClosure)
	return TpClosure, nil
}
