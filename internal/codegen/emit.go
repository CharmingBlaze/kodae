package codegen

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"kodae/internal/ast"
	"kodae/internal/check"
	"kodae/internal/cruntime"
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
	case check.KF32:
		return "float"
	case check.KI32:
		return "int32_t"
	case check.KU32:
		return "uint32_t"
	case check.KU8:
		return "uint8_t"
	case check.KStr:
		return "kodae_str"
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
			return "kodae_str*"
		}
		if t.Pointee.Kind == check.KBool {
			return "bool*"
		}
		return "const void*"
	case check.KList:
		return "kodae_list"
	case check.KResult:
		return cResultCName(t.Res)
	case check.KTuple:
		return cTupleCName(t)
	case check.KAny:
		return "kodae_any"
	case check.KClosure:
		return "void"
	default:
		return "int64_t"
	}
}

// cResultCName is the C typedef for result[T] (value, err, ok — see build-spec).
func cResultCName(res *check.Type) string {
	if res == nil {
		return "kodae_res_i64"
	}
	switch res.Kind {
	case check.KInt, check.KEnum:
		return "kodae_res_i64"
	case check.KFloat:
		return "kodae_res_f64"
	case check.KStr:
		return "kodae_res_str"
	case check.KBool:
		return "kodae_res_bool"
	default:
		return "kodae_res_i64"
	}
}

// cStructCName is the C typedef name for a Kodae struct.
func cStructCName(n string) string { return "S_" + cid(n) }

func cStructTagName(n string) string { return "s_" + cid(n) + "_" }

var tupleNames = make(map[string]string)
var tupleTypes []*check.Type

func cTupleCName(t *check.Type) string {
	sig := t.String()
	if n, ok := tupleNames[sig]; ok {
		return n
	}
	n := fmt.Sprintf("kodae_tuple_%d", len(tupleNames))
	tupleNames[sig] = n
	tupleTypes = append(tupleTypes, t)
	return n
}

// cParamT is the C type for a function parameter (struct types are by pointer).
func cParamT(t *check.Type) string {
	if t != nil && t.Kind == check.KStruct {
		return cStructCName(t.StructName) + "*"
	}
	return cT(t)
}

func cOptStruct(inner *check.Type) string {
	if inner == nil {
		return "kodae_opt_i64"
	}
	switch inner.Kind {
	case check.KInt, check.KEnum:
		return "kodae_opt_i64"
	case check.KFloat:
		return "kodae_opt_f64"
	case check.KStr:
		return "kodae_opt_str"
	case check.KBool:
		return "kodae_opt_bool"
	default:
		return "kodae_opt_i64"
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
	locals          []map[string]struct{}
	closureLocals   []map[string]*check.ClosureInfo
	retWant *check.Type
	// fnDefers: C lines to run at return / end of function (LIFO; see defer).
	fnDefers []string
	// tryPre: C lines for result? (error propagation) — flushed before the surrounding statement.
	tryPre []string
	tryN   int
	rcN    int // c_rc_* temps for `result catch` lowering
}

type EmitOptions struct {
	LibraryMode bool
}

func (em *emitter) pushScope() {
	em.locals = append(em.locals, make(map[string]struct{}))
	em.closureLocals = append(em.closureLocals, make(map[string]*check.ClosureInfo))
}

func (em *emitter) popScope() {
	if len(em.locals) == 0 {
		return
	}
	em.locals = em.locals[:len(em.locals)-1]
	if len(em.closureLocals) > 0 {
		em.closureLocals = em.closureLocals[:len(em.closureLocals)-1]
	}
}

func (em *emitter) addClosureLocal(name string, ci *check.ClosureInfo) {
	if len(em.closureLocals) == 0 {
		em.pushScope()
	}
	em.closureLocals[len(em.closureLocals)-1][name] = ci
}

func (em *emitter) lookupClosure(name string) *check.ClosureInfo {
	for i := len(em.closureLocals) - 1; i >= 0; i-- {
		if ci, ok := em.closureLocals[i][name]; ok {
			return ci
		}
	}
	return nil
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
	if tx.ListInner != nil {
		inner, err := em.resolveTypeExpr(tx.ListInner)
		if err != nil {
			return nil, err
		}
		lt := &check.Type{Kind: check.KList, Elem: inner}
		if tx.Optional {
			return &check.Type{Kind: check.KOptional, Opt: lt}, nil
		}
		return lt, nil
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
	if tx.TupleInner != nil {
		var elems []*check.Type
		for _, x := range tx.TupleInner {
			inner, err := em.resolveTypeExpr(x)
			if err != nil {
				return nil, err
			}
			elems = append(elems, inner)
		}
		return &check.Type{Kind: check.KTuple, TupleElems: elems}, nil
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
	case "f32":
		return check.TpF32, nil
	case "i32":
		return check.TpI32, nil
	case "u32":
		return check.TpU32, nil
	case "u8":
		return check.TpU8, nil
	case "float", "f64", "float64", "double":
		return check.TpFloat, nil
	case "str", "string":
		return check.TpStr, nil
	case "bool":
		return check.TpBool, nil
	case "byte":
		return check.TpByte, nil
	case "Any":
		return check.TpAny, nil
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

// cExternDeclParamT is the C type written in `extern` declarations: struct by value
// (Raylib-style), not pointer like normal Kodae user functions.
func (em *emitter) cExternDeclParamT(t *ast.TypeExpr) (string, error) {
	ty, err := em.resolveTypeExpr(t)
	if err != nil {
		return "", err
	}
	return cT(ty), nil
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
		// C stdio uses `int` for some returns; int64 in Kodae is fine at ABI level but redeclare must match.
		if ret == "int64_t" {
			switch ex.Name {
			case "printf", "puts", "putchar", "scanf", "perror", "remove":
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
			s, err := em.cExternDeclParamT(p.T)
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
	return EmitCWithOptions(p, inf, EmitOptions{})
}

// structEmitOrder returns struct names so each struct appears after any struct type
// referenced by its fields (needed for Raylib-style interop structs in one TU).
func structEmitOrder(inf *check.Info) []string {
	remaining := make(map[string]*check.Struct, len(inf.Struct))
	for n, s := range inf.Struct {
		remaining[n] = s
	}
	var out []string
	for len(remaining) > 0 {
		var ready []string
		for name, sdef := range remaining {
			ok := true
			for _, fn := range sdef.Order {
				ft := sdef.Fields[fn]
				if ft != nil && ft.Kind == check.KStruct && ft.StructName != "" && ft.StructName != name {
					if _, pending := remaining[ft.StructName]; pending {
						ok = false
						break
					}
				}
			}
			if ok {
				ready = append(ready, name)
			}
		}
		if len(ready) == 0 {
			var rest []string
			for n := range remaining {
				rest = append(rest, n)
			}
			sort.Strings(rest)
			out = append(out, rest...)
			break
		}
		sort.Strings(ready)
		for _, n := range ready {
			delete(remaining, n)
			out = append(out, n)
		}
	}
	return out
}

func EmitCWithOptions(p *ast.Program, inf *check.Info, opts EmitOptions) (string, error) {
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

	// Reset global tuple tracker for each emit (important for tests)
	tupleNames = make(map[string]string)
	tupleTypes = nil

	// Pre-scan to discover all tuple types
	for _, t := range inf.Types {
		if t != nil && t.Kind == check.KTuple {
			cTupleCName(t)
		}
	}

	out.WriteString(cruntime.ParsonH)
	out.WriteString("\n")
	out.WriteString(cruntime.BootstrapC)
	out.WriteString("\n")
	out.WriteString(cruntime.ParsonC)
	out.WriteString("\n")
	out.WriteString(cruntime.WsClientC)
	out.WriteString("\n")
	out.WriteString("#include <math.h>\n\n")
	out.WriteString(`typedef struct { bool has; int64_t v; } kodae_opt_i64;
typedef struct { bool has; double v; } kodae_opt_f64;
typedef struct { bool has; kodae_str v; } kodae_opt_str;
typedef struct { bool has; bool v; } kodae_opt_bool;

`)

	for _, t := range tupleTypes {
		name := cTupleCName(t)
		out.WriteString("typedef struct {\n")
		for i, el := range t.TupleElems {
			fmt.Fprintf(&out, "  %s f%d;\n", cT(el), i)
		}
		fmt.Fprintf(&out, "} %s;\n\n", name)
	}

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

	for _, name := range structEmitOrder(em.inf) {
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

	var needUndef bool
	for _, d := range p.Decls {
		ex, ok := d.(*ast.ExternDecl)
		if !ok {
			continue
		}
		if stdioNoRedeclare(ex.Name) {
			continue
		}
		needUndef = true
	}
	if needUndef {
		// windows.h (via kodae bootstrap) #defines these; Raylib uses the same names as real functions.
		out.WriteString("#if defined(_WIN32)\n" +
			"#undef CloseWindow\n" +
			"#undef ShowCursor\n" +
			"#undef DrawText\n" +
			"#endif\n\n")
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
	if opts.LibraryMode {
		if err := em.emitExportWrappers(&out, p, inf); err != nil {
			return "", err
		}
	} else {
		if mainFn == nil {
			return "", fmt.Errorf("missing fn main")
		}
		if err := em.emitMain(&out, mainFn); err != nil {
			return "", err
		}
	}

	return out.String(), nil
}

func cPubType(t *check.Type) string {
	if t == nil {
		return "void"
	}
	switch t.Kind {
	case check.KInt, check.KEnum:
		return "int64_t"
	case check.KFloat:
		return "double"
	case check.KBool:
		return "bool"
	case check.KStr:
		return "const char*"
	case check.KStruct:
		return cStructCName(t.StructName)
	default:
		return cT(t)
	}
}

func pubMarshalArg(pt *check.Type, name string) string {
	if pt != nil && pt.Kind == check.KStr {
		return fmt.Sprintf("kodae_str_lit(%s)", name)
	}
	return name
}

func pubMarshalRet(rt *check.Type, expr string) string {
	if rt != nil && rt.Kind == check.KStr {
		return fmt.Sprintf("kodae_str_to_cstr(%s)", expr)
	}
	return expr
}

func (em *emitter) emitExportWrappers(out *bytes.Buffer, p *ast.Program, inf *check.Info) error {
	for _, d := range p.Decls {
		f, ok := d.(*ast.FnDecl)
		if !ok {
			continue
		}
		fd := inf.Fns[f.Name]
		if fd == nil {
			fd = f
		}
		var retT *check.Type
		var err error
		if fd.Return == nil {
			retT = check.TpVoid
		} else {
			retT, err = em.resolveTypeExpr(fd.Return)
			if err != nil {
				continue
			}
		}
		if !isCCompatible(retT) {
			continue
		}
		skipFn := false
		for _, p := range fd.Params {
			pt, err := em.resolveTypeExpr(p.T)
			if err != nil || !isCCompatible(pt) {
				skipFn = true
				break
			}
		}
		if skipFn {
			continue
		}
		fmt.Fprintf(out, "%s %s(", cPubType(retT), cid(f.Name))
		for i, p := range fd.Params {
			if i > 0 {
				out.WriteString(", ")
			}
			pt, _ := em.resolveTypeExpr(p.T)
			fmt.Fprintf(out, "%s %s", cPubType(pt), cid(p.Name))
		}
		out.WriteString(") {\n")
		var call strings.Builder
		fmt.Fprintf(&call, "f_%s(", cid(f.Name))
		for i, p := range fd.Params {
			if i > 0 {
				call.WriteString(", ")
			}
			pt, err := em.resolveTypeExpr(p.T)
			if err != nil {
				return err
			}
			call.WriteString(pubMarshalArg(pt, cid(p.Name)))
		}
		call.WriteString(")")
		if retT != nil && retT.Kind == check.KVoid {
			fmt.Fprintf(out, "%s;\n", call.String())
		} else {
			fmt.Fprintf(out, "return %s;\n", pubMarshalRet(retT, call.String()))
		}
		out.WriteString("}\n\n")
	}
	return nil
}

func isCCompatible(t *check.Type) bool {
	if t == nil {
		return true
	}
	switch t.Kind {
	case check.KInt, check.KFloat, check.KBool, check.KStr, check.KVoid, check.KEnum, check.KStruct:
		return true
	default:
		return false
	}
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
		return "(kodae_str){0}"
	case check.KBool:
		return "false"
	case check.KStruct:
		return fmt.Sprintf("((%s){0})", cStructCName(t.StructName))
	case check.KList:
		if t.Elem == nil {
			return "kodae_list_new((int64_t)sizeof(int64_t))"
		}
		return fmt.Sprintf("kodae_list_new((int64_t)sizeof(%s))", cT(t.Elem))
	case check.KResult:
		return fmt.Sprintf("((%s){0})", cResultCName(t.Res))
	case check.KTuple:
		return fmt.Sprintf("((%s){0})", cTupleCName(t))
	case check.KVoid:
		return "0"
	default:
		return "0"
	}
}

func (em *emitter) zeroOptional(t *check.Type) string {
	switch cOptStruct(t.Opt) {
	case "kodae_opt_str":
		return "{false, (kodae_str){NULL, 0}}"
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
	var ret = em.retWant
	flat := flattenFuncLitsPostOrder(f.Body.Stmts)
	for _, fl := range flat {
		ci := em.inf.Closures[check.ExprKey(fl)]
		if ci != nil {
			out.WriteString(closureForwardDecl(ci))
		}
	}
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
			em.closureLocals = nil
			return err
		}
		fmt.Fprintf(out, "%s %s", cParamT(pt), cid(p.Name))
	}
	out.WriteString(") ")
	out.WriteString("{\n")
	em.pushScope()
	for _, s := range f.Body.Stmts {
		if err := em.emitStmt(out, s); err != nil {
			em.curFn = nil
			em.params = nil
			em.structPtrParam = nil
			em.locals = nil
			em.closureLocals = nil
			em.fnDefers = nil
			return err
		}
	}
	em.emitFnDefers(out)
	em.popScope()
	out.WriteString("}\n")
	for _, fl := range flat {
		if err := em.emitOneClosureDef(out, fl); err != nil {
			em.curFn = nil
			em.params = nil
			em.structPtrParam = nil
			em.locals = nil
			em.closureLocals = nil
			em.fnDefers = nil
			return err
		}
	}
	em.curFn = nil
	em.params = nil
	em.structPtrParam = nil
	em.locals = nil
	em.closureLocals = nil
	return nil
}

func closureForwardDecl(ci *check.ClosureInfo) string {
	if ci.CapturesThis {
		return fmt.Sprintf("static void %s(%s*);\n", ci.Mangled, cStructCName(ci.RecvStruct))
	}
	return fmt.Sprintf("static void %s(void);\n", ci.Mangled)
}

func (em *emitter) emitOneClosureDef(out *bytes.Buffer, fl *ast.FuncLit) error {
	k := check.ExprKey(fl)
	ci := em.inf.Closures[k]
	if ci == nil {
		return fmt.Errorf("closure: missing closure metadata")
	}
	savedFn := em.curFn
	savedParams := em.params
	savedStruct := em.structPtrParam
	savedLocals := em.locals
	savedClosureLocals := em.closureLocals
	savedRet := em.retWant
	savedDefers := em.fnDefers
	savedTry := em.tryPre
	defer func() {
		em.curFn = savedFn
		em.params = savedParams
		em.structPtrParam = savedStruct
		em.locals = savedLocals
		em.closureLocals = savedClosureLocals
		em.retWant = savedRet
		em.fnDefers = savedDefers
		em.tryPre = savedTry
	}()

	em.curFn = nil
	em.params = make(map[string]struct{})
	em.structPtrParam = make(map[string]bool)
	if ci.CapturesThis {
		em.params["self"] = struct{}{}
		em.structPtrParam["self"] = true
	}
	em.locals = nil
	em.closureLocals = nil
	em.retWant = check.TpVoid
	em.fnDefers = em.fnDefers[:0]
	em.tryPre = em.tryPre[:0]

	out.WriteString("static void ")
	out.WriteString(ci.Mangled)
	out.WriteByte('(')
	if ci.CapturesThis {
		fmt.Fprintf(out, "%s* %s", cStructCName(ci.RecvStruct), cid("self"))
	} else {
		out.WriteString("void")
	}
	out.WriteString(") {\n")
	em.pushScope()
	for _, s := range fl.Body.Stmts {
		if err := em.emitStmt(out, s); err != nil {
			return err
		}
	}
	em.emitFnDefers(out)
	em.popScope()
	out.WriteString("}\n")
	return nil
}

func (em *emitter) emitStructUpdateExpr(x *ast.StructUpdateExpr) (string, error) {
	bt, err := em.typeOf(x.Base)
	if err != nil {
		return "", err
	}
	if bt == nil || bt.Kind != check.KStruct || bt.StructDef == nil {
		return "", fmt.Errorf("`with`: internal struct type")
	}
	baseStr, err := em.emitExpr(x.Base)
	if err != nil {
		return "", err
	}
	tn := cStructCName(bt.StructName)
	var b strings.Builder
	b.WriteString("({ ")
	b.WriteString(tn)
	b.WriteString(" _uw = (")
	b.WriteString(baseStr)
	b.WriteString(");\n")
	for _, in := range x.Inits {
		val, err := em.emitExpr(in.Init)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, " _uw.u_%s = %s;\n", cid(in.Name), val)
	}
	b.WriteString(" _uw; })")
	return b.String(), nil
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
	out.WriteString("kodae_console_utf8_init();\n")
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
		sub, err := em.emitExpr(rce.Subj)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		em.addLocal(ls.Name)
		if ls.Const {
			fmt.Fprintf(out, "const %s %s = %s;\n", cT(ty), cid(ls.Name), sub)
		} else {
			fmt.Fprintf(out, "%s %s = %s;\n", cT(ty), cid(ls.Name), sub)
		}
		return nil
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
	fmt.Fprintf(out, "kodae_str %s = (%s).err;\n", eid, tmp)
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
		if ok && bin.Op == ".." {
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
		}
		it, err := em.typeOf(x.In)
		if err != nil {
			return err
		}
		if it == nil || it.Kind != check.KList || it.Elem == nil {
			return fmt.Errorf("for-in: expected int range or list")
		}
		lv, err := em.emitExpr(x.In)
		if err != nil {
			return err
		}
		em.tryN++
		idx := fmt.Sprintf("c_i_%d", em.tryN)
		iv := cid(x.Var)
		fmt.Fprintf(out, "for (int64_t %s = 0; %s < (%s).len; %s++) ", idx, idx, lv, idx)
		if x.Body == nil {
			out.WriteString(";\n")
			return nil
		}
		em.pushScope()
		em.addLocal(x.Var)
		out.WriteString("{\n")
		fmt.Fprintf(out, "%s %s = (*((%s*)kodae_list_at_ptr(&(%s), %s)));\n", cT(it.Elem), iv, cT(it.Elem), lv, idx)
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
		if len(x.Destruct) > 1 {
			init, err := em.emitExpr(x.Init)
			if err != nil {
				return err
			}
			em.flushTryPre(out)
			em.rcN++
			tmp := fmt.Sprintf("c_tup_%d", em.rcN)
			fmt.Fprintf(out, "%s %s = %s;\n", cT(ty), tmp, init)
			for i, n := range x.Destruct {
				em.addLocal(n)
				if x.Const {
					fmt.Fprintf(out, "const %s %s = %s.f%d;\n", cT(ty.TupleElems[i]), cid(n), tmp, i)
				} else {
					fmt.Fprintf(out, "%s %s = %s.f%d;\n", cT(ty.TupleElems[i]), cid(n), tmp, i)
				}
			}
			return nil
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
			if fl, ok := stripExprToFuncLit(x.Init); ok {
				ci := em.inf.Closures[check.ExprKey(fl)]
				if ci == nil {
					return fmt.Errorf("closure: missing closure metadata")
				}
				em.addClosureLocal(x.Name, ci)
				return nil
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
	case *ast.RepeatStmt:
		c, err := em.emitExpr(x.Count)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		em.rcN++
		tmp := fmt.Sprintf("c_rep_%d", em.rcN)
		fmt.Fprintf(out, "for (int64_t %s = 0; %s < (int64_t)(%s); %s++) {\n", tmp, tmp, c, tmp)
		if err := em.emitBlock(out, x.Body.Stmts, false); err != nil {
			return err
		}
		out.WriteString("}\n")
		return nil
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
			return fmt.Sprintf("((%s)->u_%s)", cid(em.maybeThisAlias(id.Name)), cid(v.Field)), nil
		}
		base, err := em.emitLvalue(v.Left)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("((%s).u_%s)", base, cid(v.Field)), nil
	case *ast.IndexExpr:
		lt, err := em.typeOf(v.Left)
		if err != nil {
			return "", err
		}
		if lt == nil || lt.Kind != check.KList || lt.Elem == nil {
			return "", fmt.Errorf("assign index: left must be list value")
		}
		base, err := em.emitLvalue(v.Left)
		if err != nil {
			return "", err
		}
		ix, err := em.emitExpr(v.Index)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(*((%s*)kodae_list_at_ptr(&(%s), (int64_t)(%s))))", cT(lt.Elem), base, ix), nil
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
		sub, err := em.emitExpr(rce.Subj)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		fmt.Fprintf(out, "return %s;\n", sub)
		return nil
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
	fmt.Fprintf(out, "kodae_str %s = (%s).err;\n", eid, tmp)
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
		sub, err := em.emitExpr(rce.Subj)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		fmt.Fprintf(out, "((void)(%s));\n", sub)
		return nil
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
	fmt.Fprintf(out, "kodae_str %s = (%s).err;\n", eid, tmp)
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
		lhs, err := em.emitLvalue(a.Left)
		if err != nil {
			return err
		}
		sub, err := em.emitExpr(rce.Subj)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		fmt.Fprintf(out, "%s = %s;\n", lhs, sub)
		return nil
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
	fmt.Fprintf(out, "kodae_str %s = (%s).err;\n", eid, tmp)
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
	if tup, ok := x.Left.(*ast.TupleExpr); ok {
		rv, err := em.emitExpr(x.Right)
		if err != nil {
			return err
		}
		rt, err := em.typeOf(x.Right)
		if err != nil {
			return err
		}
		em.flushTryPre(out)
		em.rcN++
		tmp := fmt.Sprintf("c_tup_%d", em.rcN)
		fmt.Fprintf(out, "%s %s = %s;\n", cT(rt), tmp, rv)
		for i, lexpr := range tup.Exprs {
			lhs, err := em.emitLvalue(lexpr)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "%s = %s.f%d;\n", lhs, tmp, i)
		}
		return nil
	}
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
			fmt.Fprintf(out, "%s = kodae_str_concat(%s, %s);\n", lhs, lhs, rv)
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
	name = em.maybeThisAlias(name)
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
	name = em.maybeThisAlias(name)
	if em == nil || em.structPtrParam == nil {
		return false
	}
	return em.structPtrParam[name]
}

func (em *emitter) maybeThisAlias(name string) string {
	if name == "this" && em.isParam("self") {
		return "self"
	}
	return name
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

func (em *emitter) emitTryUnwrap(x *ast.TryUnwrapExpr) (string, error) {
	_ = x
	return "", fmt.Errorf("? is not supported in Kodae v1; use catch")
}

func (em *emitter) emitExpr(e ast.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("nil expression")
	}
	switch x := e.(type) {
	case *ast.IdentExpr:
		if em.lookupClosure(x.Name) != nil {
			return "", fmt.Errorf("%q is a lambda binding — call it with %s(), do not use it as a value", x.Name, x.Name)
		}
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
		return fmt.Sprintf("kodae_str_lit(%s)", cQuote(x.Val)), nil
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
	case *ast.ListLit:
		return em.emitListLit(x)
	case *ast.IndexExpr:
		return em.emitIndexExpr(x)
	case *ast.TupleExpr:
		return em.emitTupleExpr(x)
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
	case *ast.StructUpdateExpr:
		return em.emitStructUpdateExpr(x)
	case *ast.FuncLit:
		return "", fmt.Errorf("lambda expression must be assigned (let cb = fn() {{ ... }})")
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
			return "", fmt.Errorf("struct %s init: missing %s", slit.TypeName, fn)
		}
		val, err := em.emitExpr(in)
		if err != nil {
			return "", err
		}
		b.WriteString(val)
	}
	b.WriteString("}")
	return b.String(), nil
}

func (em *emitter) emitTupleExpr(x *ast.TupleExpr) (string, error) {
	t, err := em.typeOf(x)
	if err != nil {
		return "", err
	}
	if t.Kind != check.KTuple {
		return "", fmt.Errorf("tuple expr expected tuple type, got %v", t)
	}
	name := cTupleCName(t)
	var b strings.Builder
	b.WriteString("(")
	b.WriteString(name)
	b.WriteString("){")
	for i, ex := range x.Exprs {
		if i > 0 {
			b.WriteString(", ")
		}
		val, err := em.emitExpr(ex)
		if err != nil {
			return "", err
		}
		b.WriteString(val)
	}
	b.WriteString("}")
	return b.String(), nil
}


func (em *emitter) emitIdentLoad(name string, _ *check.Type) string {
	name = em.maybeThisAlias(name)
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
	case "~":
		return "(~((int64_t)(" + inner + ")))", nil
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
	case "&", "|", "^":
		return "((" + l + ") " + b.Op + " (" + r + "))", nil
	case "..":
		return "", fmt.Errorf("range .. only valid in for-in")
	}

	if b.Op == "+" && lt.Kind == check.KStr && rt.Kind == check.KStr {
		return fmt.Sprintf("(kodae_str_concat((%s), (%s)))", l, r), nil
	}
	if b.Op == "+" && lt.Kind == check.KStr && checkCoercesToString(rt) {
		rs, err := em.emitToStr(b.R)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(kodae_str_concat((%s), (%s)))", l, rs), nil
	}
	if b.Op == "+" && checkCoercesToString(lt) && rt.Kind == check.KStr {
		ls, err := em.emitToStr(b.L)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(kodae_str_concat((%s), (%s)))", ls, r), nil
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
	if lt != nil && lt.Kind == check.KList {
		if m.Field == "len" {
			inner, e2 := em.emitExpr(m.Left)
			if e2 != nil {
				return "", e2
			}
			return fmt.Sprintf("((int64_t)((%s).len))", inner), nil
		}
		return "", fmt.Errorf("list has no field %q (use [i], len(x), or .len)", m.Field)
	}
	if lt == nil || lt.Kind != check.KStruct {
		if m.Field == "ok" || m.Field == "value" || m.Field == "err" {
			return "", fmt.Errorf("result field access (.ok/.value/.err) is not part of Kodae v1; use catch")
		}
		return "", fmt.Errorf("member: expected struct value on left, got %v", lt)
	}
	if id, ok := m.Left.(*ast.IdentExpr); ok && em.isStructParam(id.Name) {
		return fmt.Sprintf("((%s)->u_%s)", cid(em.maybeThisAlias(id.Name)), cid(m.Field)), nil
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
			return fmt.Sprintf("(kodae_int_to_str((int64_t)(%s)))", arg), nil
		case check.KFloat:
			return fmt.Sprintf("(kodae_float_to_str(%s))", arg), nil
		case check.KBool:
			return fmt.Sprintf("(kodae_bool_to_str(%s))", arg), nil
		case check.KList:
			return fmt.Sprintf("kodae_str_lit(\"[list]\")"), nil
		case check.KStruct:
			return fmt.Sprintf("kodae_str_lit(\"[struct]\")"), nil
		default:
			return fmt.Sprintf("(kodae_int_to_str((int64_t)(%s)))", arg), nil
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
	case check.KInt, check.KFloat, check.KBool, check.KEnum, check.KList, check.KStruct:
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
	if lf != nil && lf.Kind == check.KList {
		return em.emitListMethodCall(c, me, lf)
	}
	if lf != nil && lf.Kind == check.KStr {
		return em.emitStringMethodCall(c, me)
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

func (em *emitter) emitIndexExpr(ix *ast.IndexExpr) (string, error) {
	lt, err := em.typeOf(ix.Left)
	if err != nil {
		return "", err
	}
	if lt == nil || lt.Kind != check.KList || lt.Elem == nil {
		return "", fmt.Errorf("indexing requires list value")
	}
	base, err := em.emitLvalue(ix.Left)
	if err != nil {
		return "", err
	}
	i, err := em.emitExpr(ix.Index)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(*((%s*)kodae_list_at_ptr(&(%s), (int64_t)(%s))))", cT(lt.Elem), base, i), nil
}

func (em *emitter) emitListElemAddr(et *check.Type, arg string) string {
	// Use C99 compound literal array trick to get a pointer to the value (rvalue or lvalue).
	// This avoids the double-wrap issue and works for structs initialized from structs in C.
	return fmt.Sprintf("((%s[]){%s})", cT(et), arg)
}

func (em *emitter) emitListLit(ll *ast.ListLit) (string, error) {
	if em.curFn == nil {
		return "", fmt.Errorf("global list literals are not supported yet")
	}
	tt, err := em.typeOf(ll)
	if err != nil {
		return "", err
	}
	if tt == nil || tt.Kind != check.KList || tt.Elem == nil {
		return "", fmt.Errorf("list literal: missing inferred list type")
	}
	em.tryN++
	tmp := fmt.Sprintf("c_list_%d", em.tryN)
	em.tryPre = append(em.tryPre, fmt.Sprintf("kodae_list %s = kodae_list_new((int64_t)sizeof(%s));\n", tmp, cT(tt.Elem)))
	for _, el := range ll.Elems {
		ev, err := em.emitExpr(el)
		if err != nil {
			return "", err
		}
		em.tryPre = append(em.tryPre, fmt.Sprintf("kodae_list_push(&%s, %s);\n", tmp, em.emitListElemAddr(tt.Elem, ev)))
	}
	return tmp, nil
}

func (em *emitter) emitListMethodCall(c *ast.CallExpr, me *ast.MemberExpr, lt *check.Type) (string, error) {
	if lt == nil || lt.Kind != check.KList || lt.Elem == nil {
		return "", fmt.Errorf("list method: receiver must be list")
	}
	base, err := em.emitLvalue(me.Left)
	if err != nil {
		return "", err
	}
	switch me.Field {
	case "push":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("push: need 1 argument")
		}
		v, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("kodae_list_push(&(%s), %s)", base, em.emitListElemAddr(lt.Elem, v)), nil
	case "append":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("append: need 1 argument")
		}
		r, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("kodae_list_append(&(%s), &(%s))", base, r), nil
	case "pop":
		if len(c.Args) != 0 {
			return "", fmt.Errorf("pop: no arguments")
		}
		em.tryN++
		tmp := fmt.Sprintf("c_pop_%d", em.tryN)
		em.tryPre = append(em.tryPre,
			fmt.Sprintf("%s %s;\n", cT(lt.Elem), tmp),
			fmt.Sprintf("kodae_list_pop(&(%s), &%s);\n", base, tmp))
		return tmp, nil
	case "remove":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("remove: need index argument")
		}
		i, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		em.tryN++
		tmp := fmt.Sprintf("c_rem_%d", em.tryN)
		em.tryPre = append(em.tryPre,
			fmt.Sprintf("%s %s;\n", cT(lt.Elem), tmp),
			fmt.Sprintf("kodae_list_remove_at(&(%s), (int64_t)(%s), &%s);\n", base, i, tmp))
		return tmp, nil
	case "shuffle":
		return fmt.Sprintf("kodae_list_shuffle(&(%s))", base), nil
	case "first":
		return fmt.Sprintf("(*((%s*)kodae_list_at_ptr(&(%s), 0)))", cT(lt.Elem), base), nil
	case "last":
		return fmt.Sprintf("(*((%s*)kodae_list_at_ptr(&(%s), (int64_t)((%s).len - 1))))", cT(lt.Elem), base, base), nil
	case "reverse":
		return fmt.Sprintf("kodae_list_reverse(&(%s))", base), nil
	case "sort":
		return fmt.Sprintf("kodae_list_sort(&(%s))", base), nil
	default:
		return "", fmt.Errorf("list has no method %q", me.Field)
	}
}

func (em *emitter) emitStringMethodCall(c *ast.CallExpr, me *ast.MemberExpr) (string, error) {
	base, err := em.emitExpr(me.Left)
	if err != nil {
		return "", err
	}
	switch me.Field {
	case "len":
		return fmt.Sprintf("((int64_t)(%s).len)", base), nil
	case "upper", "lower", "trim", "reverse", "is_empty", "is_number":
		return fmt.Sprintf("kodae_str_%s(%s)", me.Field, base), nil
	case "repeat":
		a0, _ := em.emitExpr(c.Args[0])
		return fmt.Sprintf("kodae_str_repeat(%s, (int64_t)%s)", base, a0), nil
	case "contains", "starts", "ends":
		a0, _ := em.emitExpr(c.Args[0])
		return fmt.Sprintf("kodae_str_%s(%s, %s)", me.Field, base, a0), nil
	case "replace":
		a0, _ := em.emitExpr(c.Args[0])
		a1, _ := em.emitExpr(c.Args[1])
		return fmt.Sprintf("kodae_str_replace(%s, %s, %s)", base, a0, a1), nil
	case "split":
		a0, _ := em.emitExpr(c.Args[0])
		return fmt.Sprintf("kodae_str_split(%s, %s)", base, a0), nil
	case "slice":
		a0, _ := em.emitExpr(c.Args[0])
		a1, _ := em.emitExpr(c.Args[1])
		return fmt.Sprintf("kodae_str_slice(%s, (int64_t)%s, (int64_t)%s)", base, a0, a1), nil
	default:
		return "", fmt.Errorf("str has no method %q", me.Field)
	}
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

func (em *emitter) emitCall(c *ast.CallExpr) (string, error) {
	if me, ok := c.Fun.(*ast.MemberExpr); ok {
		return em.emitMethodCall(c, me)
	}
	name, ok := emitPeelCallFunc(c.Fun)
	if !ok {
		return "", fmt.Errorf("indirect call: use f(...) or (f)(...) or m.method(...)")
	}
	if ci := em.lookupClosure(name); ci != nil {
		if len(c.Args) != 0 {
			return "", fmt.Errorf("closure call takes no arguments")
		}
		if ci.CapturesThis {
			return fmt.Sprintf("%s(%s)", ci.Mangled, cid("self")), nil
		}
		return fmt.Sprintf("%s()", ci.Mangled), nil
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
		return fmt.Sprintf("(kodae_input((%s)))", p), nil
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
		return fmt.Sprintf("(kodae_random((int64_t)(%s), (int64_t)(%s)))", a0, a1), nil
	case "len":
		if len(c.Args) != 1 {
			return "", fmt.Errorf("len: need one argument")
		}
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("((int64_t)((%s).len))", a0), nil
	case "clear_screen":
		if len(c.Args) != 0 {
			return "", fmt.Errorf("clear_screen: no arguments")
		}
		return "kodae_clear_screen()", nil
	case "ok", "err":
		return "", fmt.Errorf("%s(...) is not supported in Kodae v1; use catch", name)
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
	case "sqrt", "floor", "ceil", "round", "sin", "cos", "tan", "log":
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		if name == "log" {
			t, _ := em.typeOf(c.Args[0])
			if t.Kind == check.KStr {
				return fmt.Sprintf("kodae_log(%s)", a0), nil
			}
		}
		cname := name
		return fmt.Sprintf("(%s((double)(%s)))", cname, a0), nil
	case "pow", "atan2":
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		a1, err := em.emitExpr(c.Args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s((double)(%s), (double)(%s)))", name, a0, a1), nil
	case "clamp":
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		a1, err := em.emitExpr(c.Args[1])
		if err != nil {
			return "", err
		}
		a2, err := em.emitExpr(c.Args[2])
		if err != nil {
			return "", err
		}
		t, _ := em.typeOf(c)
		if t.Kind == check.KFloat {
			return fmt.Sprintf("kodae_clamp_float(%s, %s, %s)", a0, a1, a2), nil
		}
		return fmt.Sprintf("kodae_clamp_int(%s, %s, %s)", a0, a1, a2), nil
	case "lerp":
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		a1, err := em.emitExpr(c.Args[1])
		if err != nil {
			return "", err
		}
		a2, err := em.emitExpr(c.Args[2])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("kodae_lerp(%s, %s, %s)", a0, a1, a2), nil
	case "map":
		args := make([]string, 5)
		for i := 0; i < 5; i++ {
			s, err := em.emitExpr(c.Args[i])
			if err != nil {
				return "", err
			}
			args[i] = s
		}
		return fmt.Sprintf("kodae_map(%s, %s, %s, %s, %s)", args[0], args[1], args[2], args[3], args[4]), nil
	case "distance":
		args := make([]string, 4)
		for i := 0; i < 4; i++ {
			s, err := em.emitExpr(c.Args[i])
			if err != nil {
				return "", err
			}
			args[i] = s
		}
		return fmt.Sprintf("kodae_distance(%s, %s, %s, %s)", args[0], args[1], args[2], args[3]), nil
	case "angle_to":
		args := make([]string, 4)
		for i := 0; i < 4; i++ {
			s, err := em.emitExpr(c.Args[i])
			if err != nil {
				return "", err
			}
			args[i] = s
		}
		return fmt.Sprintf("kodae_angle_to(%s, %s, %s, %s)", args[0], args[1], args[2], args[3]), nil
	case "time", "time_ms", "timer_start":
		return "kodae_" + name + "()", nil
	case "timer_elapsed", "countdown", "wait", "wait_ms":
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("kodae_%s(%s)", name, a0), nil
	case "countdown_done":
		a0, err := em.emitExpr(c.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("kodae_countdown_done(%s)", a0), nil
	case "random_float", "random_bool":
		if name == "random_bool" {
			return "kodae_random_bool()", nil
		}
		a0, _ := em.emitExpr(c.Args[0])
		a1, _ := em.emitExpr(c.Args[1])
		return fmt.Sprintf("kodae_random_float(%s, %s)", a0, a1), nil
	case "chance":
		a0, _ := em.emitExpr(c.Args[0])
		return fmt.Sprintf("kodae_chance(%s)", a0), nil
	case "random_pick":
		a0, _ := em.emitExpr(c.Args[0])
		lt, _ := em.typeOf(c.Args[0])
		return fmt.Sprintf("(*((%s*)kodae_random_pick(&(%s))))", cT(lt.Elem), a0), nil
	case "read_file", "file_exists", "delete_file", "make_folder", "delete_folder", "folder_exists", "os_name", "clipboard_get":
		if len(c.Args) == 0 {
			return "kodae_" + name + "()", nil
		}
		a0, _ := em.emitExpr(c.Args[0])
		return fmt.Sprintf("kodae_%s(%s)", name, a0), nil
	case "write_file", "append_file", "copy_file", "move_file", "clipboard_set", "open_url", "run":
		args := make([]string, len(c.Args))
		for i := 0; i < len(c.Args); i++ {
			args[i], _ = em.emitExpr(c.Args[i])
		}
		return "kodae_" + name + "(" + strings.Join(args, ", ") + ")", nil
	case "list_files":
		a0, _ := em.emitExpr(c.Args[0])
		return fmt.Sprintf("kodae_list_files(%s)", a0), nil
	case "printn":
		return em.emitPrintInternal(c, false)
	case "input_int", "input_float":
		a0, _ := em.emitExpr(c.Args[0])
		return "kodae_" + name + "(" + a0 + ")", nil
	case "swap":
		a0, _ := em.emitLvalue(c.Args[0])
		a1, _ := em.emitLvalue(c.Args[1])
		t, _ := em.typeOf(c.Args[0])
		return fmt.Sprintf("({ %s _tmp = %s; %s = %s; %s = _tmp; })", cT(t), a0, a0, a1, a1), nil
	case "in_range":
		args := make([]string, 3)
		for i := 0; i < 3; i++ {
			args[i], _ = em.emitExpr(c.Args[i])
		}
		return fmt.Sprintf("kodae_in_range(%s, %s, %s)", args[0], args[1], args[2]), nil
	case "in_rect":
		args := make([]string, 6)
		for i := 0; i < 6; i++ {
			args[i], _ = em.emitExpr(c.Args[i])
		}
		return fmt.Sprintf("kodae_in_rect(%s, %s, %s, %s, %s, %s)", args[0], args[1], args[2], args[3], args[4], args[5]), nil
	case "exit":
		a0, _ := em.emitExpr(c.Args[0])
		return "exit((int)(" + a0 + "))", nil
	case "args":
		return "kodae_args()", nil
	case "env":
		a0, _ := em.emitExpr(c.Args[0])
		return "kodae_env(" + a0 + ")", nil
	case "is_windows", "is_mac", "is_linux":
		return "kodae_" + name + "()", nil
	case "assert":
		a0, _ := em.emitExpr(c.Args[0])
		a1, _ := em.emitExpr(c.Args[1])
		return "kodae_assert(" + a0 + ", " + a1 + ")", nil
	case "debug":
		a0, _ := em.emitExpr(c.Args[0])
		at, _ := em.typeOf(c.Args[0])
		var b strings.Builder
		if err := em.appendShowValueC(&b, a0, at); err != nil {
			return "", err
		}
		b.WriteString("printf(\"\\n\");")
		return b.String(), nil
	case "todo":
		a0, _ := em.emitExpr(c.Args[0])
		return "kodae_" + name + "(" + a0 + ")", nil
	case "benchmark_start":
		return "kodae_time()", nil
	case "benchmark_end":
		a0, _ := em.emitExpr(c.Args[0])
		a1, _ := em.emitExpr(c.Args[1])
		return "kodae_benchmark_end(" + a0 + ", " + a1 + ")", nil
	case "json_read":
		a0, _ := em.emitExpr(c.Args[0])
		return "kodae_json_read(" + a0 + ")", nil
	case "json_write":
		a0, _ := em.emitExpr(c.Args[0])
		a1, _ := em.emitExpr(c.Args[1])
		// For now, only structs/lists/scalars are supported in json_write
		return "kodae_json_write(" + a0 + ", " + a1 + ")", nil
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

// emitStrAsCStr passes a kodae_str value to a C `const char*` / ptr[byte] param.
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
	if pWant != nil && pWant.Kind == check.KF32 {
		s, err := em.emitExpr(a)
		if err != nil {
			return "", err
		}
		return "((float)(" + s + "))", nil
	}
	if pWant != nil && pWant.Kind == check.KI32 {
		s, err := em.emitExpr(a)
		if err != nil {
			return "", err
		}
		return "((int32_t)(int64_t)(" + s + "))", nil
	}
	if pWant != nil && pWant.Kind == check.KU32 {
		s, err := em.emitExpr(a)
		if err != nil {
			return "", err
		}
		return "((uint32_t)(uint64_t)(int64_t)(" + s + "))", nil
	}
	if pWant != nil && pWant.Kind == check.KU8 {
		s, err := em.emitExpr(a)
		if err != nil {
			return "", err
		}
		return "((uint8_t)(int64_t)(" + s + "))", nil
	}
	if pWant != nil && pWant.Kind == check.KStruct {
		return em.emitExternByValueStructArg(a)
	}
	return em.emitExpr(a)
}

// emitExternByValueStructArg passes a struct to a C extern expecting a value (not S_* pointer).
func (em *emitter) emitExternByValueStructArg(a ast.Expr) (string, error) {
	t, err := em.typeOf(a)
	if err != nil {
		return "", err
	}
	if t == nil || t.Kind != check.KStruct {
		return em.emitExpr(a)
	}
	if sl, ok := a.(*ast.StructLit); ok {
		return em.emitStructLit(sl)
	}
	if id, ok := a.(*ast.IdentExpr); ok {
		return em.emitIdentLoad(id.Name, t), nil
	}
	inner, err2 := em.emitExpr(a)
	if err2 != nil {
		return "", err2
	}
	return inner, nil
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
	out := b.String()
	if ex.Return != nil {
		rt, err := em.resolveTypeExpr(ex.Return)
		if err == nil && rt != nil {
			switch rt.Kind {
			case check.KF32:
				return "((double)(" + out + "))", nil
			case check.KI32, check.KU32, check.KU8:
				return "((int64_t)(" + out + "))", nil
			}
		}
	}
	return out, nil
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
		fmt.Fprintf(b, "kodae_show_int((int64_t)(%s));\n", cexpr)
	case check.KFloat:
		fmt.Fprintf(b, "kodae_show_float((double)(%s));\n", cexpr)
	case check.KBool:
		fmt.Fprintf(b, "kodae_show_bool((bool)(%s));\n", cexpr)
	case check.KStr:
		fmt.Fprintf(b, "kodae_show_str((%s));\n", cexpr)
	case check.KList:
		if t.Elem == nil {
			return fmt.Errorf("list: missing element type")
		}
		if err := em.appendShowListBlock(b, cexpr, t); err != nil {
			return err
		}
	default:
		return fmt.Errorf("print: value type %s not supported", t)
	}
	return nil
}

func (em *emitter) appendShowListBlock(b *strings.Builder, varExpr string, t *check.Type) error {
	em.rcN++
	idx := fmt.Sprintf("c_li_%d", em.rcN)
	fmt.Fprintf(b, "printf(\"[\"); ")
	fmt.Fprintf(b, "for (int64_t %s = 0; %s < (int64_t)(%s).len; %s++) { ", idx, idx, varExpr, idx)
	fmt.Fprintf(b, "if (%s > 0) printf(\", \"); ", idx)
	elemAddr := fmt.Sprintf("(*((%s*)kodae_list_at_ptr(&(%s), %s)))", cT(t.Elem), varExpr, idx)
	if err := em.appendShowValueC(b, elemAddr, t.Elem); err != nil {
		return err
	}
	fmt.Fprintf(b, "} printf(\"]\");")
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
	return em.emitPrintInternal(c, true)
}

func (em *emitter) emitPrintInternal(c *ast.CallExpr, newline bool) (string, error) {
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
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\"); } else { kodae_print_int((%s).v); }\n", av, av)
			case check.KFloat:
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\"); } else { kodae_print_float((%s).v); }\n", av, av)
			case check.KStr:
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\"); } else { kodae_print_str((%s).v); }\n", av, av)
			case check.KBool:
				fmt.Fprintf(&b, "if (!(%s).has) { printf(\"none\"); } else { kodae_print_bool((%s).v); }\n", av, av)
			case check.KStruct:
				innerT := in
				if innerT.StructDef == nil {
					return "", fmt.Errorf("print: bad optional struct")
				}
				b.WriteString("if (!(" + av + ").has) { printf(\"none\"); } else { ")
				if err := em.appendShowStructBlock(&b, "("+av+").v", innerT); err != nil {
					return "", err
				}
				b.WriteString("}\n")
			default:
				return "", fmt.Errorf("print: cannot print optional of %s", in)
			}
			continue
		}
		switch t.Kind {
		case check.KInt, check.KEnum:
			fmt.Fprintf(&b, "kodae_print_int((int64_t)(%s));\n", av)
		case check.KFloat:
			fmt.Fprintf(&b, "kodae_print_float(%s);\n", av)
		case check.KBool:
			fmt.Fprintf(&b, "kodae_print_bool(%s);\n", av)
		case check.KStr:
			fmt.Fprintf(&b, "kodae_print_str(%s);\n", av)
		case check.KStruct:
			if t.StructDef == nil {
				return "", fmt.Errorf("struct %s has no metadata", t)
			}
			if err := em.appendShowStructBlock(&b, av, t); err != nil {
				return "", err
			}
		case check.KList:
			if err := em.appendShowListBlock(&b, av, t); err != nil {
				return "", err
			}
		case check.KVoid:
			return "", fmt.Errorf("print: void argument")
		default:
			return "", fmt.Errorf("print: type %s is not supported", t)
		}
	}
	if newline {
		b.WriteString("printf(\"\\n\");\n")
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
			return fmt.Sprintf("(kodae_int_to_str((int64_t)(%s)))", av), nil
		case check.KFloat:
			return fmt.Sprintf("(kodae_float_to_str(%s))", av), nil
		case check.KBool:
			return fmt.Sprintf("(kodae_bool_to_str(%s))", av), nil
		default:
			return fmt.Sprintf("(kodae_int_to_str((int64_t)(%s)))", av), nil
		}
	}
	return "", fmt.Errorf("builtin %s", name)
}

