package bindgen

type ClangASTNode struct {
	ID                  string         `json:"id"`
	Kind                string         `json:"kind"`
	Name                string         `json:"name"`
	Type                *ClangASTType  `json:"type"`
	Inner               []ClangASTNode `json:"inner"`
	CompleteDefinition  bool           `json:"completeDefinition"`
	TagUsed             string         `json:"tagUsed"`
	IsImplicit          bool           `json:"isImplicit"`
}

type ClangASTType struct {
	QualType string `json:"qualType"`
}

func (n *ClangASTNode) FindNodes(kind string) []ClangASTNode {
	var res []ClangASTNode
	if n.Kind == kind {
		res = append(res, *n)
	}
	for _, child := range n.Inner {
		res = append(res, child.FindNodes(kind)...)
	}
	return res
}

type GenerateResult struct {
	Content string
	Structs int
	Externs int
	Skipped int
}
