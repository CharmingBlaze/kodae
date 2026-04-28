package ccdriver

import (
	"path/filepath"
	"testing"
)

func TestZigExeName(t *testing.T) {
	t.Parallel()
	n := zigExeName()
	if n == "" {
		t.Fatal("zig executable name must not be empty")
	}
}

func TestBundledZigCandidatesIncludesBundleFromBinCwd(t *testing.T) {
	t.Parallel()
	exe := `C:\kodae\bin\kodae.exe`
	cwd := `C:\kodae\bin`
	want := filepath.Clean(`C:\kodae\toolchain\zig\` + zigExeName())

	got := bundledZigCandidates(exe, cwd)
	for _, c := range got {
		if c == want {
			return
		}
	}
	t.Fatalf("expected candidate %q in %v", want, got)
}
