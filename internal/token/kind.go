// Package token defines Clio lexemes.
package token

// Type is a token/keyword kind.
type Type int

const (
	ILLEGAL Type = iota
	EOF
	NEWLINE
	IDENT
	INTLIT
	FLOATLIT
	STRLIT
	FN
	LET
	CONST
	IF
	ELSE
	WHILE
	FOR
	IN
	LOOP
	BREAK
	RETURN
	STRUCT
	MODULE
	USE
	PUB
	EXTERN
	TRUE
	FALSE
	AS
	RESULT
	CATCH
	OK
	ERR
	ENUM
	MATCH
	NONE
	THIS
	COMMA
	COLON
	SEMI
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACK
	RBRACK
	PLUS
	MINUS
	MUL
	DIV
	MOD
	AND
	OR
	NOT
	ASSIGN
	EQ
	NEQ
	LT
	GT
	LEQ
	GEQ
	ARROW
	DOT
	DOTDOT
	QUEST
	HASH
	PLUSEQ
	MINUSEQ
	MULEQ
	DIVEQ
	MODEQ
	FATARROW
	PLUSPLUS
	MINUSMINUS
	ELLIPSIS
	CONTINUE
	DEFER
)

// String is a short debug name for a few token kinds; many kinds share a generic name.
func (t Type) String() string {
	switch t {
	case EOF:
		return "EOF"
	case NEWLINE:
		return "NEWLINE"
	case IDENT:
		return "IDENT"
	case INTLIT:
		return "INTLIT"
	case FLOATLIT:
		return "FLOATLIT"
	case STRLIT:
		return "STRLIT"
	case ARROW:
		return "->"
	case DOTDOT:
		return ".."
	case HASH:
		return "#"
	case PLUSEQ:
		return "+="
	case MINUSEQ:
		return "-="
	case MULEQ:
		return "*="
	case DIVEQ:
		return "/="
	case MODEQ:
		return "%="
	case FATARROW:
		return "=>"
	default:
		return "TOKEN"
	}
}
