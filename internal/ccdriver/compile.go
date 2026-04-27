// Compile links one generated C99 source to an executable using a GNU/Clang driver
// (clang, Apple clang, gcc, or "zig cc"). The same line works on Windows, Linux, and
// macOS; users may point CLIO_CC (or clio's --cc) to their toolchain if needed.
package ccdriver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Compile builds cSrcPath to outPath. outPath is typically an absolute or cwd-relative
// file name (e.g. "a.out" or "app" or "app.exe" on Windows). extra is appended after -lm
// (e.g. -lraylib, -L/path) from # link in source or the CLI.
func Compile(ccc CCmd, cSrcPath, outPath string, extra []string) error {
	outAbs, err := filepath.Abs(outPath)
	if err != nil {
		return err
	}
	cAbs, err := filepath.Abs(cSrcPath)
	if err != nil {
		return err
	}
	argv := make([]string, 0, len(ccc.Prefix)+10+len(extra))
	argv = append(argv, ccc.Prog)
	argv = append(argv, ccc.Prefix...)
	// LLVM, GCC, and "zig cc" all accept this GNU-style CLI on common targets.
	argv = append(argv, "-std=c99", "-O2", "-o", outAbs, cAbs, "-lm")
	argv = append(argv, extra...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler: %w", err)
	}
	return nil
}
