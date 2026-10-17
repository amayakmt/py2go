package pyast

// Node is implemented by every AST node type.
type Node interface {
	Line() int
}

type Module struct {
	Body []Node
}

// TypeParam is one parameter of PEP 695 syntax: def f[T: int = str]. Kind is
// "TypeVar", "TypeVarTuple" or "ParamSpec".
type TypeParam struct {
	Kind    string
	Name    string
	Bound   Node // a type, or a Tuple of constraints
	Default Node
}

// TypeAlias is "type Name[params] = value".
type TypeAlias struct {
	Name       string
	TypeParams []*TypeParam
	Value      Node
	Lineno     int
}

func (n *TypeAlias) Line() int { return n.Lineno }

type FunctionDef struct {
	TypeParams    []*TypeParam
	Name          string
	Args          *Arguments
	Body          []Node
	Returns       Node
	Decorators    []Node
	HasParams     bool
	HasDecorators bool // len(Decorators) > 0
	HasTypeParams bool // PEP 695: def f[T](x: T)
	IsAsync       bool // async def
	Lineno        int
}

// ClassDef is a class statement. Keywords holds metaclass=... and friends,
// which codegen rejects.
type ClassDef struct {
	TypeParams    []*TypeParam
	Name          string
	Bases         []Node
	Keywords      []*Keyword
	Body          []Node
	Decorators    []Node
	HasTypeParams bool
	Lineno        int

	// Display is the name Python shows for a class the linker moved and
	// renamed (a class defined inside a class); empty otherwise.
	Display string

	// Abstract lists the methods decorated with abc.abstractmethod (the
	// linker removes the decorator), and ABCMeta is set for a class whose
	// metaclass= was abc.ABCMeta (the linker drops the keyword).
	Abstract []string
	ABCMeta  bool
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
	KwDefaults []Node // one entry per keyword-only parameter, nil if it has no default
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
	Keywords    []*Keyword
	HasKeywords bool // len(Keywords) > 0
	Lineno      int
}

// Keyword is "name=value" in a call. Arg is empty for "**mapping".
type Keyword struct {
	Arg    string
	Value  Node
	Lineno int
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
	Target  Node
	Iter    Node
	Body    []Node
	Orelse  []Node
	IsAsync bool // async for
	Lineno  int
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

// Attribute is "value.attr", as in the call "text.upper()".
type Attribute struct {
	Value  Node
	Attr   string
	Lineno int
}

// Subscript is "value[index]". Index is a Slice node for "value[a:b]".
type Subscript struct {
	Value  Node
	Index  Node
	Lineno int
}

// Slice is "lower:upper:step". Any part may be nil.
type Slice struct {
	Lower, Upper, Step Node
	Lineno             int
}

type List struct {
	Elts   []Node
	Lineno int
}

type Tuple struct {
	Elts   []Node
	Lineno int
}

// JoinedStr is an f-string. Values alternate between Constant text and
// FormattedValue nodes.
type JoinedStr struct {
	Values []Node
	Lineno int
}

// FormattedValue is one {replacement} field of an f-string.
// Conversion is -1 for none, or the code of 's', 'r' or 'a'.
// FormatSpec is a JoinedStr, or nil without a spec.
type FormattedValue struct {
	Value      Node
	Conversion int
	FormatSpec Node
	Lineno     int
}

// Try is try/except/else/finally. Handlers, Orelse and Finalbody may be empty.
type Try struct {
	Star      bool // try ... except*
	Body      []Node
	Handlers  []*ExceptHandler
	Orelse    []Node
	Finalbody []Node
	Lineno    int
}

// ExceptHandler is one "except Type as Name:" clause. Type is nil for a
// bare "except:", and Name is empty without "as".
type ExceptHandler struct {
	Type   Node
	Name   string
	Body   []Node
	Lineno int
}

// Raise is "raise Exc from Cause". Both are nil for a bare "raise".
type Raise struct {
	Exc    Node
	Cause  Node
	Lineno int
}

// Assert is "assert Test, Msg". Msg is nil when omitted.
type Assert struct {
	Test   Node
	Msg    Node
	Lineno int
}

// Dict is a dict display. A nil entry in Keys marks "**mapping", whose
// mapping is the matching entry of Values.
type Dict struct {
	Keys   []Node
	Values []Node
	Lineno int
}

type Set struct {
	Elts   []Node
	Lineno int
}

// Comprehension is one "for target in iter if cond ..." clause of a
// comprehension or generator expression.
type Comprehension struct {
	Target  Node
	Iter    Node
	Ifs     []Node
	IsAsync bool
}

type ListComp struct {
	Elt        Node
	Generators []*Comprehension
	Lineno     int
}

type SetComp struct {
	Elt        Node
	Generators []*Comprehension
	Lineno     int
}

type DictComp struct {
	Key        Node
	Value      Node
	Generators []*Comprehension
	Lineno     int
}

type GeneratorExp struct {
	Elt        Node
	Generators []*Comprehension
	Lineno     int
}

type Delete struct {
	Targets []Node
	Lineno  int
}

// Global is "global a, b".
type Global struct {
	Names  []string
	Lineno int
}

// Lambda is "lambda args: body". Its parameters have no annotations.
type Lambda struct {
	Args   *Arguments
	Body   Node
	Lineno int
}

// Starred is "*value", in a call, a display or an assignment target.
type Starred struct {
	Value  Node
	Lineno int
}

// Nonlocal is "nonlocal a, b".
type Nonlocal struct {
	Names  []string
	Lineno int
}

// Alias is one imported name, "name as asname".
type Alias struct {
	Name   string
	AsName string
}

// Import is "import a, b as c".
type Import struct {
	Names  []Alias
	Lineno int
}

// ImportFrom is "from module import a, b as c". Level counts leading dots.
type ImportFrom struct {
	Module string
	Names  []Alias
	Level  int
	Lineno int
}

// IfExp is "body if test else orelse".
type IfExp struct {
	Test, Body, Orelse Node
	Lineno             int
}

// NamedExpr is "target := value".
type NamedExpr struct {
	Target Node
	Value  Node
	Lineno int
}

// WithItem is one "context as vars" of a with statement; Vars may be nil.
type WithItem struct {
	Context Node
	Vars    Node
}

// With is "with a as x, b: body".
type With struct {
	Items   []*WithItem
	Body    []Node
	IsAsync bool // async with
	Lineno  int
}

// Yield is "yield value"; Value may be nil.
type Yield struct {
	Value  Node
	Lineno int
}

// Await is "await value".
type Await struct {
	Value  Node
	Lineno int
}

// YieldFrom is "yield from value".
type YieldFrom struct {
	Value  Node
	Lineno int
}

// Match is "match subject: case ...".
type Match struct {
	Subject Node
	Cases   []*MatchCase
	Lineno  int
}

// MatchCase is one case: a pattern, an optional guard and a body.
type MatchCase struct {
	Pattern Node
	Guard   Node
	Body    []Node
}

// The patterns of a case. Value is a literal or dotted name compared with ==.
type MatchValue struct {
	Value  Node
	Lineno int
}

// MatchSingleton is the pattern None, True or False; Value is "None",
// "True" or "False".
type MatchSingleton struct {
	Value  string
	Lineno int
}

type MatchSequence struct {
	Patterns []Node
	Lineno   int
}

// MatchMapping is {key: pattern, **rest}; Rest may be empty.
type MatchMapping struct {
	Keys     []Node
	Patterns []Node
	Rest     string
	Lineno   int
}

// MatchClass is Cls(p1, p2, attr=p3).
type MatchClass struct {
	Cls         Node
	Patterns    []Node
	KwdAttrs    []string
	KwdPatterns []Node
	Lineno      int
}

// MatchStar is *name in a sequence pattern; Name is empty for *_.
type MatchStar struct {
	Name   string
	Lineno int
}

// MatchAs is "pattern as name", a capture "name", or the wildcard "_"
// (both Pattern and Name empty).
type MatchAs struct {
	Pattern Node
	Name    string
	Lineno  int
}

type MatchOr struct {
	Patterns []Node
	Lineno   int
}

// Unsupported holds any node type we haven't implemented.
type Unsupported struct {
	Type   string
	Lineno int
}

func (n *IfExp) Line() int          { return n.Lineno }
func (n *NamedExpr) Line() int      { return n.Lineno }
func (n *With) Line() int           { return n.Lineno }
func (n *Yield) Line() int          { return n.Lineno }
func (n *Await) Line() int          { return n.Lineno }
func (n *YieldFrom) Line() int      { return n.Lineno }
func (n *Match) Line() int          { return n.Lineno }
func (n *MatchValue) Line() int     { return n.Lineno }
func (n *MatchSingleton) Line() int { return n.Lineno }
func (n *MatchSequence) Line() int  { return n.Lineno }
func (n *MatchMapping) Line() int   { return n.Lineno }
func (n *MatchClass) Line() int     { return n.Lineno }
func (n *MatchStar) Line() int      { return n.Lineno }
func (n *MatchAs) Line() int        { return n.Lineno }
func (n *MatchOr) Line() int        { return n.Lineno }
func (n *Module) Line() int         { return 0 }
func (n *FunctionDef) Line() int    { return n.Lineno }
func (n *ClassDef) Line() int       { return n.Lineno }
func (n *Expr) Line() int           { return n.Lineno }
func (n *Call) Line() int           { return n.Lineno }
func (n *Name) Line() int           { return n.Lineno }
func (n *Constant) Line() int       { return n.Lineno }
func (n *Unsupported) Line() int    { return n.Lineno }
func (n *Arguments) Line() int      { return 0 }
func (n *Arg) Line() int            { return n.Lineno }
func (n *Return) Line() int         { return n.Lineno }
func (n *Assign) Line() int         { return n.Lineno }
func (n *AnnAssign) Line() int      { return n.Lineno }
func (n *AugAssign) Line() int      { return n.Lineno }
func (n *If) Line() int             { return n.Lineno }
func (n *While) Line() int          { return n.Lineno }
func (n *For) Line() int            { return n.Lineno }
func (n *Break) Line() int          { return n.Lineno }
func (n *Continue) Line() int       { return n.Lineno }
func (n *Pass) Line() int           { return n.Lineno }
func (n *BinOp) Line() int          { return n.Lineno }
func (n *UnaryOp) Line() int        { return n.Lineno }
func (n *Compare) Line() int        { return n.Lineno }
func (n *BoolOp) Line() int         { return n.Lineno }
func (n *Keyword) Line() int        { return n.Lineno }
func (n *Attribute) Line() int      { return n.Lineno }
func (n *Subscript) Line() int      { return n.Lineno }
func (n *Slice) Line() int          { return n.Lineno }
func (n *List) Line() int           { return n.Lineno }
func (n *Tuple) Line() int          { return n.Lineno }
func (n *JoinedStr) Line() int      { return n.Lineno }
func (n *FormattedValue) Line() int { return n.Lineno }
func (n *Try) Line() int            { return n.Lineno }
func (n *ExceptHandler) Line() int  { return n.Lineno }
func (n *Raise) Line() int          { return n.Lineno }
func (n *Assert) Line() int         { return n.Lineno }
func (n *Dict) Line() int           { return n.Lineno }
func (n *Set) Line() int            { return n.Lineno }
func (n *Comprehension) Line() int  { return n.Target.Line() }
func (n *ListComp) Line() int       { return n.Lineno }
func (n *SetComp) Line() int        { return n.Lineno }
func (n *DictComp) Line() int       { return n.Lineno }
func (n *GeneratorExp) Line() int   { return n.Lineno }
func (n *Delete) Line() int         { return n.Lineno }
func (n *Global) Line() int         { return n.Lineno }
func (n *Lambda) Line() int         { return n.Lineno }
func (n *Starred) Line() int        { return n.Lineno }
func (n *Nonlocal) Line() int       { return n.Lineno }
func (n *Import) Line() int         { return n.Lineno }
func (n *ImportFrom) Line() int     { return n.Lineno }
