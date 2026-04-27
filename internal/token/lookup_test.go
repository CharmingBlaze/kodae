package token

import "testing"

func TestLookup_Keywords(t *testing.T) {
	cases := map[string]Type{
		"fn":     FN,
		"let":    LET,
		"return": RETURN,
		"struct": STRUCT,
		"result": RESULT,
		"ok":     OK,
		"err":    ERR,
		"enum":   ENUM,
		"match":  MATCH,
		"none":   NONE,
		"true":   TRUE,
		"false":  FALSE,
	}
	for s, want := range cases {
		if got := Lookup(s); got != want {
			t.Errorf("Lookup(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestLookup_IdentFallsBack(t *testing.T) {
	if got := Lookup("main"); got != IDENT {
		t.Errorf("expected IDENT for 'main', got %v", got)
	}
	if got := Lookup("print"); got != IDENT {
		t.Errorf("expected IDENT for 'print', got %v", got)
	}
}
