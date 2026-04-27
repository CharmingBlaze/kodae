package ast

// BlockStmt is both a container and a bare `{ ... }` statement.
type BlockStmt struct{ Stmts []Stmt }

func (s *BlockStmt) stmt() {}

// IfStmt: if (cond) { th } [ else { el } or else if ... ]
type IfStmt struct {
	Cond Expr
	Thn  *BlockStmt
	Els  Stmt // *BlockStmt or *IfStmt for `else if`
}

func (s *IfStmt) stmt() {}

type WhileStmt struct{ Cond Expr; Body *BlockStmt }

func (s *WhileStmt) stmt() {}

type LoopStmt struct{ Body *BlockStmt }

func (s *LoopStmt) stmt() {}

// ForInStmt: for (i in expr) { }
type ForInStmt struct{ Var string; In Expr; Body *BlockStmt }

func (s *ForInStmt) stmt() {}

type ReturnStmt struct{ V Expr } // V may be nil

func (s *ReturnStmt) stmt() {}

type BreakStmt struct{}

func (s *BreakStmt) stmt() {}

type ContinueStmt struct{}

func (s *ContinueStmt) stmt() {}

// DeferStmt: expression runs at function exit (emitted in reverse with other defers) — only at function top level.
type DeferStmt struct{ E Expr }

func (s *DeferStmt) stmt() {}

// LetStmt / const inside blocks
type LetStmt struct{ Const bool; Name string; T *TypeExpr; Init Expr }

func (s *LetStmt) stmt() {}

type ExprStmt struct{ E Expr }

func (s *ExprStmt) stmt() {}

// AssignStmt: =, +=, etc.
type AssignStmt struct{ Left Expr; Op string; Right Expr }

func (s *AssignStmt) stmt() {}

// MatchStmt
type MatchStmt struct {
	Scrutinee Expr
	Arms      []MatchArm
}

func (s *MatchStmt) stmt() {}

// MatchArm: pattern and body
type MatchArm struct {
	Pat  Expr
	Body *BlockStmt
}
