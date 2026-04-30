// Compile links one generated C99 source to an executable using a GNU/Clang driver,
// TinyCC (TCC), or "zig cc".
package ccdriver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Compile builds cSrcPath to outPath. outPath is typically an absolute or cwd-relative
// file name (e.g. "a.out" or "app" or "app.exe" on Windows). extra is appended after -lm
// (e.g. -lraylib, -L/path) from # link in source or the CLI.
func Compile(ccc CCmd, cSrcPath, outPath string, extra []string, gui bool) error {
	if ccc.TCC {
		return compileTCC(ccc, cSrcPath, outPath, extra, gui)
	}
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
	argv = append(argv, "-std=c99", "-O2", "-o", outAbs, cAbs, "-lm")
	if runtime.GOOS == "windows" {
		if gui {
			argv = append(argv, "-mwindows")
		}
		argv = append(argv, "-lws2_32")
	}
	argv = append(argv, extra...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler: %w", err)
	}
	return nil
}

func compileTCC(ccc CCmd, cSrcPath, outPath string, extra []string, gui bool) error {
	outAbs, err := filepath.Abs(outPath)
	if err != nil {
		return err
	}
	cAbs, err := filepath.Abs(cSrcPath)
	if err != nil {
		return err
	}
	// TCC accepts a GCC-like subset; avoid -O2 (not always meaningful for TCC).
	argv := []string{ccc.Prog, "-std=c99", "-o", outAbs, cAbs}
	if runtime.GOOS == "windows" {
		if gui {
			argv = append(argv, "-mwindows")
		}
		argv = append(argv, "-lws2_32")
	} else {
		argv = append(argv, "-lm")
	}
	argv = append(argv, extra...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("TCC: %w", err)
	}
	return nil
}

func CompileObject(ccc CCmd, cSrcPath, objPath string, extra []string) error {
	if ccc.TCC {
		return compileTCCObject(ccc, cSrcPath, objPath, extra)
	}
	objAbs, err := filepath.Abs(objPath)
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
	argv = append(argv, "-std=c99", "-O2", "-c", "-o", objAbs, cAbs)
	argv = append(argv, extra...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler (object): %w", err)
	}
	return nil
}

func compileTCCObject(ccc CCmd, cSrcPath, objPath string, extra []string) error {
	objAbs, err := filepath.Abs(objPath)
	if err != nil {
		return err
	}
	cAbs, err := filepath.Abs(cSrcPath)
	if err != nil {
		return err
	}
	argv := []string{ccc.Prog, "-std=c99", "-c", "-o", objAbs, cAbs}
	argv = append(argv, extra...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("TCC (object): %w", err)
	}
	return nil
}

func ArchiveStatic(objPath, libPath string) error {
	ar, err := findArchiver()
	if err != nil {
		return err
	}
	libAbs, err := filepath.Abs(libPath)
	if err != nil {
		return err
	}
	objAbs, err := filepath.Abs(objPath)
	if err != nil {
		return err
	}
	cmd := exec.Command(ar, "rcs", libAbs, objAbs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ar: %w", err)
	}
	return nil
}

func findArchiver() (string, error) {
	cands := []string{"ar", "llvm-ar", "gcc-ar"}
	for _, c := range cands {
		if p, err := exec.LookPath(c); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("static library requires ar/llvm-ar/gcc-ar on PATH")
}

func LinkShared(ccc CCmd, cSrcPath, outPath string, extra []string, gui bool) error {
	if ccc.TCC {
		return linkSharedTCC(ccc, cSrcPath, outPath, extra, gui)
	}
	outAbs, err := filepath.Abs(outPath)
	if err != nil {
		return err
	}
	cAbs, err := filepath.Abs(cSrcPath)
	if err != nil {
		return err
	}
	argv := make([]string, 0, len(ccc.Prefix)+12+len(extra))
	argv = append(argv, ccc.Prog)
	argv = append(argv, ccc.Prefix...)
	argv = append(argv, "-std=c99", "-O2")
	if runtime.GOOS == "windows" {
		argv = append(argv, "-shared")
	} else {
		argv = append(argv, "-shared", "-fPIC")
	}
	argv = append(argv, "-o", outAbs, cAbs, "-lm")
	if runtime.GOOS == "windows" {
		if gui {
			argv = append(argv, "-mwindows")
		}
		argv = append(argv, "-lws2_32")
	}
	argv = append(argv, extra...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C linker (shared): %w", err)
	}
	return nil
}

func linkSharedTCC(ccc CCmd, cSrcPath, outPath string, extra []string, gui bool) error {
	outAbs, err := filepath.Abs(outPath)
	if err != nil {
		return err
	}
	cAbs, err := filepath.Abs(cSrcPath)
	if err != nil {
		return err
	}
	argv := []string{ccc.Prog, "-std=c99", "-shared", "-o", outAbs, cAbs}
	if runtime.GOOS != "windows" {
		argv = append(argv, "-fPIC", "-lm")
	} else {
		if gui {
			argv = append(argv, "-mwindows")
		}
		argv = append(argv, "-lws2_32")
	}
	argv = append(argv, extra...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("TCC (shared): %w", err)
	}
	return nil
}
