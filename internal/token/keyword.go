package token

var keywords = map[string]Type{
	"and":    AND,
	"or":     OR,
	"not":    NOT,
	"NOT":    NOT,
	"fn":     FN,
	"let":    LET,
	"const":  CONST,
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
	"for":    FOR,
	"in":     IN,
	"loop":   LOOP,
	"break":  BREAK,
	"return": RETURN,
	"struct": STRUCT,
	"module": MODULE,
	"use":    USE,
	"pub":    PUB,
	"extern": EXTERN,
	"true":   TRUE,
	"false":  FALSE,
	"as":     AS,
	"result": RESULT,
	"catch":  CATCH,
	"ok":     OK,
	"err":    ERR,
	"enum":   ENUM,
	"match":  MATCH,
	"none":     NONE,
	"continue": CONTINUE,
	"defer":    DEFER,
}

// Lookup returns a keyword type or IDENT.
func Lookup(ident string) Type {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return IDENT
}
