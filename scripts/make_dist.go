package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	platforms := []struct {
		os     string
		arch   string
		zigUrl string
		zigExe string
	}{
		{"windows", "amd64", "https://ziglang.org/download/0.12.0/zig-windows-x86_64-0.12.0.zip", "zig.exe"},
		{"linux", "amd64", "https://ziglang.org/download/0.12.0/zig-linux-x86_64-0.12.0.tar.xz", "zig"},
		{"darwin", "arm64", "https://ziglang.org/download/0.12.0/zig-macos-aarch64-0.12.0.tar.xz", "zig"},
	}

	fmt.Println("Clio Distribution & Bundling Tool")
	fmt.Println("================================")

	for _, p := range platforms {
		fmt.Printf("\n--- Creating Bundle: %s/%s ---\n", p.os, p.arch)
		
		// 1. Run clio bundle (builds binary and sets up folder structure)
		cmd := exec.Command("go", "run", "./cmd/clio", "bundle", p.os, p.arch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error bundling for %s/%s: %v\n", p.os, p.arch, err)
			continue
		}

		bundleDir := filepath.Join("dist", fmt.Sprintf("clio-%s-%s", p.os, p.arch))
		zigDstDir := filepath.Join(bundleDir, "toolchain", "zig")
		zigDstPath := filepath.Join(zigDstDir, p.zigExe)

		// 2. Inform user about Zig
		fmt.Printf("Bundle structure ready at %s\n", bundleDir)
		fmt.Printf("To finish this bundle, download Zig from:\n  %s\n", p.zigUrl)
		fmt.Printf("And place the '%s' binary into:\n  %s\n", p.zigExe, zigDstPath)
	}

	fmt.Println("\nBundling complete (folder structures created).")
	fmt.Println("Distribute the folders as .zip or .tar.gz to users.")
}
