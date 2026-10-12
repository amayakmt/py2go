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
	switch name {
	case "print":
		return g.printCall(n)
	case "range", "enumerate", "zip":
		return value{}, unsupported(name+"() outside a for loop", n.Line())
	case "sorted":
		return g.sortedCall(n)
	case "len", "int", "float", "str", "repr", "bool", "abs", "round", "sum", "divmod", "min", "max":
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

	switch name {
	case "len":
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		a := args[0]
		switch {
		case a.typ == tStr:
			return atom(call(g.rt("StrLen"), a.code), tInt)
		case a.typ.isList() || a.typ.isTuple():
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
		return truth(args[0]), nil

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
		if len(n.Args) == 2 {
			return value{}, unsupported("sum() with a start value", n.Line())
		}
		if err := arity(name, n, 1, 1); err != nil {
			return value{}, err
		}
		switch a := args[0]; a.typ {
		case listOf(tInt):
			return atom(call(g.rt("SumInt"), a.code), tInt)
		case listOf(tFloat):
			return atom(call(g.rt("SumFloat"), a.code), tFloat)
		default:
			return badArg(a)
		}

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
		a := args[0]
		if !a.typ.isList() || !ordered(a.typ.elem()) {
			return value{}, unsupported(fmt.Sprintf("%s() of %s", name, a.typ), n.Line())
		}
		return value{code: call(g.rt(title+"List"), a.code), typ: a.typ.elem(), prec: precAtom}, nil
	}
	if len(args) < 2 {
		return value{}, unsupported(fmt.Sprintf("%s() with %d arguments", name, len(args)), n.Line())
	}

	t := args[0].typ
	codes := make([]string, len(args))
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
	if !a.typ.isList() {
		return value{}, unsupported(fmt.Sprintf("sorted() of %s", a.typ), n.Line())
	}
	rev, err := g.reverseKeyword(n, "sorted")
	if err != nil {
		return value{}, err
	}
	return value{code: call(g.rt("Sorted"), a.code, rev), typ: a.typ, prec: precAtom}, nil
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
