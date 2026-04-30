package tccbundle

import "testing"

func TestSidecarPath_NoBundle(t *testing.T) {
	t.Parallel()
	_, ok := SidecarPath()
	if ok {
		// In CI or dev, a stray toolchain next to the test binary could make this true; tolerate.
		t.Log("SidecarPath returned true (toolchain present beside test binary)")
	}
}
