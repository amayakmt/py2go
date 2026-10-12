package codegen

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Go operator precedence, lowest to highest. Parentheses are added only
// where a child binds looser than its parent, since the Python tree is
// already explicit about grouping.
const (
	precOr = iota + 1
	precAnd
	precCmp
	precAdd
	precMul
	precUnary
	precAtom
)

// value is the result of generating one expression: Go source plus the
// static Python type that codegen computed bottom-up.
type value struct {
	code string
	typ  pyType
	prec int
	// konst marks a Go constant expression. Go folds constants with exact
	// arithmetic, so 0.1 + 0.2 would print 0.3 where Python prints
	// 0.30000000000000004. lit marks a bare, untyped float literal; it
	// must be wrapped in float64() before it meets another constant.
	konst bool
	lit   bool
	// intLit marks an integer literal and holds its value. A tuple index
	// and the exponent of ** must be literals, because their result type
	// depends on the value.
	intLit bool
	intVal int64
}

// at returns the code, parenthesized if it binds looser than min.
func (v value) at(min int) string {
	if v.prec < min {
		return "(" + v.code + ")"
	}
	return v.code
}

// typed gives an untyped float literal the float64 type, so Go rounds it
// before any arithmetic instead of folding it exactly.
func (v value) typed() value {
	if v.lit {
		return value{code: "float64(" + v.code + ")", typ: tFloat, prec: precAtom, konst: true}
	}
	return v
}

// operands prepares two values for one Go operator. Literals are typed only
// when both sides are constants, the one case where Go would fold them.
func operands(l, r value) (value, value) {
	if l.konst && r.konst {
		return l.typed(), r.typed()
	}
	return l, r
}

// toFloat converts an int to float64. Go will not promote on its own.
func toFloat(v value) value {
	if v.typ == tInt {
		return value{code: "float64(" + v.code + ")", typ: tFloat, prec: precAtom, konst: v.konst}
	}
	return v
}

func isNumeric(t pyType) bool { return t == tInt || t == tFloat }

func (g *generator) expr(n pyast.Node) (value, error) {
	switch n := n.(type) {
	case *pyast.Constant:
		return g.constant(n)
	case *pyast.Name:
		return g.name(n)
	case *pyast.Call:
		return g.call(n)
	case *pyast.BinOp:
		return g.binOp(n)
	case *pyast.UnaryOp:
		return g.unaryOp(n)
	case *pyast.Compare:
		return g.compare(n)
	case *pyast.BoolOp:
		return g.boolOp(n)
	case *pyast.List:
		return g.listLit(n, "")
	case *pyast.Tuple:
		return g.tupleLit(n, "")
	case *pyast.Subscript:
		return g.subscript(n)
	case *pyast.JoinedStr:
		return g.fstring(n)
	case *pyast.Attribute:
		return value{}, unsupported("attribute access '."+n.Attr+"'", n.Line())
	case *pyast.Unsupported:
		return value{}, unsupported(n.Type, n.Line())
	default:
		return value{}, unsupported(nodeName(n), n.Line())
	}
}

// typedExpr generates an expression whose type must be a value type, which
// rules out the result of a call to a None function.
func (g *generator) typedExpr(n pyast.Node) (value, error) { return g.exprAs(n, "") }

// exprAs is typedExpr for a place that needs a known type, such as the
// right side of "xs: list[int] = ...". The expectation matters only for
// literals that carry no type of their own: [] is a list of anything.
func (g *generator) exprAs(n pyast.Node, want pyType) (value, error) {
	var v value
	var err error
	switch n := n.(type) {
	case *pyast.List:
		v, err = g.listLit(n, want)
	case *pyast.Tuple:
		v, err = g.tupleLit(n, want)
	default:
		v, err = g.expr(n)
	}
	if err != nil {
		return v, err
	}
	if v.typ == tNone {
		return v, unsupported("use of a None value", n.Line())
	}
	return v, nil
}

func (g *generator) constant(n *pyast.Constant) (value, error) {
	switch n.PyType {
	case "int":
		v := n.Value.(int64)
		return value{code: strconv.FormatInt(v, 10), typ: tInt, prec: precAtom, konst: true, intLit: true, intVal: v}, nil
	case "float":
		f := n.Value.(float64)
		if math.IsInf(f, 1) {
			return value{code: g.rt("Inf"), typ: tFloat, prec: precAtom}, nil
		}
		return value{code: floatLiteral(f), typ: tFloat, prec: precAtom, konst: true, lit: true}, nil
	case "bool":
		return value{code: strconv.FormatBool(n.Value.(bool)), typ: tBool, prec: precAtom, konst: true}, nil
	case "str":
		return value{code: strconv.Quote(n.Value.(string)), typ: tStr, prec: precAtom, konst: true}, nil
	}
	return value{}, unsupported("use of "+constantText(n), n.Line())
}

// floatLiteral spells a float so Go reads it as a float, not an integer:
// a bare 3 would be an int in print(3.0).
func floatLiteral(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".e") {
		s += ".0"
	}
	return s
}

func (g *generator) name(n *pyast.Name) (value, error) {
	if t, ok := g.env.lookup(n.ID); ok {
		return value{code: n.ID, typ: t, prec: precAtom}, nil
	}
	if _, isFunc := g.sigs[n.ID]; isFunc {
		return value{}, unsupported(fmt.Sprintf("function '%s' used as a value", n.ID), n.Line())
	}
	if g.fnName != "" {
		if _, ok := g.mainEnv.lookup(n.ID); ok {
			return value{}, unsupported(fmt.Sprintf("module-level variable '%s' used inside a function", n.ID), n.Line())
		}
	}
	return value{}, unsupported(fmt.Sprintf("undefined name '%s'", n.ID), n.Line())
}

func (g *generator) call(n *pyast.Call) (value, error) {
	switch fn := n.Func.(type) {
	case *pyast.Name:
		return g.namedCall(n, fn)
	case *pyast.Attribute:
		return g.methodCall(n, fn)
	}
	return value{}, unsupported("call target that is not a plain name", n.Line())
}

func (g *generator) namedCall(n *pyast.Call, callee *pyast.Name) (value, error) {
	// A local variable hides the function of the same name.
	if _, isVar := g.env.lookup(callee.ID); isVar {
		return value{}, unsupported(fmt.Sprintf("call of variable '%s'", callee.ID), n.Line())
	}
	// A function the program defines hides a builtin of the same name.
	if err := g.sigErrs[callee.ID]; err != nil {
		return value{}, err
	}
	if sig, ok := g.sigs[callee.ID]; ok {
		return g.userCall(n, callee, sig)
	}
	return g.builtinCall(n, callee)
}

func (g *generator) userCall(n *pyast.Call, callee *pyast.Name, sig *funcSig) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	// At module level, Python runs top to bottom, so the def must come first.
	if g.fnName == "" && !g.defined[callee.ID] {
		return value{}, unsupported(fmt.Sprintf("call to '%s' before its definition", callee.ID), n.Line())
	}
	if len(n.Args) != len(sig.params) {
		return value{}, unsupported(fmt.Sprintf("call to '%s' with %d arguments, expected %d",
			callee.ID, len(n.Args), len(sig.params)), n.Line())
	}

	args := make([]string, 0, len(n.Args))
	for i, a := range n.Args {
		v, err := g.exprAs(a, sig.params[i])
		if err != nil {
			return value{}, err
		}
		if v.typ != sig.params[i] {
			return value{}, mismatch(fmt.Sprintf("argument %d of '%s'", i+1, callee.ID), sig.params[i], v.typ, a.Line())
		}
		args = append(args, v.code)
	}
	return value{code: callee.ID + "(" + strings.Join(args, ", ") + ")", typ: sig.ret, prec: precAtom}, nil
}

// opSymbols spells operators for error messages.
var opSymbols = map[string]string{
	"Add": "+", "Sub": "-", "Mult": "*", "Div": "/", "FloorDiv": "//", "Mod": "%",
	"Pow": "**", "MatMult": "@", "LShift": "<<", "RShift": ">>",
	"BitOr": "|", "BitXor": "^", "BitAnd": "&",
	"UAdd": "+", "USub": "-", "Not": "not", "Invert": "~",
	"Eq": "==", "NotEq": "!=", "Lt": "<", "LtE": "<=", "Gt": ">", "GtE": ">=",
	"Is": "is", "IsNot": "is not", "In": "in", "NotIn": "not in",
}

func (g *generator) binOp(n *pyast.BinOp) (value, error) {
	l, r, err := g.operandPair(n.Left, n.Right)
	if err != nil {
		return value{}, err
	}
	return g.binary(n.Op, l, r, n.Line())
}

// binary applies a binary operator to two generated operands. It decides
// the Go operator, inserts int to float conversions, and routes the
// operators whose Python and Go meanings differ through the runtime.
func (g *generator) binary(op string, l, r value, line int) (value, error) {
	bad := func() (value, error) {
		return value{}, unsupported(fmt.Sprintf("operator %s on %s and %s", opSymbols[op], l.typ, r.typ), line)
	}
	if op == "MatMult" {
		return value{}, unsupported("operator "+opSymbols[op], line)
	}

	// str and list operands: concatenation and repetition.
	if l.typ == tStr || r.typ == tStr || l.typ.isList() || r.typ.isList() {
		if v, ok := g.sequenceBinary(op, l, r); ok {
			return v, nil
		}
		return bad()
	}
	if !isNumeric(l.typ) || !isNumeric(r.typ) {
		return bad()
	}
	bothInt := l.typ == tInt && r.typ == tInt

	switch op {
	case "Pow":
		return g.power(l, r, line)

	case "Add", "Sub", "Mult":
		goOp, prec := map[string]string{"Add": "+", "Sub": "-", "Mult": "*"}[op], precAdd
		if op == "Mult" {
			prec = precMul
		}
		if bothInt {
			return infix(l, goOp, r, tInt, prec), nil
		}
		l, r = operands(toFloat(l), toFloat(r))
		return infix(l, goOp, r, tFloat, prec), nil

	case "Div":
		// "/" always yields a float, so both sides convert first.
		if bothInt {
			return value{code: g.rt("DivInt") + "(" + l.code + ", " + r.code + ")", typ: tFloat, prec: precAtom}, nil
		}
		l, r = operands(toFloat(l), toFloat(r))
		return value{code: g.rt("Div") + "(" + l.code + ", " + r.code + ")", typ: tFloat, prec: precAtom}, nil

	case "FloorDiv", "Mod":
		helper := map[string]string{"FloorDiv": "FloorDiv", "Mod": "Mod"}[op]
		if bothInt {
			return value{code: g.rt(helper) + "(" + l.code + ", " + r.code + ")", typ: tInt, prec: precAtom}, nil
		}
		l, r = operands(toFloat(l), toFloat(r))
		return value{code: g.rt(helper+"Float") + "(" + l.code + ", " + r.code + ")", typ: tFloat, prec: precAtom}, nil

	case "BitAnd", "BitOr", "BitXor", "LShift", "RShift":
		if !bothInt {
			return bad()
		}
		switch op {
		case "LShift", "RShift":
			return value{code: g.rt(op) + "(" + l.code + ", " + r.code + ")", typ: tInt, prec: precAtom}, nil
		case "BitAnd":
			return infix(l, "&", r, tInt, precMul), nil
		case "BitOr":
			return infix(l, "|", r, tInt, precAdd), nil
		default:
			return infix(l, "^", r, tInt, precAdd), nil
		}
	}
	return value{}, unsupported("operator "+op, line)
}

// sequenceBinary handles the operators that accept a str or list:
// "+" joins two of the same type, "*" repeats one by an int.
func (g *generator) sequenceBinary(op string, l, r value) (value, bool) {
	same := l.typ == r.typ
	switch {
	case op == "Add" && same && l.typ == tStr:
		return infix(l, "+", r, tStr, precAdd), true
	case op == "Add" && same && l.typ.isList():
		return value{code: l.at(precAtom) + ".Concat(" + r.code + ")", typ: l.typ, prec: precAtom}, true
	case op == "Mult" && l.typ == tStr && r.typ == tInt:
		return value{code: g.rt("StrRepeat") + "(" + l.code + ", " + r.code + ")", typ: tStr, prec: precAtom}, true
	case op == "Mult" && l.typ == tInt && r.typ == tStr:
		return value{code: g.rt("StrRepeat") + "(" + r.code + ", " + l.code + ")", typ: tStr, prec: precAtom}, true
	case op == "Mult" && l.typ.isList() && r.typ == tInt:
		return value{code: l.at(precAtom) + ".Repeat(" + r.code + ")", typ: l.typ, prec: precAtom}, true
	case op == "Mult" && l.typ == tInt && r.typ.isList():
		return value{code: r.at(precAtom) + ".Repeat(" + l.code + ")", typ: r.typ, prec: precAtom}, true
	}
	return value{}, false
}

// power is **. With int operands the result is an int for a non-negative
// exponent and a float for a negative one, so the exponent has to be a
// literal for the result type to be known. Any float operand gives a float.
func (g *generator) power(l, r value, line int) (value, error) {
	if l.typ == tInt && r.typ == tInt {
		switch {
		case !r.intLit:
			return value{}, unsupported("operator ** with a variable exponent on int operands", line)
		case r.intVal >= 0:
			return value{code: g.rt("PowInt") + "(" + l.code + ", " + r.code + ")", typ: tInt, prec: precAtom}, nil
		}
	}
	// The operands are function arguments, so Go does not fold them.
	l, r = toFloat(l), toFloat(r)
	return value{code: g.rt("PowFloat") + "(" + l.code + ", " + r.code + ")", typ: tFloat, prec: precAtom}, nil
}

// infix joins two operands with a Go binary operator of the given precedence.
// The left operand may share the precedence (Go is left associative),
// the right one may not, so that Python's explicit grouping is preserved.
func infix(l value, op string, r value, t pyType, prec int) value {
	return value{
		code:  l.at(prec) + " " + op + " " + r.at(prec+1),
		typ:   t,
		prec:  prec,
		konst: l.konst && r.konst,
	}
}

func (g *generator) unaryOp(n *pyast.UnaryOp) (value, error) {
	v, err := g.typedExpr(n.Operand)
	if err != nil {
		return value{}, err
	}
	bad := func() (value, error) {
		return value{}, unsupported(fmt.Sprintf("operator %s on %s", opSymbols[n.Op], v.typ), n.Line())
	}

	switch n.Op {
	case "Not":
		// "not x" is a bool for any x, so truthiness applies to every type.
		t := truth(v)
		return value{code: "!" + t.at(precUnary), typ: tBool, prec: precUnary, konst: v.konst}, nil

	case "UAdd":
		if !isNumeric(v.typ) {
			return bad()
		}
		return v, nil

	case "USub":
		if !isNumeric(v.typ) {
			return bad()
		}
		// Go constants have no negative zero, so -0.0 would print as 0.0.
		if c, ok := n.Operand.(*pyast.Constant); ok && c.PyType == "float" && c.Value.(float64) == 0 {
			return value{code: g.rt("NegZero"), typ: tFloat, prec: precAtom}, nil
		}
		neg := unary("-", v, v.typ)
		neg.intLit, neg.intVal = v.intLit, -v.intVal
		return neg, nil

	case "Invert":
		if v.typ != tInt {
			return bad()
		}
		return unary("^", v, tInt), nil
	}
	return value{}, unsupported("operator "+n.Op, n.Line())
}

// unary prefixes an operator. Negating a literal keeps it a literal.
func unary(op string, v value, t pyType) value {
	operand := v.at(precUnary)
	// "- -x" must not collapse into the decrement token.
	if v.prec == precUnary && op == "-" {
		operand = "(" + v.code + ")"
	}
	return value{code: op + operand, typ: t, prec: precUnary, konst: v.konst, lit: v.lit && op == "-"}
}

// truth converts a value of any type to a Go bool the way Python's
// truthiness does: zero and the empty string are false.
func truth(v value) value {
	switch v.typ {
	case tInt, tFloat:
		return value{code: v.at(precCmp+1) + " != 0", typ: tBool, prec: precCmp, konst: v.konst}
	case tStr:
		return value{code: v.at(precCmp+1) + ` != ""`, typ: tBool, prec: precCmp, konst: v.konst}
	}
	if v.typ.isList() || v.typ.isTuple() {
		return value{code: v.at(precAtom) + ".Len() != 0", typ: tBool, prec: precCmp}
	}
	return v
}

// operandPair generates the two sides of a binary operator. A list or
// tuple literal on one side takes its type from the other, so that
// "xs == []" and "xs + [1]" work without annotations.
func (g *generator) operandPair(left, right pyast.Node) (l, r value, err error) {
	isLiteral := func(n pyast.Node) bool {
		switch n.(type) {
		case *pyast.List, *pyast.Tuple:
			return true
		}
		return false
	}
	if isLiteral(left) && !isLiteral(right) {
		if r, err = g.typedExpr(right); err != nil {
			return
		}
		l, err = g.exprAs(left, r.typ)
		return
	}
	if l, err = g.typedExpr(left); err != nil {
		return
	}
	r, err = g.exprAs(right, l.typ)
	return
}

func (g *generator) compare(n *pyast.Compare) (value, error) {
	if len(n.Ops) != 1 {
		return value{}, unsupported("chained comparison", n.Line())
	}
	op := n.Ops[0]
	if op == "In" || op == "NotIn" {
		return g.contains(n, op == "NotIn")
	}
	goOp := map[string]string{"Eq": "==", "NotEq": "!=", "Lt": "<", "LtE": "<=", "Gt": ">", "GtE": ">="}[op]
	if goOp == "" {
		return value{}, unsupported("operator '"+opSymbols[op]+"'", n.Line())
	}

	l, r, err := g.operandPair(n.Left, n.Comparators[0])
	if err != nil {
		return value{}, err
	}

	// An int compared with a float converts to float, as in Python.
	// Anything else must match exactly: Python allows 1 == "a" or True < 2,
	// but those are almost always mistakes.
	switch {
	case isNumeric(l.typ) && isNumeric(r.typ) && l.typ != r.typ:
		l, r = toFloat(l), toFloat(r)
	case l.typ != r.typ:
		return value{}, unsupported(fmt.Sprintf("comparison of %s with %s", l.typ, r.typ), n.Line())
	case l.typ == tBool && op != "Eq" && op != "NotEq":
		return value{}, unsupported("ordering comparison of bool values", n.Line())
	case l.typ.isList() || l.typ.isTuple():
		// Containers compare element by element in the runtime.
		fn := map[string]string{"Eq": "Eq", "NotEq": "Eq", "Lt": "Lt", "LtE": "Le", "Gt": "Gt", "GtE": "Ge"}[op]
		code := g.rt(fn) + "(" + l.code + ", " + r.code + ")"
		if op == "NotEq" {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, nil
	}
	l, r = operands(l, r)

	// Comparison operators do not chain in Go the way they would read,
	// so a comparison nested in a comparison always gets parentheses.
	return value{
		code:  l.at(precCmp+1) + " " + goOp + " " + r.at(precCmp+1),
		typ:   tBool,
		prec:  precCmp,
		konst: l.konst && r.konst,
	}, nil
}

// contains is "x in container" and "x not in container" for str, list
// and tuple containers.
func (g *generator) contains(n *pyast.Compare, negate bool) (value, error) {
	r, err := g.typedExpr(n.Comparators[0])
	if err != nil {
		return value{}, err
	}
	var want pyType
	if r.typ.isList() {
		want = r.typ.elem()
	}
	l, err := g.exprAs(n.Left, want)
	if err != nil {
		return value{}, err
	}

	var code string
	switch {
	case r.typ == tStr && l.typ == tStr:
		code = g.rt("StrContains") + "(" + r.code + ", " + l.code + ")"
	case r.typ.isList() && l.typ == r.typ.elem():
		code = r.at(precAtom) + ".Contains(" + l.code + ")"
	case r.typ.isTuple() && allEqual(r.typ.elems(), l.typ):
		// The item is boxed as an interface, so give a constant its Go type.
		code = g.rt("SeqContains") + "(" + r.code + ", " + l.typ.goType() + "(" + l.code + "))"
	default:
		op := "in"
		if negate {
			op = "not in"
		}
		return value{}, unsupported(fmt.Sprintf("operator '%s' on %s and %s", op, l.typ, r.typ), n.Line())
	}
	if negate {
		return value{code: "!" + code, typ: tBool, prec: precUnary}, nil
	}
	return value{code: code, typ: tBool, prec: precAtom}, nil
}

func allEqual(types []pyType, t pyType) bool {
	for _, x := range types {
		if x != t {
			return false
		}
	}
	return true
}

func (g *generator) boolOp(n *pyast.BoolOp) (value, error) {
	goOp, prec := "&&", precAnd
	if n.Op == "Or" {
		goOp, prec = "||", precOr
	}
	word := strings.ToLower(n.Op)

	parts := make([]string, 0, len(n.Values))
	konst := true
	for _, e := range n.Values {
		v, err := g.typedExpr(e)
		if err != nil {
			return value{}, err
		}
		// Python's and/or return one of their operands, Go's return a bool.
		// The two agree only when every operand is a bool.
		if v.typ != tBool {
			return value{}, unsupported(fmt.Sprintf("'%s' on %s operand (only bool is supported)", word, v.typ), e.Line())
		}
		konst = konst && v.konst
		// Same-operator children can share the parent's precedence.
		parts = append(parts, v.at(prec))
	}
	return value{code: strings.Join(parts, " "+goOp+" "), typ: tBool, prec: prec, konst: konst}, nil
}
