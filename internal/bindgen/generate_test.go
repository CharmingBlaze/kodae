package bindgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateBindings(t *testing.T) {
	header := `
struct TestStruct {
    int x;
    float y;
};
enum TestEnum { A, B };
void TestFunc(struct TestStruct s, enum TestEnum e);
`
	tmpDir, err := os.MkdirTemp("", "kodae-bind-test")
	old, _ := os.Getwd()
	os.Chdir(tmpDir)
	t.Cleanup(func() { os.Chdir(old) })
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	headerPath := filepath.Join(tmpDir, "test.h")
	if err := os.WriteFile(headerPath, []byte(header), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := GenerateBindings(headerPath, "testlib")
	if err != nil {
		t.Fatalf("GenerateBindings failed: %v", err)
	}

	if res.Structs != 1 {
		t.Errorf("expected 1 struct, got %d", res.Structs)
	}
	if res.Externs != 1 {
		t.Errorf("expected 1 extern, got %d", res.Externs)
	}

	expected := "pub enum TestEnum { A, B }"
	if !contains(res.Content, expected) {
		t.Errorf("output missing expected enum: %s", expected)
	}

	expectedStruct := "pub struct TestStruct {"
	if !contains(res.Content, expectedStruct) {
		t.Errorf("output missing expected struct: %s", expectedStruct)
	}

	expectedFunc := "extern fn TestFunc(s: TestStruct, e: TestEnum) -> void"
	if !contains(res.Content, expectedFunc) {
		t.Errorf("output missing expected function: %s", expectedFunc)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
