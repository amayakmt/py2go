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
	case *pyast.Unsupported:
		return value{}, unsupported(n.Type, n.Line())
	default:
		return value{}, unsupported(nodeName(n), n.Line())
	}
}

// typedExpr generates an expression whose type must be one of the four
// value types, which rules out the result of a call to a None function.
func (g *generator) typedExpr(n pyast.Node) (value, error) {
	v, err := g.expr(n)
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
		return value{code: strconv.FormatInt(n.Value.(int64), 10), typ: tInt, prec: precAtom, konst: true}, nil
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
	callee, ok := n.Func.(*pyast.Name)
	if !ok {
		return value{}, unsupported("call target that is not a plain name", n.Line())
	}
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}

	// A local variable hides the function of the same name.
	if _, isVar := g.env.lookup(callee.ID); isVar {
		return value{}, unsupported(fmt.Sprintf("call of variable '%s'", callee.ID), n.Line())
	}

	switch callee.ID {
	case "print":
		args := make([]string, 0, len(n.Args))
		for _, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return value{}, err
			}
			args = append(args, v.code)
		}
		return value{code: g.rt("Print") + "(" + strings.Join(args, ", ") + ")", typ: tNone, prec: precAtom}, nil
	case "range":
		return value{}, unsupported("range() outside a for loop", n.Line())
	}

	if err := g.sigErrs[callee.ID]; err != nil {
		return value{}, err
	}
	sig, ok := g.sigs[callee.ID]
	if !ok {
		return value{}, unsupported(fmt.Sprintf("call to undefined function '%s'", callee.ID), n.Line())
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
		v, err := g.typedExpr(a)
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
	l, err := g.typedExpr(n.Left)
	if err != nil {
		return value{}, err
	}
	r, err := g.typedExpr(n.Right)
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
	if op == "Pow" || op == "MatMult" {
		return value{}, unsupported("operator "+opSymbols[op], line)
	}

	if l.typ == tStr && r.typ == tStr && op == "Add" {
		return infix(l, "+", r, tStr, precAdd), nil
	}
	if !isNumeric(l.typ) || !isNumeric(r.typ) {
		return bad()
	}
	bothInt := l.typ == tInt && r.typ == tInt

	switch op {
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
		return unary("-", v, v.typ), nil

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
	return v
}

func (g *generator) compare(n *pyast.Compare) (value, error) {
	if len(n.Ops) != 1 {
		return value{}, unsupported("chained comparison", n.Line())
	}
	op := n.Ops[0]
	goOp := map[string]string{"Eq": "==", "NotEq": "!=", "Lt": "<", "LtE": "<=", "Gt": ">", "GtE": ">="}[op]
	if goOp == "" {
		return value{}, unsupported("operator '"+opSymbols[op]+"'", n.Line())
	}

	l, err := g.typedExpr(n.Left)
	if err != nil {
		return value{}, err
	}
	r, err := g.typedExpr(n.Comparators[0])
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
