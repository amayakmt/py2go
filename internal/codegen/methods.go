package codegen

import (
	"fmt"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// methodCall generates "receiver.name(args)" for the str and list methods
// the transpiler knows.
func (g *generator) methodCall(n *pyast.Call, m *pyast.Attribute) (value, error) {
	if (m.Attr == "startswith" || m.Attr == "endswith") && len(n.Args) > 0 && !n.HasKeywords {
		// A tuple of prefixes is searched by the runtime.
		if _, isTuple := n.Args[0].(*pyast.Tuple); isTuple {
			if recv, err := g.recvExpr(m.Value, m.Attr); err == nil && recv.typ == tStr {
				return g.dynMethodCall(n, m, recv)
			}
		}
	}
	v, err := g.methodCallTyped(n, m)
	if err != nil && g.recordJoins && strings.Contains(err.Error(), " expects ") && g.widenVar(m.Value) {
		// The variable holds other items than it was given: the next round
		// compiles it as a container of objects. For now the call is dynamic.
		if recv, rerr := g.recvExpr(m.Value, m.Attr); rerr == nil {
			if dv, derr := g.dynMethodCall(n, m, recv); derr == nil {
				return dv, nil
			}
		}
	}
	if err != nil && (!g.infer || g.recordJoins) {
		// A method the static rules do not know may still exist at run time.
		// So can a str method with arguments of the wrong type: Python raises
		// the TypeError when the call runs.
		if recv, rerr := g.recvExpr(m.Value, m.Attr); rerr == nil && g.dynMethodOK(recv) && (fallbackOK(err) || recv.typ == tStr) {
			if dv, derr := g.dynMethodCall(n, m, recv); derr == nil {
				return dv, nil
			}
		}
	}
	return v, err
}

func (g *generator) methodCallTyped(n *pyast.Call, m *pyast.Attribute) (value, error) {
	// dict.fromkeys(keys, value) is called on the type itself.
	if cls, ok := m.Value.(*pyast.Name); ok && cls.ID == "dict" && !g.isVariable("dict") {
		return g.dictClassMethod(n, m)
	}
	if sup, isSuper := g.superTarget(m.Value); isSuper {
		return g.superCall(n, sup, m.Attr)
	}
	// Exception.__init__(self, ...) and object.__init__(self): the base
	// class called by name.
	if base, ok := m.Value.(*pyast.Name); ok && !g.isVariable(base.ID) && g.classes[base.ID] == nil &&
		(base.ID == "object" || runtime.IsExceptionClass(base.ID)) && g.method != nil && g.method.kind == mMethod &&
		len(n.Args) >= 1 && !n.HasKeywords {
		if first, ok := n.Args[0].(*pyast.Name); ok && first.ID == g.selfName() {
			rest := *n
			rest.Args = n.Args[1:]
			sup := superInfo{cls: g.method.cls, definer: g.method.definer, self: g.selfName()}
			return g.superBuiltin(&rest, sup, m.Attr)
		}
	}
	recv, err := g.recvExpr(m.Value, m.Attr)
	if err != nil {
		return value{}, err
	}
	switch {
	case recv.typ.isGen():
		return g.genMethod(n, m, recv)
	case recv.typ.isIter():
		return g.iterGenMethod(n, m, recv)
	case recv.typ.isClassRef():
		return g.classLevelCall(n, m, recv)
	case recv.typ.isClass():
		return g.instanceCall(n, m, recv)
	case recv.typ == tObject:
		return g.dynMethodCall(n, m, recv)
	case recv.typ == tExc:
		return g.excMethod(n, m, recv)
	case recv.typ == tStr:
		return g.strMethod(n, m, recv)
	case recv.typ.isList():
		return g.listMethod(n, m, recv)
	case recv.typ.isVarTuple():
		return g.varTupleMethod(n, m, recv)
	case recv.typ.isDict():
		return g.dictMethod(n, m, recv)
	case recv.typ.isSet():
		return g.setMethod(n, m, recv)
	}
	return value{}, unsupported(fmt.Sprintf("method '%s' on %s", m.Attr, recv.typ), n.Line())
}

// methodArgs generates the arguments of a method call, each of which must
// have the given type, in order. A list literal argument takes its type
// from the parameter.
func (g *generator) methodArgs(n *pyast.Call, m *pyast.Attribute, types ...pyType) ([]string, error) {
	if n.HasKeywords {
		return nil, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) != len(types) {
		return nil, unsupported(fmt.Sprintf("%s() with %d arguments, expected %d", m.Attr, len(n.Args), len(types)), n.Line())
	}
	codes := make([]string, len(types))
	for i, a := range n.Args {
		v, err := g.exprAs(a, types[i])
		if err != nil {
			return nil, err
		}
		v, err = g.coerce(v, types[i], fmt.Sprintf("argument %d of %s()", i+1, m.Attr), a.Line())
		if err != nil {
			return nil, err
		}
		codes[i] = v.code
	}
	return codes, nil
}

func (g *generator) strMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	// One entry per method: its name in the runtime, parameter types, result.
	type sig struct {
		fn     string
		params []pyType
		ret    pyType
	}
	methods := map[string][]sig{
		"upper":      {{"StrUpper", nil, tStr}},
		"lower":      {{"StrLower", nil, tStr}},
		"strip":      {{"StrStrip", nil, tStr}, {"StrStripChars", []pyType{tStr}, tStr}},
		"split":      {{"StrSplit", nil, listOf(tStr)}, {"StrSplitSep", []pyType{tStr}, listOf(tStr)}},
		"join":       {{"StrJoin", []pyType{listOf(tStr)}, tStr}},
		"replace":    {{"StrReplace", []pyType{tStr, tStr}, tStr}},
		"find":       {{"StrFind", []pyType{tStr}, tInt}},
		"startswith": {{"StrStartsWith", []pyType{tStr}, tBool}},
		"endswith":   {{"StrEndsWith", []pyType{tStr}, tBool}},
		"count":      {{"StrCount", []pyType{tStr}, tInt}},
	}
	overloads, ok := methods[m.Attr]
	if !ok {
		return value{}, unsupported(fmt.Sprintf("method '%s' on str", m.Attr), n.Line())
	}
	if m.Attr == "join" {
		return g.strJoin(n, m, recv)
	}

	// join is called on the separator: sep.join(parts) is StrJoin(sep, parts).
	for _, s := range overloads {
		if len(s.params) != len(n.Args) {
			continue
		}
		args, err := g.methodArgs(n, m, s.params...)
		if err != nil {
			return value{}, err
		}
		return value{code: call(g.rt(s.fn), append([]string{recv.code}, args...)...), typ: s.ret, prec: precAtom}, nil
	}
	return value{}, unsupported(fmt.Sprintf("%s() with %d arguments", m.Attr, len(n.Args)), n.Line())
}

func (g *generator) listMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	elem := recv.typ.elem()
	r := recv.at(precAtom)
	method := func(name string, ret pyType, params ...pyType) (value, error) {
		args, err := g.methodArgs(n, m, params...)
		if err != nil {
			return value{}, err
		}
		return value{code: r + "." + name + "(" + joinArgs(args) + ")", typ: ret, prec: precAtom}, nil
	}

	switch m.Attr {
	case "append":
		return method("Append", tNone, elem)
	case "pop":
		if len(n.Args) == 0 {
			return method("Pop", elem)
		}
		return method("PopAt", elem, tInt)
	case "insert":
		return method("Insert", tNone, tInt, elem)
	case "extend":
		return method("Extend", tNone, recv.typ)
	case "reverse":
		return method("Reverse", tNone)
	case "index":
		return method("Index", tInt, elem)
	case "count":
		return method("Count", tInt, elem)
	case "remove":
		return method("Remove", tNone, elem)
	case "clear":
		return method("Clear", tNone)
	case "copy":
		return method("Copy", recv.typ)
	case "sort":
		if len(n.Args) > 0 {
			return value{}, unsupported(fmt.Sprintf("sort() with %d arguments", len(n.Args)), n.Line())
		}
		rev, keyNode, err := g.sortKeywords(n, "sort")
		if err != nil {
			return value{}, err
		}
		if keyNode != nil {
			key, keyT, err := g.keyFunc(keyNode, elem)
			if err != nil {
				return value{}, err
			}
			fn := g.rt("SortKey") + "[" + g.goType(elem) + ", " + g.goType(keyT) + "]"
			return value{code: call(fn, recv.code, key, rev), typ: tNone, prec: precAtom}, nil
		}
		return value{code: r + ".Sort(" + rev + ")", typ: tNone, prec: precAtom}, nil
	}
	return value{}, unsupported(fmt.Sprintf("method '%s' on %s", m.Attr, recv.typ), n.Line())
}

func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += ", "
		}
		out += a
	}
	return out
}

// strJoin is sep.join(iterable of str). The runtime joins a list, so another
// iterable, such as a generator expression, is collected first.
func (g *generator) strJoin(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) != 1 {
		return value{}, unsupported(fmt.Sprintf("join() with %d arguments, expected 1", len(n.Args)), n.Line())
	}
	v, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	list, err := g.asList(v, "join()", n.Line())
	if err != nil {
		return value{}, err
	}
	if list.typ != listOf(tStr) {
		return value{}, mismatch("argument 1 of join()", listOf(tStr), list.typ, n.Args[0].Line())
	}
	return value{code: call(g.rt("StrJoin"), recv.code, list.code), typ: tStr, prec: precAtom}, nil
}

// dictClassMethod is dict.fromkeys(keys, value).
func (g *generator) dictClassMethod(n *pyast.Call, m *pyast.Attribute) (value, error) {
	if m.Attr != "fromkeys" {
		return value{}, unsupported(fmt.Sprintf("method '%s' of dict", m.Attr), n.Line())
	}
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) != 2 {
		return value{}, unsupported(fmt.Sprintf("dict.fromkeys() with %d arguments (the value is required)", len(n.Args)), n.Line())
	}
	keys, elem, err := g.iterableArg(n.Args[0], "dict.fromkeys()")
	if err != nil {
		return value{}, err
	}
	if err := g.hashableOrError(elem, "dict key", n.Line()); err != nil {
		return value{}, err
	}
	val, err := g.typedExpr(n.Args[1])
	if err != nil {
		return value{}, err
	}
	elem = g.keyType(elem)
	t := dictOf(elem, val.typ)
	return value{code: call(g.rt("FromKeys")+"["+g.dictGoArgs(t)+"]", keys, val.code), typ: t, prec: precAtom}, nil
}

func (g *generator) dictMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	k, v := recv.typ.keyType(), recv.typ.valType()
	r := recv.at(precAtom)
	method := func(name string, ret pyType, params ...pyType) (value, error) {
		args, err := g.methodArgs(n, m, params...)
		if err != nil {
			return value{}, err
		}
		return value{code: r + "." + name + "(" + joinArgs(args) + ")", typ: ret, prec: precAtom}, nil
	}

	switch m.Attr {
	case "get":
		if len(n.Args) == 1 {
			return method("GetOpt", optionalOf(v), k)
		}
		if c, ok := n.Args[1].(*pyast.Constant); ok && c.PyType == "none" {
			args, err := g.methodArgs(&pyast.Call{Func: n.Func, Args: n.Args[:1], Lineno: n.Line()}, m, k)
			if err != nil {
				return value{}, err
			}
			return value{code: r + ".GetOpt(" + joinArgs(args) + ")", typ: optionalOf(v), prec: precAtom}, nil
		}
		return method("GetDefault", v, k, v)
	case "pop":
		if len(n.Args) == 1 {
			return method("Pop", v, k)
		}
		return method("PopDefault", v, k, v)
	case "popitem":
		return method("PopItem", tupleOf([]pyType{k, v}))
	case "setdefault":
		if len(n.Args) == 1 {
			return value{}, unsupported("setdefault() with one argument, which stores None", n.Line())
		}
		return method("SetDefault", v, k, v)
	case "clear":
		return method("Clear", tNone)
	case "copy":
		return method("Copy", recv.typ)
	case "keys":
		return method("Keys", pyType("dict_keys["+string(k)+", "+string(v)+"]"))
	case "values":
		return method("Values", pyType("dict_values["+string(k)+", "+string(v)+"]"))
	case "items":
		return method("Items", pyType("dict_items["+string(k)+", "+string(v)+"]"))
	case "update":
		return g.dictUpdate(n, m, recv)
	}
	return value{}, unsupported(fmt.Sprintf("method '%s' on %s", m.Attr, recv.typ), n.Line())
}

// dictUpdate is d.update(other_dict) and d.update(iterable of pairs).
func (g *generator) dictUpdate(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	if n.HasKeywords {
		return value{}, unsupported("keyword arguments", n.Line())
	}
	if len(n.Args) != 1 {
		return value{}, unsupported(fmt.Sprintf("update() with %d arguments, expected 1", len(n.Args)), n.Line())
	}
	arg, err := g.exprAs(n.Args[0], recv.typ)
	if err != nil {
		return value{}, err
	}
	r := recv.at(precAtom)
	if arg.typ == recv.typ {
		return value{code: r + ".Update(" + arg.code + ")", typ: tNone, prec: precAtom}, nil
	}
	pair := tupleOf([]pyType{recv.typ.keyType(), recv.typ.valType()})
	if code, elem, ok := g.iterableOf(arg); ok && elem == pair {
		return value{code: r + ".UpdatePairs(" + code + ")", typ: tNone, prec: precAtom}, nil
	}
	return value{}, mismatch("argument 1 of update()", recv.typ, arg.typ, n.Args[0].Line())
}

func (g *generator) setMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	elem := recv.typ.elem()
	r := recv.at(precAtom)
	method := func(name string, ret pyType, params ...pyType) (value, error) {
		args, err := g.methodArgs(n, m, params...)
		if err != nil {
			return value{}, err
		}
		return value{code: r + "." + name + "(" + joinArgs(args) + ")", typ: ret, prec: precAtom}, nil
	}
	// The set-valued methods take any iterable of the element type.
	others := func(name string, ret pyType, min, max int) (value, error) {
		if n.HasKeywords {
			return value{}, unsupported("keyword arguments", n.Line())
		}
		if len(n.Args) < min || len(n.Args) > max {
			return value{}, unsupported(fmt.Sprintf("%s() with %d arguments", m.Attr, len(n.Args)), n.Line())
		}
		args := make([]string, len(n.Args))
		for i, a := range n.Args {
			v, err := g.exprAs(a, recv.typ)
			if err != nil {
				return value{}, err
			}
			code, e, ok := g.iterableOf(v)
			if !ok || e != elem {
				return value{}, mismatch(fmt.Sprintf("argument %d of %s()", i+1, m.Attr), recv.typ, v.typ, a.Line())
			}
			args[i] = code
		}
		return value{code: r + "." + name + "(" + joinArgs(args) + ")", typ: ret, prec: precAtom}, nil
	}

	switch m.Attr {
	case "add":
		return method("Add", tNone, elem)
	case "remove":
		return method("Remove", tNone, elem)
	case "discard":
		return method("Discard", tNone, elem)
	case "pop":
		return method("Pop", elem)
	case "clear":
		return method("Clear", tNone)
	case "copy":
		return method("Copy", recv.typ)
	case "union":
		return others("Union", recv.typ, 0, maxTupleArity*100)
	case "intersection":
		return others("Intersection", recv.typ, 0, maxTupleArity*100)
	case "difference":
		return others("Difference", recv.typ, 0, maxTupleArity*100)
	case "symmetric_difference":
		return others("SymmetricDifference", recv.typ, 1, 1)
	case "update":
		return others("Update", tNone, 0, maxTupleArity*100)
	case "intersection_update":
		return others("IntersectionUpdate", tNone, 0, maxTupleArity*100)
	case "difference_update":
		return others("DifferenceUpdate", tNone, 0, maxTupleArity*100)
	case "symmetric_difference_update":
		return others("SymmetricDifferenceUpdate", tNone, 1, 1)
	case "issubset":
		return others("IsSubset", tBool, 1, 1)
	case "issuperset":
		return others("IsSuperset", tBool, 1, 1)
	case "isdisjoint":
		return others("IsDisjoint", tBool, 1, 1)
	}
	return value{}, unsupported(fmt.Sprintf("method '%s' on %s", m.Attr, recv.typ), n.Line())
}

func (g *generator) varTupleMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	elem := recv.typ.elem()
	r := recv.at(precAtom)
	method := func(name string, ret pyType, params ...pyType) (value, error) {
		args, err := g.methodArgs(n, m, params...)
		if err != nil {
			return value{}, err
		}
		return value{code: r + "." + name + "(" + joinArgs(args) + ")", typ: ret, prec: precAtom}, nil
	}
	switch m.Attr {
	case "count":
		return method("Count", tInt, elem)
	case "index":
		return method("Index", tInt, elem)
	}
	return value{}, unsupported(fmt.Sprintf("method '%s' on %s", m.Attr, recv.typ), n.Line())
}
