package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// builtinCall generates a call to a Python builtin function. User-defined
// functions of the same name are tried first, in namedCall.
func (g *generator) builtinCall(n *pyast.Call, callee *pyast.Name) (value, error) {
	name := callee.ID
	if v, ok, err := g.dynBuiltin(name, n, false); ok || err != nil {
		return v, err
	}
	v, err := g.builtinCallTyped(n, callee)
	if err != nil && !g.infer && fallbackOK(err) {
		// What the static rules refuse, the run-time ones may do.
		if dv, ok, derr := g.dynBuiltin(name, n, true); ok && derr == nil {
			return dv, nil
		}
	}
	return v, err
}

func (g *generator) builtinCallTyped(n *pyast.Call, callee *pyast.Name) (value, error) {
	name := callee.ID
	switch name {
	case "next":
		return g.nextCall(n)
	case "iter":
		return g.iterCall(n)
	case "isinstance":
		return g.isinstanceCall(n)
	case "issubclass":
		return g.issubclassCall(n)
	case "type":
		return g.typeCall(n)
	case "callable":
		return g.callableCall(n)
	case "print":
		return g.printCall(n)
	case "sorted":
		return g.sortedCall(n)
	case "range":
		return g.rangeCall(n)
	case "enumerate":
		return g.enumerateCall(n)
	case "zip":
		return g.zipCall(n)
	case "reversed":
		return g.reversedCall(n)
	case "list", "set", "dict", "tuple":
		return g.convertCall(name, n)
	case "any", "all":
		return g.anyAllCall(name, n)
	case "map":
		return g.mapCall(n)
	case "filter":
		return g.filterCall(n)
	case "min", "max":
		if len(n.Keywords) > 0 {
			return g.minMaxKey(name, n)
		}
		args := make([]value, len(n.Args))
		for i, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return value{}, err
			}
			args[i] = v
		}
		return g.minMax(name, n, args)
	case "len", "int", "float", "str", "repr", "bool", "abs", "round", "sum", "divmod", "hash", "id":
		if n.HasKeywords {
			return value{}, unsupported("keyword arguments", n.Line())
		}
		args := make([]value, len(n.Args))
		for i, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return value{}, err
			}
			args[i] = v
		}
		return g.scalarBuiltin(name, n, args)
	}
	return value{}, unsupported(fmt.Sprintf("call to undefined function '%s'", name), n.Line())
}

// arity reports a call with the wrong number of arguments.
func arity(name string, n *pyast.Call, min, max int) error {
	if len(n.Args) >= min && len(n.Args) <= max {
		return nil
	}
	return unsupported(fmt.Sprintf("%s() with %d arguments", name, len(n.Args)), n.Line())
}

func call(fn string, args ...string) string { return fn + "(" + strings.Join(args, ", ") + ")" }

func (g *generator) scalarBuiltin(name string, n *pyast.Call, args []value) (value, error) {
	atom := func(code string, t pyType) (value, error) {
		return value{code: code, typ: t, prec: precAtom}, nil
	}
	badArg := func(a value) (value, error) {
		return value{}, unsupported(fmt.Sprintf("%s() of %s", name, a.typ), n.Line())
	}
	if len(args) == 1 {
		if v, ok, err := g.classBuiltin(name, n, args[0]); ok || err != nil {
			return v, err
		}
	}
	if g.bigMode && len(args) == 1 && (name == "int" || name == "round") && (args[0].typ == tStr || args[0].typ == tFloat) {
		// The result can be any size.
		fn := "DynIntAny"
		if name == "round" {
			fn = "DynRound"
			return g.dynValue(g.rt(fn)+"("+g.boxed(args[0])+", nil)", tObject), nil
		}
		return g.dynValue(g.rt(fn)+"("+g.boxed(args[0])+")", tObject), nil
	}
	if name == "id" {
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		return atom(call(g.rt("Id"), g.boxed(args[0])), tInt)
	}
	if name == "hash" {
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		a := args[0]
		if a.typ == tInt || a.typ == tStr || a.typ == tBool || a.typ.isTuple() {
			return atom(call(g.rt("Hash"), g.relabel(a, tObject).code), tInt)
		}
		return value{}, unsupported(fmt.Sprintf("hash() of %s", a.typ), n.Line())
	}

	switch name {
	case "len":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		a := args[0]
		switch {
		case a.typ == tStr:
			return atom(call(g.rt("StrLen"), a.code), tInt)
		case a.typ.isList() || a.typ.isTuple() || a.typ.isVarTuple() || a.typ.isDict() || a.typ.isSet() || a.typ.isView() || a.typ == tRange:
			return atom(a.at(precAtom)+".Len()", tInt)
		}
		return badArg(a)

	case "int":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		switch a := args[0]; a.typ {
		case tInt:
			return a, nil
		case tBool:
			return atom(call(g.rt("BoolToInt"), a.code), tInt)
		case tFloat:
			return atom(call(g.rt("IntFromFloat"), a.code), tInt)
		case tStr:
			return atom(call(g.rt("IntFromStr"), a.code), tInt)
		default:
			return badArg(a)
		}

	case "float":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		switch a := args[0]; a.typ {
		case tFloat:
			return a, nil
		case tInt:
			return toFloat(a), nil
		case tBool:
			return atom("float64("+call(g.rt("BoolToInt"), a.code)+")", tFloat)
		case tStr:
			return atom(call(g.rt("FloatFromStr"), a.code), tFloat)
		default:
			return badArg(a)
		}

	case "str":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		if args[0].typ == tStr {
			return args[0], nil
		}
		return atom(call(g.rt("Str"), args[0].code), tStr)

	case "repr":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		return atom(call(g.rt("Repr"), args[0].code), tStr)

	case "bool":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		return g.truth(args[0]), nil

	case "abs":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		switch a := args[0]; a.typ {
		case tInt:
			return atom(call(g.rt("AbsInt"), a.code), tInt)
		case tFloat:
			return atom(call(g.rt("AbsFloat"), a.code), tFloat)
		default:
			return badArg(a)
		}

	case "round":
		if len(n.Args) == 2 {
			return value{}, unsupported("round() with 2 arguments", n.Line())
		}
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		switch a := args[0]; a.typ {
		case tInt:
			return a, nil
		case tFloat:
			return atom(call(g.rt("RoundFloat"), a.code), tInt)
		default:
			return badArg(a)
		}

	case "divmod":
		if err := arity(name, n, 2, 2); err != nil {
			return value{}, err
		}
		a, b := args[0], args[1]
		if !isNumeric(a.typ) || !isNumeric(b.typ) {
			return value{}, unsupported(fmt.Sprintf("divmod() of %s and %s", a.typ, b.typ), n.Line())
		}
		if a.typ == tInt && b.typ == tInt {
			return atom(call(g.rt("DivModInt"), a.code, b.code), tupleOf([]pyType{tInt, tInt}))
		}
		a, b = operands(toFloat(a), toFloat(b))
		return atom(call(g.rt("DivModFloat"), a.code, b.code), tupleOf([]pyType{tFloat, tFloat}))

	case "sum":
		if err := arity(name, n, 1, 2); err != nil {
			return value{}, err
		}
		a, err := g.asList(args[0], "sum()", n.Line())
		if err != nil {
			return value{}, err
		}
		var start *value
		if len(args) == 2 {
			start = &args[1]
		}
		if g.classOf(a.typ.elem()) != nil || (start != nil && start.typ.isClass()) {
			return g.sumInstances(n, a, start)
		}
		if start != nil && start.typ == tStr {
			// Python refuses to sum into a string, with a message of its own.
			return g.dynValue(g.rt("DynSum")+"("+g.boxed(a)+", "+g.boxed(*start)+")", tObject), nil
		}
		var total value
		switch a.typ {
		case listOf(tInt):
			total = value{code: call(g.rt("SumInt"), a.code), typ: tInt, prec: precAtom}
		case listOf(tFloat):
			total = value{code: call(g.rt("SumFloat"), a.code), typ: tFloat, prec: precAtom}
		default:
			return badArg(args[0])
		}
		if start == nil {
			return total, nil
		}
		return g.binary("Add", *start, total, n.Line())

	case "min", "max":
		return g.minMax(name, n, args)
	}
	return value{}, unsupported("builtin "+name, n.Line())
}

// minMax is min/max over several arguments or over one list. The arguments
// must share a type: Python returns the winning object itself, so
// max(3, 2.5) is the int 3, and a Go function cannot return int or float.
func (g *generator) minMax(name string, n *pyast.Call, args []value) (value, error) {
	title := strings.ToUpper(name[:1]) + name[1:]
	ordered := func(t pyType) bool { return t == tInt || t == tFloat || t == tStr }

	if len(args) == 1 {
		a, err := g.asList(args[0], name+"()", n.Line())
		if err != nil {
			return value{}, err
		}
		if g.classOf(a.typ.elem()) != nil {
			return value{code: call(g.rt(title+"AnyList"), a.code), typ: a.typ.elem(), prec: precAtom}, nil
		}
		if !ordered(a.typ.elem()) {
			return value{}, unsupported(fmt.Sprintf("%s() of %s", name, args[0].typ), n.Line())
		}
		return value{code: call(g.rt(title+"List"), a.code), typ: a.typ.elem(), prec: precAtom}, nil
	}
	if len(args) < 2 {
		return value{}, unsupported(fmt.Sprintf("%s() with %d arguments", name, len(args)), n.Line())
	}

	t := args[0].typ
	codes := make([]string, len(args))
	if g.classOf(t) != nil {
		for i, a := range args {
			if a.typ != t {
				return value{}, unsupported(fmt.Sprintf("%s() of %s and %s", name, t, a.typ), n.Line())
			}
			codes[i] = a.code
		}
		return value{code: g.rt(title+"Any") + "[" + g.goType(t) + "](" + strings.Join(codes, ", ") + ")", typ: t, prec: precAtom}, nil
	}
	for i, a := range args {
		if a.typ != t || !ordered(t) {
			return value{}, unsupported(fmt.Sprintf("%s() of %s and %s (arguments must share one of int, float, str)", name, t, a.typ), n.Line())
		}
		codes[i] = a.code
	}
	return value{code: g.rt(title) + "[" + t.goType() + "](" + strings.Join(codes, ", ") + ")", typ: t, prec: precAtom}, nil
}

// reverseKeyword reads the optional reverse=... keyword of sorted and
// list.sort, and rejects every other keyword.
func (g *generator) reverseKeyword(n *pyast.Call, fn string) (string, error) {
	code := "false"
	for _, kw := range n.Keywords {
		if kw.Arg != "reverse" {
			return "", unsupported(fmt.Sprintf("keyword argument '%s' of %s()", kw.Arg, fn), kw.Line())
		}
		v, err := g.typedExpr(kw.Value)
		if err != nil {
			return "", err
		}
		if v.typ != tBool {
			return "", mismatch("reverse", tBool, v.typ, kw.Line())
		}
		code = v.code
	}
	return code, nil
}

func (g *generator) sortedCall(n *pyast.Call) (value, error) {
	if err := arity("sorted", n, 1, 1); err != nil {
		return value{}, err
	}
	a, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	list, err := g.asList(a, "sorted()", n.Line())
	if err != nil {
		return value{}, err
	}
	rev, keyNode, err := g.sortKeywords(n, "sorted")
	if err != nil {
		return value{}, err
	}
	if keyNode == nil {
		return value{code: call(g.rt("Sorted"), list.code, rev), typ: list.typ, prec: precAtom}, nil
	}
	key, keyT, err := g.keyFunc(keyNode, list.typ.elem())
	if err != nil {
		return value{}, err
	}
	fn := g.rt("SortedKey") + "[" + g.goType(list.typ.elem()) + ", " + g.goType(keyT) + "]"
	return value{code: call(fn, list.code, key, rev), typ: list.typ, prec: precAtom}, nil
}

// sortKeywords reads reverse=... and key=... of sorted() and list.sort().
func (g *generator) sortKeywords(n *pyast.Call, fn string) (reverse string, key pyast.Node, err error) {
	reverse = "false"
	for _, kw := range n.Keywords {
		switch kw.Arg {
		case "reverse":
			v, err := g.typedExpr(kw.Value)
			if err != nil {
				return "", nil, err
			}
			if v.typ != tBool {
				return "", nil, mismatch("reverse", tBool, v.typ, kw.Line())
			}
			reverse = v.code
		case "key":
			key = kw.Value
		default:
			return "", nil, unsupported(fmt.Sprintf("keyword argument '%s' of %s()", kw.Arg, fn), kw.Line())
		}
	}
	return reverse, key, nil
}

// keyFunc generates the key function of sorted, min, max or sort for items
// of type item. A lambda takes its parameter type from the items.
func (g *generator) keyFunc(node pyast.Node, item pyType) (code string, key pyType, err error) {
	var v value
	if lam := g.builtinAsLambda(node, funcOf([]pyType{item}, tNone)); lam != nil {
		node = lam
	}
	if lam, ok := node.(*pyast.Lambda); ok {
		v, err = g.lambda(lam, []pyType{item}, "")
	} else {
		v, err = g.exprAs(node, funcOf([]pyType{item}, tNone))
	}
	if err != nil {
		return "", "", err
	}
	if !v.typ.isFunc() || len(v.typ.funcParams()) != 1 || v.typ.funcParams()[0] != item || v.typ.funcRet() == tNone {
		return "", "", unsupported(fmt.Sprintf("key function of type %s for items of type %s", v.typ, item), node.Line())
	}
	return v.code, v.typ.funcRet(), nil
}

// minMaxKey is min(...) and max(...) with a key function.
func (g *generator) minMaxKey(name string, n *pyast.Call) (value, error) {
	var keyNode pyast.Node
	for _, kw := range n.Keywords {
		if kw.Arg != "key" {
			return value{}, unsupported(fmt.Sprintf("keyword argument '%s' of %s()", kw.Arg, name), kw.Line())
		}
		keyNode = kw.Value
	}
	var list value
	if len(n.Args) == 1 {
		v, err := g.typedExpr(n.Args[0])
		if err != nil {
			return value{}, err
		}
		if list, err = g.asList(v, name+"()", n.Line()); err != nil {
			return value{}, err
		}
	} else if len(n.Args) >= 2 {
		codes := make([]string, len(n.Args))
		var t pyType
		for i, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return value{}, err
			}
			if i == 0 {
				t = v.typ
			} else if v.typ != t {
				return value{}, unsupported(fmt.Sprintf("%s() of %s and %s", name, t, v.typ), a.Line())
			}
			codes[i] = v.code
		}
		list = value{code: g.rt("NewList") + "[" + g.goType(t) + "](" + strings.Join(codes, ", ") + ")", typ: listOf(t), prec: precAtom}
	} else {
		return value{}, unsupported(fmt.Sprintf("%s() with no arguments", name), n.Line())
	}
	item := list.typ.elem()
	key, keyT, err := g.keyFunc(keyNode, item)
	if err != nil {
		return value{}, err
	}
	fn := map[string]string{"min": "MinKey", "max": "MaxKey"}[name]
	return value{
		code: call(g.rt(fn)+"["+g.goType(item)+", "+g.goType(keyT)+"]", list.code, key),
		typ:  item, prec: precAtom,
	}, nil
}

// functionArg generates the function argument of map() or filter(): a lambda
// takes its parameter types from the iterables.
func (g *generator) functionArg(node pyast.Node, params []pyType) (value, error) {
	var v value
	var err error
	if lam := g.builtinAsLambda(node, funcOf(params, tNone)); lam != nil {
		node = lam
	}
	if lam, ok := node.(*pyast.Lambda); ok {
		v, err = g.lambda(lam, params, "")
	} else {
		v, err = g.exprAs(node, funcOf(params, tNone))
	}
	if err != nil {
		return value{}, err
	}
	if !v.typ.isFunc() || len(v.typ.funcParams()) != len(params) {
		return value{}, unsupported(fmt.Sprintf("function of type %s where one taking %d argument(s) is expected", v.typ, len(params)), node.Line())
	}
	for i, p := range params {
		if v.typ.funcParams()[i] != p {
			return value{}, mismatch(fmt.Sprintf("function argument %d", i+1), p, v.typ.funcParams()[i], node.Line())
		}
	}
	return v, nil
}

// mapCall is map(f, iterable) and map(f, a, b).
func (g *generator) mapCall(n *pyast.Call) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) < 2 || len(n.Args) > 3 {
		return value{}, unsupported(fmt.Sprintf("map() with %d arguments", len(n.Args)), n.Line())
	}
	codes := make([]string, len(n.Args)-1)
	elems := make([]pyType, len(n.Args)-1)
	for i, a := range n.Args[1:] {
		code, elem, err := g.iterableArg(a, "map()")
		if err != nil {
			return value{}, err
		}
		codes[i], elems[i] = code, elem
	}
	fn, err := g.functionArg(n.Args[0], elems)
	if err != nil {
		return value{}, err
	}
	ret := fn.typ.funcRet()
	if ret == tNone {
		return value{}, unsupported("map() of a function that returns None", n.Line())
	}
	if len(elems) == 1 {
		f := g.rt("Map") + "[" + g.goType(elems[0]) + ", " + g.goType(ret) + "]"
		return value{code: call(f, fn.code, codes[0]), typ: iterOf(ret), prec: precAtom}, nil
	}
	f := g.rt("Map2") + "[" + g.goType(elems[0]) + ", " + g.goType(elems[1]) + ", " + g.goType(ret) + "]"
	return value{code: call(f, fn.code, codes[0], codes[1]), typ: iterOf(ret), prec: precAtom}, nil
}

// filterCall is filter(f, iterable).
func (g *generator) filterCall(n *pyast.Call) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) != 2 {
		return value{}, unsupported(fmt.Sprintf("filter() with %d arguments", len(n.Args)), n.Line())
	}
	if c, ok := n.Args[0].(*pyast.Constant); ok && c.PyType == "none" {
		return value{}, unsupported("filter(None, ...)", n.Line())
	}
	code, elem, err := g.iterableArg(n.Args[1], "filter()")
	if err != nil {
		return value{}, err
	}
	fn, err := g.functionArg(n.Args[0], []pyType{elem})
	if err != nil {
		return value{}, err
	}
	if fn.typ.funcRet() != tBool {
		return value{}, mismatch("filter() function result", tBool, fn.typ.funcRet(), n.Line())
	}
	f := g.rt("Filter") + "[" + g.goType(elem) + "]"
	return value{code: call(f, fn.code, code), typ: iterOf(elem), prec: precAtom}, nil
}

// printCall is print(*args, sep=..., end=...).
func (g *generator) printCall(n *pyast.Call) (value, error) {
	sep, end := strconv.Quote(" "), strconv.Quote("\n")
	custom := false
	for _, kw := range n.Keywords {
		if kw.Arg != "sep" && kw.Arg != "end" {
			return value{}, unsupported(fmt.Sprintf("keyword argument '%s' of print()", kw.Arg), kw.Line())
		}
		v, err := g.typedExpr(kw.Value)
		if err != nil {
			return value{}, err
		}
		if v.typ != tStr {
			return value{}, mismatch(kw.Arg, tStr, v.typ, kw.Line())
		}
		if kw.Arg == "sep" {
			sep = v.code
		} else {
			end = v.code
		}
		custom = true
	}

	// print(*items) spreads the items into the arguments.
	spread := false
	for _, a := range n.Args {
		if _, ok := a.(*pyast.Starred); ok {
			spread = true
		}
	}
	if spread {
		var segments []string // the arguments as a []any expression, built left to right
		var run []string
		flush := func() {
			if len(run) > 0 {
				segments = append(segments, "[]any{"+strings.Join(run, ", ")+"}")
				run = nil
			}
		}
		for _, a := range n.Args {
			if st, ok := a.(*pyast.Starred); ok {
				code, elem, err := g.iterableArg(st.Value, "print(*...)")
				if err != nil {
					return value{}, err
				}
				flush()
				segments = append(segments, call(g.rt("Spread")+"["+g.goType(elem)+"]", code))
				continue
			}
			v, err := g.typedExpr(a)
			if err != nil {
				return value{}, err
			}
			run = append(run, v.code)
		}
		flush()
		all := segments[0]
		for _, seg := range segments[1:] {
			all = "append(" + all + ", " + seg + "...)"
		}
		return value{code: g.rt("PrintOpts") + "(" + sep + ", " + end + ", " + all + "...)", typ: tNone, prec: precAtom}, nil
	}

	args := make([]string, 0, len(n.Args)+2)
	if custom {
		args = append(args, sep, end)
	}
	for _, a := range n.Args {
		v, err := g.typedExpr(a)
		if err != nil {
			return value{}, err
		}
		args = append(args, v.code)
	}
	fn := "Print"
	if custom {
		fn = "PrintOpts"
	}
	return value{code: call(g.rt(fn), args...), typ: tNone, prec: precAtom}, nil
}

// rangeCall is range(stop), range(start, stop) and range(start, stop, step)
// as a value. A for loop over a literal range() call does not come here.
func (g *generator) rangeCall(n *pyast.Call) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if err := arity("range", n, 1, 3); err != nil {
		return value{}, err
	}
	args := make([]string, len(n.Args))
	vals := make([]value, len(n.Args))
	anyObject := false
	for i, a := range n.Args {
		v, err := g.typedExpr(a)
		if err != nil {
			return value{}, err
		}
		vals[i] = v
		anyObject = anyObject || v.typ == tObject
	}
	if anyObject && g.bigMode {
		// The ends may not fit a machine word: the range is an object.
		return g.dynValue(g.rt("DynRange")+"("+g.dynArgs(vals)+")", tObject), nil
	}
	for i, a := range n.Args {
		v := vals[i]
		var err error
		if v.typ == tObject {
			v, err = g.coerce(v, tInt, "range() argument", a.Line())
			if err != nil {
				return value{}, err
			}
		}
		if v.typ != tInt {
			return value{}, mismatch("range() argument", tInt, v.typ, a.Line())
		}
		args[i] = v.code
	}
	switch len(args) {
	case 1:
		args = []string{"0", args[0], "1"}
	case 2:
		args = append(args, "1")
	}
	return value{code: call(g.rt("NewRange"), args...), typ: tRange, prec: precAtom}, nil
}

// iterableArg generates an argument that must be iterable and returns it as a
// runtime.Iterable together with its item type.
func (g *generator) iterableArg(node pyast.Node, what string) (code string, elem pyType, err error) {
	v, err := g.typedExpr(node)
	if err != nil {
		return "", "", err
	}
	code, elem, ok := g.iterableOf(v)
	if !ok {
		return "", "", unsupported(fmt.Sprintf("%s of %s", what, v.typ), node.Line())
	}
	return code, elem, nil
}

// enumerateCall is enumerate(iterable) and enumerate(iterable, start).
func (g *generator) enumerateCall(n *pyast.Call) (value, error) {
	start := "0"
	args := n.Args
	for _, kw := range n.Keywords {
		if kw.Arg != "start" || len(args) != 1 {
			return value{}, unsupported(fmt.Sprintf("keyword argument '%s' of enumerate()", kw.Arg), kw.Line())
		}
		args = append(append([]pyast.Node{}, args...), kw.Value)
	}
	if len(args) < 1 || len(args) > 2 {
		return value{}, unsupported(fmt.Sprintf("enumerate() with %d arguments", len(args)), n.Line())
	}
	code, elem, err := g.iterableArg(args[0], "enumerate()")
	if err != nil {
		return value{}, err
	}
	if len(args) == 2 {
		v, err := g.typedExpr(args[1])
		if err != nil {
			return value{}, err
		}
		if v.typ != tInt {
			return value{}, mismatch("enumerate() start", tInt, v.typ, args[1].Line())
		}
		start = v.code
	}
	t := iterOf(tupleOf([]pyType{tInt, elem}))
	return value{code: call(g.rt("Enumerate")+"["+g.goType(elem)+"]", code, start), typ: t, prec: precAtom}, nil
}

// zipCall is zip(a, b, ...) with one to six inputs.
func (g *generator) zipCall(n *pyast.Call) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) < 1 || len(n.Args) > maxTupleArity {
		return value{}, unsupported(fmt.Sprintf("zip() with %d arguments", len(n.Args)), n.Line())
	}
	codes := make([]string, len(n.Args))
	elems := make([]pyType, len(n.Args))
	goElems := make([]string, len(n.Args))
	for i, a := range n.Args {
		code, elem, err := g.iterableArg(a, "zip()")
		if err != nil {
			return value{}, err
		}
		codes[i], elems[i], goElems[i] = code, elem, g.goType(elem)
	}
	fn := fmt.Sprintf("%s%d[%s]", g.rt("Zip"), len(codes), strings.Join(goElems, ", "))
	return value{code: call(fn, codes...), typ: iterOf(tupleOf(elems)), prec: precAtom}, nil
}

// reversedCall is reversed(x) for a list, str, range or dict.
func (g *generator) reversedCall(n *pyast.Call) (value, error) {
	if err := arity("reversed", n, 1, 1); err != nil {
		return value{}, err
	}
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	v, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	switch {
	case v.typ.isList():
		return value{code: call(g.rt("Reversed")+"["+g.goType(v.typ.elem())+"]", v.code), typ: iterOf(v.typ.elem()), prec: precAtom}, nil
	case v.typ == tStr:
		return value{code: call(g.rt("ReversedStr"), v.code), typ: iterOf(tStr), prec: precAtom}, nil
	case v.typ == tRange:
		return value{code: call(g.rt("ReversedRange"), v.code), typ: tRange, prec: precAtom}, nil
	case v.typ.isDict():
		keys := call(g.rt("ListFrom")+"["+g.goType(v.typ.keyType())+"]", v.code)
		return value{code: call(g.rt("Reversed")+"["+g.goType(v.typ.keyType())+"]", keys), typ: iterOf(v.typ.keyType()), prec: precAtom}, nil
	}
	return value{}, unsupported(fmt.Sprintf("reversed() of %s", v.typ), n.Line())
}

// anyAllCall is any(iterable) and all(iterable). It hands the iterable to the
// runtime as it is, so a generator behind it is advanced only as far as needed.
func (g *generator) anyAllCall(name string, n *pyast.Call) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if err := arity(name, n, 1, 1); err != nil {
		return value{}, err
	}
	code, elem, err := g.iterableArg(n.Args[0], name+"()")
	if err != nil {
		return value{}, err
	}
	fn := map[string]string{"any": "Any", "all": "All"}[name]
	return value{code: call(g.rt(fn)+"["+g.goType(elem)+"]", code), typ: tBool, prec: precAtom}, nil
}

// convertCall is list(x), set(x) and dict(x).
func (g *generator) convertCall(name string, n *pyast.Call) (value, error) {
	if name == "dict" && n.HasKeywords {
		return g.dictFromKeywords(n)
	}
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) == 0 {
		switch name {
		case "list":
			return value{code: g.rt("NewList") + "[any]()", typ: listOf(tObject), prec: precAtom}, nil
		case "set":
			return value{code: g.rt("NewSet") + "[any]()", typ: setOf(tObject), prec: precAtom}, nil
		case "dict":
			return value{code: g.rt("NewDict") + "[any, any]()", typ: dictOf(tObject, tObject), prec: precAtom}, nil
		case "tuple":
			return value{code: g.rt("NewVarTuple") + "[any]()", typ: varTupleOf(tObject), prec: precAtom}, nil
		}
		return value{}, unsupported(fmt.Sprintf("empty %s without a type annotation", name), n.Line())
	}
	if len(n.Args) != 1 {
		return value{}, unsupported(fmt.Sprintf("%s() with %d arguments", name, len(n.Args)), n.Line())
	}
	v, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	code, elem, ok := g.iterableOf(v)
	if !ok {
		return value{}, unsupported(fmt.Sprintf("%s() of %s", name, v.typ), n.Line())
	}
	switch name {
	case "tuple":
		return value{code: call(g.rt("TupleFrom")+"["+g.goType(elem)+"]", code), typ: varTupleOf(elem), prec: precAtom}, nil
	case "list":
		if v.typ.isList() {
			return value{code: v.at(precAtom) + ".Copy()", typ: v.typ, prec: precAtom}, nil
		}
		return value{code: call(g.rt("ListFrom")+"["+g.goType(elem)+"]", code), typ: listOf(elem), prec: precAtom}, nil
	case "set":
		if err := g.hashableOrError(elem, "set element", n.Line()); err != nil {
			return value{}, err
		}
		return value{code: call(g.rt("SetFrom")+"["+g.goType(elem)+"]", code), typ: setOf(elem), prec: precAtom}, nil
	}
	// dict
	if v.typ.isDict() {
		return value{code: v.at(precAtom) + ".Copy()", typ: v.typ, prec: precAtom}, nil
	}
	if !elem.isTuple() || len(elem.elems()) != 2 {
		return value{}, unsupported(fmt.Sprintf("dict() of %s (needs pairs)", v.typ), n.Line())
	}
	k, val := elem.elems()[0], elem.elems()[1]
	if err := g.hashableOrError(k, "dict key", n.Line()); err != nil {
		return value{}, err
	}
	t := dictOf(k, val)
	return value{code: call(g.rt("DictFrom")+"["+g.dictGoArgs(t)+"]", code), typ: t, prec: precAtom}, nil
}

// dictFromKeywords is dict(a=1, b=2), a dict with str keys.
func (g *generator) dictFromKeywords(n *pyast.Call) (value, error) {
	if len(n.Args) != 0 {
		return value{}, unsupported("dict() with positional and keyword arguments", n.Line())
	}
	var val pyType
	pairs := make([]string, len(n.Keywords))
	for i, kw := range n.Keywords {
		if kw.Arg == "" {
			return value{}, unsupported("dict(**mapping)", kw.Line())
		}
		v, err := g.typedExpr(kw.Value)
		if err != nil {
			return value{}, err
		}
		if i == 0 {
			val = v.typ
		} else if v.typ != val {
			return value{}, unsupported(fmt.Sprintf("dict() keywords mix value types (%s and %s)", val, v.typ), kw.Line())
		}
		pairs[i] = fmt.Sprintf("%s[string, %s]{Key: %s, Val: %s}", g.rt("KV"), g.goType(val), strconv.Quote(pyName(kw.Arg)), v.code)
	}
	t := dictOf(tStr, val)
	return value{code: g.rt("NewDict") + "[string, " + g.goType(val) + "](" + strings.Join(pairs, ", ") + ")", typ: t, prec: precAtom}, nil
}
