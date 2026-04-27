package ast

// TypeExpr is a (possibly optional) type reference, e.g. "Player", "int", "Player?".
// If PtrInner is set, this type is ptr[PtrInner] (Name is ignored).
// If ResultInner is set, this is result[ResultInner] (Name is "result").
type TypeExpr struct {
	Name         string
	Optional     bool
	PtrInner     *TypeExpr
	ResultInner  *TypeExpr
}

func (t *TypeExpr) String() string {
	if t.ResultInner != nil {
		return "result[" + t.ResultInner.String() + "]"
	}
	if t.PtrInner != nil {
		return "ptr[" + t.PtrInner.String() + "]"
	}
	if t.Optional {
		return t.Name + "?"
	}
	return t.Name
}
