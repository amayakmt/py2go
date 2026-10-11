package pyast

// Node is implemented by every AST node type.
type Node interface {
	Line() int
}

type Module struct {
	Body []Node
}

type FunctionDef struct {
	Name      string
	Body      []Node
	Returns   Node
	HasParams bool
	Lineno    int
}

type Expr struct {
	Value  Node
	Lineno int
}

type Call struct {
	Func        Node
	Args        []Node
	HasKeywords bool
	Lineno      int
}
type Name struct {
	ID     string
	Lineno int
}

type Constant struct {
	Value  any
	Lineno int
}

// Unsupported holds any node type we haven't implemented.
type Unsupported struct {
	Type   string
	Lineno int
}

func (n *Module) Line() int      { return 0 }
func (n *FunctionDef) Line() int { return n.Lineno }
func (n *Expr) Line() int        { return n.Lineno }
func (n *Call) Line() int        { return n.Lineno }
func (n *Name) Line() int        { return n.Lineno }
func (n *Constant) Line() int    { return n.Lineno }
func (n *Unsupported) Line() int { return n.Lineno }
