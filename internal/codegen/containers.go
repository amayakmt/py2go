package codegen

import (
	"fmt"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// dictLit generates {k: v, ...}. The type comes from the first pair, or from
// want for an empty dict.
func (g *generator) dictLit(n *pyast.Dict, want pyType) (value, error) {
	var keyWant, valWant pyType
	if want.isDict() {
		keyWant, valWant = want.keyType(), want.valType()
	}
	if len(n.Keys) == 0 {
		if !want.isDict() {
			return value{}, unsupported("empty dict without a type annotation", n.Line())
		}
		return value{code: g.rt("NewDict") + "[" + g.dictGoArgs(want) + "]()", typ: want, prec: precAtom}, nil
	}

	var key, val pyType
	type entry struct {
		k, v     string
		unpacked bool
	}
	entries := make([]entry, 0, len(n.Keys))
	spread := false
	for i, kn := range n.Keys {
		if kn == nil { // {**other}
			v, err := g.exprAs(n.Values[i], want)
			if err != nil {
				return value{}, err
			}
			if !v.typ.isDict() {
				return value{}, unsupported("'**' of "+string(v.typ)+" in a dict display", n.Values[i].Line())
			}
			if key == "" {
				key, val = v.typ.keyType(), v.typ.valType()
			} else if v.typ != dictOf(key, val) {
				return value{}, unsupported(fmt.Sprintf("dict display mixes %s and %s", dictOf(key, val), v.typ), n.Values[i].Line())
			}
			entries = append(entries, entry{k: v.code, unpacked: true})
			spread = true
			continue
		}
		k, err := g.exprAs(kn, keyWant)
		if err != nil {
			return value{}, err
		}
		v, err := g.exprAs(n.Values[i], valWant)
		if err != nil {
			return value{}, err
		}
		if k.typ == tExc || v.typ == tExc {
			return value{}, unsupported("exception stored in a dict", kn.Line())
		}
		if key == "" {
			if !k.typ.hashable() {
				return value{}, unsupported(fmt.Sprintf("dict key of type %s (not hashable)", k.typ), kn.Line())
			}
			key, val = k.typ, v.typ
		} else if k.typ != key || v.typ != val {
			return value{}, unsupported(fmt.Sprintf("dict display mixes key and value types (%s: %s and %s: %s)", key, val, k.typ, v.typ), kn.Line())
		}
		entries = append(entries, entry{k: k.code, v: v.code})
	}
	t := dictOf(key, val)
	args := g.dictGoArgs(t)

	if !spread {
		pairs := make([]string, len(entries))
		for i, e := range entries {
			pairs[i] = fmt.Sprintf("%s[%s]{Key: %s, Val: %s}", g.rt("KV"), args, e.k, e.v)
		}
		return value{code: g.rt("NewDict") + "[" + args + "](" + strings.Join(pairs, ", ") + ")", typ: t, prec: precAtom}, nil
	}

	// With "**" the entries are applied in order to a new dict.
	var b strings.Builder
	d := g.tmp("d")
	fmt.Fprintf(&b, "func() %s {\nvar %s = %s[%s]()\n", g.goType(t), d, g.rt("NewDict"), args)
	for _, e := range entries {
		if e.unpacked {
			fmt.Fprintf(&b, "%s.Update(%s)\n", d, e.k)
		} else {
			fmt.Fprintf(&b, "%s.Set(%s, %s)\n", d, e.k, e.v)
		}
	}
	fmt.Fprintf(&b, "return %s\n}()", d)
	return value{code: b.String(), typ: t, prec: precAtom}, nil
}

// dictGoArgs spells the type arguments of a dict: "string, int64".
func (g *generator) dictGoArgs(t pyType) string {
	return g.goType(t.keyType()) + ", " + g.goType(t.valType())
}

// setLit generates {a, b, c}.
func (g *generator) setLit(n *pyast.Set, want pyType) (value, error) {
	var elemWant pyType
	if want.isSet() {
		elemWant = want.elem()
	}
	var elem pyType
	codes := make([]string, len(n.Elts))
	for i, e := range n.Elts {
		v, err := g.exprAs(e, elemWant)
		if err != nil {
			return value{}, err
		}
		if i == 0 {
			if !v.typ.hashable() {
				return value{}, unsupported(fmt.Sprintf("set element of type %s (not hashable)", v.typ), e.Line())
			}
			elem = v.typ
		} else if v.typ != elem {
			return value{}, unsupported(fmt.Sprintf("set with mixed element types (%s and %s)", elem, v.typ), e.Line())
		}
		codes[i] = v.code
	}
	ctor := "NewSet"
	if len(n.Elts) > 2 && allConstants(n.Elts) {
		ctor = "NewSetOfConstants"
	}
	return value{
		code: g.rt(ctor) + "[" + g.goType(elem) + "](" + strings.Join(codes, ", ") + ")",
		typ:  setOf(elem), prec: precAtom,
	}, nil
}

// allConstants reports whether every node is a literal that CPython's
// compiler can fold into a constant: numbers, strings, signed or combined
// literals, and tuples of those.
func allConstants(nodes []pyast.Node) bool {
	for _, n := range nodes {
		if !isConstant(n) {
			return false
		}
	}
	return true
}

func isConstant(n pyast.Node) bool {
	switch n := n.(type) {
	case *pyast.Constant:
		return true
	case *pyast.UnaryOp:
		return n.Op != "Not" && isConstant(n.Operand)
	case *pyast.BinOp:
		return isConstant(n.Left) && isConstant(n.Right)
	case *pyast.Tuple:
		return allConstants(n.Elts)
	}
	return false
}

// emptyConstructor handles list(), set() and dict() with no arguments, which
// take their type from the context like the literals [] and {} do.
func (g *generator) emptyConstructor(n *pyast.Call, want pyType) (value, bool) {
	name, ok := callName(n)
	if !ok || len(n.Args) != 0 || n.HasKeywords || g.isVariable(name) {
		return value{}, false
	}
	if _, shadowed := g.sigs[name]; shadowed {
		return value{}, false
	}
	switch {
	case name == "list" && want.isList():
		return value{code: g.rt("NewList") + "[" + g.goType(want.elem()) + "]()", typ: want, prec: precAtom}, true
	case name == "tuple" && want.isVarTuple():
		return value{code: g.rt("NewVarTuple") + "[" + g.goType(want.elem()) + "]()", typ: want, prec: precAtom}, true
	case name == "set" && want.isSet():
		return value{code: g.rt("NewSet") + "[" + g.goType(want.elem()) + "]()", typ: want, prec: precAtom}, true
	case name == "dict" && want.isDict():
		return value{code: g.rt("NewDict") + "[" + g.dictGoArgs(want) + "]()", typ: want, prec: precAtom}, true
	}
	return value{}, false
}

// hashableOrError rejects a type that cannot be a dict key or set element.
func hashableOrError(t pyType, what string, line int) error {
	if t.hashable() {
		return nil
	}
	return unsupported(fmt.Sprintf("%s of type %s (not hashable)", what, t), line)
}

// iterableOf describes v as a runtime.Iterable: the Go expression and the
// type of the items it yields. A str iterates as a list of its characters.
func (g *generator) iterableOf(v value) (code string, elem pyType, ok bool) {
	elem, ok = v.typ.iterElem()
	if !ok {
		return "", "", false
	}
	if v.typ == tStr {
		return g.rt("Chars") + "(" + v.code + ")", elem, true
	}
	return v.code, elem, true
}

// asList converts any iterable value into a list. Lists are returned as
// they are, so callers that only read need no copy.
func (g *generator) asList(v value, what string, line int) (value, error) {
	if v.typ.isList() {
		return v, nil
	}
	code, elem, ok := g.iterableOf(v)
	if !ok {
		return value{}, unsupported(fmt.Sprintf("%s of %s", what, v.typ), line)
	}
	return value{code: call(g.rt("ListFrom")+"["+g.goType(elem)+"]", code), typ: listOf(elem), prec: precAtom}, nil
}

// iterSource is a for loop's walk over a sequence: a Go expression that
// creates a runtime.Cursor, and the type of the items it yields.
type iterSource struct {
	cursor string
	typ    pyType
}

func (g *generator) iterOf(node pyast.Node) (iterSource, error) {
	v, err := g.typedExpr(node)
	if err != nil {
		return iterSource{}, err
	}
	code, elem, ok := g.iterableOf(v)
	if !ok {
		return iterSource{}, unsupported(fmt.Sprintf("for loop over %s", v.typ), node.Line())
	}
	return iterSource{cursor: code + ".Iter()", typ: elem}, nil
}
