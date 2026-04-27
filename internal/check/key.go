package check

import (
	"clio/internal/ast"
	"reflect"
)

func exprKey(e ast.Expr) uintptr {
	return ExprKey(e)
}

// ExprKey returns a stable key for the concrete expr node; must match the map used in Check/Info
func ExprKey(e ast.Expr) uintptr {
	if e == nil {
		return 0
	}
	v := reflect.ValueOf(e)
	if v.Kind() != reflect.Ptr {
		return 0
	}
	return v.Pointer()
}
