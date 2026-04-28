package ast

// TypeExpr is a type reference.
// If PtrInner is set, this type is ptr[PtrInner] (Name is ignored).
// If ResultInner is set, this is result[ResultInner] (Name is "result").
// If ListInner is set, this is list[ListInner] (Name is ignored).
type TypeExpr struct {
	Name         string
	Optional     bool
	PtrInner     *TypeExpr
	ResultInner  *TypeExpr
	ListInner    *TypeExpr
	TupleInner   []*TypeExpr
}

func (t *TypeExpr) String() string {
	if t.ResultInner != nil {
		return "result[" + t.ResultInner.String() + "]"
	}
	if t.PtrInner != nil {
		return "ptr[" + t.PtrInner.String() + "]"
	}
	if t.ListInner != nil {
		return "list[" + t.ListInner.String() + "]"
	}
	if t.TupleInner != nil {
		s := "("
		for i, x := range t.TupleInner {
			if i > 0 {
				s += ", "
			}
			s += x.String()
		}
		return s + ")"
	}
	if t.Optional {
		return t.Name + "?"
	}
	return t.Name
}
