package check

import (
	"fmt"

	"clio/internal/ast"
)

func (c *Checker) typeExpr(e ast.Expr) (*Type, error) {
	if c.err != nil {
		return TpInt, c.err
	}
	switch x := e.(type) {
	case *ast.IdentExpr:
		t := c.get(x.Name)
		if t == nil {
			if c.enums[x.Name] != nil {
				return nil, fmt.Errorf("expected %q as value (use %q.Variant for enum value)", x.Name, x.Name)
			}
			if c.fns[x.Name] != nil {
				return nil, fmt.Errorf("cannot use function name %q as a value (did you mean to call it?)", x.Name)
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
		if tL != nil && tL.Kind == KResult {
			var ft *Type
			switch x.Field {
			case "ok":
				ft = TpBool
			case "value":
				if tL.Res == nil {
					return nil, fmt.Errorf("result: value field needs inner type")
				}
				ft = tL.Res
			case "err":
				ft = TpStr
			default:
				return nil, fmt.Errorf("result has no field %q (use .ok, .value, .err)", x.Field)
			}
			c.setType(e, ft)
			return ft, nil
		}
		if tL == nil || tL.Kind != KStruct || tL.StructDef == nil {
			return nil, fmt.Errorf("field access .%q on %s (need struct or result value)", x.Field, tL)
		}
		fdt, ok := tL.StructDef.Fields[x.Field]
		if !ok {
			return nil, fmt.Errorf("struct %s has no field %q", tL.StructName, x.Field)
		}
		c.setType(e, fdt)
		return fdt, nil
	case *ast.TryUnwrapExpr:
		if !c.tryOK {
			return nil, fmt.Errorf("? on result is only valid in: let init, return value, = right-hand side, or expression statement")
		}
		if c.returnWant == nil || c.returnWant.Kind != KResult {
			return nil, fmt.Errorf("? requires the enclosing function to return a result type (propagate errors to caller)")
		}
		t, err := c.typeExpr(x.X)
		if err != nil {
			return nil, err
		}
		if t == nil || t.Kind != KResult {
			return nil, fmt.Errorf("? expects a result[...] value, got %v", t)
		}
		if t.Res == nil {
			return nil, fmt.Errorf("? on malformed result type")
		}
		inner := t.Res
		c.setType(e, inner)
		return inner, nil
	case *ast.ResultCatchExpr:
		if !c.tryOK {
			return nil, fmt.Errorf("result catch: only valid in let init, return value, = right-hand side, or expression statement")
		}
		t, err := c.typeExpr(x.Subj)
		if err != nil {
			return nil, err
		}
		if t == nil || t.Kind != KResult || t.Res == nil {
			return nil, fmt.Errorf("catch: left side must be a result[...] value, got %v", t)
		}
		inner := t.Res
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
	if _, ok = en.Index[variant]; !ok {
		return nil, fmt.Errorf("enum %q has no variant %q", enName, variant)
	}
	et := c.enumTypeFor(en)
	c.setType(e, et)
	return et, nil
}

// typeBinary: assign case handled elsewhere (AssignStmt) — for expr we don't see raw =
func (c *Checker) typeBinary(b *ast.BinaryExpr) (*Type, error) {
	// disambiguation: parse produced (a = b) as binary with op "=" for assign — actually assign is AssignStmt, not in typeExpr. Good.
	// a +=  b as Assign. Good.
	// a == b, etc.
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
			// C-style: allow && on ints? no — for Clio, bool
			// if user used word `and` token maps to && — we still require bool? Spec says boolean logic. For MVP require bool; user can use comparisons.
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
			tOpt, err = c.typeExpr(b.L) // t == none, t != none: optional t on left
		} else {
			tOpt, err = c.typeExpr(b.R) // none == t, none != t
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

// typeCompare: numeric or enum same type? Spec: only numeric comparison for < — enums?
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
		// str + (int|float|bool|enum) or symmetric — value coerced to str
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
			return nil, fmt.Errorf("cannot + types %s and %s (use + only for str+str or number+number)", l, r)
		}
		return nil, fmt.Errorf("arithmetic: expected numbers, have %s and %s", l, r)
	}
	// promote to float
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

// peelCallFunc strips grouping parens and returns the callee name for a direct call f(...).
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

// typeMethodCall: x.m(args) -> fn mangle(Struct, m) with (receiver, args...)
func (c *Checker) typeMethodCall(x *ast.CallExpr, me *ast.MemberExpr) (*Type, error) {
	recvT, err := c.typeExpr(me.Left)
	if err != nil {
		return nil, err
	}
	if recvT == nil || recvT.Kind != KStruct || recvT.StructDef == nil {
		return nil, fmt.Errorf("method call: receiver must be a struct, got %v", recvT)
	}
	mangled := recvT.StructName + "_" + me.Field
	f := c.fns[mangled]
	if f == nil {
		return nil, fmt.Errorf("no method %q for struct %s (expected fn %q)", me.Field, recvT.StructName, mangled)
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
	c.setType(x, rt)
	return rt, nil
}

// assignableExtern allows str to ptr[byte] for C interop.
func (c *Checker) assignableExtern(want, got *Type) error {
	if want == nil || got == nil {
		return nil
	}
	if want.Kind == KPtr && want.Pointee != nil && want.Pointee.Kind == KByte && got.equal(TpStr) {
		return nil
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
	// built-ins
	switch name {
	case "print":
		// 1..n args, each any printable, entire thing is a statement; type void — use void as "unit"
		if len(x.Args) == 0 {
			return nil, fmt.Errorf("print needs at least one argument")
		}
		for _, a := range x.Args {
			if _, err := c.typeExpr(a); err != nil {
				return nil, err
			}
		}
		// "void" call used as expression — disallow?
		// in Clio, print(...) as ExprStmt, not in expression position usually; if in `let x = print(1)` — error at assign
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "input":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("input: need one string (prompt)")
		}
		pT, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if pT == nil || pT.Kind != KStr {
			return nil, fmt.Errorf("input: prompt must be str")
		}
		c.setType(x, TpStr)
		return TpStr, nil
	case "random":
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("random: need two int bounds (inclusive)")
		}
		a, e1 := c.typeExpr(x.Args[0])
		if e1 != nil {
			return nil, e1
		}
		b, e2 := c.typeExpr(x.Args[1])
		if e2 != nil {
			return nil, e2
		}
		if a == nil || a.Kind != KInt || b == nil || b.Kind != KInt {
			return nil, fmt.Errorf("random: both bounds must be int")
		}
		c.setType(x, TpInt)
		return TpInt, nil
	case "clear_screen":
		if len(x.Args) != 0 {
			return nil, fmt.Errorf("clear_screen: no arguments")
		}
		c.setType(x, TpVoid)
		return TpVoid, nil
	case "ok":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("ok: need one argument")
		}
		at, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if c.inReturn && c.returnWant != nil && c.returnWant.Kind == KResult {
			if want := c.returnWant.Res; want != nil {
				if e := c.assignable(want, at); e != nil {
					return nil, fmt.Errorf("ok: %v", e)
				}
			}
			c.setType(x, c.returnWant)
			return c.returnWant, nil
		}
		r := &Type{Kind: KResult, Res: at}
		c.setType(x, r)
		return r, nil
	case "err":
		if len(x.Args) != 1 {
			return nil, fmt.Errorf("err: need one string argument")
		}
		st, err := c.typeExpr(x.Args[0])
		if err != nil {
			return nil, err
		}
		if st == nil || st.Kind != KStr {
			return nil, fmt.Errorf("err: message must be str")
		}
		if c.inReturn && c.returnWant != nil && c.returnWant.Kind == KResult {
			c.setType(x, c.returnWant)
			return c.returnWant, nil
		}
		r := &Type{Kind: KResult, Res: TpStr}
		c.setType(x, r)
		return r, nil
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
	case "min", "max", "abs":
		if name == "abs" {
			if len(x.Args) != 1 {
				return nil, fmt.Errorf("abs: need one number")
			}
			tt, err := c.typeExpr(x.Args[0])
			if err != nil {
				return nil, err
			}
			if !tt.isNumeric() {
				return nil, fmt.Errorf("abs: number expected")
			}
			c.setType(x, tt)
			return tt, nil
		}
		if len(x.Args) != 2 {
			return nil, fmt.Errorf("%s: need two numbers", name)
		}
		a, e1 := c.typeExpr(x.Args[0])
		if e1 != nil {
			return nil, e1
		}
		b, e2 := c.typeExpr(x.Args[1])
		if e2 != nil {
			return nil, e2
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
	default:
		if ex := c.externs[name]; ex != nil {
			return c.typeExternCall(x, ex)
		}
		// user fn
		f := c.fns[name]
		if f == nil {
			return nil, fmt.Errorf("unknown function %q", name)
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
