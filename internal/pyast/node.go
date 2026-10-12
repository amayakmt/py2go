package pyast

// Node is implemented by every AST node type.
type Node interface {
	Line() int
}

type Module struct {
	Body []Node
}

type FunctionDef struct {
	Name          string
	Args          *Arguments
	Body          []Node
	Returns       Node
	HasParams     bool
	HasDecorators bool
	HasTypeParams bool // PEP 695: def f[T](x: T)
	Lineno        int
}

// Arguments is a function signature. Codegen accepts only plain positional
// parameters, but the rest is kept so it can reject with a line number.
type Arguments struct {
	PosOnly    []*Arg
	Args       []*Arg
	Vararg     *Arg
	Kwonly     []*Arg
	Kwarg      *Arg
	Defaults   []Node // defaults of positional parameters
	KwDefaults []Node // defaults of keyword-only parameters, nil entries dropped
}

type Arg struct {
	Name       string
	Annotation Node // nil when the parameter has no annotation
	Lineno     int
}

// Positional returns posonly and regular parameters in call order.
func (a *Arguments) Positional() []*Arg {
	return append(append([]*Arg{}, a.PosOnly...), a.Args...)
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

// Constant is a literal. PyType is one of "int", "float", "str", "bool",
// "none". Value holds int64, float64, string, bool, or nil to match.
type Constant struct {
	PyType string
	Value  any
	Lineno int
}

type Return struct {
	Value  Node // nil for a bare return
	Lineno int
}

// Assign is "a = b". Targets has more than one entry for "a = b = c".
type Assign struct {
	Targets []Node
	Value   Node
	Lineno  int
}

// AnnAssign is "x: int = 5". Value is nil for a bare "x: int".
// Simple is true when the target is a plain name, not "(x): int" or "a.b: int".
type AnnAssign struct {
	Target     Node
	Annotation Node
	Value      Node
	Simple     bool
	Lineno     int
}

// AugAssign is "x += 1". Op is the operator name: "Add", "Sub", ...
type AugAssign struct {
	Target Node
	Op     string
	Value  Node
	Lineno int
}

// If covers elif too: an elif is an If alone in Orelse.
type If struct {
	Test   Node
	Body   []Node
	Orelse []Node
	Lineno int
}

type While struct {
	Test   Node
	Body   []Node
	Orelse []Node
	Lineno int
}

type For struct {
	Target Node
	Iter   Node
	Body   []Node
	Orelse []Node
	Lineno int
}

type Break struct{ Lineno int }
type Continue struct{ Lineno int }
type Pass struct{ Lineno int }

// Operators are nodes in Python's AST. Here they decode to their class
// name: "Add", "Sub", "Mult", "Div", "FloorDiv", "Mod", "Pow", "LShift",
// "RShift", "BitOr", "BitXor", "BitAnd", "MatMult" for BinOp;
// "UAdd", "USub", "Not", "Invert" for UnaryOp; "And", "Or" for BoolOp;
// "Eq", "NotEq", "Lt", "LtE", "Gt", "GtE", "Is", "IsNot", "In", "NotIn"
// for Compare.
type BinOp struct {
	Left   Node
	Op     string
	Right  Node
	Lineno int
}

type UnaryOp struct {
	Op      string
	Operand Node
	Lineno  int
}

// Compare is "a < b < c": Ops[i] sits between operand i and Comparators[i].
type Compare struct {
	Left        Node
	Ops         []string
	Comparators []Node
	Lineno      int
}

type BoolOp struct {
	Op     string
	Values []Node
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
func (n *Arguments) Line() int   { return 0 }
func (n *Arg) Line() int         { return n.Lineno }
func (n *Return) Line() int      { return n.Lineno }
func (n *Assign) Line() int      { return n.Lineno }
func (n *AnnAssign) Line() int   { return n.Lineno }
func (n *AugAssign) Line() int   { return n.Lineno }
func (n *If) Line() int          { return n.Lineno }
func (n *While) Line() int       { return n.Lineno }
func (n *For) Line() int         { return n.Lineno }
func (n *Break) Line() int       { return n.Lineno }
func (n *Continue) Line() int    { return n.Lineno }
func (n *Pass) Line() int        { return n.Lineno }
func (n *BinOp) Line() int       { return n.Lineno }
func (n *UnaryOp) Line() int     { return n.Lineno }
func (n *Compare) Line() int     { return n.Lineno }
func (n *BoolOp) Line() int      { return n.Lineno }
