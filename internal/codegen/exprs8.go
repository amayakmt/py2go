package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// ifExp is "a if test else b". Only the branch that is taken runs, so it is
// a function literal with an if.
func (g *generator) ifExp(n *pyast.IfExp, want pyType) (value, error) {
	cond, err := g.condition(n.Test)
	if err != nil {
		return value{}, err
	}
	pos, neg := g.narrowTest(n.Test)
	branch := func(node pyast.Node, nars []*narrowing) (value, string, error) {
		outer := g.env
		g.env = newScope(outer)
		defer func() { g.env = outer }()
		decls := g.bindNarrowExpr(nars)
		v, err := g.exprAs(node, want)
		return v, decls, err
	}
	a, adecl, err := branch(n.Body, pos)
	if err != nil {
		return value{}, err
	}
	b, bdecl, err := branch(n.Orelse, neg)
	if err != nil {
		return value{}, err
	}
	a, b = a.typed(), b.typed()
	t := a.typ
	if a.typ != b.typ {
		var ok bool
		if t, ok = g.joinTypes(a.typ, b.typ); !ok {
			return value{}, unsupported(fmt.Sprintf("conditional expression with branches of type %s and %s", a.typ, b.typ), n.Line())
		}
		if a, err = g.coerce(a, t, "conditional expression", n.Line()); err != nil {
			return value{}, err
		}
		if b, err = g.coerce(b, t, "conditional expression", n.Line()); err != nil {
			return value{}, err
		}
	}
	code := fmt.Sprintf("func() %s {\nif %s {\n%sreturn %s\n}\n%sreturn %s\n}()", g.goType(t), cond.code, adecl, a.code, bdecl, b.code)
	return value{code: code, typ: t, prec: precAtom}, nil
}

// joinTypes finds the type two branches share: the nearest common class.
func (g *generator) joinTypes(a, b pyType) (pyType, bool) {
	if a == b {
		return a, true
	}
	switch {
	case a == tNoneType && b != tNoneType:
		return optionalOf(b), true
	case b == tNoneType:
		return optionalOf(a), true
	case a.isOptional() && g.isSubtype(b, a):
		return a, true
	case b.isOptional() && g.isSubtype(a, b):
		return b, true
	case a.isOptional() || b.isOptional():
		inner, ok := g.joinTypes(strip(a), strip(b))
		return optionalOf(inner), ok
	}
	if t, ok := g.commonClass(a, b); ok {
		return t, true
	}
	return "", false
}

// chainedCompare is "a < b < c": each operand is evaluated once, and the
// chain stops at the first comparison that fails.
func (g *generator) chainedCompare(n *pyast.Compare) (value, error) {
	nodes := append([]pyast.Node{n.Left}, n.Comparators...)
	outer := g.env
	g.env = newScope(outer)
	defer func() { g.env = outer }()

	var b strings.Builder
	b.WriteString("func() bool {\n")
	names := make([]*pyast.Name, len(nodes))
	declare := func(i int) error {
		v, err := g.typedExpr(nodes[i])
		if err != nil {
			return err
		}
		id := g.tmp("c")
		g.env.declare(id, v.typ)
		names[i] = &pyast.Name{ID: id, Lineno: n.Line()}
		fmt.Fprintf(&b, "var %s %s = %s\n_ = %s\n", id, g.goType(v.typ), v.typed().code, id)
		return nil
	}
	if err := declare(0); err != nil {
		return value{}, err
	}
	for i, op := range n.Ops {
		if err := declare(i + 1); err != nil {
			return value{}, err
		}
		pair, err := g.compare(&pyast.Compare{Left: names[i], Ops: []string{op}, Comparators: []pyast.Node{names[i+1]}, Lineno: n.Line()})
		if err != nil {
			return value{}, err
		}
		if i == len(n.Ops)-1 {
			fmt.Fprintf(&b, "return %s\n", pair.code)
		} else {
			fmt.Fprintf(&b, "if !(%s) {\nreturn false\n}\n", pair.code)
		}
	}
	b.WriteString("}()")
	return value{code: b.String(), typ: tBool, prec: precAtom}, nil
}

// namedExpr is "(name := value)": it assigns the variable and has its value.
func (g *generator) namedExpr(n *pyast.NamedExpr) (value, error) {
	target, ok := n.Target.(*pyast.Name)
	if !ok {
		return value{}, unsupported("assignment expression target that is not a name", n.Line())
	}
	if err := checkName(target.ID, n.Line()); err != nil {
		return value{}, err
	}
	// In a comprehension the variable belongs to the scope around it.
	inComp := g.compRoot != nil
	var t pyType
	var declared, global bool
	if inComp {
		if t, declared = g.compRoot.parent.lookup(target.ID); !declared {
			if gt, ok := g.globalTypes[target.ID]; ok && g.fnName != "" {
				return value{}, unsupported(fmt.Sprintf("assignment expression to '%s', a module variable, inside a comprehension (type %s)", target.ID, gt), n.Line())
			}
		}
	} else {
		t, declared, global = g.lookupBinding(target.ID)
	}
	var v value
	var err error
	if declared {
		if v, err = g.exprAs(n.Value, t); err != nil {
			return value{}, err
		}
		if v, err = g.coerce(v, t, fmt.Sprintf("variable '%s'", target.ID), n.Line()); err != nil {
			return value{}, err
		}
	} else {
		if v, err = g.typedExpr(n.Value); err != nil {
			return value{}, err
		}
		t = v.typ
		if t == tExc || t.isClassRef() {
			return value{}, unsupported("assignment expression of "+string(t), n.Line())
		}
		if g.hoist == nil {
			return value{}, unsupported("assignment expression outside a statement", n.Line())
		}
		if inComp {
			if g.fnName == "" {
				g.declareGlobal(g.hoist, target.ID, t, "")
			} else {
				g.compRoot.parent.declare(target.ID, t)
				fmt.Fprintf(g.hoist, "var %s %s\n_ = %s\n", target.ID, g.goType(t), target.ID)
			}
		} else {
			g.declare(g.hoist, target.ID, t, "")
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "func() %s {\n%s = %s\n", g.goType(t), target.ID, v.typed().code)
	g.afterAssign(&b, target.ID, global)
	if !declared && g.atModuleLevel() {
		g.markGlobalSet(&b, target.ID)
	}
	fmt.Fprintf(&b, "return %s\n}()", target.ID)
	return value{code: b.String(), typ: t, prec: precAtom}, nil
}

func strip(t pyType) pyType {
	if t.isOptional() {
		return t.optElem()
	}
	return t
}

// unwrapOperands lets an Optional take part in an operator. Python fails when
// the operand turns out to be None, so the generated code does too.
func (g *generator) unwrapOperands(msg func(lname, rname string) string, l, r value) (value, value) {
	lname, rname := g.typeNameOf(strip(l.typ)), g.typeNameOf(strip(r.typ))
	if l.typ.isOptional() {
		l = value{code: l.at(precAtom) + ".Unwrap(\"TypeError\", " + strconv.Quote(msg("NoneType", rname)) + ")", typ: l.typ.optElem(), prec: precAtom}
	}
	if r.typ.isOptional() {
		r = value{code: r.at(precAtom) + ".Unwrap(\"TypeError\", " + strconv.Quote(msg(lname, "NoneType")) + ")", typ: r.typ.optElem(), prec: precAtom}
	}
	return l, r
}

// genInfo describes a generator function: what it yields, is sent and
// returns. iterator marks one annotated Iterator[T] or Iterable[T], whose
// caller gets the plain iterator type.
type genInfo struct {
	yield, send, ret pyType
	iterator         bool
}

const yielderVar = "_pyy"

// isGeneratorBody reports whether a function body contains a yield, not
// counting the functions nested in it.
func isGeneratorBody(body []pyast.Node) bool {
	found := false
	for _, st := range body {
		walkPrune(st, func(n pyast.Node) bool {
			switch n.(type) {
			case *pyast.Yield, *pyast.YieldFrom:
				found = true
			case *pyast.FunctionDef, *pyast.Lambda, *pyast.ClassDef:
				return false
			}
			return !found
		})
	}
	return found
}

// yieldExpr is "yield v". It has the value that was sent in.
func (g *generator) yieldExpr(n *pyast.Yield) (value, error) {
	if g.gen == nil {
		return value{}, unsupported("'yield' outside function", n.Line())
	}
	if n.Value == nil {
		if g.gen.yield != tNoneType {
			return value{}, unsupported("yield without a value in a generator of "+string(g.gen.yield), n.Line())
		}
		return value{code: yielderVar + ".Yield(" + g.rt("NoneV") + ")", typ: g.gen.send, prec: precAtom}, nil
	}
	v, err := g.exprAs(n.Value, g.gen.yield)
	if err != nil {
		return value{}, err
	}
	v, err = g.coerce(v, g.gen.yield, "yielded value", n.Line())
	if err != nil {
		return value{}, err
	}
	return value{code: yielderVar + ".Yield(" + v.typed().code + ")", typ: g.gen.send, prec: precAtom}, nil
}

// yieldFrom is "yield from x": a generator is delegated to, and any other
// iterable is walked.
func (g *generator) yieldFrom(n *pyast.YieldFrom) (value, error) {
	if g.gen == nil {
		return value{}, unsupported("'yield from' outside function", n.Line())
	}
	v, err := g.typedExpr(n.Value)
	if err != nil {
		return value{}, err
	}
	if v.typ.isGen() {
		a := v.typ.typeArgs()
		if a[0] != g.gen.yield || a[1] != g.gen.send {
			return value{}, unsupported(fmt.Sprintf("yield from %s in a generator of %s sent %s", v.typ, g.gen.yield, g.gen.send), n.Line())
		}
		return value{code: g.rt("YieldFromGen") + "(" + yielderVar + ", " + v.code + ")", typ: a[2], prec: precAtom}, nil
	}
	code, elem, ok := g.iterableOf(v)
	if !ok || elem != g.gen.yield {
		return value{}, unsupported(fmt.Sprintf("yield from %s in a generator of %s", v.typ, g.gen.yield), n.Line())
	}
	return value{code: g.rt("YieldFromIter") + "(" + yielderVar + ", " + code + ")", typ: tNoneType, prec: precAtom}, nil
}

// genMethod is send, throw, close and __next__ of a generator.
func (g *generator) genMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	a := recv.typ.typeArgs()
	r := recv.at(precAtom)
	switch m.Attr {
	case "send":
		args, err := g.methodArgs(n, m, a[1])
		if err != nil {
			return value{}, err
		}
		return value{code: r + ".Send(" + args[0] + ")", typ: a[0], prec: precAtom}, nil
	case "throw":
		if len(n.Args) != 1 || n.HasKeywords {
			return value{}, unsupported("throw() with something other than one exception", n.Line())
		}
		exc, err := g.excValue(n.Args[0])
		if err != nil {
			return value{}, err
		}
		return value{code: r + ".Throw(" + exc + ")", typ: a[0], prec: precAtom}, nil
	case "close":
		return value{code: r + ".Close()", typ: tNone, prec: precAtom}, nil
	case "__next__":
		return value{code: r + ".NextValue()", typ: a[0], prec: precAtom}, nil
	}
	return value{}, unsupported(fmt.Sprintf("method '%s' on a generator", m.Attr), n.Line())
}

// iterGenMethod is close and throw of the iterator a generator function
// annotated Iterator[T] makes; it is a generator all the same.
func (g *generator) iterGenMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	r := recv.at(precAtom)
	switch m.Attr {
	case "close":
		return value{code: r + ".Close()", typ: tNone, prec: precAtom}, nil
	case "throw":
		if len(n.Args) != 1 || n.HasKeywords {
			return value{}, unsupported("throw() with something other than one exception", n.Line())
		}
		exc, err := g.excValue(n.Args[0])
		if err != nil {
			return value{}, err
		}
		return value{code: r + ".Throw(" + exc + ")", typ: recv.typ.elem(), prec: precAtom}, nil
	case "__next__":
		return value{code: r + ".NextValue()", typ: recv.typ.elem(), prec: precAtom}, nil
	}
	return value{}, unsupported(fmt.Sprintf("method '%s' on %s", m.Attr, recv.typ), n.Line())
}

// nextCall is next(it) and next(it, default).
func (g *generator) nextCall(n *pyast.Call) (value, error) {
	if n.HasKeywords || len(n.Args) < 1 || len(n.Args) > 2 {
		return value{}, unsupported(fmt.Sprintf("next() with %d arguments", len(n.Args)), n.Line())
	}
	it, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	var elem pyType
	var plain string
	switch {
	case it.typ.isGen(), it.typ.isIter():
		elem = it.typ.elem()
		plain = it.at(precAtom) + ".NextValue()"
	default:
		if c := g.classOf(it.typ); c != nil {
			if m := c.lookup("__next__"); m != nil && m.kind == mMethod && len(m.sig.pos) == 0 && len(n.Args) == 1 {
				return value{code: it.at(precAtom) + ".M___next__()", typ: m.sig.ret, prec: precAtom}, nil
			}
		}
		return value{}, unsupported(fmt.Sprintf("next() of %s, which is not an iterator", it.typ), n.Line())
	}
	if len(n.Args) == 1 {
		return value{code: plain, typ: elem, prec: precAtom}, nil
	}
	if c, ok := n.Args[1].(*pyast.Constant); ok && c.PyType == "none" {
		return value{code: g.rt("NextOpt") + "[" + g.goType(elem) + "](" + it.code + ")", typ: optionalOf(elem), prec: precAtom}, nil
	}
	def, err := g.exprAs(n.Args[1], elem)
	if err != nil {
		return value{}, err
	}
	def, err = g.coerce(def, elem, "default of next()", n.Line())
	if err != nil {
		return value{}, err
	}
	return value{code: g.rt("NextOr") + "[" + g.goType(elem) + "](" + it.code + ", " + def.code + ")", typ: elem, prec: precAtom}, nil
}

// iterCall is iter(x).
func (g *generator) iterCall(n *pyast.Call) (value, error) {
	if n.HasKeywords || len(n.Args) != 1 {
		return value{}, unsupported(fmt.Sprintf("iter() with %d arguments", len(n.Args)), n.Line())
	}
	x, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	if x.typ.isIter() || x.typ.isGen() {
		return x, nil
	}
	if c := g.classOf(x.typ); c != nil {
		if m := c.lookup("__iter__"); m != nil && m.kind == mMethod && len(m.sig.pos) == 0 {
			return value{code: x.at(precAtom) + ".M___iter__()", typ: m.sig.ret, prec: precAtom}, nil
		}
	}
	code, elem, ok := g.iterableOf(x)
	if !ok {
		return value{}, unsupported(fmt.Sprintf("'%s' object is not iterable", pythonTypeName(x.typ)), n.Line())
	}
	return value{code: g.rt("IterOf") + "[" + g.goType(elem) + "](" + code + ")", typ: iterOf(elem), prec: precAtom}, nil
}

// starIndex is the position of a starred element, or -1.
func starIndex(nodes []pyast.Node) int {
	for i, n := range nodes {
		if _, ok := n.(*pyast.Starred); ok {
			return i
		}
	}
	return -1
}

// starUnpack is "a, *rest, b = value": the starred target gets a list of what
// is left.
func (g *generator) starUnpack(b *strings.Builder, tg *pyast.Tuple, star int, src string, t pyType, line int) error {
	if starIndex(tg.Elts[star+1:]) >= 0 {
		return unsupported("two starred expressions in assignment", line)
	}
	before, after := star, len(tg.Elts)-star-1
	fixed := before + after
	starTarget := tg.Elts[star].(*pyast.Starred).Value

	if t.isTuple() {
		elems := t.elems()
		n := len(elems)
		if n < fixed {
			return unsupported(fmt.Sprintf("not enough values to unpack (expected at least %d, got %d)", fixed, n), line)
		}
		restTypes := elems[before : n-after]
		var rest pyType
		for _, e := range restTypes {
			if rest != "" && e != rest {
				return unsupported(fmt.Sprintf("starred target for a tuple of mixed types %s", t), line)
			}
			rest = e
		}
		if rest == "" {
			return unsupported("starred target that gets nothing from "+string(t), line)
		}
		for i := 0; i < before; i++ {
			if err := g.assignTarget(b, tg.Elts[i], fmt.Sprintf("%s.F%d", src, i), elems[i], line); err != nil {
				return err
			}
		}
		var items []string
		for i := before; i < n-after; i++ {
			items = append(items, fmt.Sprintf("%s.F%d", src, i))
		}
		restCode := g.rt("NewList") + "[" + g.goType(rest) + "](" + strings.Join(items, ", ") + ")"
		if err := g.assignTarget(b, starTarget, restCode, listOf(rest), line); err != nil {
			return err
		}
		for k := 0; k < after; k++ {
			i := n - after + k
			if err := g.assignTarget(b, tg.Elts[star+1+k], fmt.Sprintf("%s.F%d", src, i), elems[i], line); err != nil {
				return err
			}
		}
		return nil
	}

	list, err := g.asList(value{code: src, typ: t, prec: precAtom}, "unpacking", line)
	if err != nil {
		return err
	}
	tmp := g.tmp("u")
	g.temp(b, tmp, list.typ, list.code)
	elem := list.typ.elem()
	fmt.Fprintf(b, "%s(%s.Len(), %d)\n", g.rt("CheckUnpackStar"), tmp, fixed)
	for i := 0; i < before; i++ {
		if err := g.assignTarget(b, tg.Elts[i], fmt.Sprintf("%s.Get(%d)", tmp, i), elem, line); err != nil {
			return err
		}
	}
	restCode := fmt.Sprintf("%s.Slice(%s{Lo: %d, HasLo: true, Hi: %s.Len() - %d, HasHi: true})", tmp, g.rt("Bounds"), before, tmp, after)
	if err := g.assignTarget(b, starTarget, restCode, listOf(elem), line); err != nil {
		return err
	}
	for k := 0; k < after; k++ {
		if err := g.assignTarget(b, tg.Elts[star+1+k], fmt.Sprintf("%s.Get(%s.Len() - %d)", tmp, tmp, after-k), elem, line); err != nil {
			return err
		}
	}
	return nil
}

// starredItems builds a display that has *iterable elements: a list, set or
// tuple of the items, in order. build is "NewList", "NewSet" or the like.
func (g *generator) starredDisplay(elts []pyast.Node, want pyType, kind string, line int) (value, error) {
	var parts []struct {
		star bool
		v    value
	}
	var elem pyType
	for _, e := range elts {
		var v value
		var err error
		star := false
		if st, ok := e.(*pyast.Starred); ok {
			star = true
			it, err := g.typedExpr(st.Value)
			if err != nil {
				return value{}, err
			}
			code, el, ok := g.iterableOf(it)
			if !ok {
				return value{}, unsupported(fmt.Sprintf("*%s in a display", it.typ), e.Line())
			}
			v = value{code: code, typ: el, prec: precAtom}
		} else {
			v, err = g.exprAs(e, want)
			if err != nil {
				return value{}, err
			}
		}
		if elem == "" {
			elem = v.typ
		} else if v.typ != elem {
			j, ok := g.joinTypes(elem, v.typ)
			if !ok {
				return value{}, unsupported(fmt.Sprintf("display with mixed element types (%s and %s)", elem, v.typ), e.Line())
			}
			elem = j
		}
		parts = append(parts, struct {
			star bool
			v    value
		}{star, v})
	}
	if want != "" {
		elem = want
	}
	goElem := g.goType(elem)
	var b strings.Builder
	switch kind {
	case "set":
		fmt.Fprintf(&b, "func() *%s.Set[%s] {\nout := %s[%s]()\n", "runtime", goElem, g.rt("NewSet"), goElem)
	default:
		fmt.Fprintf(&b, "func() *%s.List[%s] {\nout := %s[%s]()\n", "runtime", goElem, g.rt("NewList"), goElem)
	}
	g.usesRuntime = true
	for _, p := range parts {
		switch {
		case p.star && kind == "set":
			fmt.Fprintf(&b, "for c := %s.Iter(); c.Next(); {\nout.Add(c.Value())\n}\n", p.v.code)
		case p.star:
			fmt.Fprintf(&b, "for c := %s.Iter(); c.Next(); {\nout.Append(c.Value())\n}\n", p.v.code)
		case kind == "set":
			fmt.Fprintf(&b, "out.Add(%s)\n", g.relabel(p.v, elem).code)
		default:
			fmt.Fprintf(&b, "out.Append(%s)\n", g.relabel(p.v, elem).code)
		}
	}
	b.WriteString("return out\n}()")
	switch kind {
	case "set":
		if !g.hashable(elem) {
			return value{}, unsupported(fmt.Sprintf("set element of type %s (not hashable)", elem), line)
		}
		return value{code: b.String(), typ: setOf(elem), prec: precAtom}, nil
	case "tuple":
		return value{code: g.rt("TupleFrom") + "[" + goElem + "](" + b.String() + ")", typ: varTupleOf(elem), prec: precAtom}, nil
	}
	return value{code: b.String(), typ: listOf(elem), prec: precAtom}, nil
}

// tryExceptStar is "try ... except* T as e": each clause takes the part of the
// exception group that its classes match, and what no clause takes is raised
// again when all clauses have run.
func (g *generator) tryExceptStar(b *strings.Builder, n *pyast.Try) error {
	for _, h := range n.Handlers {
		if h.Type == nil {
			return unsupported("'except*' without a type", h.Line())
		}
	}
	exc := g.tmp("exc")
	code, ctx, err := g.withFlow(func(cb *strings.Builder) error { return g.block(cb, n.Body) })
	if err != nil {
		return err
	}
	g.declareFlow(b, ctx)
	fmt.Fprintf(b, "%s := %s(func() {\n%s})\n", exc, g.rt("Catch"), code)
	fmt.Fprintf(b, "if %s != nil {\nrem := %s\n_ = rem\n", exc, exc)
	for _, h := range n.Handlers {
		classes, err := g.exceptClasses(h.Type)
		if err != nil {
			return err
		}
		m, rest := g.tmp("m"), g.tmp("rest")
		fmt.Fprintf(b, "%s, %s := %s(rem, %s)\nrem = %s\nif %s != nil {\n", m, rest, g.rt("SplitStar"), strings.Join(classes, ", "), rest, m)
		if err := g.handlerBody(b, h, m); err != nil {
			return err
		}
		b.WriteString("}\n")
	}
	b.WriteString("if rem != nil {\npanic(rem)\n}\n")
	if len(n.Orelse) > 0 {
		if ctx.usedBreak || ctx.usedContinue || ctx.usedReturn {
			fmt.Fprintf(b, "} else if %s == %d {\n", ctx.flowVar, flowNone)
		} else {
			b.WriteString("} else {\n")
		}
		if err := g.block(b, n.Orelse); err != nil {
			return err
		}
	}
	b.WriteString("}\n")
	return g.dispatchFlow(b, ctx)
}

// excMethod is split and subgroup of an exception group.
func (g *generator) excMethod(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	if m.Attr != "split" && m.Attr != "subgroup" {
		return value{}, unsupported(fmt.Sprintf("method '%s' of an exception", m.Attr), n.Line())
	}
	if n.HasKeywords || len(n.Args) != 1 {
		return value{}, unsupported(m.Attr+"() needs one argument", n.Line())
	}
	var split string
	arg := n.Args[0]
	if _, isLambda := arg.(*pyast.Lambda); isLambda {
		f, err := g.exprAs(arg, funcOf([]pyType{tExc}, tBool))
		if err != nil {
			return value{}, err
		}
		split = g.rt("Split") + "(" + recv.code + ", " + f.code + ")"
	} else if classes, err := g.exceptClasses(arg); err == nil {
		split = g.rt("SplitClasses") + "(" + recv.code + ", " + strings.Join(classes, ", ") + ")"
	} else {
		f, err2 := g.exprAs(arg, funcOf([]pyType{tExc}, tBool))
		if err2 != nil || f.typ != funcOf([]pyType{tExc}, tBool) {
			return value{}, err
		}
		split = g.rt("Split") + "(" + recv.code + ", " + f.code + ")"
	}
	opt := optionalOf(tExc)
	if m.Attr == "subgroup" {
		return value{code: "func() " + g.goType(opt) + " {\nm, _ := " + split + "\nreturn " + g.rt("OptExc") + "(m)\n}()", typ: opt, prec: precAtom}, nil
	}
	t := tupleOf([]pyType{opt, opt})
	code := fmt.Sprintf("func() %s {\nm, r := %s\nreturn %s\n}()", g.goType(t), split, g.tupleCtor(t, []string{g.rt("OptExc") + "(m)", g.rt("OptExc") + "(r)"}))
	return value{code: code, typ: t, prec: precAtom}, nil
}

// callsGenerator reports whether a call is of a generator function or method
// of the program, whose result is a generator that can be closed.
func (g *generator) callsGenerator(call *pyast.Call) bool {
	switch f := call.Func.(type) {
	case *pyast.Name:
		if sig, ok := g.env.lookupSig(f.ID); ok && sig != nil {
			return sig.gen != nil
		}
		if sig := g.sigs[f.ID]; sig != nil && !g.isVariable(f.ID) {
			return sig.gen != nil
		}
	case *pyast.Attribute:
		recv, err := g.typedExpr(f.Value)
		if err != nil {
			return false
		}
		if c := g.classOf(recv.typ); c != nil {
			if m := c.lookup(g.mangled(f.Attr)); m != nil && m.sig != nil {
				return m.sig.gen != nil
			}
		}
	}
	return false
}
