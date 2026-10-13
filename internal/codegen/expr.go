package codegen

import (
	"fmt"
	"github.com/amayakmt/py2go/runtime"
	"math"
	"math/big"
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
	// fsig is the def behind a function used as a value. It lets a
	// function with defaults go where a Callable with fewer parameters
	// is expected.
	fsig *funcSig
	// intLit marks an integer literal and holds its value. A tuple index
	// and the exponent of ** must be literals, because their result type
	// depends on the value.
	intLit bool
	intVal int64

	// classRef is set for a class used as a value (a class name, cls, or
	// type(x)). code is empty unless the value is a variable, and
	// dynamicRef marks type(x) of an instance whose exact class is not known.
	classRef   *classInfo
	exactRef   bool
	dynamicRef bool

	// fresh marks a container that was just built by a display or a
	// comprehension, which nothing else refers to yet. varName is the
	// variable a value was read from.
	fresh   bool
	varName string
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

// guardConst keeps Go from folding a float operation on two constants that
// overflows float64, which Go rejects where Python has inf: the right side
// becomes a run-time value.
func guardConst(op string, l, r value) value {
	lf, ok1 := floatConstOf(l)
	rf, ok2 := floatConstOf(r)
	if !ok1 || !ok2 {
		return r
	}
	var res float64
	switch op {
	case "+":
		res = lf + rf
	case "-":
		res = lf - rf
	case "*":
		res = lf * rf
	case "/":
		if rf == 0 {
			return r
		}
		res = lf / rf
	default:
		return r
	}
	if math.IsInf(res, 0) {
		return value{code: "func() float64 { return " + r.code + " }()", typ: tFloat, prec: precAtom}
	}
	return r
}

// floatConstOf reads the value of float64(literal).
func floatConstOf(v value) (float64, bool) {
	if !v.konst || v.typ != tFloat || !strings.HasPrefix(v.code, "float64(") || !strings.HasSuffix(v.code, ")") {
		return 0, false
	}
	f, err := strconv.ParseFloat(v.code[len("float64("):len(v.code)-1], 64)
	return f, err == nil
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
	case *pyast.Dict:
		return g.dictLit(n, "")
	case *pyast.Set:
		return g.setLit(n, "")
	case *pyast.ListComp:
		return g.listComp(n, "")
	case *pyast.SetComp:
		return g.setComp(n, "")
	case *pyast.DictComp:
		return g.dictComp(n, "")
	case *pyast.GeneratorExp:
		return g.genExp(n)
	case *pyast.Lambda:
		return g.dynLambda(n)
	case *pyast.IfExp:
		return g.ifExp(n, "")
	case *pyast.Yield:
		return g.yieldExpr(n)
	case *pyast.YieldFrom:
		return g.yieldFrom(n)
	case *pyast.NamedExpr:
		return g.namedExpr(n)
	case *pyast.Subscript:
		return g.subscript(n)
	case *pyast.JoinedStr:
		return g.fstring(n)
	case *pyast.Attribute:
		return g.attribute(n)
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
	keepObject := g.keepObject
	g.keepObject = false
	target := want
	if want.isOptional() {
		// A display for an Optional[list[int]] is a list[int].
		want = want.optElem()
	}
	switch n := n.(type) {
	case *pyast.List:
		v, err = g.listLit(n, want)
	case *pyast.Tuple:
		v, err = g.tupleLit(n, want)
	case *pyast.Dict:
		v, err = g.dictLit(n, want)
	case *pyast.Set:
		v, err = g.setLit(n, want)
	case *pyast.ListComp:
		v, err = g.listComp(n, want)
	case *pyast.SetComp:
		v, err = g.setComp(n, want)
	case *pyast.DictComp:
		v, err = g.dictComp(n, want)
	case *pyast.Name:
		if lam := g.builtinAsLambda(n, want); lam != nil {
			v, err = g.lambda(lam, want.funcParams(), "")
		} else {
			v, err = g.expr(n)
		}
	case *pyast.Lambda:
		if !want.isFunc() {
			return g.dynLambda(n)
		}
		v, err = g.lambda(n, want.funcParams(), want.funcRet())
	case *pyast.IfExp:
		v, err = g.ifExp(n, want)
	case *pyast.Call:
		if ev, ok := g.emptyConstructor(n, want); ok {
			return ev, nil
		}
		saved := g.callWant
		g.callWant = target
		v, err = g.expr(n)
		g.callWant = saved
	default:
		v, err = g.expr(n)
	}
	if err != nil {
		return v, err
	}
	if v.typ.isClassRef() && v.classRef != nil {
		if want.isFunc() {
			if v, err = g.constructorValue(v, n.Line()); err != nil {
				return v, err
			}
		} else {
			v = g.classObject(v)
		}
	}
	switch n.(type) {
	case *pyast.List, *pyast.Dict, *pyast.Set, *pyast.ListComp, *pyast.SetComp, *pyast.DictComp:
		v.fresh = true
	}
	if target == tObject {
		v = g.escape(v)
	}
	if target != "" && !(keepObject && v.typ == tObject) {
		v = g.relabel(v, target)
	}
	if v.fsig != nil && v.typ.isFunc() && !target.isFunc() && (v.fsig.hasDefaults() || !v.fsig.plain()) {
		// A function with defaults or keywords, kept in a variable, is
		// called with whatever the call gives; its parameters travel with it.
		v = g.relabel(v, tObject)
	}
	if v.typ == tNone {
		// A call that returns nothing, used as a value, is None once it ran.
		v = value{code: "func() " + g.rt("NoneT") + " {\n" + v.code + "\nreturn " + g.rt("NoneV") + "\n}()", typ: tNoneType, prec: precAtom}
	}
	return v, nil
}

func (g *generator) constant(n *pyast.Constant) (value, error) {
	switch n.PyType {
	case "int":
		if b, ok := n.Value.(*big.Int); ok {
			return value{code: g.rt("BigLit") + "(" + strconv.Quote(b.String()) + ")", typ: tObject, prec: precAtom}, nil
		}
		v := n.Value.(int64)
		return value{code: strconv.FormatInt(v, 10), typ: tInt, prec: precAtom, konst: true, intLit: true, intVal: v}, nil
	case "bytes":
		return value{code: g.rt("Bytes") + "(" + strconv.Quote(n.Value.(string)) + ")", typ: tObject, prec: precAtom}, nil
	case "complex":
		z := n.Value.(complex128)
		return value{code: fmt.Sprintf("complex128(complex(%s, %s))", floatLiteral(real(z)), floatLiteral(imag(z))), typ: tObject, prec: precAtom}, nil
	case "ellipsis":
		return value{code: g.rt("Ellipsis"), typ: tObject, prec: precAtom}, nil
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
	case "none":
		return value{code: g.rt("NoneV"), typ: tNoneType, prec: precAtom}, nil
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

// isVariable reports whether name is a variable in the current scopes or,
// inside a function, a module variable.
func (g *generator) isVariable(name string) bool {
	if _, ok := g.env.lookup(name); ok {
		return true
	}
	if g.fnName != "" {
		_, ok := g.globalTypes[name]
		return ok
	}
	return false
}

func (g *generator) name(n *pyast.Name) (value, error) {
	if t, ok := g.env.lookup(n.ID); ok {
		if t.isClassRef() {
			return g.classRefValue(n.ID, t), nil
		}
		if a, ok := g.env.lookupAlias(n.ID); ok {
			return value{code: a.code, typ: a.typ, prec: precAtom}, nil
		}
		sig, _ := g.env.lookupSig(n.ID)
		if sig != nil && len(sig.decorators) > 0 {
			sig = nil
		}
		return value{code: n.ID, typ: t, prec: precAtom, fsig: sig, varName: n.ID}, nil
	}
	if v, ok := g.readGlobal(n.ID); ok {
		v.varName = n.ID
		return v, nil
	}
	if v, ok := g.classNameValue(n.ID); ok {
		return v, nil
	}
	if sig, ok := g.sigs[n.ID]; ok && len(sig.decorators) == 0 {
		return g.functionValue(sig)
	}
	if n.ID == "__name__" {
		return value{code: `"__main__"`, typ: tStr, prec: precAtom}, nil
	}
	switch n.ID {
	case "Ellipsis":
		return value{code: g.rt("Ellipsis"), typ: tObject, prec: precAtom}, nil
	case "NotImplemented":
		return value{code: g.rt("NotImplementedV"), typ: tObject, prec: precAtom}, nil
	}
	if dynBuiltinValues[n.ID] {
		return value{code: g.rt("BuiltinFunc") + "(" + strconv.Quote(n.ID) + ")", typ: tObject, prec: precAtom}, nil
	}
	if dynBuiltinTypes[n.ID] || runtime.IsExceptionClass(n.ID) {
		return value{code: g.rt("BuiltinClassOf") + "(" + strconv.Quote(n.ID) + ")", typ: tObject, prec: precAtom}, nil
	}
	if !g.infer && !pythonBuiltins[n.ID] {
		// A name nothing defines is a NameError when the program gets there.
		return value{code: g.rt("UndefinedName") + "(" + strconv.Quote(n.ID) + ")", typ: tObject, prec: precAtom}, nil
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
	// A lambda called on the spot takes its parameter types from the arguments.
	if lam, ok := n.Func.(*pyast.Lambda); ok {
		if n.HasKeywords {
			return value{}, unsupported("keyword arguments in a call of a lambda", n.Line())
		}
		types := make([]pyType, len(n.Args))
		for i, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return value{}, err
			}
			types[i] = v.typ
		}
		fv, err := g.lambda(lam, types, "")
		if err != nil {
			return value{}, err
		}
		return g.callValue(n, fv)
	}
	// Box[int](...) and identity[int](...) give the type arguments.
	if sub, ok := n.Func.(*pyast.Subscript); ok {
		if head, ok := sub.Value.(*pyast.Name); ok && !g.isVariable(head.ID) {
			if gf := g.genericFns[head.ID]; gf != nil {
				targs, err := g.typeArgs(sub.Index)
				if err != nil {
					return value{}, err
				}
				return g.genericCall(n, gf, targs)
			}
			if gc := g.genericClasses[head.ID]; gc != nil {
				targs, err := g.typeArgs(sub.Index)
				if err != nil {
					return value{}, err
				}
				return g.genericClassCall(n, gc, targs)
			}
		}
	}
	// Any other expression, such as make_adder(1)(2) or handlers[0](x), is
	// called through its value.
	fv, err := g.recvExpr(n.Func)
	if err != nil {
		return value{}, err
	}
	switch {
	case fv.typ.isClass():
		return g.callInstance(n, fv)
	case fv.typ.isClassRef() && fv.classRef != nil:
		if fv.dynamicRef {
			return value{}, unsupported("call of the class of a value that is not exactly known (type(x)(...) works on self only)", n.Line())
		}
		return g.constructCall(n, fv.classRef)
	case fv.typ == tObject:
		return g.dynCall(n, fv)
	case !fv.typ.isFunc() && g.boxable(fv):
		return g.dynCall(n, fv)
	case !fv.typ.isFunc():
		return value{}, unsupported(fmt.Sprintf("call of a value of type %s", fv.typ), n.Line())
	}
	return g.callValue(n, fv)
}

func (g *generator) namedCall(n *pyast.Call, callee *pyast.Name) (value, error) {
	name := callee.ID
	// Variables and nested functions hide module-level functions and builtins.
	if sig, found := g.env.lookupSig(name); found {
		if sig != nil && len(sig.decorators) == 0 {
			return g.userCallOrDynamic(n, sig)
		}
		t, _ := g.env.lookup(name)
		if t.isFunc() {
			return g.callValue(n, value{code: name, typ: t, prec: precAtom})
		}
		if t.isClass() {
			return g.callInstance(n, value{code: name, typ: t, prec: precAtom})
		}
		if t.isClassRef() {
			return g.constructCall(n, g.classes[t.classRefName()])
		}
		if g.boxable(value{typ: t}) {
			return g.dynCall(n, value{code: name, typ: t, prec: precAtom})
		}
		return value{}, unsupported(fmt.Sprintf("call of variable '%s'", name), n.Line())
	}
	if g.fnName != "" {
		if t, ok := g.globalTypes[name]; ok {
			v, _ := g.readGlobal(name)
			if t.isClass() {
				return g.callInstance(n, v)
			}
			if t == tObject {
				return g.dynCall(n, v)
			}
			if !t.isFunc() {
				return value{}, unsupported(fmt.Sprintf("call of variable '%s' of type %s", name, t), n.Line())
			}
			return g.callValue(n, v)
		}
	}
	if cv, ok := g.classNameValue(name); ok {
		return g.constructCall(n, cv.classRef)
	}
	if gf := g.genericFns[name]; gf != nil && !g.isVariable(name) {
		return g.genericCall(n, gf, nil)
	}
	if gc := g.genericClasses[name]; gc != nil && !g.isVariable(name) {
		return g.genericClassCall(n, gc, nil)
	}
	// A function the program defines hides a builtin of the same name.
	if err := g.sigErrs[name]; err != nil {
		return value{}, err
	}
	if sig, ok := g.sigs[name]; ok && len(sig.decorators) == 0 {
		return g.userCallOrDynamic(n, sig)
	}
	if g.isExceptionClass(name) && g.classes[name] == nil {
		// ValueError("x") as a value, to store, pass or raise later.
		code, err := g.excValue(n)
		if err != nil {
			return value{}, err
		}
		return value{code: code, typ: tExc, prec: precAtom}, nil
	}
	return g.builtinCall(n, callee)
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
	switch n.Op {
	case "BitAnd", "BitOr", "BitXor", "LShift", "RShift":
		// An int literal next to a float stays an int: the operator is
		// refused with the types the program has.
		if err == nil && l.typ == tFloat && r.typ == tFloat {
			if rv, rerr := g.typedExpr(n.Right); rerr == nil {
				r = rv
			}
		}
	}
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
		if l.typ != tNone && r.typ != tNone && l.typ != "" && r.typ != "" && !l.typ.isClassRef() && !r.typ.isClassRef() && l.typ != tExc && r.typ != tExc {
			// The operands' types are known to be unsuitable; Python
			// raises the TypeError when the expression runs.
			return g.dynBinary(op, l, r), nil
		}
		return value{}, unsupported(fmt.Sprintf("operator %s on %s and %s", opSymbols[op], l.typ, r.typ), line)
	}
	if involvesObject(l, r) {
		return g.dynBinary(op, l, r), nil
	}
	if g.bigMode && l.typ == tInt && r.typ == tInt && growthOp(op) {
		// Ints in code without hints do not stop at 64 bits.
		return g.dynBinary(op, l, r), nil
	}
	if l.typ.isOptional() || r.typ.isOptional() {
		l, r = g.unwrapOperands(func(a, b string) string {
			return fmt.Sprintf("unsupported operand type(s) for %s: '%s' and '%s'", opSymbols[op], a, b)
		}, l, r)
	}
	if l.typ.isClass() || r.typ.isClass() {
		return g.classBinary(op, l, r, line)
	}
	if op == "MatMult" {
		return value{}, unsupported("operator "+opSymbols[op], line)
	}

	if v, ok := g.setBinary(op, l, r); ok {
		return v, nil
	}

	// str, list and tuple[T, ...] operands: concatenation and repetition.
	if l.typ == tStr || r.typ == tStr || l.typ.isList() || r.typ.isList() || l.typ.isVarTuple() || r.typ.isVarTuple() {
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
		v, err := g.power(l, r, line)
		if err != nil && g.boxable(l) && g.boxable(r) {
			return g.dynBinary(op, l, r), nil
		}
		return v, err

	case "Add", "Sub", "Mult":
		goOp, prec := map[string]string{"Add": "+", "Sub": "-", "Mult": "*"}[op], precAdd
		if op == "Mult" {
			prec = precMul
		}
		if bothInt {
			if op == "Mult" && !(l.konst && r.konst) {
				// A product that does not fit raises OverflowError.
				return value{code: g.rt("MulInt") + "(" + l.code + ", " + r.code + ")", typ: tInt, prec: precAtom}, nil
			}
			return infix(l, goOp, r, tInt, prec), nil
		}
		l, r = operands(toFloat(l), toFloat(r))
		r = guardConst(goOp, l, r)
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

// setBinary handles the operators of sets, dicts and keys views: |, &, - and ^
// on sets and keys views, and | on dicts.
func (g *generator) setBinary(op string, l, r value) (value, bool) {
	if l.typ.isDict() && l.typ == r.typ && op == "BitOr" {
		return value{code: l.at(precAtom) + ".Union(" + r.code + ")", typ: l.typ, prec: precAtom}, true
	}
	method := map[string]string{"BitOr": "Union", "BitAnd": "Intersection", "Sub": "Difference", "BitXor": "SymmetricDifference"}[op]
	if method == "" {
		return value{}, false
	}
	switch {
	case l.typ.isSet() && l.typ == r.typ:
		return value{code: l.at(precAtom) + "." + method + "(" + r.code + ")", typ: l.typ, prec: precAtom}, true
	case l.typ.isKeysView():
		// d.keys() | other works on a set made from the keys.
		viewMethod := map[string]string{"BitOr": "Or", "BitAnd": "And", "Sub": "Sub", "BitXor": "Xor"}[op]
		code, elem, ok := g.iterableOf(r)
		if !ok || elem != l.typ.keyType() || r.typ.isTuple() {
			return value{}, false
		}
		return value{code: l.at(precAtom) + "." + viewMethod + "(" + code + ")", typ: setOf(elem), prec: precAtom}, true
	}
	return value{}, false
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
	case op == "Add" && same && l.typ.isVarTuple():
		return value{code: l.at(precAtom) + ".Concat(" + r.code + ")", typ: l.typ, prec: precAtom}, true
	case op == "Mult" && l.typ.isVarTuple() && r.typ == tInt:
		return value{code: l.at(precAtom) + ".Repeat(" + r.code + ")", typ: l.typ, prec: precAtom}, true
	case op == "Mult" && l.typ == tInt && r.typ.isVarTuple():
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
		case l.intLit && r.intLit && r.intVal >= 0 && r.intVal <= 4096:
			// Both are literals: the power is computed now, as an int when it
			// fits and as a big int otherwise.
			p := new(big.Int).Exp(big.NewInt(l.intVal), big.NewInt(r.intVal), nil)
			if p.IsInt64() {
				return value{code: strconv.FormatInt(p.Int64(), 10), typ: tInt, prec: precAtom, konst: true, intLit: true, intVal: p.Int64()}, nil
			}
			return value{code: g.rt("BigLit") + "(" + strconv.Quote(p.String()) + ")", typ: tObject, prec: precAtom}, nil
		case !r.intLit:
			return g.dynBinary("Pow", l, r), nil
		case r.intVal >= 0:
			return value{code: g.rt("PowInt") + "(" + l.code + ", " + r.code + ")", typ: tInt, prec: precAtom}, nil
		}
	}
	// A negative base with a fractional exponent is a complex number, so the
	// result type depends on the values unless the base is a positive literal
	// or the exponent a whole number.
	if !(l.konst && (l.typ == tInt && l.intVal > 0 || l.typ == tFloat && l.lit)) && !(r.typ == tInt) {
		return g.dynBinary("Pow", l, r), nil
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
	if (v.typ == tObject || v.typ == tBool) && n.Op != "Not" {
		return g.dynUnary(n.Op, v), nil
	}
	if v.typ.isClass() && n.Op != "Not" {
		return g.classUnary(n.Op, v, n.Line())
	}

	switch n.Op {
	case "Not":
		// "not x" is a bool for any x, so truthiness applies to every type.
		t := g.truth(v)
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
	if v.typ.isList() || v.typ.isTuple() || v.typ.isVarTuple() || v.typ.isDict() || v.typ.isSet() || v.typ.isView() || v.typ == tRange {
		return value{code: v.at(precAtom) + ".Len() != 0", typ: tBool, prec: precCmp}
	}
	if v.typ.isIter() {
		return value{code: "true", typ: tBool, prec: precAtom} // an iterator is always truthy
	}
	if v.typ == tExc {
		return value{code: "true", typ: tBool, prec: precAtom} // exceptions are always truthy
	}
	return v
}

// operandPair generates the two sides of a binary operator. A list or
// tuple literal on one side takes its type from the other, so that
// "xs == []" and "xs + [1]" work without annotations.
func (g *generator) operandPair(left, right pyast.Node) (l, r value, err error) {
	isLiteral := func(n pyast.Node) bool {
		switch n.(type) {
		case *pyast.List, *pyast.Tuple, *pyast.Dict, *pyast.Set:
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
		return g.chainedCompare(n)
	}
	op := n.Ops[0]
	if op == "In" || op == "NotIn" {
		return g.contains(n, op == "NotIn")
	}
	if v, ok, err := g.typeCompare(n); ok || err != nil {
		return v, err
	}
	if v, ok, err := g.noneCompare(n); ok || err != nil {
		return v, err
	}
	goOp := map[string]string{"Eq": "==", "NotEq": "!=", "Lt": "<", "LtE": "<=", "Gt": ">", "GtE": ">="}[op]
	if goOp == "" && op != "Is" && op != "IsNot" {
		return value{}, unsupported("operator '"+opSymbols[op]+"'", n.Line())
	}

	l, r, err := g.operandPair(n.Left, n.Comparators[0])
	if err != nil {
		return value{}, err
	}
	if v, ok, err := g.objectCompare(op, l, r, n.Line()); ok || err != nil {
		return v, err
	}
	if goOp == "" {
		return value{}, unsupported("operator '"+opSymbols[op]+"'", n.Line())
	}
	if (l.typ.isOptional() || r.typ.isOptional()) && op != "Eq" && op != "NotEq" && op != "Is" && op != "IsNot" {
		l, r = g.unwrapOperands(func(a, b string) string {
			return fmt.Sprintf("'%s' not supported between instances of '%s' and '%s'", opSymbols[op], a, b)
		}, l, r)
	}
	if l.typ.isOptional() || r.typ.isOptional() {
		t := l.typ
		if !t.isOptional() {
			t = r.typ
		}
		if !g.isSubtype(l.typ, t) || !g.isSubtype(r.typ, t) {
			return value{}, unsupported(fmt.Sprintf("comparison of %s with %s", l.typ, r.typ), n.Line())
		}
		code := g.rt("Eq") + "(" + g.relabel(l, t).code + ", " + g.relabel(r, t).code + ")"
		if op == "NotEq" {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, nil
	}

	if l.typ == tExc || r.typ == tExc {
		return value{}, unsupported("comparison of exceptions", n.Line())
	}
	if l.typ.isFunc() || r.typ.isFunc() {
		return value{}, unsupported("comparison of function values", n.Line())
	}

	// A keys view equals a set (or another keys view) with the same items.
	if (op == "Eq" || op == "NotEq") && keysViewAndSet(l.typ, r.typ) {
		code := g.rt("Eq") + "(" + l.code + ", " + r.code + ")"
		if op == "NotEq" {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, nil
	}

	// An int compared with a float converts to float, as in Python.
	// Anything else must match exactly: Python allows 1 == "a" or True < 2,
	// but those are almost always mistakes.
	switch {
	case isNumeric(l.typ) && isNumeric(r.typ) && l.typ != r.typ:
		l, r = toFloat(l), toFloat(r)
	case l.typ != r.typ && g.boxable(l) && g.boxable(r):
		return g.dynCompare(op, l, r), nil
	case l.typ != r.typ:
		return value{}, unsupported(fmt.Sprintf("comparison of %s with %s", l.typ, r.typ), n.Line())
	case l.typ == tBool && op != "Eq" && op != "NotEq":
		return g.dynCompare(op, l, r), nil
	case l.typ.isDict() || l.typ == tRange || l.typ.isView() || l.typ.isIter():
		// These compare for equality only.
		if op != "Eq" && op != "NotEq" {
			return value{}, unsupported(fmt.Sprintf("ordering comparison of %s", l.typ), n.Line())
		}
		code := g.rt("Eq") + "(" + l.code + ", " + r.code + ")"
		if op == "NotEq" {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, nil
	case l.typ.isList() || l.typ.isTuple() || l.typ.isVarTuple() || l.typ.isSet():
		// Containers compare element by element in the runtime. For sets,
		// the ordering operators mean "is a subset".
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
	if r.typ.isClass() {
		return g.classContains(n, r, negate)
	}
	var want pyType
	switch {
	case r.typ.isList() || r.typ.isSet() || r.typ.isVarTuple():
		want = r.typ.elem()
	case r.typ.isDict() || r.typ.isKeysView():
		want = r.typ.keyType()
	case r.typ.isValuesView():
		want = r.typ.valType()
	}
	l, err := g.exprAs(n.Left, want)
	if err != nil {
		return value{}, err
	}

	var code string
	switch {
	case r.typ == tStr && l.typ == tStr:
		code = g.rt("StrContains") + "(" + r.code + ", " + l.code + ")"
	case (r.typ.isList() || r.typ.isVarTuple()) && l.typ == r.typ.elem():
		code = r.at(precAtom) + ".Contains(" + l.code + ")"
	case (r.typ.isSet() && l.typ == r.typ.elem()), (r.typ.isDict() && l.typ == r.typ.keyType()),
		(r.typ.isKeysView() && l.typ == r.typ.keyType()), (r.typ.isValuesView() && l.typ == r.typ.valType()),
		(r.typ.isItemsView() && l.typ == tupleOf([]pyType{r.typ.keyType(), r.typ.valType()})),
		(r.typ == tRange && l.typ == tInt):
		code = r.at(precAtom) + ".Has(" + l.code + ")"
	case r.typ.isTuple() && allEqual(r.typ.elems(), l.typ):
		// The item is boxed as an interface, so give a constant its Go type.
		code = g.rt("SeqContains") + "(" + r.code + ", " + l.typ.goType() + "(" + l.code + "))"
	default:
		if r.typ.isOptional() || r.typ == tNone || r.typ == tNoneType || l.typ == tNone {
			op := "in"
			if negate {
				op = "not in"
			}
			return value{}, unsupported(fmt.Sprintf("operator '%s' on %s and %s", op, l.typ, r.typ), n.Line())
		}
		code = g.rt("DynContains") + "(" + g.boxed(r) + ", " + g.boxed(l) + ")"
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
	return g.boolOperands(n.Op, n.Values)
}

// boolOperands generates "a and b and c" or "a or b or c". After
// "isinstance(x, C) and ..." (or "not isinstance(x, C) or ...") the operands
// that follow see x as a C, which the generated code binds in a closure.
func (g *generator) boolOperands(op string, values []pyast.Node) (value, error) {
	goOp, prec := "&&", precAnd
	if op == "Or" {
		goOp, prec = "||", precOr
	}
	word := strings.ToLower(op)
	_ = word

	parts := make([]string, 0, len(values))
	konst := true
	for i, e := range values {
		v, err := g.typedExpr(e)
		if err != nil {
			return value{}, err
		}
		// Python's and/or return one of their operands, Go's return a bool.
		// The two agree only when every operand is a bool.
		if v.typ != tBool {
			if !g.boxable(v) {
				return value{}, unsupported(fmt.Sprintf("'%s' on %s operand", word, v.typ), e.Line())
			}
			return g.valueBoolOp(op, values)
		}
		konst = konst && v.konst
		// Same-operator children can share the parent's precedence.
		parts = append(parts, v.at(prec))

		if i == len(values)-1 {
			break
		}
		pos, neg := g.narrowTest(e)
		nars := pos
		if op == "Or" {
			nars = neg
		}
		if len(nars) == 0 {
			continue
		}
		outer := g.env
		g.env = newScope(outer)
		decls := g.bindNarrowExpr(nars)
		rest, err := g.boolOperands(op, values[i+1:])
		g.env = outer
		if err != nil {
			return value{}, err
		}
		if decls == "" {
			parts = append(parts, rest.at(prec))
		} else {
			parts = append(parts, fmt.Sprintf("func() bool {\n%sreturn %s\n}()", decls, rest.code))
		}
		konst = false
		break
	}
	return value{code: strings.Join(parts, " "+goOp+" "), typ: tBool, prec: prec, konst: konst}, nil
}

// keysViewAndSet is whether a and b are a keys view and a set, or two keys
// views, over the same item type.
func keysViewAndSet(a, b pyType) bool {
	item := func(t pyType) (pyType, bool) {
		switch {
		case t.isKeysView():
			return t.keyType(), true
		case t.isSet():
			return t.elem(), true
		}
		return "", false
	}
	ea, oka := item(a)
	eb, okb := item(b)
	return oka && okb && ea == eb && (a.isKeysView() || b.isKeysView())
}

// growthOp is an operator whose result can need more bits than its operands.
func growthOp(op string) bool {
	switch op {
	case "Add", "Sub", "Mult", "Pow", "LShift":
		return true
	}
	return false
}
