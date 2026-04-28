package check

import (
	"fmt"

	"kodae/internal/ast"
)

func (c *Checker) typeExpr(e ast.Expr) (*Type, error) {
	if c.err != nil {
		return TpInt, c.err
	}
	switch x := e.(type) {
	case *ast.IdentExpr:
		t := c.get(x.Name)
		if t == nil {
			if x.Name == "this" {
				return nil, fmt.Errorf("this can only be used inside a method")
			}
			if x.Name == "ok" || x.Name == "err" {
				return nil, fmt.Errorf("%s(...) is not supported in Kodae v1; use catch", x.Name)
			}
			if c.enums[x.Name] != nil {
				return nil, fmt.Errorf("expected %q as value (use %q.Variant for enum value)", x.Name, x.Name)
			}
			if c.fns[x.Name] != nil {
				return nil, fmt.Errorf("cannot use function name %q as a value (did you mean to call it?)", x.Name)
			}
			if sug, ok := suggestName(x.Name, c.visibleNameCandidates()); ok {
				return nil, fmt.Errorf("unknown name %q — did you mean %q?", x.Name, sug)
			}
			return nil, fmt.Errorf("unknown name %q", x.Name)
		}
		c.setType(e, t)
		return t, nil
	case *ast.IntLit:
		c.setType(e, TpInt)
		return TpInt, nil
	case *ast.FloatLit:
		c.setType(e, TpFloat)
		return TpFloat, nil
	case *ast.StringLit:
		c.setType(e, TpStr)
		return TpStr, nil
	case *ast.BoolLit:
		c.setType(e, TpBool)
		return TpBool, nil
	case *ast.NoneExpr:
		c.setType(e, TpNil)
		return TpNil, nil
	case *ast.ParenExpr:
		t, err := c.typeExpr(x.Inner)
		if err != nil {
			return nil, err
		}
		c.setType(e, t)
		return t, err
	case *ast.UnaryExpr:
		inner, err := c.typeExpr(x.X)
		if err != nil {
			return nil, err
		}
		switch x.Op {
		case "!":
			if inner.Kind != KBool {
				return nil, fmt.Errorf("! expects bool, got %s", inner)
			}
			c.setType(e, TpBool)
			return TpBool, nil
		case "-", "+":
			if !inner.isNumeric() {
				return nil, fmt.Errorf("unary %s expects a number, got %s", x.Op, inner)
			}
			c.setType(e, inner)
			return inner, nil
		case "~":
			if inner.Kind != KInt {
				return nil, fmt.Errorf("~ expects int, got %s", inner)
			}
			c.setType(e, TpInt)
			return TpInt, nil
		}
		return nil, fmt.Errorf("bad unary %q", x.Op)
	case *ast.BinaryExpr:
		return c.typeBinary(x)
	case *ast.CallExpr:
		return c.typeCall(x)
	case *ast.CastExpr:
		a, err := c.typeExpr(x.Arg)
		if err != nil {
			return nil, err
		}
		var out *Type
		switch x.To {
		case "int":
			out = TpInt
		case "float":
			out = TpFloat
		case "str":
			out = TpStr
		case "bool":
			out = TpBool
		default:
			return nil, fmt.Errorf("cast to %q not supported", x.To)
		}
		c.setType(e, out)
		_ = a
		return out, nil
	case *ast.StructLit:
		return c.typeStructLit(x, e)
	case *ast.ListLit:
		if len(x.Elems) == 0 {
			return nil, fmt.Errorf("list literal: cannot infer element type from [] (use an annotation)")
		}
		var elemT *Type
		for i, el := range x.Elems {
			et, err := c.typeExpr(el)
			if err != nil {
				return nil, err
			}
			if et == nil || et.Kind == KNil {
				return nil, fmt.Errorf("list literal: element %d cannot be none without an explicit list type", i)
			}
			if elemT == nil {
				elemT = et
				continue
			}
			if err := c.assignable(elemT, et); err != nil {
				return nil, fmt.Errorf("list literal: element %d: %v", i, err)
			}
		}
		out := &Type{Kind: KList, Elem: elemT}
		c.setType(e, out)
		return out, nil
	case *ast.IndexExpr:
		lt, err := c.typeExpr(x.Left)
		if err != nil {
			return nil, err
		}
		if lt == nil || lt.Kind != KList || lt.Elem == nil {
			return nil, fmt.Errorf("indexing requires list value, got %v", lt)
		}
		it, err := c.typeExpr(x.Index)
		if err != nil {
			return nil, err
		}
		if it == nil || it.Kind != KInt {
			return nil, fmt.Errorf("list index must be int, got %v", it)
		}
		c.setType(e, lt.Elem)
		return lt.Elem, nil
	case *ast.MemberExpr:
		// enum const: EnumName.Variant (EnumName is not a binding)
		if li, isId := x.Left.(*ast.IdentExpr); isId {
			if en, ok := c.enums[li.Name]; ok {
				if _, vok := en.Index[x.Field]; vok {
					return c.checkEnumConst(li.Name, x.Field, e)
				}
				return nil, fmt.Errorf("enum %s has no variant %q", li.Name, x.Field)
			}
		}
		tL, err := c.typeExpr(x.Left)
		if err != nil {
			return nil, err
		}
		if tL != nil && tL.Kind == KList {
			if x.Field == "len" {
				c.setType(e, TpInt)
				return TpInt, nil
			}
			if sug, ok := suggestName(x.Field, []string{"len"}); ok {
				return nil, fmt.Errorf("list has no field %q — did you mean .%q? (or call len(yourList))", x.Field, sug)
			}
			return nil, fmt.Errorf("list has no field %q (use [index], len(list), or .len for length)", x.Field)
		}
		if tL == nil || tL.Kind != KStruct || tL.StructDef == nil {
			if x.Field == "ok" || x.Field == "value" || x.Field == "err" {
				return nil, fmt.Errorf("result field access (.ok/.value/.err) is not part of Kodae v1; use catch")
			}
			return nil, fmt.Errorf("field access .%q on %s (need struct value)", x.Field, tL)
		}
		fdt, ok := tL.StructDef.Fields[x.Field]
		if !ok {
			if sug, ok2 := suggestName(x.Field, tL.StructDef.Order); ok2 {
				return nil, fmt.Errorf("struct %s has no field %q — did you mean %q?", tL.StructName, x.Field, sug)
			}
			return nil, fmt.Errorf("struct %s has no field %q", tL.StructName, x.Field)
		}
		c.setType(e, fdt)
		return fdt, nil
	case *ast.TryUnwrapExpr:
		return nil, fmt.Errorf("? is not supported in Kodae v1; use catch")
	case *ast.ResultCatchExpr:
		if !c.tryOK {
			return nil, fmt.Errorf("result catch: only valid in let init, return value, = right-hand side, or expression statement")
		}
		t, err := c.typeExpr(x.Subj)
		if err != nil {
			return nil, err
		}
		if t == nil || t.Kind == KVoid {
			return nil, fmt.Errorf("catch: left side must produce a value, got %v", t)
		}
		inner := t
		c.push()
		if x.ErrName == "" {
			return nil, fmt.Errorf("catch: need (name) for the error str")
		}
		c.set(x.ErrName, TpStr)
		was := c.tryOK
		c.tryOK = false
		if x.Body != nil {
			c.stmt(x.Body)
		}
		c.tryOK = was
		c.pop()
		c.setType(e, inner)
		return inner, nil
	case *ast.PostfixExpr:
		if x.Op != "++" && x.Op != "--" {
			return nil, fmt.Errorf("postfix op %q not supported", x.Op)
		}
		t, err := c.typeExpr(x.X)
		if err != nil {
			return nil, err
		}
		if t == nil || !t.isNumeric() {
			return nil, fmt.Errorf("%s: need a numeric lvalue, got %v", x.Op, t)
		}
		c.setType(e, t)
		return t, nil
	default:
		return nil, fmt.Errorf("typeExpr: unhandled %T", e)
	}
}

func (c *Checker) typeStructLit(x *ast.StructLit, e ast.Expr) (*Type, error) {
	sdef, ok := c.structs[x.TypeName]
	if !ok {
		return nil, fmt.Errorf("unknown struct type %q in literal", x.TypeName)
	}
	if !c.canSeeRemote(sdef.Pub, sdef.SrcFile) {
		return nil, fmt.Errorf("struct %q is not visible in this file (use pub struct in the defining file)", sdef.Name)
	}
	dup := make(map[string]int, len(sdef.Order))
	present := make(map[string]bool, len(sdef.Order))
	for i, in := range x.Inits {
		if in.Name == "" {
			return nil, fmt.Errorf("struct %s: empty field in literal", x.TypeName)
		}
		if j, has := dup[in.Name]; has {
			return nil, fmt.Errorf("struct %s: duplicate init for field %q (inits at %d and %d)", x.TypeName, in.Name, j, i)
		}
		dup[in.Name] = i
		present[in.Name] = true
	}
	for _, n := range sdef.Order {
		if !present[n] {
			return nil, fmt.Errorf("struct %s: literal must set field %q", x.TypeName, n)
		}
	}
	for _, in := range x.Inits {
		_, fok := sdef.Fields[in.Name]
		if !fok {
			if sug, ok := suggestName(in.Name, sdef.Order); ok {
				return nil, fmt.Errorf("struct %s: unknown field %q in literal — did you mean %q?", x.TypeName, in.Name, sug)
			}
			return nil, fmt.Errorf("struct %s: unknown field %q in literal", x.TypeName, in.Name)
		}
	}
	for _, in := range x.Inits {
		et, err := c.typeExpr(in.Init)
		if err != nil {
			return nil, err
		}
		want := sdef.Fields[in.Name]
		if err := c.assignable(want, et); err != nil {
			return nil, err
		}
	}
	out := StructType(sdef)
	if out == nil {
		return nil, fmt.Errorf("struct %q", x.TypeName)
	}
	c.setType(e, out)
	return out, nil
}

func (c *Checker) checkEnumConst(enName, variant string, e ast.Expr) (*Type, error) {
	en, ok := c.enums[enName]
	if !ok {
		return nil, fmt.Errorf("not an enum: %q", enName)
	}
	if !c.canSeeRemote(en.Pub, en.File) {
		return nil, fmt.Errorf("enum %q is not visible in this file (use pub enum in the defining file)", enName)
	}
	if _, ok = en.Index[variant]; !ok {
		return nil, fmt.Errorf("enum %q has no variant %q", enName, variant)
	}
	et := c.enumTypeFor(en)
	c.setType(e, et)
	return et, nil
}

func (c *Checker) typeBinary(b *ast.BinaryExpr) (*Type, error) {
	switch b.Op {
	case "&&", "||":
		l, err := c.typeExpr(b.L)
		if err != nil {
			return nil, err
		}
		r, err2 := c.typeExpr(b.R)
		if err2 != nil {
			return nil, err2
		}
		if l.Kind != KBool || r.Kind != KBool {
			return nil, fmt.Errorf("%s: expected bool, got %s and %s", b.Op, l, r)
		}
		c.setType(b, TpBool)
		return TpBool, nil
	case "==", "!=":
		return c.typeEq(b)
	case "<", ">", "<=", ">=":
		return c.typeCompare(b)
	case "+", "-", "*", "/", "%":
		return c.typeArith(b)
	case "&", "|", "^":
		l, err := c.typeExpr(b.L)
		if err != nil {
			return nil, err
		}
		r, err2 := c.typeExpr(b.R)
		if err2 != nil {
			return nil, err2
		}
		if l.Kind != KInt || r.Kind != KInt {
			return nil, fmt.Errorf("%s: expected int, got %s and %s", b.Op, l, r)
		}
		c.setType(b, TpInt)
		return TpInt, nil
	case "..":
		l, err := c.typeExpr(b.L)
		if err != nil {
			return nil, err
		}
		r, err2 := c.typeExpr(b.R)
		if err2 != nil {
			return nil, err2
		}
		if l.Kind != KInt || r.Kind != KInt {
			return nil, fmt.Errorf(".. requires int bounds, have %s .. %s", l, r)
		}
		c.setType(b, TpRange)
		return TpRange, nil
	default:
		return nil, fmt.Errorf("internal: unknown op %q", b.Op)
	}
}

func (c *Checker) typeEq(b *ast.BinaryExpr) (*Type, error) {
	_, lNone := b.L.(*ast.NoneExpr)
	_, rNone := b.R.(*ast.NoneExpr)
	if lNone || rNone {
		if lNone && rNone {
			c.setType(b, TpBool)
			return TpBool, nil
		}
		var tOpt *Type
		var err error
		if rNone {
			tOpt, err = c.typeExpr(b.L)
		} else {
			tOpt, err = c.typeExpr(b.R)
		}
		if err != nil {
			return nil, err
		}
		if tOpt.Kind == KOptional {
			c.setType(b, TpBool)
			return TpBool, nil
		}
		return nil, fmt.Errorf("compare to none: the other side must be optional, got %s", tOpt)
	}

	l, e1 := c.typeExpr(b.L)
	if e1 != nil {
		return nil, e1
	}
	r, e2 := c.typeExpr(b.R)
	if e2 != nil {
		return nil, e2
	}
	if l.Kind == KEnum && r.Kind == KEnum && l.EnumRef == r.EnumRef {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.isNumeric() && r.isNumeric() {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.equal(r) && (l.Kind == KStr || l.Kind == KBool) {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.equal(r) && l.Kind == KInt {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.equal(r) && l.Kind == KFloat {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.equal(r) && l.Kind == KStr {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.equal(r) && l.Kind == KEnum {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.Kind == KStruct && r.Kind == KStruct && l.equal(r) {
		if b.Op == "==" || b.Op == "!=" {
			c.setType(b, TpBool)
			return TpBool, nil
		}
	}
	return nil, fmt.Errorf("incompatible in %s: %s vs %s", b.Op, l, r)
}

func (c *Checker) typeCompare(b *ast.BinaryExpr) (*Type, error) {
	l, e1 := c.typeExpr(b.L)
	if e1 != nil {
		return nil, e1
	}
	r, e2 := c.typeExpr(b.R)
	if e2 != nil {
		return nil, e2
	}
	if l.isNumeric() && r.isNumeric() {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	if l.Kind == KEnum && r.Kind == KEnum && l.EnumRef == r.EnumRef {
		c.setType(b, TpBool)
		return TpBool, nil
	}
	return nil, fmt.Errorf("invalid comparison: %s vs %s", l, r)
}

func (c *Checker) typeArith(b *ast.BinaryExpr) (*Type, error) {
	if b.Op == "+" {
		ls, e1 := c.typeExpr(b.L)
		if e1 != nil {
			return nil, e1
		}
		rs, e2 := c.typeExpr(b.R)
		if e2 != nil {
			return nil, e2
		}
		if ls.Kind == KStr && rs.Kind == KStr {
			c.setType(b, TpStr)
			return TpStr, nil
		}
		if ls.Kind == KStr && coercesToString(rs) {
			c.setType(b, TpStr)
			return TpStr, nil
		}
		if rs.Kind == KStr && coercesToString(ls) {
			c.setType(b, TpStr)
			return TpStr, nil
		}
	}
	l, e1 := c.typeExpr(b.L)
	if e1 != nil {
		return nil, e1
	}
	r, e2 := c.typeExpr(b.R)
	if e2 != nil {
		return nil, e2
	}
	if !l.isNumeric() || !r.isNumeric() {
		if b.Op == "+" {
			if l.Kind == KStr && (r.isNumeric() || r.Kind == KEnum) {
				return nil, fmt.Errorf("cannot combine %s and %s with + in this order (put the string on the right, or use str() on the number)", l, r)
			}
			if r.Kind == KStr && (l.isNumeric() || l.Kind == KEnum) {
				return nil, fmt.Errorf("cannot add %s and %s (use str(...) on the number, or use \"...\" for text)", l, r)
			}
			if (l.isNumeric() || l.Kind == KEnum) && (r.isNumeric() || r.Kind == KEnum) {
				return nil, fmt.Errorf("cannot + types %s and %s (int and float are ok together; for strings use \"...\" or str(...))", l, r)
			}
			return nil, fmt.Errorf("cannot + types %s and %s (str+str, str+number, or two numbers)", l, r)
		}
		return nil, fmt.Errorf("arithmetic: need two numbers, got %s and %s", l, r)
	}
	k := KInt
	if l.Kind == KFloat || r.Kind == KFloat {
		k = KFloat
	}
	var t *Type
	if k == KFloat {
		t = TpFloat
	} else {
		t = TpInt
	}
	c.setType(b, t)
	return t, nil
}

func peelCallFunc(e ast.Expr) (name string, ok bool) {
	for {
		if pe, o := e.(*ast.ParenExpr); o {
			e = pe.Inner
			continue
		}
		if id, o := e.(*ast.IdentExpr); o {
			return id.Name, true
		}
		return "", false
	}
}

func (c *Checker) typeMethodCall(x *ast.CallExpr, me *ast.MemberExpr) (*Type, error) {
	recvT, err := c.typeExpr(me.Left)
	if err != nil {
		return nil, err
	}
	if recvT != nil && recvT.Kind == KStr {
		switch me.Field {
		case "len":
			if len(x.Args) != 0 {
				return nil, fmt.Errorf("str.len: no arguments")
			}
			c.setType(x, TpInt)
			return TpInt, nil
		case "upper", "lower", "trim", "reverse", "repeat":
			if (me.Field == "repeat" && len(x.Args) != 1) || (me.Field != "repeat" && len(x.Args) != 0) {
				return nil, fmt.Errorf("str.%s: wrong number of arguments", me.Field)
			}
			if me.Field == "repeat" {
				at, err := c.typeExpr(x.Args[0])
				if err != nil {
					return nil, err
				}
				if at == nil || at.Kind != KInt {
					return nil, fmt.Errorf("str.repeat: int expected")
				}
			}
			c.setType(x, TpStr)
			return TpStr, nil
		case "contains", "starts", "ends":
			if len(x.Args) != 1 {
				return nil, fmt.Errorf("str.%s: need one string", me.Field)
			}
			at, err := c.typeExpr(x.Args[0])
			if err != nil {
				return nil, err
			}
			if at == nil || at.Kind != KStr {
				return nil, fmt.Errorf("str.%s: string expected", me.Field)
			}
			c.setType(x, TpBool)
			return TpBool, nil
		case "replace":
			if len(x.Args) != 2 {
				return nil, fmt.Errorf("str.replace: need two strings (old, new)")
			}
			for _, a := range x.Args {
				at, err := c.typeExpr(a)
				if err != nil {
					return nil, err
				}
				if at == nil || at.Kind != KStr {
					return nil, fmt.Errorf("str.replace: strings expected")
				}
			}
			c.setType(x, TpStr)
			return TpStr, nil
		case "split":
			if len(x.Args) != 1 {
				return nil, fmt.Errorf("str.split: need delimiter string")
			}
			at, err := c.typeExpr(x.Args[0])
			if err != nil {
				return nil, err
			}
			if at == nil || at.Kind != KStr {
				return nil, fmt.Errorf("str.split: string expected")
			}
			c.setType(x, TpListStr)
			return TpListStr, nil
		case "slice":
			if len(x.Args) != 2 {
				return nil, fmt.Errorf("str.slice: need start and end indices")
			}
			for _, a := range x.Args {
				at, err := c.typeExpr(a)
				if err != nil {
					return nil, err
				}
				if at == nil || at.Kind != KInt {
					return nil, fmt.Errorf("str.slice: indices must be int")
				}
			}
			c.setType(x, TpStr)
			return TpStr, nil
		case "is_empty", "is_number":
			if len(x.Args) != 0 {
				return nil, fmt.Errorf("str.%s: no arguments", me.Field)
			}
			c.setType(x, TpBool)
			return TpBool, nil
		default:
			return nil, fmt.Errorf("str has no method %q", me.Field)
		}
	}
	if recvT != nil && recvT.Kind == KList {
		switch me.Field {
		case "push", "pop", "append", "remove", "first", "last", "reverse", "sort", "shuffle":
			if me.Field == "pop" || me.Field == "first" || me.Field == "last" || me.Field == "reverse" || me.Field == "sort" || me.Field == "shuffle" {
				if len(x.Args) != 0 {
					return nil, fmt.Errorf("list.%s: no arguments", me.Field)
				}
				res := TpVoid
				if me.Field == "pop" || me.Field == "first" || me.Field == "last" {
					res = recvT.Elem
				}
				c.setType(x, res)
				return res, nil
			}
			if me.Field == "remove" {
				if len(x.Args) != 1 {
					return nil, fmt.Errorf("list.remove: need index")
				}
				at, err := c.typeExpr(x.Args[0])
				if err != nil {
					return nil, err
				}
				if at == nil || at.Kind != KInt {
					return nil, fmt.Errorf("list.remove: index must be int")
				}
				c.setType(x, recvT.Elem)
				return recvT.Elem, nil
			}
			if len(x.Args) != 1 {
				return nil, fmt.Errorf("list.%s: need one argument", me.Field)
			}
			at, err := c.typeExpr(x.Args[0])
			if err != nil {
				return nil, err
			}
			want := recvT.Elem
			if me.Field == "append" {
				want = recvT
			}
			if err := c.assignable(want, at); err != nil {
				return nil, fmt.Errorf("list.%s: %v", me.Field, err)
			}
			c.setType(x, TpVoid)
			return TpVoid, nil
		case "find", "filter", "map", "count", "any", "all":
			// lambda/callback methods - only basic signatures for now
			if len(x.Args) != 1 {
				return nil, fmt.Errorf("list.%s: need one callback", me.Field)
			}
			// ... placeholder for advanced lambda check ...
			c.setType(x, TpVoid) // actually map returns list[U] etc.
			return TpVoid, nil
		case "sum":
			if len(x.Args) != 0 {
				return nil, fmt.Errorf("list.sum: no arguments")
			}
			if !recvT.Elem.isNumeric() {
				return nil, fmt.Errorf("list.sum: numeric elements expected")
			}
			c.setType(x, recvT.Elem)
			return recvT.Elem, nil
		}
	}
	if recvT == nil || recvT.Kind != KStruct || recvT.StructDef == nil {
		return nil, fmt.Errorf("method call: receiver must be a struct, got %v", recvT)
	}
	mangled := recvT.StructName + "_" + me.Field
	f := c.fns[mangled]
	if f == nil {
		return nil, fmt.Errorf("no method %q for struct %s (expected fn %q)", me.Field, recvT.StructName, mangled)
	}
	if !c.canSeeRemote(f.Pub, f.File) {
		return nil, fmt.Errorf("method %q is not visible in this file (use pub fn in the defining file)", me.Field)
	}
	if f.Params == nil {
		return nil, fmt.Errorf("method %q: internal error", me.Field)
	}
	if len(f.Params) < 1 || f.Params[0].Name != "self" {
		return nil, fmt.Errorf("method %q: first parameter must be self", me.Field)
	}
	wantSelf, err := c.resolveType(f.Params[0].T)
	if err != nil {
		return nil, err
	}
	if err := c.assignable(wantSelf, recvT); err != nil {
		return nil, fmt.Errorf("self: %v", err)
	}
	if 1+len(x.Args) != len(f.Params) {
		return nil, fmt.Errorf("%s: need %d args, got %d (including self)", me.Field, len(f.Params), 1+len(x.Args))
	}
	for i := 1; i < len(f.Params); i++ {
		at, e2 := c.typeExpr(x.Args[i-1])
		if e2 != nil {
			return nil, e2
		}
		pw, perr := c.resolveType(f.Params[i].T)
		if perr == nil {
			if aerr := c.assignable(pw, at); aerr != nil {
				return nil, fmt.Errorf("arg: %v", aerr)
			}
		}
	}
	if f.Return == nil {
		c.setType(x, TpVoid)
		return TpVoid, nil
	}
	rt, err := c.resolveType(f.Return)
	if err != nil {
		return nil, err
	}
	c.setType(x, rt)
	return rt, nil
}

func (c *Checker) typeExternCall(x *ast.CallExpr, ex *ast.ExternDecl) (*Type, error) {
	c.externTypeCtx++
	defer func() { c.externTypeCtx-- }()
	var wantFixed int
	variadic := false
	for _, p := range ex.Params {
		if p.Dots {
			variadic = true
			break
		}
		wantFixed++
	}
	if !variadic {
		if len(x.Args) != wantFixed {
			return nil, fmt.Errorf("%s: need %d args, got %d", ex.Name, wantFixed, len(x.Args))
		}
	} else if len(x.Args) < wantFixed {
		return nil, fmt.Errorf("%s: need at least %d args, got %d", ex.Name, wantFixed, len(x.Args))
	}
	argI := 0
	for i, p := range ex.Params {
		if p.Dots {
			for j := argI; j < len(x.Args); j++ {
				if _, err := c.typeExpr(x.Args[j]); err != nil {
					return nil, err
				}
			}
			break
		}
		at, err := c.typeExpr(x.Args[argI])
		if err != nil {
			return nil, err
		}
		pwant, perr := c.resolveType(p.T)
		if perr == nil {
			if aerr := c.assignableExtern(pwant, at); aerr != nil {
				return nil, fmt.Errorf("arg %d: %v", i, aerr)
			}
		}
		argI++
	}
	var rt *Type
	var err error
	if ex.Return != nil {
		rt, err = c.resolveType(ex.Return)
		if err != nil {
			return nil, err
		}
	} else {
		rt = TpVoid
	}
	if rt == nil {
		rt = TpVoid
	}
	if rt != nil && rt.Kind == KF32 {
		c.setType(x, TpFloat)
		return TpFloat, nil
	}
	if rt != nil && (rt.Kind == KI32 || rt.Kind == KU32 || rt.Kind == KU8) {
		c.setType(x, TpInt)
		return TpInt, nil
	}
	c.setType(x, rt)
	return rt, nil
}

func (c *Checker) assignableExtern(want, got *Type) error {
	if want == nil || got == nil {
		return nil
	}
	if want.Kind == KPtr && want.Pointee != nil && want.Pointee.Kind == KByte && got.equal(TpStr) {
		return nil
	}
	if want.Kind == KF32 {
		if got.Kind == KInt || got.Kind == KFloat {
			return nil
		}
		return c.assignable(want, got)
	}
	if want.Kind == KI32 || want.Kind == KU32 || want.Kind == KU8 {
		if got.Kind == KInt || got.Kind == KFloat {
			return nil
		}
		return c.assignable(want, got)
	}
	return c.assignable(want, got)
}

func (c *Checker) typeCall(x *ast.CallExpr) (*Type, error) {
	if me, ok := x.Fun.(*ast.MemberExpr); ok {
		return c.typeMethodCall(x, me)
	}
	name, ok := peelCallFunc(x.Fun)
	if !ok {
		return nil, fmt.Errorf("only direct calls f(...), (f)(...), or o.method(...)")
	}

	checkStrArg := func(name string) error {
		if len(x.Args) != 1 {
			return fmt.Errorf("%s: need one string argument", name)
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return err
		}
		if at == nil || at.Kind != KStr {
			return fmt.Errorf("%s: string expected", name)
		}
		return nil
	}

	// built-ins
	switch name {
	case "print", "printn":
		c.inf.UsesConsole = true
		if name == "print" && len(x.Args) == 0 {
			return nil, fmt.Errorf("print needs at least one argument")
		}
		for _, a := range x.Args {
			if _, err := c.typeExpr(a); err != nil {
				return nil, err
			}
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "input", "input_int", "input_float":
		c.inf.UsesConsole = true
		if err := checkStrArg(name); err != nil {
			return nil, err
		}
		res := TpStr
		switch name {
		case "input_int":
			res = TpInt
		case "input_float":
			res = TpFloat
		}
		c.setType(x, res)
		return res, nil
	case "random":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("random: need two int bounds (inclusive)")
		}
		a, e1 := c.typeExpr(x.Args[0])
		b, e2 := c.typeExpr(x.Args[1])
		if e1 != nil || e2 != nil {
			return nil, firstErr(e1, e2)
		}
		if a == nil || a.Kind != KInt || b == nil || b.Kind != KInt {
			return nil, fmt.Errorf("random: both bounds must be int")
		}
		c.setType(x, TpInt)
		return TpInt, nil
	case "swap":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("swap: need two variables")
		}
		a, e1 := c.typeExpr(x.Args[0])
		b, e2 := c.typeExpr(x.Args[1])
		if e1 != nil || e2 != nil {
			return nil, firstErr(e1, e2)
		}
		if !a.equal(b) {
			return nil, fmt.Errorf("swap: types must match, got %v and %v", a, b)
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "in_range":
		if len(x.Args) != 3 {
			return nil, fmt.Errorf("in_range: need (val, min, max)")
		}
		for _, a := range x.Args {
			if _, err := c.typeExpr(a); err != nil {
				return nil, err
			}
		}
		c.setType(x, TpBool)
		return TpBool, nil
	case "in_rect":
		if len(x.Args) != 6 {
			return nil, fmt.Errorf("in_rect: need (px, py, rx, ry, rw, rh)")
		}
		for _, a := range x.Args {
			if _, err := c.typeExpr(a); err != nil {
				return nil, err
			}
		}
		c.setType(x, TpBool)
		return TpBool, nil
	case "random_float":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("random_float: need two float bounds")
		}
		for _, a := range x.Args {
			at, err := c.typeExpr(a)
			if err != nil {
				return nil, err
			}
			if !at.isNumeric() {
				return nil, fmt.Errorf("random_float: numbers expected")
			}
		}
		c.setType(x, TpFloat)
		return TpFloat, nil
	case "chance":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("chance: need one int (percentage 0-100)")
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if at == nil || at.Kind != KInt {
			return nil, fmt.Errorf("chance: int expected")
		}
		c.setType(x, TpBool)
		return TpBool, nil
	case "random_bool":
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("random_bool: no arguments")
		}
		c.setType(x, TpBool)
		return TpBool, nil
	case "random_pick":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("random_pick: need one list")
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if at == nil || at.Kind != KList || at.Elem == nil {
			return nil, fmt.Errorf("random_pick: list expected")
		}
		c.setType(x, at.Elem)
		return at.Elem, nil
	case "time", "timer_start":
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("%s: no arguments", name)
		}
		c.setType(x, TpFloat)
		return TpFloat, nil
	case "time_ms":
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("time_ms: no arguments")
		}
		c.setType(x, TpInt)
		return TpInt, nil
	case "wait", "countdown":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("%s: need one number", name)
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if !at.isNumeric() {
			return nil, fmt.Errorf("%s: number expected", name)
		}
		res := TpVoid
		if name == "countdown" {
			res = TpFloat
		}
		c.setType(x, res)
		return res, nil
	case "wait_ms":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("wait_ms: need one int (milliseconds)")
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if at == nil || at.Kind != KInt {
			return nil, fmt.Errorf("wait_ms: int expected")
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "timer_elapsed", "countdown_done":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("%s: need one argument", name)
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if name == "timer_elapsed" {
			if at == nil || at.Kind != KFloat {
				return nil, fmt.Errorf("timer_elapsed: float expected")
			}
			c.setType(x, TpFloat)
			return TpFloat, nil
		} else {
			if at == nil || at.Kind != KFloat {
				return nil, fmt.Errorf("countdown_done: float expected")
			}
			c.setType(x, TpBool)
			return TpBool, nil
		}
	case "clear_screen":
		c.inf.UsesConsole = true
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("clear_screen: no arguments")
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "len":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("len: need one argument")
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if at == nil || at.Kind != KList {
			return nil, fmt.Errorf("len: argument must be list")
		}
		c.setType(x, TpInt)
		return TpInt, nil
	case "ok", "err":
		return nil, fmt.Errorf("%s(...) is not supported in Kodae v1; use catch", name)
	case "int", "float", "str", "bool":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("%s: need one argument", name)
		}
		_, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		var t *Type
		switch name {
		case "int":
			t = TpInt
		case "float":
			t = TpFloat
		case "str":
			t = TpStr
		case "bool":
			t = TpBool
		}
		c.setType(x, t)
		return t, nil
	case "min", "max", "abs", "sqrt", "floor", "ceil", "round", "sin", "cos", "tan", "log":
		unary := name == "abs" || name == "sqrt" || name == "floor" || name == "ceil" || name == "round" || name == "sin" || name == "cos" || name == "tan" || name == "log"
		if unary {
			if len(x.Args) != 1 {
				return nil, fmt.Errorf("%s: need one number", name)
			}
			tt, err := c.typeExpr(x.Args[0])
			if err != nil {
				return nil, err
			}
			if name == "log" && tt.Kind == KStr {
				c.setType(x, TpVoid)
				return TpVoid, nil
			}
			if !tt.isNumeric() {
				return nil, fmt.Errorf("%s: number expected", name)
			}
			res := TpFloat
			if name == "abs" && tt.Kind == KInt {
				res = TpInt
			}
			c.setType(x, res)
			return res, nil
		}
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("%s: need two numbers", name)
		}
		a, e1 := c.typeExpr(x.Args[0])
		b, e2 := c.typeExpr(x.Args[1])
		if e1 != nil || e2 != nil {
			return nil, firstErr(e1, e2)
		}
		if !a.isNumeric() || !b.isNumeric() {
			return nil, fmt.Errorf("%s: need numbers", name)
		}
		tt := TpInt
		if a.Kind == KFloat || b.Kind == KFloat {
			tt = TpFloat
		}
		c.setType(x, tt)
		return tt, nil
	case "pow", "atan2":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("%s: need two arguments", name)
		}
		for _, a := range x.Args {
			at, err := c.typeExpr(a)
			if err != nil {
				return nil, err
			}
			if !at.isNumeric() {
				return nil, fmt.Errorf("%s: numbers expected", name)
			}
		}
		c.setType(x, TpFloat)
		return TpFloat, nil
	case "clamp":
		if len(x.Args) != 3 {
			return nil, fmt.Errorf("clamp: need 3 numbers (val, min, max)")
		}
		isFloat := false
		for _, a := range x.Args {
			at, err := c.typeExpr(a)
			if err != nil {
				return nil, err
			}
			if !at.isNumeric() {
				return nil, fmt.Errorf("clamp: numbers expected")
			}
			if at.Kind == KFloat {
				isFloat = true
			}
		}
		res := TpInt
		if isFloat {
			res = TpFloat
		}
		c.setType(x, res)
		return res, nil
	case "lerp", "map":
		expected := 3
		if name == "map" {
			expected = 5
		}
		if len(x.Args) != expected {
			return nil, fmt.Errorf("%s: wrong number of arguments", name)
		}
		for _, a := range x.Args {
			at, err := c.typeExpr(a)
			if err != nil {
				return nil, err
			}
			if !at.isNumeric() {
				return nil, fmt.Errorf("%s: numbers expected", name)
			}
		}
		c.setType(x, TpFloat)
		return TpFloat, nil
	case "distance", "angle_to":
		if len(x.Args) != 4 {
			return nil, fmt.Errorf("%s: need 4 numbers (x1, y1, x2, y2)", name)
		}
		for _, a := range x.Args {
			at, err := c.typeExpr(a)
			if err != nil {
				return nil, err
			}
			if !at.isNumeric() {
				return nil, fmt.Errorf("%s: numbers expected", name)
			}
		}
		c.setType(x, TpFloat)
		return TpFloat, nil
	case "read_file", "delete_file", "make_folder", "delete_folder", "folder_exists", "file_exists", "os_name", "clipboard_get":
		if name == "os_name" || name == "clipboard_get" {
			if len(x.Args) != 0 {
				return nil, fmt.Errorf("%s: no arguments", name)
			}
		} else {
			if err := checkStrArg(name); err != nil {
				return nil, err
			}
		}
		res := TpStr
		if name == "file_exists" || name == "folder_exists" || name == "delete_file" || name == "delete_folder" || name == "make_folder" {
			res = TpBool
		}
		c.setType(x, res)
		return res, nil
	case "write_file", "append_file", "copy_file", "move_file", "clipboard_set", "open_url", "run":
		argCount := 1
		if name == "write_file" || name == "append_file" || name == "copy_file" || name == "move_file" {
			argCount = 2
		}
		if len(x.Args) != argCount {
			return nil, fmt.Errorf("%s: need %d arguments", name, argCount)
		}
		for _, a := range x.Args {
			at, err := c.typeExpr(a)
			if err != nil {
				return nil, err
			}
			if at == nil || at.Kind != KStr {
				return nil, fmt.Errorf("%s: strings expected", name)
			}
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "list_files":
		if err := checkStrArg(name); err != nil {
			return nil, err
		}
		c.setType(x, TpListStr)
		return TpListStr, nil
	case "exit":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("exit: need exit code (int)")
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if at == nil || at.Kind != KInt {
			return nil, fmt.Errorf("exit: int expected")
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "args":
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("args: no arguments")
		}
		c.setType(x, TpListStr)
		return TpListStr, nil
	case "env":
		if err := checkStrArg(name); err != nil {
			return nil, err
		}
		c.setType(x, TpStr)
		return TpStr, nil
	case "is_windows", "is_mac", "is_linux":
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("%s: no arguments", name)
		}
		c.setType(x, TpBool)
		return TpBool, nil
	case "assert":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("assert: need condition (bool) and message (str)")
		}
		cond, e1 := c.typeExpr(x.Args[0])
		msg, e2 := c.typeExpr(x.Args[1])
		if e1 != nil || e2 != nil {
			return nil, firstErr(e1, e2)
		}
		if cond == nil || cond.Kind != KBool || msg == nil || msg.Kind != KStr {
			return nil, fmt.Errorf("assert: need (bool, str)")
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "debug":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("debug: need one value")
		}
		if _, err := c.typeExpr(x.Args[0]); err != nil {
			return nil, err
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "todo":
		if err := checkStrArg(name); err != nil {
			return nil, err
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "benchmark_start":
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("benchmark_start: no arguments")
		}
		c.setType(x, TpFloat)
		return TpFloat, nil
	case "benchmark_end":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("benchmark_end: need timer (float) and label (str)")
		}
		t, e1 := c.typeExpr(x.Args[0])
		l, e2 := c.typeExpr(x.Args[1])
		if e1 != nil || e2 != nil {
			return nil, firstErr(e1, e2)
		}
		if t == nil || t.Kind != KFloat || l == nil || l.Kind != KStr {
			return nil, fmt.Errorf("benchmark_end: need (float, str)")
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "json_read":
		if err := checkStrArg(name); err != nil {
			return nil, err
		}
		c.setType(x, TpAny)
		return TpAny, nil
	case "json_write":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("json_write: need path string and value")
		}
		path, e1 := c.typeExpr(x.Args[0])
		if e1 != nil {
			return nil, e1
		}
		if path == nil || path.Kind != KStr {
			return nil, fmt.Errorf("json_write: path must be string")
		}
		if _, err := c.typeExpr(x.Args[1]); err != nil {
			return nil, err
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	default:
		if ex := c.externs[name]; ex != nil {
			return c.typeExternCall(x, ex)
		}
		f := c.fns[name]
		if f == nil {
			if sug, ok := suggestName(name, c.callableNameCandidates()); ok {
				return nil, fmt.Errorf("unknown function %q — did you mean %q()?", name, sug)
			}
			return nil, fmt.Errorf("unknown function %q", name)
		}
		if !c.canSeeRemote(f.Pub, f.File) {
			return nil, fmt.Errorf("function %q is not visible in this file (use pub fn in the defining file)", name)
		}
		if f.Params == nil {
			f.Params = []ast.Param{}
		}
		if len(f.Params) != len(x.Args) {
			return nil, fmt.Errorf("%s: need %d args, got %d", name, len(f.Params), len(x.Args))
		}
		for i := range f.Params {
			at, err := c.typeExpr(x.Args[i])
			if err != nil {
				return nil, err
			}
			pwant, perr := c.resolveType(f.Params[i].T)
			if perr == nil {
				if aerr := c.assignable(pwant, at); aerr != nil {
					return nil, fmt.Errorf("arg %d: %v", i, aerr)
				}
			}
		}
		if f.Return == nil {
			c.setType(x, TpVoid)
			return TpVoid, nil
		}
		rt, err := c.resolveType(f.Return)
		if err != nil {
			return nil, err
		}
		c.setType(x, rt)
		return rt, nil
	}
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
