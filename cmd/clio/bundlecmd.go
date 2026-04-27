package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func runBundle(args []string) error {
	fmt.Println("clio bundle — creating portable distribution...")

	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH

	// Override if requested (e.g. clio bundle linux amd64)
	if len(args) >= 1 {
		targetOS = args[0]
	}
	if len(args) >= 2 {
		targetArch = args[1]
	}

	distDir := "dist"
	bundleName := fmt.Sprintf("clio-%s-%s", targetOS, targetArch)
	bundleDir := filepath.Join(distDir, bundleName)

	fmt.Printf("target: %s/%s\n", targetOS, targetArch)
	
	if err := os.MkdirAll(filepath.Join(bundleDir, "bin"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(bundleDir, "toolchain", "zig"), 0755); err != nil {
		return err
	}

	// 1. Build clio
	fmt.Println("building clio...")
	clioName := "clio"
	if targetOS == "windows" {
		clioName = "clio.exe"
	}
	outPath := filepath.Join(bundleDir, "bin", clioName)
	
	cmd := exec.Command("go", "build", "-o", outPath, "./cmd/clio")
	cmd.Env = append(os.Environ(), "GOOS="+targetOS, "GOARCH="+targetArch)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %v\n%s", err, string(b))
	}

	// 2. Obtain Zig (copy from PATH if current platform matches)
	if targetOS == runtime.GOOS && targetArch == runtime.GOARCH {
		fmt.Println("searching for zig on PATH to bundle...")
		if p, err := exec.LookPath("zig"); err == nil {
			fmt.Printf("found zig at %s, copying...\n", p)
			zigDst := filepath.Join(bundleDir, "toolchain", "zig", filepath.Base(p))
			if err := copyFile(p, zigDst); err != nil {
				fmt.Printf("warning: failed to copy zig: %v\n", err)
			} else {
				fmt.Println("zig bundled successfully")
			}
		} else {
			fmt.Println("zig not found on PATH, skipping bundle inclusion")
		}
	} else {
		fmt.Println("cross-compiling: skipping automatic zig bundling (download manually)")
	}
	
	fmt.Printf("\nbundle created at: %s\n", bundleDir)
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// Copy permissions
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, si.Mode())
}
