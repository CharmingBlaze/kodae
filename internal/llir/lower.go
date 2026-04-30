package llir

import (
	"fmt"
	"strconv"
	"strings"

	"kodae/internal/ast"
	"kodae/internal/check"
)

// LowerToLLVM emits textual LLVM IR for a small supported subset of Kodae (see errors for limits).
func LowerToLLVM(p *ast.Program, inf *check.Info) (string, error) {
	if p == nil || inf == nil {
		return "", fmt.Errorf("llir: nil program or info")
	}
	var head, globs, main strings.Builder
	lw := &lower{
		inf:     inf,
		locals:  make(map[string]string),
		globBuf: &globs,
		mainBuf: &main,
	}
	if err := lw.validateProgram(p); err != nil {
		return "", err
	}
	mainFn := lw.findMain(p)
	if mainFn == nil || mainFn.Body == nil {
		return "", fmt.Errorf("llir: missing fn main() body")
	}
	head.WriteString("; kodae llir (experimental)\n")
	head.WriteString("declare void @rt_init()\n")
	head.WriteString("declare void @rt_print_int64(i64)\n")
	head.WriteString("declare void @rt_sb_begin_print()\n")
	head.WriteString("declare void @rt_sb_append_lit(i8*)\n")
	head.WriteString("declare void @rt_sb_append_int(i64)\n")
	head.WriteString("declare void @rt_sb_finish_print()\n")
	main.WriteString("define i32 @main() {\n")
	main.WriteString("entry:\n")
	main.WriteString("  call void @rt_init()\n")
	for _, st := range mainFn.Body.Stmts {
		if err := lw.emitStmt(st); err != nil {
			return "", err
		}
	}
	main.WriteString("  ret i32 0\n")
	main.WriteString("}\n")
	return head.String() + globs.String() + main.String(), nil
}

type lower struct {
	inf      *check.Info
	locals   map[string]string
	globBuf  *strings.Builder
	mainBuf  *strings.Builder
	strID    int
	regID    int
}

func (lw *lower) freshReg(prefix string) string {
	lw.regID++
	return fmt.Sprintf("%%%s%d", prefix, lw.regID)
}

func (lw *lower) validateProgram(p *ast.Program) error {
	for _, d := range p.Decls {
		switch t := d.(type) {
		case *ast.FnDecl:
			if t.Name != "main" {
				return fmt.Errorf("llir: only fn main is supported for --backend=llvm (found fn %s)", t.Name)
			}
			if len(t.Params) > 0 {
				return fmt.Errorf("llir: main must take no parameters")
			}
		default:
			return fmt.Errorf("llir: unsupported top-level decl %T for LLVM MVP", d)
		}
	}
	return nil
}

func (lw *lower) findMain(p *ast.Program) *ast.FnDecl {
	for _, d := range p.Decls {
		if f, ok := d.(*ast.FnDecl); ok && f.Name == "main" {
			return f
		}
	}
	return nil
}

func (lw *lower) typeOf(e ast.Expr) (*check.Type, error) {
	if e == nil {
		return nil, fmt.Errorf("nil expr")
	}
	t, ok := lw.inf.Types[check.ExprKey(e)]
	if !ok || t == nil {
		return nil, fmt.Errorf("missing type for %s", ast.ExprString(e))
	}
	return t, nil
}

func (lw *lower) emitStmt(s ast.Stmt) error {
	b := lw.mainBuf
	switch st := s.(type) {
	case *ast.LetStmt:
		if len(st.Destruct) > 0 {
			return fmt.Errorf("llir: multi-let not supported")
		}
		ty, err := lw.typeOf(st.Init)
		if err != nil {
			return err
		}
		if ty.Kind != check.KInt && ty.Kind != check.KEnum {
			return fmt.Errorf("llir: only int locals supported (got %s)", ty)
		}
		v, err := lw.emitIntRValue(st.Init)
		if err != nil {
			return err
		}
		slot := lw.freshReg("loc_")
		lw.locals[st.Name] = slot
		fmt.Fprintf(b, "  %s = alloca i64\n", slot)
		fmt.Fprintf(b, "  store i64 %s, i64* %s\n", v, slot)
		return nil
	case *ast.ExprStmt:
		return lw.emitExprStmt(st)
	case *ast.ReturnStmt:
		if st.V != nil {
			return fmt.Errorf("llir: return with value not supported")
		}
		return nil
	default:
		return fmt.Errorf("llir: unsupported stmt %T", s)
	}
}

func (lw *lower) emitExprStmt(st *ast.ExprStmt) error {
	b := lw.mainBuf
	c, ok := stripParenExpr(st.E).(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("llir: only call expression statements supported")
	}
	name, ok := callName(c)
	if !ok || name != "print" {
		return fmt.Errorf("llir: only print(...) calls supported in expr stmt")
	}
	if len(c.Args) == 0 {
		return fmt.Errorf("llir: print needs at least one argument")
	}
	for _, a := range c.Args {
		t, err := lw.typeOf(a)
		if err != nil {
			return err
		}
		switch t.Kind {
		case check.KInt, check.KEnum:
			v, err := lw.emitIntRValue(a)
			if err != nil {
				return err
			}
			fmt.Fprintf(b, "  call void @rt_print_int64(i64 %s)\n", v)
		case check.KStr:
			fmt.Fprintf(b, "  call void @rt_sb_begin_print()\n")
			if err := lw.emitStrPiecesToSB(a); err != nil {
				return err
			}
			fmt.Fprintf(b, "  call void @rt_sb_finish_print()\n")
		default:
			return fmt.Errorf("llir: print argument type %s not supported", t)
		}
	}
	return nil
}

func stripParenExpr(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.Inner
	}
}

func callName(c *ast.CallExpr) (string, bool) {
	id, ok := c.Fun.(*ast.IdentExpr)
	if !ok {
		return "", false
	}
	return id.Name, true
}

func (lw *lower) emitStrPiecesToSB(e ast.Expr) error {
	b := lw.mainBuf
	e = stripParenExpr(e)
	switch x := e.(type) {
	case *ast.StringLit:
		ptr := lw.globalCString(x.Val)
		fmt.Fprintf(b, "  call void @rt_sb_append_lit(i8* %s)\n", ptr)
		return nil
	case *ast.BinaryExpr:
		if x.Op != "+" {
			return fmt.Errorf("llir: str concat only supports +")
		}
		lt, err := lw.typeOf(x.L)
		if err != nil {
			return err
		}
		rt, err := lw.typeOf(x.R)
		if err != nil {
			return err
		}
		if lt.Kind != check.KStr || rt.Kind != check.KStr {
			return fmt.Errorf("llir: string + requires str on both sides")
		}
		if err := lw.emitStrPiecesToSB(x.L); err != nil {
			return err
		}
		return lw.emitStrPiecesToSB(x.R)
	case *ast.CallExpr:
		n, ok := callName(x)
		if !ok || n != "str" || len(x.Args) != 1 {
			return fmt.Errorf("llir: unsupported str expression %s", ast.ExprString(e))
		}
		v, err := lw.emitIntRValue(x.Args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(b, "  call void @rt_sb_append_int(i64 %s)\n", v)
		return nil
	case *ast.CastExpr:
		if strings.EqualFold(x.To, "str") {
			v, err := lw.emitIntRValue(x.Arg)
			if err != nil {
				return err
			}
			fmt.Fprintf(b, "  call void @rt_sb_append_int(i64 %s)\n", v)
			return nil
		}
		return lw.emitStrPiecesToSB(stripParenExpr(x.Arg))
	case *ast.IdentExpr:
		return fmt.Errorf("llir: str-typed locals in print not supported yet")
	default:
		return fmt.Errorf("llir: unsupported str expr %T", e)
	}
}

func (lw *lower) globalCString(s string) (ptrReg string) {
	gpre := lw.globBuf
	bMain := lw.mainBuf
	lw.strID++
	name := fmt.Sprintf("@.str.%d", lw.strID)
	n := len(s) + 1
	escaped := llvmCStringBytes(s)
	fmt.Fprintf(gpre, "%s = private unnamed_addr constant [%d x i8] c\"%s\\00\", align 1\n", name, n, escaped)
	ptr := lw.freshReg("sptr")
	fmt.Fprintf(bMain, "  %s = getelementptr inbounds [%d x i8], [%d x i8]* %s, i64 0, i64 0\n", ptr, n, n, name)
	return ptr
}

func llvmCStringBytes(s string) string {
	var o strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\', '"':
			o.WriteByte('\\')
			o.WriteByte(c)
		case '\n':
			o.WriteString("\\0A")
		case '\r':
			o.WriteString("\\0D")
		case '\t':
			o.WriteString("\\09")
		default:
			if c < 32 || c == 127 {
				o.WriteString(fmt.Sprintf("\\%02X", c))
			} else {
				o.WriteByte(c)
			}
		}
	}
	return o.String()
}

func (lw *lower) emitIntRValue(e ast.Expr) (string, error) {
	b := lw.mainBuf
	e = stripParenExpr(e)
	switch x := e.(type) {
	case *ast.IntLit:
		if x.Raw != "" {
			return x.Raw, nil
		}
		return strconv.FormatInt(x.Val, 10), nil
	case *ast.IdentExpr:
		slot, ok := lw.locals[x.Name]
		if !ok {
			return "", fmt.Errorf("llir: unknown local %q", x.Name)
		}
		r := lw.freshReg("v")
		fmt.Fprintf(b, "  %s = load i64, i64* %s\n", r, slot)
		return r, nil
	case *ast.UnaryExpr:
		if x.Op != "-" {
			return "", fmt.Errorf("llir: unary %q not supported", x.Op)
		}
		inner, err := lw.emitIntRValue(x.X)
		if err != nil {
			return "", err
		}
		r := lw.freshReg("v")
		fmt.Fprintf(b, "  %s = sub i64 0, %s\n", r, inner)
		return r, nil
	case *ast.BinaryExpr:
		lt, err := lw.typeOf(x.L)
		if err != nil {
			return "", err
		}
		if lt.Kind != check.KInt && lt.Kind != check.KEnum {
			return "", fmt.Errorf("llir: binary int expected")
		}
		lhs, err := lw.emitIntRValue(x.L)
		if err != nil {
			return "", err
		}
		rhs, err := lw.emitIntRValue(x.R)
		if err != nil {
			return "", err
		}
		r := lw.freshReg("v")
		switch x.Op {
		case "+":
			fmt.Fprintf(b, "  %s = add i64 %s, %s\n", r, lhs, rhs)
		case "-":
			fmt.Fprintf(b, "  %s = sub i64 %s, %s\n", r, lhs, rhs)
		case "*":
			fmt.Fprintf(b, "  %s = mul i64 %s, %s\n", r, lhs, rhs)
		case "/":
			fmt.Fprintf(b, "  %s = sdiv i64 %s, %s\n", r, lhs, rhs)
		case "%":
			fmt.Fprintf(b, "  %s = srem i64 %s, %s\n", r, lhs, rhs)
		default:
			return "", fmt.Errorf("llir: unsupported int op %q", x.Op)
		}
		return r, nil
	default:
		return "", fmt.Errorf("llir: unsupported int expr %T", e)
	}
}
