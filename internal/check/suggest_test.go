package check

import "testing"

func TestLevenshtein(t *testing.T) {
	t.Parallel()
	if levenshtein("scre", "score") != 1 {
		t.Fatalf("got %d", levenshtein("scre", "score"))
	}
	if levenshtein("scroe", "score") != 2 {
		t.Fatalf("got %d for scroe→score", levenshtein("scroe", "score"))
	}
	if levenshtein("a", "b") != 1 {
		t.Fatal()
	}
	if levenshtein("", "abc") != 3 {
		t.Fatal()
	}
}

func TestSuggestName(t *testing.T) {
	t.Parallel()
	cands := []string{"score", "alive", "main", "x"}
	if s, ok := suggestName("scre", cands); !ok || s != "score" {
		t.Fatalf("got %q, %v", s, ok)
	}
	if _, ok := suggestName("nope", cands); ok {
		t.Fatal("expected no suggest")
	}
}
