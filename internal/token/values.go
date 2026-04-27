package token

// Token is a single lexeme in source order.
type Token struct {
	Type    Type
	Literal string
	Line    int
	Col     int
}
