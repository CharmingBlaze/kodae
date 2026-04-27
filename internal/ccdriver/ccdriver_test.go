package ccdriver

import "testing"

func TestZigExeName(t *testing.T) {
	t.Parallel()
	n := zigExeName()
	if n == "" {
		t.Fatal("zig executable name must not be empty")
	}
}
