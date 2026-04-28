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
	fmt.Println("kodae bundle — creating portable distribution...")

	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH

	// Override if requested (e.g. kodae bundle linux amd64)
	if len(args) >= 1 {
		targetOS = args[0]
	}
	if len(args) >= 2 {
		targetArch = args[1]
	}

	distDir := "dist"
	bundleName := fmt.Sprintf("kodae-%s-%s", targetOS, targetArch)
	bundleDir := filepath.Join(distDir, bundleName)

	fmt.Printf("target: %s/%s\n", targetOS, targetArch)
	
	if err := os.MkdirAll(filepath.Join(bundleDir, "bin"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(bundleDir, "toolchain", "zig"), 0755); err != nil {
		return err
	}

	// 1. Build kodae
	fmt.Println("building kodae...")
	kodaeName := "kodae"
	if targetOS == "windows" {
		kodaeName = "kodae.exe"
	}
	outPath := filepath.Join(bundleDir, "bin", kodaeName)
	
	cmd := exec.Command("go", "build", "-o", outPath, "./cmd/kodae")
	cmd.Env = append(os.Environ(), "GOOS="+targetOS, "GOARCH="+targetArch)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %v\n%s", err, string(b))
	}

	// 2. Obtain Zig (copy from PATH if current platform matches)
	zigExe := "zig"
	if targetOS == "windows" {
		zigExe = "zig.exe"
	}
	if targetOS == runtime.GOOS && targetArch == runtime.GOARCH {
		fmt.Println("searching for zig on PATH to bundle...")
		if p, err := exec.LookPath("zig"); err == nil {
			fmt.Printf("found zig at %s, copying...\n", p)
			zigDst := filepath.Join(bundleDir, "toolchain", "zig", zigExe)
			if err := copyFile(p, zigDst); err != nil {
				fmt.Printf("warning: failed to copy zig: %v\n", err)
			} else {
				fmt.Println("zig bundled successfully")
			}
		} else {
			fmt.Println("zig not found on PATH, skipping bundle inclusion")
		}
	} else {
		fmt.Println("cross-compiling: skipping automatic zig bundling (download manually and place in toolchain/zig/)")
	}

	// 3. Copy include and examples
	fmt.Println("bundling include/ and examples/...")
	if err := copyDir("include", filepath.Join(bundleDir, "include")); err != nil {
		fmt.Printf("warning: failed to bundle include/: %v\n", err)
	}
	if err := copyDir("examples", filepath.Join(bundleDir, "examples")); err != nil {
		fmt.Printf("warning: failed to bundle examples/: %v\n", err)
	}
	if err := copyFile("README.md", filepath.Join(bundleDir, "README.md")); err != nil {
		fmt.Printf("warning: failed to bundle README.md: %v\n", err)
	}
	
	fmt.Printf("\nbundle created at: %s\n", bundleDir)
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
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
