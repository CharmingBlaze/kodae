package codegen

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"clio/internal/ast"
	"clio/internal/check"
	"clio/internal/cruntime"
)

var cReserved = map[string]struct{}{
	"if": {}, "for": {}, "return": {}, "int": {}, "void": {}, "static": {}, "const": {},
	"char": {}, "long": {}, "break": {}, "case": {}, "default": {}, "do": {}, "else": {},
	"struct": {}, "switch": {}, "auto": {}, "double": {}, "float": {}, "short": {},
	"register": {}, "unsigned": {}, "sizeof": {}, "goto": {}, "continue": {}, "enum": {},
	"union": {}, "extern": {}, "inline": {}, "restrict": {}, "typeof": {}, "and": {},
	"or": {}, "not": {},
}

var cidBuiltin = map[string]struct{}{
	"print": {}, "int": {}, "float": {}, "str": {}, "bool": {}, "min": {}, "max": {}, "abs": {},
	"input": {}, "random": {}, "clear_screen": {},
	"main": {}, "printf": {}, "malloc": {}, "free": {}, "memcpy": {}, "memset": {},
	"strlen": {}, "true": {}, "false": {},
}

func cid(s string) string {
	if _, ok := cReserved[s]; ok {
		return "u_" + s
	}
	if _, ok := cidBuiltin[s]; ok {
		return "u_" + s
	}
	return s
}

func cT(t *check.Type) string {
	if t == nil {
		return "void"
	}
	switch t.Kind {
	case check.KInt:
		return "int64_t"
	case check.KFloat:
		return "double"
	case check.KStr:
		return "clio_str"
	case check.KBool:
		return "bool"
	case check.KEnum:
		return "int64_t"
	case check.KVoid:
		return "void"
	case check.KNil:
		return "int64_t"
	case check.KRange:
		return "int64_t"
	case check.KStruct:
		return cStructCName(t.StructName)
	case check.KOptional:
		return cOptStruct(t.Opt)
	case check.KByte:
		return "uint8_t"
	case check.KPtr:
		if t.Pointee != nil {
			if t.Pointee.Kind == check.KByte {
				return "const char*"
			}
		}
		if t.Pointee == nil {
			return "const void*"
		}
		if t.Pointee.Kind == check.KInt {
			return "int64_t*"
		}
		if t.Pointee.Kind == check.KFloat {
			return "double*"
		}
		if t.Pointee.Kind == check.KStr {
			return "clio_str*"
		}
		if t.Pointee.Kind == check.KBool {
			return "bool*"
		}
		return "const void*"
	case check.KResult:
		return cResultCName(t.Res)
	default:
		return "int64_t"
	}
}

// cResultCName is the C typedef for result[T] (value, err, ok — see build-spec).
func cResultCName(res *check.Type) string {
	if res == nil {
		return "clio_res_i64"
	}
	switch res.Kind {
	case check.KInt, check.KEnum:
		return "clio_res_i64"
	case check.KFloat:
		return "clio_res_f64"
	case check.KStr:
		return "clio_res_str"
	case check.KBool:
		return "clio_res_bool"
	default:
		return "clio_res_i64"
	}
}

// cStructCName is the C typedef name for a Clio struct.
func cStructCName(n string) string { return "S_" + cid(n) }

func cStructTagName(n string) string { return "s_" + cid(n) + "_" }

// cParamT is the C type for a function parameter (struct types are by pointer).
func cParamT(t *check.Type) string {
	if t != nil && t.Kind == check.KStruct {
		return cStructCName(t.StructName) + "*"
	}
	return cT(t)
}

func cOptStruct(inner *check.Type) string {
	if inner == nil {
		return "clio_opt_i64"
	}
	switch inner.Kind {
	case check.KInt, check.KEnum:
		return "clio_opt_i64"
	case check.KFloat:
		return "clio_opt_f64"
	case check.KStr:
		return "clio_opt_str"
	case check.KBool:
		return "clio_opt_bool"
	default:
		return "clio_opt_i64"
	}
}

func cQuote(s string) string {
	return strconv.Quote(s)
}

type emitter struct {
	inf     *check.Info
	globals map[string]struct{}

	curFn         *ast.FnDecl
	inMain        bool
	params        map[string]struct{}
	structPtrParam map[string]bool
	locals        []map[string]struct{}
	retWant *check.Type
	// fnDefers: C lines to run at return / end of function (LIFO; see defer).
	fnDefers []string
	// tryPre: C lines for result? (error propagation) — flushed before the surrounding statement.
	tryPre []string
	tryN   int
	rcN    int // c_rc_* temps for `result catch` lowering
}

func (em *emitter) pushScope() {
	em.locals = append(em.locals, make(map[string]struct{}))
}

func (em *emitter) popScope() {
	if len(em.locals) == 0 {
		return
	}
	em.locals = em.locals[:len(em.locals)-1]
}

// emitFnDefers runs and clears the pending defer list (LIFO, C order).
func (em *emitter) emitFnDefers(out *bytes.Buffer) {
	for i := len(em.fnDefers) - 1; i >= 0; i-- {
		out.WriteString(em.fnDefers[i])
		if !strings.HasSuffix(em.fnDefers[i], ";") {
			out.WriteString(";")
		}
		out.WriteByte('\n')
	}
	em.fnDefers = em.fnDefers[:0]
}

func (em *emitter) addLocal(name string) {
	if len(em.locals) == 0 {
		em.pushScope()
	}
	em.locals[len(em.locals)-1][name] = struct{}{}
}

func (em *emitter) isLocal(name string) bool {
	for i := len(em.locals) - 1; i >= 0; i-- {
		if _, ok := em.locals[i][name]; ok {
			return true
		}
	}
	return false
}

func (em *emitter) isParam(name string) bool {
	if em.params == nil {
		return false
	}
	_, ok := em.params[name]
	return ok
}

func (em *emitter) resolveTypeExpr(tx *ast.TypeExpr) (*check.Type, error) {
	if tx == nil {
		return check.TpInt, nil
	}
	if tx.PtrInner != nil {
		inner, err := em.resolveTypeExpr(tx.PtrInner)
		if err != nil {
			return nil, err
		}
		pt := &check.Type{Kind: check.KPtr, Pointee: inner}
		if tx.Optional {
			return &check.Type{Kind: check.KOptional, Opt: pt}, nil
		}
		return pt, nil
	}
	if tx.ResultInner != nil {
		inner, err := em.resolveTypeExpr(tx.ResultInner)
		if err != nil {
			return nil, err
		}
		rt := &check.Type{Kind: check.KResult, Res: inner}
		if tx.Optional {
			return &check.Type{Kind: check.KOptional, Opt: rt}, nil
		}
		return rt, nil
	}
	if tx.Optional {
		inner, err := em.resolveTypeExpr(&ast.TypeExpr{Name: tx.Name, Optional: false, PtrInner: nil})
		if err != nil {
			return nil, err
		}
		return &check.Type{Kind: check.KOptional, Opt: inner}, nil
	}
	if tx.Name == "void" {
		if tx.Optional {
			return nil, fmt.Errorf("void? is invalid")
		}
		return check.TpVoid, nil
	}
	if en, ok := em.inf.Enums[tx.Name]; ok {
		return &check.Type{Kind: check.KEnum, EnumName: en.Name, EnumRef: en}, nil
	}
	if sd, ok := em.inf.Struct[tx.Name]; ok {
		t := check.StructType(sd)
		if t == nil {
			return nil, fmt.Errorf("struct %q", tx.Name)
		}
		if tx.Optional {
			return nil, fmt.Errorf("optional of struct is not supported")
		}
		return t, nil
	}
	switch tx.Name {
	case "int":
		return check.TpInt, nil
	case "float", "f64", "float64", "double":
		return check.TpFloat, nil
	case "str", "string":
		return check.TpStr, nil
	case "bool":
		return check.TpBool, nil
	case "byte":
		return check.TpByte, nil
	}
	return nil, fmt.Errorf("unknown type %q", tx.Name)
}

func (em *emitter) typeOf(e ast.Expr) (*check.Type, error) {
	if e == nil {
		return nil, fmt.Errorf("nil expression")
	}
	if _, ok := e.(*ast.NoneExpr); ok {
		return check.TpNil, nil
	}
	t, ok := em.inf.Types[check.ExprKey(e)]
	if !ok || t == nil {
		return nil, fmt.Errorf("missing type for expression %s", ast.ExprString(e))
	}
	return t, nil
}

// stdioNoRedeclare: <stdio.h> in bootstrap already provides these; skip duplicate `extern` lines.
func stdioNoRedeclare(name string) bool {
	switch name {
	case "printf", "fprintf", "snprintf", "puts", "fputs", "getchar", "putchar", "scanf", "fscanf":
		return true
	default:
		return false
	}
}

// cExternIdent is the C name for a foreign symbol (not f_; only escape C keywords).
func cExternIdent(name string) string {
	if _, ok := cReserved[name]; ok {
		return "u_" + name
	}
	return name
}

func (em *emitter) cParamTFromExpr(t *ast.TypeExpr) (string, error) {
	ty, err := em.resolveTypeExpr(t)
	if err != nil {
		return "", err
	}
	return cParamT(ty), nil
}

// emitExternCDecl returns one line `extern T name(param_list);`
func (em *emitter) emitExternCDecl(ex *ast.ExternDecl) (string, error) {
	var ret string
	if ex.Return != nil {
		rt, err := em.resolveTypeExpr(ex.Return)
		if err != nil {
			return "", err
		}
		ret = cT(rt)
		// C stdio uses `int` for some returns; int64 in Clio is fine at ABI level but redeclare must match.
		if ret == "int64_t" {
			if ex.Name == "printf" || ex.Name == "puts" || ex.Name == "putchar" || ex.Name == "scanf" {
				ret = "int"
			} else if ex.Name == "perror" || ex.Name == "remove" {
				ret = "int"
			}
		}
	} else {
		ret = "void"
	}
	var b strings.Builder
	b.WriteString("extern ")
	b.WriteString(ret)
	b.WriteByte(' ')
	b.WriteString(cExternIdent(ex.Name))
	b.WriteByte('(')
	needComma := false
	if len(ex.Params) == 0 {
		b.WriteString("void")
	} else {
		for _, p := range ex.Params {
			if p.Dots {
				if needComma {
					b.WriteString(", ")
				}
				b.WriteString("...")
				needComma = true
				break
			}
			if p.T == nil {
				return "", fmt.Errorf("extern %s: param %q has no type", ex.Name, p.Name)
			}
			if needComma {
				b.WriteString(", ")
			}
			needComma = true
			s, err := em.cParamTFromExpr(p.T)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
			b.WriteByte(' ')
			b.WriteString(cid(p.Name))
		}
	}
	b.WriteString(");\n")
	return b.String(), nil
}

func EmitC(p *ast.Program, inf *check.Info) (string, error) {
	if p == nil {
		return "", fmt.Errorf("nil program")
	}
	if inf == nil {
		return "", fmt.Errorf("nil type info")
	}
	em := &emitter{
		inf:     inf,
		globals: make(map[string]struct{}),
	}
	for _, d := range p.Decls {
		if ld, ok := d.(*ast.LetDecl); ok {
			em.globals[ld.Name] = struct{}{}
		}
	}

	var out bytes.Buffer
	out.WriteString(cruntime.BootstrapC)
	out.WriteString("#include <math.h>\n\n")
	out.WriteString(`typedef struct { bool has; int64_t v; } clio_opt_i64;
typedef struct { bool has; double v; } clio_opt_f64;
typedef struct { bool has; clio_str v; } clio_opt_str;
typedef struct { bool has; bool v; } clio_opt_bool;

`)

	for _, d := range p.Decls {
		ed, ok := d.(*ast.EnumDecl)
		if !ok {
			continue
		}
		for i, v := range ed.Variants {
			fmt.Fprintf(&out, "#define E_%s__%s %d\n", ed.Name, v, i)
		}
		fmt.Fprintf(&out, "typedef int64_t E_%s;\n\n", ed.Name)
	}

	for name := range em.inf.Struct {
		sdef := em.inf.Struct[name]
		tag := cStructTagName(sdef.Name)
		tn := cStructCName(sdef.Name)
		fmt.Fprintf(&out, "struct %s{\n", tag)
		for _, fn := range sdef.Order {
			ft := sdef.Fields[fn]
			fmt.Fprintf(&out, "  %s u_%s;\n", cT(ft), cid(fn))
		}
		fmt.Fprintf(&out, "};\ntypedef struct %s %s;\n\n", tag, tn)
	}

	for _, d := range p.Decls {
		ex, ok := d.(*ast.ExternDecl)
		if !ok {
			continue
		}
		if stdioNoRedeclare(ex.Name) {
			continue
		}
		if s, err := em.emitExternCDecl(ex); err != nil {
			return "", err
		} else {
			out.WriteString(s)
		}
	}
	out.WriteByte('\n')

	for _, d := range p.Decls {
		ld, ok := d.(*ast.LetDecl)
		if !ok {
			continue
		}
		ty, err := em.letDeclType(ld)
		if err != nil {
			return "", err
		}
		initExpr, err := em.emitGlobalInit(ld, ty)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&out, "static %s g_%s = %s;\n", cT(ty), cid(ld.Name), initExpr)
	}
	out.WriteByte('\n')

	var fns []*ast.FnDecl
	for _, d := range p.Decls {
		if f, ok := d.(*ast.FnDecl); ok {
			fns = append(fns, f)
		}
	}

	for _, f := range fns {
		if f.Name == "main" {
			continue
		}
		line, err := em.fnForwardDecl(f)
		if err != nil {
			return "", err
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	out.WriteByte('\n')

	for _, f := range fns {
		if f.Name == "main" {
			continue
		}
		if err := em.emitFn(&out, f); err != nil {
			return "", err
		}
		out.WriteByte('\n')
	}

	var mainFn *ast.FnDecl
	for _, f := range fns {
		if f.Name == "main" {
			mainFn = f
			break
		}
	}
	if mainFn == nil {
		return "", fmt.Errorf("missing fn main")
	}
	if err := em.emitMain(&out, mainFn); err != nil {
		return "", err
	}

	return out.String(), nil
}

func (em *emitter) letDeclType(ld *ast.LetDecl) (*check.Type, error) {
	if ld.T != nil {
		return em.resolveTypeExpr(ld.T)
	}
	if ld.Init == nil {
		return nil, fmt.Errorf("let %q needs type or initializer", ld.Name)
	}
	return em.typeOf(ld.Init)
}

func (em *emitter) emitGlobalInit(ld *ast.LetDecl, ty *check.Type) (string, error) {
	if _, ok := ld.Init.(*ast.NoneExpr); ok {
		if ty.Kind != check.KOptional {
			return "", fmt.Errorf("none initializer for non-optional global %q", ld.Name)
		}
		return em.zeroOptional(ty), nil
	}
	s, err := em.emitExpr(ld.Init)
	if err != nil {
		return "", err
	}
	if ty.Kind == check.KOptional {
		return em.someOptional(ty.Opt, s), nil
	}
	return s, nil
}

// cZeroValue is the C expression for a default / zero-initialized value of t.
func (em *emitter) cZeroValue(t *check.Type) string {
	if t == nil {
		return "0"
	}
	if t.Kind == check.KOptional {
		return em.zeroOptional(t)
	}
	switch t.Kind {
	case check.KInt, check.KEnum, check.KNil, check.KRange:
		return "0"
	case check.KFloat:
		return "0.0"
	case check.KStr:
		return "(clio_str){0}"
	case check.KBool:
		return "false"
	case check.KStruct:
		return fmt.Sprintf("((%s){0})", cStructCName(t.StructName))
	case check.KResult:
		return fmt.Sprintf("((%s){0})", cResultCName(t.Res))
	case check.KVoid:
		return "0"
	default:
		return "0"
	}
}

func (em *emitter) zeroOptional(t *check.Type) string {
	switch cOptStruct(t.Opt) {
	case "clio_opt_str":
		return "{false, (clio_str){NULL, 0}}"
	default:
		return "{false, 0}"
	}
}

func (em *emitter) someOptional(_ *check.Type, innerExpr string) string {
	return fmt.Sprintf("{true, %s}", innerExpr)
}

func (em *emitter) fnForwardDecl(f *ast.FnDecl) (string, error) {
	var ret *check.Type
	var err error
	if f.Return == nil {
		ret = check.TpVoid
	} else {
		ret, err = em.resolveTypeExpr(f.Return)
		if err != nil {
			return "", err
		}
	}
	var b strings.Builder
	b.WriteString("static ")
	b.WriteString(cT(ret))
	b.WriteString(" f_")
	b.WriteString(cid(f.Name))
	b.WriteByte('(')
	for i, p := range f.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		pt, err := em.resolveTypeExpr(p.T)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s %s", cParamT(pt), cid(p.Name))
	}
	b.WriteString(");\n")
	return b.String(), nil
}

func (em *emitter) emitFn(out *bytes.Buffer, f *ast.FnDecl) error {
	em.curFn = f
	em.params = make(map[string]struct{})
	em.structPtrParam = make(map[string]bool)
	for _, p := range f.Params {
		em.params[p.Name] = struct{}{}
	}
	for _, p := range f.Params {
		pt, e := em.resolveTypeExpr(p.T)
		if e == nil && pt != nil && pt.Kind == check.KStruct {
			em.structPtrParam[p.Name] = true
		}
	}
	em.locals = nil
	if f.Return == nil {
		em.retWant = check.TpVoid
	} else {
		rt, err := em.resolveTypeExpr(f.Return)
		if err != nil {
			em.curFn = nil
			em.params = nil
			em.structPtrParam = nil
			return err
		}
		em.retWant = rt
	}

	var ret = em.retWant
	out.WriteString("static ")
	out.WriteString(cT(ret))
	out.WriteString(" f_")
	out.WriteString(cid(f.Name))
	out.WriteByte('(')
	for i, p := range f.Params {
		if i > 0 {
			out.WriteString(", ")
		}
		pt, err := em.resolveTypeExpr(p.T)
		if err != nil {
			em.curFn = nil
			em.params = nil
			em.structPtrParam = nil
			return err
		}
		fmt.Fprintf(out, "%s %s", cParamT(pt), cid(p.Name))
	}
	out.WriteString(") ")
	if f.Body == nil {
		em.curFn = nil
		em.params = nil
		em.structPtrParam = nil
		em.locals = nil
		em.fnDefers = nil
		return fmt.Errorf("fn %q: empty body", f.Name)
	}
	em.fnDefers = em.fnDefers[:0]
	em.tryPre = em.tryPre[:0]
	em.tryN = 0
	em.rcN = 0
	out.WriteString("{\n")
	em.pushScope()
	for _, s := range f.Body.Stmts {
		if err := em.emitStmt(out, s); err != nil {
			em.curFn = nil
			em.params = nil
			em.structPtrParam = nil
			em.locals = nil
			em.fnDefers = nil
			return err
		}
	}
	em.emitFnDefers(out)
	em.popScope()
	out.WriteString("}\n")
	em.curFn = nil
	em.params = nil
	em.structPtrParam = nil
	em.locals = nil
	return nil
}

func (em *emitter) emitMain(out *bytes.Buffer, f *ast.FnDecl) error {
	em.curFn = f
	em.params = make(map[string]struct{})
	em.structPtrParam = make(map[string]bool)
	for _, p := range f.Params {
		em.params[p.Name] = struct{}{}
	}
	for _, p := range f.Params {
		pt, e := em.resolveTypeExpr(p.T)
		if e == nil && pt != nil && pt.Kind == check.KStruct {
			em.structPtrParam[p.Name] = true
		}
	}
	em.locals = nil
	if f.Return == nil {
		em.retWant = check.TpVoid
	} else {
		rt, err := em.resolveTypeExpr(f.Return)
		if err != nil {
			em.curFn = nil
			return err
		}
		em.retWant = rt
	}

	out.WriteString("int main(void) {\n")
	out.WriteString("clio_console_utf8_init();\n")
	em.inMain = true
	em.pushScope()
	defer em.popScope()
	if f.Body == nil {
		em.curFn = nil
		em.structPtrParam = nil
		return fmt.Errorf("main: empty body")
	}
	em.fnDefers = em.fnDefers[:0]
	em.tryPre = em.tryPre[:0]
	em.tryN = 0
	em.rcN = 0
	for _, s := range f.Body.Stmts {
		if err := em.emitStmt(out, s); err != nil {
			em.curFn = nil
			em.structPtrParam = nil
			return err
		}
	}
	em.emitFnDefers(out)
	out.WriteString("return 0;\n")
	out.WriteString("}\n")
	em.inMain = false
	em.curFn = nil
	em.params = nil
	em.structPtrParam = nil
	em.locals = nil
	return nil
}

func (em *emitter) emitBlock(out *bytes.Buffer, stmts []ast.Stmt, brace bool) error {
	if brace {
		out.WriteString("{\n")
	}
	em.pushScope()
	for _, s := range stmts {
		if err := em.emitStmt(out, s); err != nil {
			em.popScope()
			return err
		}
	}
	em.popScope()
	if brace {
		out.WriteString("}\n")
	}
	return nil
}

func stripExprParens(e ast.Expr) ast.Expr {
	for e != nil {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.Inner
	}
	return nil
}

// asResultCatchExpr returns the catch node, or nil.
func asResultCatchExpr(e ast.Expr) *ast.ResultCatchExpr {
	s := stripExprParens(e)
	if s == nil {
		return nil
	}
	r, _ := s.(*ast.ResultCatchExpr)
	return r
}

func (em *emitter) emitLetResultCatch(out *bytes.Buffer, ls *ast.LetStmt, rce *ast.ResultCatchExpr, ty *check.Type) error {
	if ls.Const {
		return fmt.Errorf("const with result catch (omit const)")
	}
	if ty != nil && ty.Kind == check.KOptional {
		return fmt.Errorf("let: optional with result catch is not supported in codegen")
	}
	rT, err := em.typeOf(rce.Subj)
	if err != nil {
		return err
	}
	if rT == nil || rT.Kind != check.KResult {
		return fmt.Errorf("catch: need result on the left (checker should catch)")
	}
	resC := cT(rT)
	em.rcN++
	tmp := fmt.Sprintf("c_rc_%d", em.rcN)
	sub, err := em.emitExpr(rce.Subj)
	if err != nil {
		return err
	}
	em.flushTryPre(out)
	vname := cid(ls.Name)
	z := em.cZeroValue(ty)
	fmt.Fprintf(out, "%s %s = %s;\n", cT(ty), vname, z)
	fmt.Fprintf(out, "%s %s = %s;\n", resC, tmp, sub)
	fmt.Fprintf(out, "if (!(%s).ok) {\n", tmp)
	eid := cid(rce.ErrName)
	fmt.Fprintf(out, "clio_str %s = (%s).err;\n", eid, tmp)
	em.pushScope()
	em.addLocal(rce.ErrName)
	if rce.Body != nil {
		for _, st := range rce.Body.Stmts {
			if err := em.emitStmt(out, st); err != nil {
				em.popScope()
				return err
			}
		}
	}
	em.popScope()
	fmt.Fprintf(out, "} else {\n")
	fmt.Fprintf(out, "%s = (%s).value;\n", vname, tmp)
	fmt.Fprintf(out, "}\n")
	em.addLocal(ls.Name)
	return nil
}

func (em *emitter) emitStmt(out *bytes.Buffer, s ast.Stmt) error {
	switch x := s.(type) {
	case *ast.BlockStmt:
		return em.emitBlock(out, x.Stmts, true)
	case *ast.IfStmt:
		cond, err := em.emitExpr(x.Cond)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "if (%s) ", cond)
		if x.Thn == nil {
			out.WriteString(";\n")
			return nil
		}
		if err := em.emitBlock(out, x.Thn.Stmts, true); err != nil {
			return err
		}
		if x.Els != nil {
			out.WriteString("else ")
			switch e := x.Els.(type) {
			case *ast.IfStmt:
				return em.emitStmt(out, e)
			case *ast.BlockStmt:
				return em.emitBlock(out, e.Stmts, true)
			default:
				return em.emitStmt(out, x.Els)
			}
		}
		return nil
	case *ast.WhileStmt:
		cond, err := em.emitExpr(x.Cond)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "while (%s) ", cond)
		if x.Body == nil {
			out.WriteString(";\n")
			return nil
		}
		return em.emitBlock(out, x.Body.Stmts, true)
	case *ast.LoopStmt:
		out.WriteString("for (;;) ")
		if x.Body == nil {
			out.WriteString(";\n")
			return nil
		}
		return em.emitBlock(out, x.Body.Stmts, true)
	case *ast.ForInStmt:
		bin, ok := x.In.(*ast.BinaryExpr)
		if !ok || bin.Op != ".." {
			return fmt.Errorf("for-in: expected int range ..")
		}
		lo, err := em.emitExpr(bin.L)
		if err != nil {
			return err
		}
		hi, err := em.emitExpr(bin.R)
		if err != nil {
			return err
		}
		v := cid(x.Var)
		fmt.Fprintf(out, "for (int64_t %s = %s; %s < %s; %s++) ", v, lo, v, hi, v)
		if x.Body == nil {
			out.WriteString(";\n")
			return nil
		}
		em.pushScope()
		em.addLocal(x.Var)
		out.WriteString("{\n")
		for _, st := range x.Body.Stmts {
			if err := em.emitStmt(out, st); err != nil {
				em.popScope()
				return err
			}
		}
		em.popScope()
		out.WriteString("}\n")
		return nil
	case *ast.ReturnStmt:
		if em.curFn != nil {
			em.emitFnDefers(out)
		}
		if x.V == nil {
			// C's int main() must not use bare return; other void fns can
			if em.inMain {
				out.WriteString("return 0;\n")
			} else {
				out.WriteString("return;\n")
			}
			return nil
		}
		if rce := asResultCatchExpr(x.V); rce != nil {
			return em.emitReturnResultCatch(out, rce)
		}
		ev, err := em.emitExpr(x.V)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		if em.retWant != nil && em.retWant.Kind == check.KFloat {
			tv, _ := em.typeOf(x.V)
			if tv != nil && tv.Kind == check.KInt {
				fmt.Fprintf(out, "return (double)(%s);\n", ev)
				return nil
			}
		}
		fmt.Fprintf(out, "return %s;\n", ev)
		return nil
	case *ast.BreakStmt:
		out.WriteString("break;\n")
		return nil
	case *ast.ContinueStmt:
		out.WriteString("continue;\n")
		return nil
	case *ast.DeferStmt:
		if em.curFn == nil {
			return fmt.Errorf("defer: no current function")
		}
		line, err := em.emitExpr(x.E)
		if err != nil {
			return err
		}
		t, _ := em.typeOf(x.E)
		if t != nil && t.Kind == check.KVoid {
			em.fnDefers = append(em.fnDefers, line)
		} else {
			em.fnDefers = append(em.fnDefers, fmt.Sprintf("((void)(%s))", line))
		}
		return nil
	case *ast.LetStmt:
		ty, err := em.letStmtType(x)
		if err != nil {
			return err
		}
		var init string
		if x.Init == nil {
			init = em.cZeroValue(ty)
		} else if _, ok := x.Init.(*ast.NoneExpr); ok {
			if ty.Kind != check.KOptional {
				return fmt.Errorf("none for non-optional let %q", x.Name)
			}
			init = em.zeroOptional(ty)
		} else {
			if rce := asResultCatchExpr(x.Init); rce != nil {
				return em.emitLetResultCatch(out, x, rce, ty)
			}
			init, err = em.emitExpr(x.Init)
			if err != nil {
				return err
			}
			if ty.Kind == check.KOptional {
				init = em.someOptional(ty.Opt, init)
			}
		}
		em.flushTryPre(out)
		em.addLocal(x.Name)
		if x.Const {
			fmt.Fprintf(out, "const %s %s = %s;\n", cT(ty), cid(x.Name), init)
		} else {
			fmt.Fprintf(out, "%s %s = %s;\n", cT(ty), cid(x.Name), init)
		}
		return nil
	case *ast.ExprStmt:
		if rce := asResultCatchExpr(x.E); rce != nil {
			return em.emitExprResultCatch(out, rce)
		}
		e, err := em.emitExpr(x.E)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		t, _ := em.typeOf(x.E)
		if t != nil && t.Kind == check.KVoid {
			out.WriteString(e)
			out.WriteString(";\n")
			return nil
		}
		fmt.Fprintf(out, "%s;\n", e)
		return nil
	case *ast.AssignStmt:
		return em.emitAssign(out, x)
	case *ast.MatchStmt:
		return em.emitMatch(out, x)
	default:
		return fmt.Errorf("unsupported statement %T", s)
	}
}

func (em *emitter) letStmtType(ls *ast.LetStmt) (*check.Type, error) {
	if ls.T != nil {
		return em.resolveTypeExpr(ls.T)
	}
	if ls.Init == nil {
		return nil, fmt.Errorf("let %q: need explicit type for zero-initialized let", ls.Name)
	}
	return em.typeOf(ls.Init)
}

func (em *emitter) emitLvalue(e ast.Expr) (string, error) {
	switch v := e.(type) {
	case *ast.IdentExpr:
		t, err := em.typeOf(e)
		if err != nil {
			return "", err
		}
		if t != nil && t.Kind == check.KStruct && em.isStructParam(v.Name) {
			return "(*" + cid(v.Name) + ")", nil
		}
		return em.lvalueName(v.Name), nil
	case *ast.MemberExpr:
		lt, err := em.typeOf(v.Left)
		if err != nil {
			return "", err
		}
		if lt == nil {
			return "", fmt.Errorf("member assign: need struct on left of .")
		}
		if lt.Kind == check.KEnum {
			return "", fmt.Errorf("cannot assign to %s.%s (enum is not a mutable lvalue)", lt.EnumName, v.Field)
		}
		if lt.Kind != check.KStruct {
			return "", fmt.Errorf("left of . in assignment must be a struct, got %v", lt)
		}
		if id, ok := v.Left.(*ast.IdentExpr); ok && em.isStructParam(id.Name) {
			return fmt.Sprintf("((%s)->u_%s)", cid(id.Name), cid(v.Field)), nil
		}
		base, err := em.emitLvalue(v.Left)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("((%s).u_%s)", base, cid(v.Field)), nil
	default:
		return "", fmt.Errorf("assign: need identifier or x.y field path on the left, got %T", e)
	}
}

// emitReturnResultCatch lowers `return subj catch (e) { ... }` to if/else with return in the else arm.
func (em *emitter) emitReturnResultCatch(out *bytes.Buffer, rce *ast.ResultCatchExpr) error {
	rT, err := em.typeOf(rce.Subj)
	if err != nil {
		return err
	}
	if rT == nil || rT.Kind != check.KResult {
		return fmt.Errorf("return catch: need result on the left (checker should catch)")
	}
	resC := cT(rT)
	em.rcN++
	tmp := fmt.Sprintf("c_rc_%d", em.rcN)
	sub, err := em.emitExpr(rce.Subj)
	if err != nil {
		return err
	}
	em.flushTryPre(out)
	fmt.Fprintf(out, "%s %s = %s;\n", resC, tmp, sub)
	fmt.Fprintf(out, "if (!(%s).ok) {\n", tmp)
	eid := cid(rce.ErrName)
	fmt.Fprintf(out, "clio_str %s = (%s).err;\n", eid, tmp)
	em.pushScope()
	em.addLocal(rce.ErrName)
	if rce.Body != nil {
		for _, st := range rce.Body.Stmts {
			if err := em.emitStmt(out, st); err != nil {
				em.popScope()
				return err
			}
		}
	}
	em.popScope()
	fmt.Fprintf(out, "} else {\n")
	v := fmt.Sprintf("(%s).value", tmp)
	if em.retWant != nil && em.retWant.Kind == check.KFloat {
		ri := rT.Res
		if ri != nil && ri.Kind == check.KInt {
			fmt.Fprintf(out, "return (double)(%s);\n", v)
		} else {
			fmt.Fprintf(out, "return %s;\n", v)
		}
	} else {
		fmt.Fprintf(out, "return %s;\n", v)
	}
	fmt.Fprintf(out, "}\n")
	return nil
}

func (em *emitter) emitExprResultCatch(out *bytes.Buffer, rce *ast.ResultCatchExpr) error {
	rT, err := em.typeOf(rce.Subj)
	if err != nil {
		return err
	}
	if rT == nil || rT.Kind != check.KResult {
		return fmt.Errorf("catch: need result (checker should catch)")
	}
	resC := cT(rT)
	em.rcN++
	tmp := fmt.Sprintf("c_rc_%d", em.rcN)
	sub, err := em.emitExpr(rce.Subj)
	if err != nil {
		return err
	}
	em.flushTryPre(out)
	fmt.Fprintf(out, "%s %s = %s;\n", resC, tmp, sub)
	fmt.Fprintf(out, "if (!(%s).ok) {\n", tmp)
	eid := cid(rce.ErrName)
	fmt.Fprintf(out, "clio_str %s = (%s).err;\n", eid, tmp)
	em.pushScope()
	em.addLocal(rce.ErrName)
	if rce.Body != nil {
		for _, st := range rce.Body.Stmts {
			if err := em.emitStmt(out, st); err != nil {
				em.popScope()
				return err
			}
		}
	}
	em.popScope()
	fmt.Fprintf(out, "} else {\n")
	fmt.Fprintf(out, "((void)(%s.value));\n", tmp)
	fmt.Fprintf(out, "}\n")
	return nil
}

func (em *emitter) emitAssignResultCatch(out *bytes.Buffer, a *ast.AssignStmt, rce *ast.ResultCatchExpr) error {
	if a.Op != "=" {
		return fmt.Errorf("result catch: only simple = (not %s)", a.Op)
	}
	rT, err := em.typeOf(rce.Subj)
	if err != nil {
		return err
	}
	if rT == nil || rT.Kind != check.KResult {
		return fmt.Errorf("catch: need result")
	}
	lhs, err := em.emitLvalue(a.Left)
	if err != nil {
		return err
	}
	resC := cT(rT)
	em.rcN++
	tmp := fmt.Sprintf("c_rc_%d", em.rcN)
	sub, err := em.emitExpr(rce.Subj)
	if err != nil {
		return err
	}
	em.flushTryPre(out)
	fmt.Fprintf(out, "%s %s = %s;\n", resC, tmp, sub)
	fmt.Fprintf(out, "if (!(%s).ok) {\n", tmp)
	eid := cid(rce.ErrName)
	fmt.Fprintf(out, "clio_str %s = (%s).err;\n", eid, tmp)
	em.pushScope()
	em.addLocal(rce.ErrName)
	if rce.Body != nil {
		for _, st := range rce.Body.Stmts {
			if err := em.emitStmt(out, st); err != nil {
				em.popScope()
				return err
			}
		}
	}
	em.popScope()
	fmt.Fprintf(out, "} else {\n")
	fmt.Fprintf(out, "%s = (%s).value;\n", lhs, tmp)
	fmt.Fprintf(out, "}\n")
	return nil
}

func (em *emitter) emitAssign(out *bytes.Buffer, x *ast.AssignStmt) error {
	if rce := asResultCatchExpr(x.Right); rce != nil {
		return em.emitAssignResultCatch(out, x, rce)
	}
	lhs, err := em.emitLvalue(x.Left)
	if err != nil {
		return err
	}
	lt, err := em.typeOf(x.Left)
	if err != nil {
		return err
	}
	rt, err := em.typeOf(x.Right)
	if err != nil {
		return err
	}
	rv, err := em.emitExpr(x.Right)
	if err != nil {
		return err
	}
	em.flushTryPre(out)
	switch x.Op {
	case "=":
		fmt.Fprintf(out, "%s = %s;\n", lhs, rv)
	case "+=":
		if lt.Kind == check.KStr && rt.Kind == check.KStr {
			fmt.Fprintf(out, "%s = clio_str_concat(%s, %s);\n", lhs, lhs, rv)
			return nil
		}
		fmt.Fprintf(out, "%s += %s;\n", lhs, rv)
	case "-=", "*=", "/=", "%=":
		fmt.Fprintf(out, "%s %s %s;\n", lhs, x.Op, rv)
	default:
		return fmt.Errorf("assign op %q not supported", x.Op)
	}
	return nil
}

func (em *emitter) lvalueName(name string) string {
	if em.isLocal(name) {
		return cid(name)
	}
	if em.isParam(name) {
		return cid(name)
	}
	if _, g := em.globals[name]; g {
		return "g_" + cid(name)
	}
	return cid(name)
}

func (em *emitter) isStructParam(name string) bool {
	if em == nil || em.structPtrParam == nil {
		return false
	}
	return em.structPtrParam[name]
}

func (em *emitter) emitMatch(out *bytes.Buffer, m *ast.MatchStmt) error {
	scr, err := em.emitExpr(m.Scrutinee)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "switch ((int64_t)(%s)) {\n", scr)
	for _, a := range m.Arms {
		pat, ok := a.Pat.(*ast.MemberExpr)
		if !ok {
			return fmt.Errorf("match arm: expected Enum.variant")
		}
		id, ok := pat.Left.(*ast.IdentExpr)
		if !ok {
			return fmt.Errorf("match arm: bad pattern")
		}
		fmt.Fprintf(out, "case E_%s__%s:\n", id.Name, pat.Field)
		if a.Body != nil {
			out.WriteString("{\n")
			if err := em.emitBlock(out, a.Body.Stmts, false); err != nil {
				return err
			}
			out.WriteString("}\n")
		}
		out.WriteString("break;\n")
	}
	out.WriteString("}\n")
	return nil
}

func (em *emitter) flushTryPre(out *bytes.Buffer) {
	for _, s := range em.tryPre {
		out.WriteString(s)
	}
	em.tryPre = em.tryPre[:0]
}

// emitTryUnwrap: append C prefix that returns a failed result on !ok, then return (.value) ref.
func (em *emitter) emitTryUnwrap(x *ast.TryUnwrapExpr) (string, error) {
	if em.retWant == nil || em.retWant.Kind != check.KResult {
		return "", fmt.Errorf("?: not in a result-returning function (codegen)")
	}
	tx, err := em.typeOf(x.X)
	if err != nil {
		return "", err
	}
	if tx == nil || tx.Kind != check.KResult {
		return "", fmt.Errorf("?: expected result value")
	}
	sub, err := em.emitExpr(x.X)
	if err != nil {
		return "", err
	}
	rC := cT(tx)
	em.tryN++
	tmp := fmt.Sprintf("c_try_%d", em.tryN)
	retC := cT(em.retWant)
	z := valueZeroC(em.retWant.Res)
	em.tryPre = append(em.tryPre,
		fmt.Sprintf("%s %s = %s;\n", rC, tmp, sub),
		fmt.Sprintf("if (!(%s).ok) { return ((%s){ .ok = false, .err = (%s).err, .value = %s }); }\n", tmp, retC, tmp, z))
	return "(" + tmp + ".value)", nil
}

func (em *emitter) emitExpr(e ast.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("nil expression")
	}
	switch x := e.(type) {
	case *ast.IdentExpr:
		t, err := em.typeOf(e)
		if err != nil {
			return "", err
		}
		_, isGlobal := em.globals[x.Name]
		if t.Kind == check.KEnum && em.inf.Enums[x.Name] != nil && !isGlobal && !em.isLocal(x.Name) && !em.isParam(x.Name) {
			return "", fmt.Errorf("enum %q used as value", x.Name)
		}
		lv := em.emitIdentLoad(x.Name, t)
		if t != nil && t.Kind == check.KStruct && em.isStructParam(x.Name) {
			return "(*" + lv + ")", nil
		}
		return lv, nil
	case *ast.IntLit:
		if _, err := em.typeOf(e); err != nil {
			return "", err
		}
		if x.Raw != "" {
			return x.Raw, nil
		}
		return strconv.FormatInt(x.Val, 10), nil
	case *ast.FloatLit:
		if _, err := em.typeOf(e); err != nil {
			return "", err
		}
		if x.Raw != "" {
			return x.Raw, nil
		}
		return "0.0", nil
	case *ast.StringLit:
		if _, err := em.typeOf(e); err != nil {
			return "", err
		}
		return fmt.Sprintf("clio_str_lit(%s)", cQuote(x.Val)), nil
	case *ast.BoolLit:
		if _, err := em.typeOf(e); err != nil {
			return "", err
		}
		if x.Val {
			return "true", nil
		}
		return "false", nil
	case *ast.NoneExpr:
		return "", fmt.Errorf("bare none is not a value in codegen")
	case *ast.ParenExpr:
		inner, err := em.emitExpr(x.Inner)
		if err != nil {
			return "", err
		}
		return "(" + inner + ")", nil
	case *ast.UnaryExpr:
		return em.emitUnary(x)
	case *ast.BinaryExpr:
		return em.emitBinary(x)
	case *ast.CallExpr:
		return em.emitCall(x)
	case *ast.CastExpr:
		return em.emitCast(x)
	case *ast.StructLit:
		return em.emitStructLit(x)
	case *ast.MemberExpr:
		return em.emitMember(x)
	case *ast.TryUnwrapExpr:
		return em.emitTryUnwrap(x)
	case *ast.ResultCatchExpr:
		return "", fmt.Errorf("result catch: internal — must be in let, assign, or expression statement")
	case *ast.PostfixExpr:
		if x.Op != "++" && x.Op != "--" {
			return "", fmt.Errorf("postfix %q", x.Op)
		}
		lv, err := em.emitLvalue(x.X)
		if err != nil {
			return "", err
		}
		return "(" + lv + ")" + x.Op, nil
	default:
		return "", fmt.Errorf("unsupported expression %T", e)
	}
}

func (em *emitter) emitStructLit(slit *ast.StructLit) (string, error) {
	sdef, ok := em.inf.Struct[slit.TypeName]
	if !ok {
		return "", fmt.Errorf("struct %q not in info", slit.TypeName)
	}
	tn := cStructCName(sdef.Name)
	var b strings.Builder
	b.WriteString("(")
	b.WriteString(tn)
	b.WriteString("){")
	for i, fn := range sdef.Order {
		if i > 0 {
			b.WriteString(", ")
		}
		var in ast.Expr
		for _, fi := range slit.Inits {
			if fi.Name == fn {
				in = fi.Init
				break
			}
		}
		if in == nil {
			return "", fmt.Errorf("struct literal: missing %q (checker should catch)", fn)
		}
		v, err := em.emitExpr(in)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, ".u_%s = %s", cid(fn), v)
	}
	b.WriteString("}")
	return b.String(), nil
}

func (em *emitter) emitIdentLoad(name string, t *check.Type) string {
	if em.isLocal(name) {
		return cid(name)
	}
	if em.isParam(name) {
		return cid(name)
	}
	if _, ok := em.globals[name]; ok {
		return "g_" + cid(name)
	}
	return cid(name)
}

func (em *emitter) emitUnary(x *ast.UnaryExpr) (string, error) {
	inner, err := em.emitExpr(x.X)
	if err != nil {
		return "", err
	}
	switch x.Op {
	case "!":
		return "(!(" + inner + "))", nil
	case "-":
		return "(-(" + inner + "))", nil
	case "+":
		return "(+(" + inner + "))", nil
	default:
		return "", fmt.Errorf("unary %q", x.Op)
	}
}

func (em *emitter) emitBinary(b *ast.BinaryExpr) (string, error) {
	if _, lN := b.L.(*ast.NoneExpr); lN {
		if _, rN := b.R.(*ast.NoneExpr); rN {
			if b.Op == "==" {
				return "true", nil
			}
			if b.Op == "!=" {
				return "false", nil
			}
		}
	}
	if _, rOk := b.R.(*ast.NoneExpr); rOk {
		return em.emitNoneCompare(b.L, b.Op, false)
	}
	if _, lOk := b.L.(*ast.NoneExpr); lOk {
		return em.emitNoneCompare(b.R, b.Op, true)
	}

	l, err := em.emitExpr(b.L)
	if err != nil {
		return "", err
	}
	r, err := em.emitExpr(b.R)
	if err != nil {
		return "", err
	}
	lt, err := em.typeOf(b.L)
	if err != nil {
		return "", err
	}
	rt, err := em.typeOf(b.R)
	if err != nil {
		return "", err
	}
	bt, err := em.typeOf(b)
	if err != nil {
		return "", err
	}

	switch b.Op {
	case "&&", "||":
		return "(" + l + ")" + b.Op + "(" + r + ")", nil
	case "..":
		return "", fmt.Errorf("range .. only valid in for-in")
	}

	if b.Op == "+" && lt.Kind == check.KStr && rt.Kind == check.KStr {
		return fmt.Sprintf("(clio_str_concat((%s), (%s)))", l, r), nil
	}
	if b.Op == "+" && lt.Kind == check.KStr && checkCoercesToString(rt) {
		rs, err := em.emitToStr(b.R)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(clio_str_concat((%s), (%s)))", l, rs), nil
	}
	if b.Op == "+" && checkCoercesToString(lt) && rt.Kind == check.KStr {
		ls, err := em.emitToStr(b.L)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(clio_str_concat((%s), (%s)))", ls, r), nil
	}

	if (b.Op == "==" || b.Op == "!=") && lt.Kind == check.KStr && rt.Kind == check.KStr {
		cmp := fmt.Sprintf("((%s).len == (%s).len && ((%s).len == 0 || memcmp((%s).data, (%s).data, (size_t)(%s).len) == 0))", l, r, l, l, r, l)
		if b.Op == "!=" {
			return "(!" + cmp + ")", nil
		}
		return "(" + cmp + ")", nil
	}

	if (b.Op == "==" || b.Op == "!=") && lt.Kind == check.KStruct && rt.Kind == check.KStruct {
		eq, err := em.emitStructFieldEq(b.Op, l, r, lt, rt)
		if err != nil {
			return "", err
		}
		return eq, nil
	}

	if lt.Kind == check.KEnum && rt.Kind == check.KEnum && b.Op == "==" {
		return fmt.Sprintf("(((int64_t)(%s)) == ((int64_t)(%s)))", l, r), nil
	}
	if lt.Kind == check.KEnum && rt.Kind == check.KEnum && b.Op == "!=" {
		return fmt.Sprintf("(((int64_t)(%s)) != ((int64_t)(%s)))", l, r), nil
	}

	if b.Op == "%" && bt != nil && bt.Kind == check.KFloat {
		return fmt.Sprintf("(fmod((double)(%s), (double)(%s)))", l, r), nil
	}

	if bt != nil && bt.Kind == check.KFloat && (b.Op == "+" || b.Op == "-" || b.Op == "*" || b.Op == "/") {
		return fmt.Sprintf("((double)(%s)) %s ((double)(%s))", l, b.Op, r), nil
	}

	return "(" + l + ") " + b.Op + " (" + r + ")", nil
}

// emitStructFieldEq compares two struct values in C (l and r are already emitted C expressions).
func (em *emitter) emitStructFieldEq(op, l, r string, lt, rt *check.Type) (string, error) {
	if lt.StructDef == nil || rt.StructDef == nil || lt.StructDef != rt.StructDef {
		return "", fmt.Errorf("struct == needs same type")
	}
	if op == "!=" {
		inner, err := em.emitStructFieldEq("==", l, r, lt, rt)
		if err != nil {
			return "", err
		}
		return "(!(" + inner + "))", nil
	}
	sdef := lt.StructDef
	if len(sdef.Order) == 0 {
		return "true", nil
	}
	var parts []string
	for _, fn := range sdef.Order {
		ft := sdef.Fields[fn]
		ls := fmt.Sprintf("((%s).u_%s)", l, cid(fn))
		rs := fmt.Sprintf("((%s).u_%s)", r, cid(fn))
		p, err := em.emitValEqParts(ls, rs, ft)
		if err != nil {
			return "", err
		}
		parts = append(parts, p)
	}
	return "(" + strings.Join(parts, " && ") + ")", nil
}

// emitValEqParts emits C for (a) == (b) for two lvalues/values of type t.
func (em *emitter) emitValEqParts(a, b string, t *check.Type) (string, error) {
	if t == nil {
		return fmt.Sprintf("((%s) == (%s))", a, b), nil
	}
	switch t.Kind {
	case check.KStr:
		cmp := fmt.Sprintf("((%s).len == (%s).len && ((%s).len == 0 || memcmp((%s).data, (%s).data, (size_t)(%s).len) == 0))", a, b, a, a, b, a)
		return "(" + cmp + ")", nil
	case check.KFloat:
		return fmt.Sprintf("((double)(%s) == (double)(%s))", a, b), nil
	case check.KStruct:
		lt := check.StructType(t.StructDef)
		return em.emitStructFieldEq("==", a, b, lt, lt)
	default:
		return fmt.Sprintf("((%s) == (%s))", a, b), nil
	}
}

func (em *emitter) emitNoneCompare(other ast.Expr, op string, noneLeft bool) (string, error) {
	t, err := em.typeOf(other)
	if err != nil {
		return "", err
	}
	if t.Kind != check.KOptional {
		return "", fmt.Errorf("compare to none requires optional type")
	}
	o, err := em.emitExpr(other)
	if err != nil {
		return "", err
	}
	has := "(" + o + ").has"
	switch op {
	case "==":
		if noneLeft {
			return "(!" + has + ")", nil
		}
		return "(!" + has + ")", nil
	case "!=":
		return "(" + has + ")", nil
	default:
		return "", fmt.Errorf("op %s with none", op)
	}
}

func (em *emitter) emitMember(m *ast.MemberExpr) (string, error) {
	tt, err := em.typeOf(m)
	if err != nil {
		return "", err
	}
	// Enum constant: EnumName.Variant
	if id, ok := m.Left.(*ast.IdentExpr); ok && em.inf.Enums[id.Name] != nil {
		if tt != nil && tt.Kind == check.KEnum {
			return fmt.Sprintf("((int64_t)(E_%s__%s))", id.Name, m.Field), nil
		}
	}
	lt, err := em.typeOf(m.Left)
	if err != nil {
		return "", err
	}
	if lt != nil && lt.Kind == check.KResult {
		inner, err2 := em.emitExpr(m.Left)
		if err2 != nil {
			return "", err2
		}
		switch m.Field {
		case "ok":
			return fmt.Sprintf("((%s).ok)", inner), nil
		case "value":
			return fmt.Sprintf("((%s).value)", inner), nil
		case "err":
			return fmt.Sprintf("((%s).err)", inner), nil
		default:
			return "", fmt.Errorf("result: no field %q", m.Field)
		}
	}
	if lt == nil || lt.Kind != check.KStruct {
		return "", fmt.Errorf("member: expected struct or result value on left, got %v", lt)
	}
	if id, ok := m.Left.(*ast.IdentExpr); ok && em.isStructParam(id.Name) {
		return fmt.Sprintf("((%s)->u_%s)", cid(id.Name), cid(m.Field)), nil
	}
	if me, ok := m.Left.(*ast.MemberExpr); ok {
		inner, err2 := em.emitMember(me)
		if err2 != nil {
			return "", err2
		}
		return fmt.Sprintf("((%s).u_%s)", inner, cid(m.Field)), nil
	}
	inner, err := em.emitExpr(m.Left)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("((%s).u_%s)", inner, cid(m.Field)), nil
}

func (em *emitter) emitCast(c *ast.CastExpr) (string, error) {
	arg, err := em.emitExpr(c.Arg)
	if err != nil {
		return "", err
	}
	at, err := em.typeOf(c.Arg)
	if err != nil {
		return "", err
	}
	switch c.To {
	case "int":
		switch at.Kind {
		case check.KInt:
			return "(" + arg + ")", nil
		case check.KFloat:
			return fmt.Sprintf("((int64_t)(%s))", arg), nil
		case check.KBool:
			return fmt.Sprintf("((int64_t)(%s))", arg), nil
		case check.KStr:
			return fmt.Sprintf("(((%s).len > 0 && (%s).data) ? strtoll((%s).data, NULL, 10) : 0)", arg, arg, arg), nil
		case check.KEnum:
			return fmt.Sprintf("((int64_t)(%s))", arg), nil
		default:
			return fmt.Sprintf("((int64_t)(%s))", arg), nil
		}
	case "float":
		switch at.Kind {
		case check.KFloat:
			return "(" + arg + ")", nil
		case check.KInt, check.KEnum:
			return fmt.Sprintf("((double)(%s))", arg), nil
		case check.KStr:
			return fmt.Sprintf("(((%s).len > 0 && (%s).data) ? strtod((%s).data, NULL) : 0.0)", arg, arg, arg), nil
		case check.KBool:
			return fmt.Sprintf("((double)(%s))", arg), nil
		default:
			return fmt.Sprintf("((double)(%s))", arg), nil
		}
	case "bool":
		switch at.Kind {
		case check.KBool:
			return "(" + arg + ")", nil
		case check.KInt, check.KEnum:
			return fmt.Sprintf("((int64_t)(%s) != 0)", arg), nil
		case check.KFloat:
			return fmt.Sprintf("((%s) != 0.0)", arg), nil
		case check.KStr:
			return fmt.Sprintf("((%s).len > 0)", arg), nil
		default:
			return fmt.Sprintf("((int64_t)(%s) != 0)", arg), nil
		}
	case "str":
		switch at.Kind {
		case check.KStr:
			return "(" + arg + ")", nil
		case check.KInt, check.KEnum:
			return fmt.Sprintf("(clio_int_to_str((int64_t)(%s)))", arg), nil
		case check.KFloat:
			return fmt.Sprintf("(clio_float_to_str(%s))", arg), nil
		case check.KBool:
			return fmt.Sprintf("(clio_bool_to_str(%s))", arg), nil
		default:
			return fmt.Sprintf("(clio_int_to_str((int64_t)(%s)))", arg), nil
		}
	default:
		return "", fmt.Errorf("cast to %s", c.To)
	}
}

func emitPeelCallFunc(e ast.Expr) (string, bool) {
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

// checkCoercesToString matches checker rules for str + T.
func checkCoercesToString(t *check.Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case check.KInt, check.KFloat, check.KBool, check.KEnum:
		return true
	default:
		return false
	}
}

// emitToStr lowers to C str(...) for printable scalars in str+T.
func (em *emitter) emitToStr(e ast.Expr) (string, error) {
	ce := &ast.CastExpr{To: "str", Arg: e}
	return em.emitCast(ce)
}

func (em *emitter) emitByRefForStructParam(a ast.Expr) (string, error) {
	t, err := em.typeOf(a)
	if err != nil {
		return "", err
	}
	if t == nil || t.Kind != check.KStruct {
		return em.emitExpr(a)
	}
	if id, ok := a.(*ast.IdentExpr); ok {
		if em.isStructParam(id.Name) {
			return em.emitIdentLoad(id.Name, t), nil
		}
		if em.isLocal(id.Name) {
			return "(&" + cid(id.Name) + ")", nil
		}
		if _, g := em.globals[id.Name]; g {
			return "(&g_" + cid(id.Name) + ")", nil
		}
	}
	if sl, ok := a.(*ast.StructLit); ok {
		s, e := em.emitStructLit(sl)
		if e != nil {
			return "", e
		}
		return "(&" + s + ")", nil
	}
	inner, err2 := em.emitExpr(a)
	if err2 != nil {
		return "", err2
	}
	return "(&" + inner + ")", nil
}

func (em *emitter) emitCallArg(pT *check.Type, a ast.Expr) (string, error) {
	if pT != nil && pT.Kind == check.KStruct {
		return em.emitByRefForStructParam(a)
	}
	return em.emitExpr(a)
}

func (em *emitter) emitMethodCall(c *ast.CallExpr, me *ast.MemberExpr) (string, error) {
	lf, err := em.typeOf(me.Left)
	if err != nil {
		return "", err
	}
	if lf == nil || lf.Kind != check.KStruct {
		return "", fmt.Errorf("method: need struct on left, got %v", lf)
	}
	mangled := lf.StructName + "_" + me.Field
	args := make([]ast.Expr, 0, 1+len(c.Args))
	args = append(args, me.Left)
	args = append(args, c.Args...)
	return em.emitUserCallMangled(mangled, args)
}

// emitUserCallMangled shares logic with top-level fns; name is the mangled symbol.
func (em *emitter) emitUserCallMangled(mangled string, args []ast.Expr) (string, error) {
	fn := em.inf.Fns[mangled]
	if fn == nil {
		return "", fmt.Errorf("no function or method %q", mangled)
	}
	if len(fn.Params) != len(args) {
		return "", fmt.Errorf("call: want %d args, got %d", len(fn.Params), len(args))
	}
	var b strings.Builder
	b.WriteString("f_")
	b.WriteString(cid(mangled))
	b.WriteByte('(')
	for i, a := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		pt, err := em.resolveTypeExpr(fn.Params[i].T)
		if err != nil {
			return "", err
		}
		s, e2 := em.emitCallArg(pt, a)
		if e2 != nil {
			return "", e2
		}
		b.WriteString(s)
	}
	b.WriteByte(')')
	return b.String(), nil
}

func valueZeroC(t *check.Type) string {
	if t == nil {
		return "0"
	}
	switch t.Kind {
	case check.KInt, check.KEnum:
		return "0"
	case check.KFloat:
		return "0.0"
	case check.KStr:
		return "(clio_str){0}"
	case check.KBool:
		return "false"
	default:
		return "0"
	}
}

func (em *emitter) emitResultOk(rt *check.Type, argC string) (string, error) {
	name := cResultCName(rt.Res)
	var v string
	if rt.Res == nil {
		v = fmt.Sprintf("((int64_t)(%s))", argC)
	} else {
		switch rt.Res.Kind {
		case check.KInt, check.KEnum:
			v = fmt.Sprintf("((int64_t)(%s))", argC)
		case check.KFloat:
			v = fmt.Sprintf("((double)(%s))", argC)
		case check.KStr:
			v = argC
		case check.KBool:
			v = fmt.Sprintf("((bool)(%s))", argC)
		default:
			v = argC
		}
	}
	return fmt.Sprintf("((%s){ .value = %s, .err = (clio_str){0}, .ok = true })", name, v), nil
}

func (em *emitter) emitResultErr(rt *check.Type, errStrC string) (string, error) {
	name := cResultCName(rt.Res)
	z := valueZeroC(rt.Res)
	return fmt.Sprintf("((%s){ .value = %s, .err = %s, .ok = false })", name, z, errStrC), nil
}

func (em *emitter) emitCall(c *ast.CallExpr) (string, error) {
	if me, ok := c.Fun.(*ast.MemberExpr); ok {
		return em.emitMethodCall(c, me)
	}
	name, ok := emitPeelCallFunc(c.Fun)
	if !ok {
		return "", fmt.Errorf("indirect call: use f(...) or (f)(...) or m.method(...)")
	}
	switch name {
	case "print":
		return em.emitPrint(c)
	case "input":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("input: need one argument")
		}
		p, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(clio_input((%s)))", p), nil
	case "random":
		if len(c.Args) != 2 {
			return "", fmt.Errorf("random: need two arguments")
		}
		a0, e1 := em.emitExpr(c.Args[0])
		if e1 != nil {
			return "", e1
		}
		a1, e2 := em.emitExpr(c.Args[1])
		if e2 != nil {
			return "", e2
		}
		return fmt.Sprintf("(clio_random((int64_t)(%s), (int64_t)(%s)))", a0, a1), nil
	case "clear_screen":
		if len(c.Args) != 0 {
			return "", fmt.Errorf("clear_screen: no arguments")
		}
		return "clio_clear_screen()", nil
	case "ok":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("ok: need one argument")
		}
		a0, e1 := em.emitExpr(c.Args[0])
		if e1 != nil {
			return "", e1
		}
		t, e2 := em.typeOf(c)
		if e2 != nil {
			return "", e2
		}
		if t == nil || t.Kind != check.KResult {
			return "", fmt.Errorf("ok: internal: expected result type")
		}
		return em.emitResultOk(t, a0)
	case "err":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("err: need one str argument")
		}
		a0, e1 := em.emitExpr(c.Args[0])
		if e1 != nil {
			return "", e1
		}
		t, e2 := em.typeOf(c)
		if e2 != nil {
			return "", e2
		}
		if t == nil || t.Kind != check.KResult {
			return "", fmt.Errorf("err: internal: expected result type")
		}
		return em.emitResultErr(t, a0)
	case "int", "float", "str", "bool":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("%s: need one argument", name)
		}
		if name == "bool" {
			ce := &ast.CastExpr{To: "bool", Arg: c.Args[0]}
			return em.emitCast(ce)
		}
		return em.emitBuiltinCast(name, c.Args[0])
	case "min", "max":
		if len(c.Args) != 2 {
			return "", fmt.Errorf("%s: need two arguments", name)
		}
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		a1, err := em.emitExpr(c.Args[1])
		if err != nil {
			return "", err
		}
		t, err := em.typeOf(c)
		if err != nil {
			return "", err
		}
		if t.Kind == check.KFloat {
			if name == "min" {
				return fmt.Sprintf("(((double)(%s)) < ((double)(%s)) ? (double)(%s) : (double)(%s))", a0, a1, a0, a1), nil
			}
			return fmt.Sprintf("(((double)(%s)) > ((double)(%s)) ? (double)(%s) : (double)(%s))", a0, a1, a0, a1), nil
		}
		if name == "min" {
			return fmt.Sprintf("(((int64_t)(%s)) < ((int64_t)(%s)) ? (int64_t)(%s) : (int64_t)(%s))", a0, a1, a0, a1), nil
		}
		return fmt.Sprintf("(((int64_t)(%s)) > ((int64_t)(%s)) ? (int64_t)(%s) : (int64_t)(%s))", a0, a1, a0, a1), nil
	case "abs":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("abs: need one argument")
		}
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		t, err := em.typeOf(c.Args[0])
		if err != nil {
			return "", err
		}
		if t.Kind == check.KFloat {
			return fmt.Sprintf("(fabs((double)(%s)))", a0), nil
		}
		return fmt.Sprintf("(llabs((long long)(int64_t)(%s)))", a0), nil
	default:
		if ex := em.inf.Externs[name]; ex != nil {
			return em.emitExternCall(ex, c)
		}
		return em.emitUserCallMangled(name, c.Args)
	}
}

// emitStrAsCStr passes a clio_str value to a C `const char*` / ptr[byte] param.
func (em *emitter) emitStrAsCDataPtr(strExpr string) string {
	return fmt.Sprintf("((const char*)((%s).data))", strExpr)
}

func (em *emitter) emitExternCallArg(pWant *check.Type, a ast.Expr) (string, error) {
	if pWant != nil && pWant.Kind == check.KPtr && pWant.Pointee != nil && pWant.Pointee.Kind == check.KByte {
		got, err := em.typeOf(a)
		if err != nil {
			return "", err
		}
		if got != nil && got.Kind == check.KStr {
			s, err2 := em.emitExpr(a)
			if err2 != nil {
				return "", err2
			}
			return em.emitStrAsCDataPtr(s), nil
		}
	}
	if pWant != nil && pWant.Kind == check.KStruct {
		return em.emitByRefForStructParam(a)
	}
	return em.emitExpr(a)
}

func (em *emitter) emitExternCall(ex *ast.ExternDecl, c *ast.CallExpr) (string, error) {
	var b strings.Builder
	b.WriteString(cExternIdent(ex.Name))
	b.WriteByte('(')
	argI := 0
	for _, p := range ex.Params {
		if p.Dots {
			for j := argI; j < len(c.Args); j++ {
				if j > argI {
					b.WriteString(", ")
				}
				s, err := em.emitExpr(c.Args[j])
				if err != nil {
					return "", err
				}
				b.WriteString(s)
			}
			break
		}
		if argI > 0 {
			b.WriteString(", ")
		}
		pwant, err := em.resolveTypeExpr(p.T)
		if err != nil {
			return "", err
		}
		s, err2 := em.emitExternCallArg(pwant, c.Args[argI])
		if err2 != nil {
			return "", err2
		}
		b.WriteString(s)
		argI++
	}
	b.WriteByte(')')
	return b.String(), nil
}

func (em *emitter) appendShowValueC(b *strings.Builder, cexpr string, t *check.Type) error {
	if t == nil {
		return fmt.Errorf("print: no type for expression")
	}
	if t.Kind == check.KStruct {
		if t.StructDef == nil {
			return fmt.Errorf("struct has no def")
		}
		if err := em.appendShowStructBlock(b, cexpr, t); err != nil {
			return err
		}
		return nil
	}
	switch t.Kind {
	case check.KInt, check.KEnum:
		fmt.Fprintf(b, "clio_show_int((int64_t)(%s));\n", cexpr)
	case check.KFloat:
		fmt.Fprintf(b, "clio_show_float((double)(%s));\n", cexpr)
	case check.KBool:
		fmt.Fprintf(b, "clio_show_bool((bool)(%s));\n", cexpr)
	case check.KStr:
		fmt.Fprintf(b, "clio_show_str((%s));\n", cexpr)
	default:
		return fmt.Errorf("print: value type %s not supported", t)
	}
	return nil
}

// appendShowStructBlock emits a printf-style tree for one struct; no final newline.
func (em *emitter) appendShowStructBlock(b *strings.Builder, varExpr string, t *check.Type) error {
	sdef := t.StructDef
	if sdef == nil {
		return fmt.Errorf("struct: missing definition in codegen")
	}
	fmt.Fprintf(b, "printf(%s); ", cStringLit("{ "))
	for i, fn := range sdef.Order {
		if i > 0 {
			fmt.Fprintf(b, "printf(%s); ", cStringLit(", "))
		}
		fmt.Fprintf(b, "printf(%s); ", cStringLit(fn+": "))
		ffield := fmt.Sprintf("((%s).u_%s)", varExpr, cid(fn))
		ft := sdef.Fields[fn]
		if err := em.appendShowValueC(b, ffield, ft); err != nil {
			return err
		}
	}
	fmt.Fprintf(b, "printf(%s);", cStringLit(" }"))
	return nil
}

func (em *emitter) emitPrint(c *ast.CallExpr) (string, error) {
	if len(c.Args) == 0 {
		return "", fmt.Errorf("print: need at least one argument")
	}
	var b strings.Builder
	for _, a := range c.Args {
		av, err := em.emitExpr(a)
		if err != nil {
			return "", err
		}
		t, err := em.typeOf(a)
		if err != nil {
			return "", err
		}
		if t.Kind == check.KOptional {
			in := t.Opt
			if in == nil {
				return "", fmt.Errorf("print: bad optional type")
			}
			switch in.Kind {
			case check.KInt, check.KEnum:
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\\n\"); } else { clio_print_int((%s).v); }\n", av, av)
			case check.KFloat:
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\\n\"); } else { clio_print_float((%s).v); }\n", av, av)
			case check.KStr:
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\\n\"); } else { clio_print_str((%s).v); }\n", av, av)
			case check.KBool:
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\\n\"); } else { clio_print_bool((%s).v); }\n", av, av)
			case check.KStruct:
				innerT := in
				if innerT.StructDef == nil {
					return "", fmt.Errorf("print: bad optional struct")
				}
				b.WriteString("if (!(" + av + ").has) { printf(\"none\\n\"); } else { ")
				if err := em.appendShowStructBlock(&b, "("+av+").v", innerT); err != nil {
					return "", err
				}
				b.WriteString("clio_show_endl();}\n")
			default:
				return "", fmt.Errorf("print: cannot print optional of %s", in)
			}
			continue
		}
		switch t.Kind {
		case check.KInt, check.KEnum:
			fmt.Fprintf(&b, "clio_print_int((int64_t)(%s));\n", av)
		case check.KFloat:
			fmt.Fprintf(&b, "clio_print_float(%s);\n", av)
		case check.KBool:
			fmt.Fprintf(&b, "clio_print_bool(%s);\n", av)
		case check.KStr:
			fmt.Fprintf(&b, "clio_print_str(%s);\n", av)
		case check.KStruct:
			if t.StructDef == nil {
				return "", fmt.Errorf("struct %s has no metadata", t)
			}
			if err := em.appendShowStructBlock(&b, av, t); err != nil {
				return "", err
			}
			b.WriteString("clio_show_endl();\n")
		case check.KVoid:
			return "", fmt.Errorf("print: void argument")
		default:
			return "", fmt.Errorf("print: type %s is not supported", t)
		}
	}
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func (em *emitter) emitBuiltinCast(name string, arg ast.Expr) (string, error) {
	av, err := em.emitExpr(arg)
	if err != nil {
		return "", err
	}
	at, err := em.typeOf(arg)
	if err != nil {
		return "", err
	}
	switch name {
	case "int":
		switch at.Kind {
		case check.KInt:
			return fmt.Sprintf("((int64_t)(%s))", av), nil
		case check.KFloat:
			return fmt.Sprintf("((int64_t)(%s))", av), nil
		case check.KBool:
			return fmt.Sprintf("((int64_t)(%s))", av), nil
		case check.KStr:
			return fmt.Sprintf("(((%s).len > 0 && (%s).data) ? strtoll((%s).data, NULL, 10) : 0)", av, av, av), nil
		case check.KEnum:
			return fmt.Sprintf("((int64_t)(%s))", av), nil
		default:
			return fmt.Sprintf("((int64_t)(%s))", av), nil
		}
	case "float":
		switch at.Kind {
		case check.KFloat:
			return av, nil
		case check.KInt, check.KEnum:
			return fmt.Sprintf("((double)(%s))", av), nil
		case check.KBool:
			return fmt.Sprintf("((double)(%s))", av), nil
		case check.KStr:
			return fmt.Sprintf("(((%s).len > 0 && (%s).data) ? strtod((%s).data, NULL) : 0.0)", av, av, av), nil
		default:
			return fmt.Sprintf("((double)(%s))", av), nil
		}
	case "str":
		switch at.Kind {
		case check.KStr:
			return av, nil
		case check.KInt, check.KEnum:
			return fmt.Sprintf("(clio_int_to_str((int64_t)(%s)))", av), nil
		case check.KFloat:
			return fmt.Sprintf("(clio_float_to_str(%s))", av), nil
		case check.KBool:
			return fmt.Sprintf("(clio_bool_to_str(%s))", av), nil
		default:
			return fmt.Sprintf("(clio_int_to_str((int64_t)(%s)))", av), nil
		}
	}
	return "", fmt.Errorf("builtin %s", name)
}

