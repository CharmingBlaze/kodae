package llir

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CompileIRWithClang writes LLVM IR to a temp file and invokes clang to produce outExe.
func CompileIRWithClang(ir string, outExe string) error {
	if outExe == "" {
		return fmt.Errorf("llir: empty output path")
	}
	d, err := os.MkdirTemp("", "kodae-llir-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(d)
	ll := filepath.Join(d, "out.ll")
	if err := os.WriteFile(ll, []byte(ir), 0644); err != nil {
		return err
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		return fmt.Errorf("llir: need clang on PATH to compile LLVM IR: %w", err)
	}
	cmd := exec.Command(clang, "-Wno-override-module", ll, "-o", outExe)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("llir: clang: %w", err)
	}
	return nil
}
