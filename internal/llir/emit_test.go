package llir

import (
	"strings"
	"testing"
)

func TestEmitMinimalMainContainsRet(t *testing.T) {
	t.Parallel()
	s := EmitMinimalMain(42)
	if !strings.Contains(s, "ret i32 42") {
		t.Fatalf("expected ret i32 42 in IR, got:\n%s", s)
	}
}
