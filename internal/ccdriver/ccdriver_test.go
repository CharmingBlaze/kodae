package ccdriver

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHintTextNonEmpty(t *testing.T) {
	t.Parallel()
	if hintText() == "" {
		t.Fatal("hintText must be non-empty")
	}
}

func TestFindArchiverContract(t *testing.T) {
	t.Parallel()
	p, err := findArchiver()
	if err != nil {
		if !strings.Contains(err.Error(), "ar/llvm-ar/gcc-ar") {
			t.Fatalf("unexpected archiver error: %v", err)
		}
		return
	}
	base := strings.ToLower(filepath.Base(p))
	if !strings.Contains(base, "ar") {
		t.Fatalf("unexpected archiver path: %q", p)
	}
}
