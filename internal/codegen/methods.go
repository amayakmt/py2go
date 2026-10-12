package codegen

import (
	"fmt"

	"github.com/amayakmt/py2go/internal/pyast"
)

// methodCall generates "receiver.name(args)" for the str and list methods
// the transpiler knows.
func (g *generator) methodCall(n *pyast.Call, m *pyast.Attribute) (value, error) {
	recv, err := g.typedExpr(m.Value)
	if err != nil {
		return value{}, err
	}
	switch {
	case recv.typ == tStr:
		return g.strMethod(n, m, recv)
	case recv.typ.isList():
		return g.listMethod(n, m, recv)
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
		if v.typ != types[i] {
			return nil, mismatch(fmt.Sprintf("argument %d of %s()", i+1, m.Attr), types[i], v.typ, a.Line())
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
		rev, err := g.reverseKeyword(n, "sort")
		if err != nil {
			return value{}, err
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
