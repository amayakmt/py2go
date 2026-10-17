package codegen

import (
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// narrowing says that, in some region of code, a variable has a more precise
// type than it was declared with: after isinstance(x, C), or after a test
// that x is not None.
type narrowing struct {
	name string
	to   pyType
	expr string // Go code of the narrowed value
	// alias narrowings give the name a new meaning in a scope without
	// declaring a variable, so assignments still reach the real one.
	alias bool
}

// aliasInfo is how a name reads in the scope of an alias narrowing.
type aliasInfo struct {
	code string
	typ  pyType
}

// bindNarrow makes the narrowings visible in the current scope and writes the
// declarations they need to b.
func (g *generator) bindNarrow(b *strings.Builder, nars []*narrowing) {
	for _, n := range nars {
		if n.alias {
			g.env.setAlias(n.name, n.expr, n.to)
			continue
		}
		g.declareLocal(b, n.name, n.to, n.expr)
	}
}

// bindNarrowExpr is bindNarrow for an expression, whose declarations go in a
// function literal. It returns them.
func (g *generator) bindNarrowExpr(nars []*narrowing) string {
	var b strings.Builder
	for _, n := range nars {
		if n.alias {
			g.env.setAlias(n.name, n.expr, n.to)
			continue
		}
		g.env.declare(n.name, n.to)
		b.WriteString("var " + n.name + " " + g.goType(n.to) + " = " + n.expr + "\n_ = " + n.name + "\n")
	}
	return b.String()
}

// narrowTest reads what a test tells about variables: pos holds when the test
// is true, neg when it is false.
func (g *generator) narrowTest(test pyast.Node) (pos, neg []*narrowing) {
	switch n := test.(type) {
	case *pyast.UnaryOp:
		if n.Op == "Not" {
			p, q := g.narrowTest(n.Operand)
			return q, p
		}
	case *pyast.BoolOp:
		var all []*narrowing
		for _, v := range n.Values {
			p, q := g.narrowTest(v)
			if n.Op == "And" {
				all = append(all, p...)
			} else {
				all = append(all, q...)
			}
		}
		if n.Op == "And" {
			return all, nil
		}
		return nil, all
	case *pyast.Name:
		if nar := g.someNarrowing(n); nar != nil {
			return []*narrowing{nar}, nil
		}
	case *pyast.Compare:
		if len(n.Ops) != 1 {
			return nil, nil
		}
		other, isNone := noneOperand(n)
		if !isNone {
			return nil, nil
		}
		name, ok := other.(*pyast.Name)
		if !ok {
			return nil, nil
		}
		nar := g.someNarrowing(name)
		if nar == nil {
			return nil, nil
		}
		switch n.Ops[0] {
		case "IsNot", "NotEq":
			return []*narrowing{nar}, nil
		case "Is", "Eq":
			return nil, []*narrowing{nar}
		}
	case *pyast.Call:
		if nar := g.isinstanceNarrowing(n); nar != nil {
			return []*narrowing{nar}, nil
		}
	}
	return nil, nil
}

// noneOperand reports which side of a comparison is None, and returns the other.
func noneOperand(n *pyast.Compare) (pyast.Node, bool) {
	isNone := func(x pyast.Node) bool {
		c, ok := x.(*pyast.Constant)
		return ok && c.PyType == "none"
	}
	switch {
	case len(n.Comparators) == 1 && isNone(n.Comparators[0]):
		return n.Left, true
	case len(n.Comparators) == 1 && isNone(n.Left):
		return n.Comparators[0], true
	}
	return nil, false
}

// variableType finds the type of a plain variable and the Go code that reads
// it, for narrowing.
func (g *generator) variableType(name string) (pyType, string, bool) {
	if g.fnGlobals[name] {
		return "", "", false
	}
	if a, ok := g.env.lookupAlias(name); ok {
		return a.typ, a.code, true
	}
	if t, ok := g.env.lookup(name); ok {
		return t, name, true
	}
	if v, ok := g.readGlobal(name); ok {
		return v.typ, v.code, true
	}
	return "", "", false
}

// someNarrowing is the narrowing "x is not None" for a variable that is optional.
func (g *generator) someNarrowing(n *pyast.Name) *narrowing {
	t, src, ok := g.variableType(n.ID)
	if !ok || !t.isOptional() {
		return nil
	}
	return &narrowing{name: n.ID, to: t.optElem(), expr: g.optionalValue(src), alias: true}
}

// optionalValue reads the value of an Optional known not to be None. The
// reading checks anyway, since an assignment can make the knowledge stale.
func (g *generator) optionalValue(src string) string { return src + ".Get()" }

// isinstanceNarrowing reads "isinstance(x, C)".
func (g *generator) isinstanceNarrowing(call *pyast.Call) *narrowing {
	if call.HasKeywords || len(call.Args) != 2 {
		return nil
	}
	if name, _ := callName(call); name != "isinstance" || g.isVariable("isinstance") {
		return nil
	}
	arg, ok := call.Args[0].(*pyast.Name)
	if !ok {
		return nil
	}
	spec, ok := call.Args[1].(*pyast.Name)
	if !ok || g.isVariable(spec.ID) {
		return nil
	}
	cur, src, found := g.variableType(arg.ID)
	if !found {
		return nil
	}
	if k := g.classes[spec.ID]; k != nil {
		if k.isABC() {
			return nil // a registered virtual subclass is not a k
		}
		switch {
		case cur == tObject:
			return &narrowing{name: arg.ID, to: k.typ(), expr: src + ".(" + k.name + ")"}
		case cur.isClass() && g.classOf(cur) != k && k.isSubclassOf(g.classOf(cur)):
			return &narrowing{name: arg.ID, to: k.typ(), expr: src + ".(" + k.name + ")"}
		case cur == tExc && k.exc:
			return &narrowing{name: arg.ID, to: k.typ(), expr: src + ".Obj.(" + k.name + ")"}
		}
		return nil
	}
	if cur == tObject {
		switch spec.ID {
		case "int":
			if g.bigMode {
				return nil // the int may be beyond 64 bits: it stays an object
			}
			return &narrowing{name: arg.ID, to: tInt, expr: g.rt("AsInt") + "(" + src + ")"}
		case "str":
			return &narrowing{name: arg.ID, to: tStr, expr: g.rt("Cast") + "[string](" + src + ")"}
		case "float":
			return &narrowing{name: arg.ID, to: tFloat, expr: g.rt("Cast") + "[float64](" + src + ")"}
		case "bool":
			return &narrowing{name: arg.ID, to: tBool, expr: src + ".(bool)"}
		}
	}
	return nil
}

// earlyExitNarrowing recognises "if test: <leave>" without an else, after
// which the rest of the block knows the test failed.
func (g *generator) earlyExitNarrowing(s pyast.Node) []*narrowing {
	n, ok := s.(*pyast.If)
	if !ok || len(n.Orelse) > 0 || len(n.Body) == 0 {
		return nil
	}
	switch n.Body[len(n.Body)-1].(type) {
	case *pyast.Return, *pyast.Raise, *pyast.Continue, *pyast.Break:
	default:
		return nil
	}
	_, neg := g.narrowTest(n.Test)
	return neg
}

// ---- scopes ----

func (s *scope) setAlias(name, code string, t pyType) {
	if s.aliases == nil {
		s.aliases = map[string]aliasInfo{}
	}
	s.aliases[name] = aliasInfo{code, t}
}

// lookupAlias finds an alias narrowing of name that is still in force: the
// nearest scope that binds the name or narrows it decides.
func (s *scope) lookupAlias(name string) (aliasInfo, bool) {
	for ; s != nil; s = s.parent {
		if a, ok := s.aliases[name]; ok {
			return a, true
		}
		if _, ok := s.vars[name]; ok {
			return aliasInfo{}, false
		}
	}
	return aliasInfo{}, false
}

// dropAlias ends the narrowing of name, after an assignment to it.
func (s *scope) dropAlias(name string) {
	for ; s != nil; s = s.parent {
		delete(s.aliases, name)
	}
}

// dropAssigned removes the narrowings of names that the code they would hold
// for assigns: the narrowed name is a copy, so an assignment to it would be
// lost, or would not have the narrowed type.
func dropAssigned(nars []*narrowing, body []pyast.Node) []*narrowing {
	if len(nars) == 0 {
		return nars
	}
	var out []*narrowing
	for _, n := range nars {
		if !assignsName(body, n.name) {
			out = append(out, n)
		}
	}
	return out
}
