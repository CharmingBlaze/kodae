package llir

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"kodae/internal/ast"
	"kodae/internal/check"
	"kodae/internal/cruntime"

	_ "embed"
)

//go:embed bridge_append.txt
var bridgeAppend string

// CompileProgramLLVM lowers supported programs to LLVM IR and links with the Kodae C
// runtime + bridge using clang (requires clang on PATH).
func CompileProgramLLVM(p *ast.Program, inf *check.Info, outExe string) error {
	ir, err := LowerToLLVM(p, inf)
	if err != nil {
		return err
	}
	return compileIRWithBridge(ir, outExe)
}

func compileIRWithBridge(ir, outExe string) error {
	if outExe == "" {
		return fmt.Errorf("llir: empty output path")
	}
	d, err := os.MkdirTemp("", "kodae-llir-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(d)

	// Same TU prelude order as internal/codegen (parson declarations, bootstrap, parson impl).
	bridgeSrc := cruntime.ParsonH + "\n" + cruntime.BootstrapC + "\n" + cruntime.ParsonC + "\n" +
		cruntime.WsClientC + "\n#include <math.h>\n\n" + bridgeAppend
	bridgeC := filepath.Join(d, "bridge.c")
	ll := filepath.Join(d, "user.ll")
	bridgeO := filepath.Join(d, "bridge.o")
	userO := filepath.Join(d, "user.o")
	if err := os.WriteFile(bridgeC, []byte(bridgeSrc), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(ll, []byte(ir), 0644); err != nil {
		return err
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		return fmt.Errorf("llir: need clang on PATH: %w", err)
	}
	outAbs, err := filepath.Abs(outExe)
	if err != nil {
		return err
	}
	c1 := exec.Command(clang, "-std=c99", "-O2", "-c", "-o", bridgeO, bridgeC)
	c1.Stderr, c1.Stdout = os.Stderr, os.Stdout
	if err := c1.Run(); err != nil {
		return fmt.Errorf("llir: clang bridge.c: %w", err)
	}
	// Silence benign triple-normalization diagnostics on some Windows clang builds.
	c2 := exec.Command(clang, "-Wno-override-module", "-c", "-o", userO, ll)
	c2.Stderr, c2.Stdout = os.Stderr, os.Stdout
	if err := c2.Run(); err != nil {
		return fmt.Errorf("llir: clang user.ll: %w", err)
	}
	linkArgs := []string{clang, userO, bridgeO, "-o", outAbs}
	if runtime.GOOS != "windows" {
		linkArgs = append(linkArgs, "-lm")
	} else {
		linkArgs = append(linkArgs, "-lws2_32")
	}
	c3 := exec.Command(linkArgs[0], linkArgs[1:]...)
	c3.Stderr, c3.Stdout = os.Stderr, os.Stdout
	if err := c3.Run(); err != nil {
		return fmt.Errorf("llir: link: %w", err)
	}
	return nil
}
