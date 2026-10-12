package codegen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

func (g *generator) stmt(b *strings.Builder, n pyast.Node) error {
	switch n := n.(type) {
	case *pyast.Expr:
		return g.exprStmt(b, n)
	case *pyast.AnnAssign:
		return g.annAssign(b, n)
	case *pyast.Assign:
		return g.assign(b, n)
	case *pyast.AugAssign:
		return g.augAssign(b, n)
	case *pyast.Return:
		return g.returnStmt(b, n)
	case *pyast.If:
		return g.ifStmt(b, n)
	case *pyast.While:
		return g.whileStmt(b, n)
	case *pyast.For:
		return g.forStmt(b, n)
	case *pyast.Raise:
		return g.raiseStmt(b, n)
	case *pyast.Assert:
		return g.assertStmt(b, n)
	case *pyast.Break:
		return g.jump(b, "break", n.Line())
	case *pyast.Continue:
		return g.jump(b, "continue", n.Line())
	case *pyast.Try:
		return g.tryStmt(b, n)
	case *pyast.Delete:
		return g.deleteStmt(b, n)
	case *pyast.Global:
		return g.globalStmt(n)
	case *compAction:
		return n.run(b)
	case *pyast.Pass:
		return nil
	case *pyast.FunctionDef:
		return unsupported("function definition inside a block", n.Line())
	case *pyast.Unsupported:
		return unsupported(n.Type, n.Line())
	default:
		return unsupported(nodeName(n), n.Line())
	}
}

// block generates statements in a nested scope. Names declared inside
// are not visible after the block, as in Go.
func (g *generator) block(b *strings.Builder, stmts []pyast.Node) error {
	outer := g.env
	g.env = newScope(outer)
	defer func() { g.env = outer }()
	for _, s := range stmts {
		if err := g.stmt(b, s); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) exprStmt(b *strings.Builder, n *pyast.Expr) error {
	// Go rejects a bare expression like "text" as a statement,
	// and go/format won't catch it. Python docstrings hit this.
	if _, ok := n.Value.(*pyast.Call); !ok {
		return unsupported("expression statement that is not a call", n.Line())
	}
	v, err := g.expr(n.Value)
	if err != nil {
		return err
	}
	// Go only accepts a call as a statement, and some builtins are
	// conversions in the generated code, so keep the value and drop it.
	if v.typ != tNone {
		b.WriteString("_ = ")
	}
	b.WriteString(v.code + "\n")
	return nil
}

// declare introduces a variable and emits its Go declaration. At module
// level it creates a module variable, which is a package-level Go variable
// shared with functions and visible after the block it was declared in, as
// Python's module scope is. Elsewhere it is local to the current block. The
// "_ = x" keeps Go from rejecting a variable that is only ever assigned,
// which Python allows.
func (g *generator) declare(b *strings.Builder, name string, t pyType, init string) {
	if g.atModuleLevel() {
		g.declareGlobal(b, name, t, init)
		return
	}
	g.declareLocal(b, name, t, init)
}

// declareLocal declares a variable in the current block, whatever the level.
func (g *generator) declareLocal(b *strings.Builder, name string, t pyType, init string) {
	g.env.declare(name, t)
	if init == "" {
		fmt.Fprintf(b, "var %s %s\n", name, g.goType(t))
	} else {
		fmt.Fprintf(b, "var %s %s = %s\n", name, g.goType(t), init)
	}
	fmt.Fprintf(b, "_ = %s\n", name)
}

func (g *generator) annAssign(b *strings.Builder, n *pyast.AnnAssign) error {
	target, ok := n.Target.(*pyast.Name)
	if !ok {
		return unsupported("annotated assignment to "+targetKind(n.Target), n.Line())
	}
	if !n.Simple {
		return unsupported("parenthesized annotation target", n.Line())
	}
	if err := checkName(target.ID, n.Line()); err != nil {
		return err
	}
	t, err := annotationType(n.Annotation, false)
	if err != nil {
		return err
	}
	if n.Value == nil {
		return unsupported(fmt.Sprintf("annotation of '%s' without a value", target.ID), n.Line())
	}
	v, err := g.exprAs(n.Value, t)
	if err != nil {
		return err
	}
	if v.typ != t {
		return mismatch(fmt.Sprintf("variable '%s'", target.ID), t, v.typ, n.Line())
	}

	// Python has one variable per name per function. A second annotation
	// of the same name assigns to it, even inside a nested block.
	if g.fnGlobals[target.ID] {
		return unsupported(fmt.Sprintf("annotated name '%s' can't be global", target.ID), n.Line())
	}
	if existing, declared, _ := g.lookupBinding(target.ID); declared {
		if existing != t {
			return unsupported(fmt.Sprintf("redeclaration of '%s' as %s (was %s)", target.ID, t, existing), n.Line())
		}
		fmt.Fprintf(b, "%s = %s\n", target.ID, v.code)
		return nil
	}
	g.declare(b, target.ID, t, v.code)
	return nil
}

func (g *generator) assign(b *strings.Builder, n *pyast.Assign) error {
	if len(n.Targets) != 1 {
		return unsupported("chained assignment", n.Line())
	}
	switch target := n.Targets[0].(type) {
	case *pyast.Name:
		t, declared, global := g.lookupBinding(target.ID)
		if !declared {
			return unsupported(fmt.Sprintf("assignment to undeclared name '%s' (needs a type annotation)", target.ID), n.Line())
		}
		v, err := g.exprAs(n.Value, t)
		if err != nil {
			return err
		}
		if v.typ != t {
			return mismatch(fmt.Sprintf("variable '%s'", target.ID), t, v.typ, n.Line())
		}
		fmt.Fprintf(b, "%s = %s\n", target.ID, v.code)
		g.afterAssign(b, target.ID, global)
		return nil
	case *pyast.Subscript:
		return g.assignItem(b, target, n.Value, n.Line())
	case *pyast.Tuple:
		return g.unpackAssign(b, target, n.Value, n.Line())
	default:
		return unsupported("assignment to "+targetKind(target), n.Line())
	}
}

// assignItem is "xs[i] = v".
func (g *generator) assignItem(b *strings.Builder, target *pyast.Subscript, valueNode pyast.Node, line int) error {
	base, idx, err := g.itemTarget(target)
	if err != nil {
		return err
	}
	elem := itemType(base.typ)
	v, err := g.exprAs(valueNode, elem)
	if err != nil {
		return err
	}
	if v.typ != elem {
		return mismatch("item", elem, v.typ, line)
	}
	fmt.Fprintf(b, "%s.Set(%s, %s)\n", base.at(precAtom), idx.code, v.code)
	return nil
}

// itemTarget generates the container and the index or key of an assignment
// to xs[i] or d[k].
func (g *generator) itemTarget(target *pyast.Subscript) (base, idx value, err error) {
	if base, err = g.typedExpr(target.Value); err != nil {
		return
	}
	if _, isSlice := target.Index.(*pyast.Slice); isSlice {
		err = unsupported("slice assignment", target.Line())
		return
	}
	switch {
	case base.typ.isList():
		if idx, err = g.typedExpr(target.Index); err != nil {
			return
		}
		if idx.typ != tInt {
			err = unsupported(fmt.Sprintf("index of type %s", idx.typ), target.Line())
		}
	case base.typ.isDict():
		if idx, err = g.exprAs(target.Index, base.typ.keyType()); err != nil {
			return
		}
		if idx.typ != base.typ.keyType() {
			err = mismatch("dict key", base.typ.keyType(), idx.typ, target.Line())
		}
	default:
		err = unsupported(fmt.Sprintf("item assignment on %s", base.typ), target.Line())
	}
	return
}

// itemType is the type of xs[i] for a list and of d[k] for a dict.
func itemType(container pyType) pyType {
	if container.isDict() {
		return container.valType()
	}
	return container.elem()
}

// temp declares a generated variable. It is always read once, but the
// "_ =" keeps Go quiet when every use turns out to be a discarded target.
func (g *generator) temp(b *strings.Builder, name string, t pyType, init string) {
	fmt.Fprintf(b, "var %s %s = %s\n_ = %s\n", name, g.goType(t), init, name)
}

// unpackTypes gives the type each of n targets receives when unpacking a
// value of type t. A tuple knows its arity; a list is checked at run time.
func unpackTypes(n int, t pyType, line int) ([]pyType, error) {
	switch {
	case t.isTuple():
		if elems := t.elems(); len(elems) != n {
			return nil, unsupported(fmt.Sprintf("unpacking %s into %d targets", t, n), line)
		} else {
			return elems, nil
		}
	case t.isList():
		types := make([]pyType, n)
		for i := range types {
			types[i] = t.elem()
		}
		return types, nil
	}
	// Any other iterable is unpacked through a list of its items.
	if elem, ok := t.iterElem(); ok && !t.isIter() {
		types := make([]pyType, n)
		for i := range types {
			types[i] = elem
		}
		return types, nil
	}
	return nil, unsupported(fmt.Sprintf("unpacking a value of type %s", t), line)
}

// targetType is the declared type of a plain name target, or "" if the
// target is anything else. It lets an empty list on the right take its type.
func (g *generator) targetType(n pyast.Node) pyType {
	if name, ok := n.(*pyast.Name); ok {
		t, _, _ := g.lookupBinding(name.ID)
		return t
	}
	return ""
}

// unpackAssign is "a, b = value".
func (g *generator) unpackAssign(b *strings.Builder, target *pyast.Tuple, valueNode pyast.Node, line int) error {
	// A tuple on the right with matching arity needs no tuple object:
	// every element is evaluated first, then assigned in order.
	if lit, ok := valueNode.(*pyast.Tuple); ok && len(lit.Elts) == len(target.Elts) {
		vals := make([]value, len(lit.Elts))
		for i, e := range lit.Elts {
			v, err := g.exprAs(e, g.targetType(target.Elts[i]))
			if err != nil {
				return err
			}
			vals[i] = v
		}

		// "a, b = b, a" is a Go parallel assignment when every target is an
		// existing name of the right type.
		names := make([]string, len(vals))
		codes := make([]string, len(vals))
		parallel := true
		for i, tg := range target.Elts {
			name, isName := tg.(*pyast.Name)
			if !isName {
				parallel = false
				break
			}
			names[i], codes[i] = name.ID, vals[i].code
			if name.ID == "_" {
				continue
			}
			if t, declared, global := g.lookupBinding(name.ID); !declared || global || t != vals[i].typ {
				parallel = false
				break
			}
		}
		if parallel {
			fmt.Fprintf(b, "%s = %s\n", strings.Join(names, ", "), strings.Join(codes, ", "))
			return nil
		}

		tmps := make([]string, len(vals))
		for i, v := range vals {
			tmps[i] = g.tmp(strconv.Itoa(i))
			g.temp(b, tmps[i], v.typ, v.code)
		}
		for i, tg := range target.Elts {
			if err := g.assignTarget(b, tg, tmps[i], vals[i].typ, line); err != nil {
				return err
			}
		}
		return nil
	}

	v, err := g.typedExpr(valueNode)
	if err != nil {
		return err
	}
	tmp := g.tmp("t")
	g.temp(b, tmp, v.typ, v.code)
	return g.assignTarget(b, target, tmp, v.typ, line)
}

// assignTarget stores src, an already evaluated Go expression of type t,
// into a target. Unpacking declares names that do not exist yet, since
// there is no way to annotate them and the tuple already has a type.
func (g *generator) assignTarget(b *strings.Builder, target pyast.Node, src string, t pyType, line int) error {
	switch tg := target.(type) {
	case *pyast.Name:
		if tg.ID == "_" {
			return nil
		}
		if existing, declared, global := g.lookupBinding(tg.ID); declared {
			if existing != t {
				return mismatch(fmt.Sprintf("variable '%s'", tg.ID), existing, t, line)
			}
			fmt.Fprintf(b, "%s = %s\n", tg.ID, src)
			g.afterAssign(b, tg.ID, global)
			return nil
		}
		if err := checkName(tg.ID, line); err != nil {
			return err
		}
		g.declare(b, tg.ID, t, src)
		return nil

	case *pyast.Tuple:
		if _, iterable := t.iterElem(); iterable && !t.isList() && !t.isIter() {
			// A str, dict, set or range unpacks through a list of its items.
			list, err := g.asList(value{code: src, typ: t, prec: precAtom}, "unpacking", line)
			if err != nil {
				return err
			}
			tmp := g.tmp("u")
			g.temp(b, tmp, list.typ, list.code)
			return g.assignTarget(b, target, tmp, list.typ, line)
		}
		types, err := unpackTypes(len(tg.Elts), t, line)
		if err != nil {
			return err
		}
		if t.isList() {
			fmt.Fprintf(b, "%s(%s.Len(), %d)\n", g.rt("CheckUnpack"), src, len(tg.Elts))
		}
		for i, elt := range tg.Elts {
			part := fmt.Sprintf("%s.F%d", src, i)
			if t.isList() {
				part = fmt.Sprintf("%s.Get(%d)", src, i)
			}
			if err := g.assignTarget(b, elt, part, types[i], line); err != nil {
				return err
			}
		}
		return nil

	case *pyast.Subscript:
		base, idx, err := g.itemTarget(tg)
		if err != nil {
			return err
		}
		if elem := itemType(base.typ); elem != t {
			return mismatch("item", elem, t, line)
		}
		fmt.Fprintf(b, "%s.Set(%s, %s)\n", base.at(precAtom), idx.code, src)
		return nil
	}
	return unsupported("assignment to "+targetKind(target), line)
}

func (g *generator) augAssign(b *strings.Builder, n *pyast.AugAssign) error {
	switch target := n.Target.(type) {
	case *pyast.Name:
		return g.augName(b, n, target)
	case *pyast.Subscript:
		return g.augItem(b, n, target)
	}
	return unsupported("augmented assignment to "+targetKind(n.Target), n.Line())
}

func (g *generator) augName(b *strings.Builder, n *pyast.AugAssign, target *pyast.Name) error {
	t, declared, global := g.lookupBinding(target.ID)
	if !declared {
		return unsupported(fmt.Sprintf("augmented assignment to undeclared name '%s'", target.ID), n.Line())
	}
	right, err := g.exprAs(n.Value, t)
	if err != nil {
		return err
	}
	cur := value{code: target.ID, typ: t, prec: precAtom}
	if global {
		cur, _ = g.readGlobal(target.ID)
	}
	res, err := g.binary(n.Op, cur, right, n.Line())
	if err != nil {
		return err
	}
	if res.typ != t {
		return mismatch(fmt.Sprintf("variable '%s'", target.ID), t, res.typ, n.Line())
	}

	// "xs += ys" on a list changes the list in place, so every name that
	// shares it sees the change. "xs = xs + ys" would build a new one. The
	// same goes for the operators |=, &=, -= and ^= on dicts and sets.
	if method, inPlace := inPlaceMethod(t, n.Op); inPlace {
		fmt.Fprintf(b, "%s.%s(%s)\n", target.ID, method, right.code)
		return nil
	}

	// Plain "x += y" when both sides already have the same type.
	if goOp, direct := map[string]string{"Add": "+", "Sub": "-", "Mult": "*", "BitAnd": "&", "BitOr": "|", "BitXor": "^"}[n.Op]; direct && right.typ == t && t != tBool && !t.isSet() {
		fmt.Fprintf(b, "%s %s= %s\n", target.ID, goOp, right.code)
		g.afterAssign(b, target.ID, global)
		return nil
	}
	fmt.Fprintf(b, "%s = %s\n", target.ID, res.code)
	g.afterAssign(b, target.ID, global)
	return nil
}

// inPlaceMethod names the runtime method that implements an augmented
// assignment by changing the container itself, as Python does for lists,
// dicts and sets.
func inPlaceMethod(t pyType, op string) (string, bool) {
	switch {
	case t.isList() && op == "Add":
		return "Extend", true
	case t.isList() && op == "Mult":
		return "RepeatInPlace", true
	case t.isDict() && op == "BitOr":
		return "Update", true
	case t.isSet():
		if m, ok := map[string]string{"BitOr": "Update", "BitAnd": "IntersectionUpdate", "Sub": "DifferenceUpdate", "BitXor": "SymmetricDifferenceUpdate"}[op]; ok {
			return m, true
		}
	}
	return "", false
}

var simpleOperand = regexp.MustCompile(`^(-?[0-9]+|[A-Za-z_][A-Za-z0-9_]*)$`)

// augItem is "xs[i] += v". The list and the index are evaluated once, so
// an expression with side effects is first stored in a temporary.
func (g *generator) augItem(b *strings.Builder, n *pyast.AugAssign, target *pyast.Subscript) error {
	base, idx, err := g.itemTarget(target)
	if err != nil {
		return err
	}
	elem := itemType(base.typ)
	right, err := g.exprAs(n.Value, elem)
	if err != nil {
		return err
	}

	list, index := base.code, idx.code
	if !simpleOperand.MatchString(list) {
		list = g.tmp("l")
		g.temp(b, list, base.typ, base.code)
	}
	if !simpleOperand.MatchString(index) {
		index = g.tmp("i")
		g.temp(b, index, idx.typ, idx.code)
	}
	cur := value{code: fmt.Sprintf("%s.Get(%s)", list, index), typ: elem, prec: precAtom}
	res, err := g.binary(n.Op, cur, right, n.Line())
	if err != nil {
		return err
	}
	if res.typ != elem {
		return mismatch("item", elem, res.typ, n.Line())
	}
	fmt.Fprintf(b, "%s.Set(%s, %s)\n", list, index, res.code)
	return nil
}

func (g *generator) returnStmt(b *strings.Builder, n *pyast.Return) error {
	if g.fnName == "" {
		return unsupported("'return' outside function", n.Line())
	}
	dest := fmt.Sprintf("return value of '%s'", g.fnName)

	// "return" and "return None" are the same thing.
	bare := n.Value == nil
	if c, ok := n.Value.(*pyast.Constant); ok && c.PyType == "none" {
		bare = true
	}
	if bare {
		if g.ret != tNone {
			return mismatch(dest, g.ret, tNone, n.Line())
		}
		g.emitReturn(b, "")
		return nil
	}

	v, err := g.exprAs(n.Value, g.ret)
	if err != nil {
		return err
	}
	if v.typ != g.ret {
		return mismatch(dest, g.ret, v.typ, n.Line())
	}
	g.emitReturn(b, v.code)
	return nil
}

// condition generates the test of an if or while as a Go bool.
func (g *generator) condition(test pyast.Node) (value, error) {
	v, err := g.typedExpr(test)
	if err != nil {
		return v, err
	}
	return truth(v), nil
}

func (g *generator) ifStmt(b *strings.Builder, n *pyast.If) error {
	b.WriteString("if ")
	for {
		cond, err := g.condition(n.Test)
		if err != nil {
			return err
		}
		b.WriteString(cond.code + " {\n")
		if err := g.block(b, n.Body); err != nil {
			return err
		}
		b.WriteString("}")

		if len(n.Orelse) == 0 {
			b.WriteString("\n")
			return nil
		}
		// An elif is an If alone in the else branch.
		if next, ok := n.Orelse[0].(*pyast.If); ok && len(n.Orelse) == 1 {
			b.WriteString(" else if ")
			n = next
			continue
		}
		b.WriteString(" else {\n")
		if err := g.block(b, n.Orelse); err != nil {
			return err
		}
		b.WriteString("}\n")
		return nil
	}
}

func (g *generator) whileStmt(b *strings.Builder, n *pyast.While) error {
	if len(n.Orelse) > 0 {
		return unsupported("loop else clause", n.Line())
	}
	cond, err := g.condition(n.Test)
	if err != nil {
		return err
	}
	// "for {" lets Go see that the loop only ends through break or return.
	if c, ok := n.Test.(*pyast.Constant); ok && c.PyType == "bool" && c.Value == true {
		b.WriteString("for {\n")
	} else {
		b.WriteString("for " + cond.code + " {\n")
	}
	g.loopDepth++
	defer func() { g.loopDepth-- }()
	if err := g.block(b, n.Body); err != nil {
		return err
	}
	b.WriteString("}\n")
	return nil
}

// intLiteral reports the value of an integer literal, with an optional minus.
func intLiteral(n pyast.Node) (int64, bool) {
	switch n := n.(type) {
	case *pyast.Constant:
		v, ok := n.Value.(int64)
		return v, ok && n.PyType == "int"
	case *pyast.UnaryOp:
		if v, ok := intLiteral(n.Operand); ok && n.Op == "USub" {
			return -v, true
		}
	}
	return 0, false
}

// forStmt translates "for i in range(...)". Python evaluates range()
// once, and the loop variable is a copy of a hidden counter, so assigning
// to i inside the body does not disturb the iteration:
//
//	var _py1_stop int64 = n
//	for _py1_i := int64(0); _py1_i < _py1_stop; _py1_i++ {
//		i = _py1_i
//		...
//	}
func (g *generator) forStmt(b *strings.Builder, n *pyast.For) error {
	if len(n.Orelse) > 0 {
		return unsupported("loop else clause", n.Line())
	}
	if call, ok := n.Iter.(*pyast.Call); ok {
		if name, isName := callName(call); isName && name == "range" {
			return g.forRange(b, n, call)
		}
	}
	return g.forEach(b, n)
}

func (g *generator) forRange(b *strings.Builder, n *pyast.For, call *pyast.Call) error {
	target, ok := n.Target.(*pyast.Name)
	if !ok {
		return unsupported("for loop target of kind "+targetKind(n.Target), n.Line())
	}
	if call.HasKeywords {
		return unsupported("keyword arguments", call.Line())
	}
	if len(call.Args) < 1 || len(call.Args) > 3 {
		return unsupported(fmt.Sprintf("range() with %d arguments", len(call.Args)), call.Line())
	}
	// "for _ in range(n)" discards the counter, so there is no variable.
	discard := target.ID == "_"
	if !discard {
		if err := checkName(target.ID, n.Line()); err != nil {
			return err
		}
	}

	// Evaluate the arguments once, left to right. Literals stay inline.
	id := g.tmp("")
	argCode := make([]string, 3)
	argCode[0], argCode[2] = "0", "1"
	slots := []int{1, 0, 2} // range(n): stop only; range(a, b): start, stop; ...
	if len(call.Args) >= 2 {
		slots = []int{0, 1, 2}
	}
	var stepLit int64 = 1
	stepKnown := true
	names := []string{"start", "stop", "step"}
	for i, a := range call.Args {
		slot := slots[i]
		v, err := g.typedExpr(a)
		if err != nil {
			return err
		}
		if v.typ != tInt {
			return mismatch("range() argument", tInt, v.typ, a.Line())
		}
		lit, isLit := intLiteral(a)
		if slot == 2 {
			stepLit, stepKnown = lit, isLit
			if isLit && lit == 0 {
				return unsupported("range() with a step of zero", a.Line())
			}
		}
		if isLit {
			argCode[slot] = v.code
			continue
		}
		hoisted := id + names[slot]
		fmt.Fprintf(b, "var %s int64 = %s\n", hoisted, v.code)
		argCode[slot] = hoisted
	}
	start, stop, step := argCode[0], argCode[1], argCode[2]

	counter := id + "i"
	var cond, post string
	switch {
	case stepKnown && stepLit > 0:
		cond = fmt.Sprintf("%s < %s", counter, stop)
	case stepKnown:
		cond = fmt.Sprintf("%s > %s", counter, stop)
	default:
		// The sign of the step is only known at run time.
		fmt.Fprintf(b, "%s(%s)\n", g.rt("CheckRangeStep"), step)
		cond = fmt.Sprintf("(%[1]s > 0 && %[2]s < %[3]s) || (%[1]s < 0 && %[2]s > %[3]s)", step, counter, stop)
	}
	if step == "1" {
		post = counter + "++"
	} else {
		post = fmt.Sprintf("%s += %s", counter, step)
	}

	// The loop variable lives in the enclosing block, because Python keeps
	// it alive after the loop.
	if !discard {
		if t, declared, _ := g.lookupBinding(target.ID); declared {
			if t != tInt {
				return mismatch(fmt.Sprintf("loop variable '%s'", target.ID), t, tInt, n.Line())
			}
		} else {
			g.declare(b, target.ID, tInt, "")
		}
	}

	fmt.Fprintf(b, "for %s := int64(%s); %s; %s {\n", counter, start, cond, post)
	if !discard {
		fmt.Fprintf(b, "%s = %s\n", target.ID, counter)
	}
	g.loopDepth++
	defer func() { g.loopDepth-- }()
	if err := g.block(b, n.Body); err != nil {
		return err
	}
	b.WriteString("}\n")
	return nil
}

func callName(c *pyast.Call) (string, bool) {
	if c == nil {
		return "", false
	}
	name, ok := c.Func.(*pyast.Name)
	if !ok {
		return "", false
	}
	return name.ID, true
}

// targetKind names an assignment target for error messages.
func targetKind(n pyast.Node) string {
	if u, ok := n.(*pyast.Unsupported); ok {
		return u.Type
	}
	return nodeName(n)
}

// leaf is one name a loop target binds, with the type it receives.
type leaf struct {
	name string
	typ  pyType
}

// leaves lists the names bound by a target pattern such as "i, (a, b)".
func leaves(target pyast.Node, t pyType, line int) ([]leaf, error) {
	switch tg := target.(type) {
	case *pyast.Name:
		if tg.ID == "_" {
			return nil, nil
		}
		return []leaf{{tg.ID, t}}, nil
	case *pyast.Tuple:
		types, err := unpackTypes(len(tg.Elts), t, line)
		if err != nil {
			return nil, err
		}
		var out []leaf
		for i, elt := range tg.Elts {
			l, err := leaves(elt, types[i], line)
			if err != nil {
				return nil, err
			}
			out = append(out, l...)
		}
		return out, nil
	}
	return nil, unsupported("for loop target of kind "+targetKind(target), line)
}

// forEach translates a for loop over anything iterable: a list, string,
// dict, set, range, view or iterator, including zip(), enumerate() and
// reversed(). The loop runs a cursor, which reads the container as it goes,
// so appending to a list during the loop behaves as in Python and changing a
// dict's size raises RuntimeError:
//
//	var _py1_c runtime.Cursor[int64] = xs.Iter()
//	for _py1_c.Next() {
//		x = _py1_c.Value()
//		...
//	}
func (g *generator) forEach(b *strings.Builder, n *pyast.For) error {
	src, err := g.iterOf(n.Iter)
	if err != nil {
		return err
	}

	// Bind the loop variables in the enclosing block before the loop,
	// because Python keeps them alive afterwards.
	bound, err := leaves(n.Target, src.typ, n.Line())
	if err != nil {
		return err
	}
	for _, l := range bound {
		if existing, declared, _ := g.lookupBinding(l.name); declared {
			if existing != l.typ {
				return mismatch(fmt.Sprintf("loop variable '%s'", l.name), existing, l.typ, n.Line())
			}
			continue
		}
		if err := checkName(l.name, n.Line()); err != nil {
			return err
		}
		g.declare(b, l.name, l.typ, "")
	}

	cur := g.tmp("c")
	fmt.Fprintf(b, "var %s %s[%s] = %s\nfor %s.Next() {\n", cur, g.rt("Cursor"), g.goType(src.typ), src.cursor, cur)
	item := cur + ".Value()"
	switch tg := n.Target.(type) {
	case *pyast.Name:
		if err := g.assignTarget(b, tg, item, src.typ, n.Line()); err != nil {
			return err
		}
	case *pyast.Tuple:
		tmp := g.tmp("e")
		g.temp(b, tmp, src.typ, item)
		if err := g.assignTarget(b, tg, tmp, src.typ, n.Line()); err != nil {
			return err
		}
	}

	g.loopDepth++
	defer func() { g.loopDepth-- }()
	if err := g.block(b, n.Body); err != nil {
		return err
	}
	b.WriteString("}\n")
	return nil
}

// deleteStmt is "del xs[i]", "del xs[a:b]" and "del d[k]".
func (g *generator) deleteStmt(b *strings.Builder, n *pyast.Delete) error {
	for _, target := range n.Targets {
		sub, ok := target.(*pyast.Subscript)
		if !ok {
			if _, isName := target.(*pyast.Name); isName {
				return unsupported("del of a variable", n.Line())
			}
			return unsupported("del of "+targetKind(target), n.Line())
		}
		base, err := g.typedExpr(sub.Value)
		if err != nil {
			return err
		}
		if sl, isSlice := sub.Index.(*pyast.Slice); isSlice {
			if !base.typ.isList() {
				return unsupported(fmt.Sprintf("del of a slice of %s", base.typ), n.Line())
			}
			bounds, err := g.bounds(sl)
			if err != nil {
				return err
			}
			fmt.Fprintf(b, "%s.DelSlice(%s)\n", base.at(precAtom), bounds)
			continue
		}
		switch {
		case base.typ.isList():
			idx, err := g.typedExpr(sub.Index)
			if err != nil {
				return err
			}
			if idx.typ != tInt {
				return unsupported(fmt.Sprintf("index of type %s", idx.typ), n.Line())
			}
			fmt.Fprintf(b, "%s.DelAt(%s)\n", base.at(precAtom), idx.code)
		case base.typ.isDict():
			key, err := g.exprAs(sub.Index, base.typ.keyType())
			if err != nil {
				return err
			}
			if key.typ != base.typ.keyType() {
				return mismatch("dict key", base.typ.keyType(), key.typ, n.Line())
			}
			fmt.Fprintf(b, "%s.Delete(%s)\n", base.at(precAtom), key.code)
		default:
			return unsupported(fmt.Sprintf("del of an item of %s", base.typ), n.Line())
		}
	}
	return nil
}

// compAction is a statement that runs generated code. Comprehensions are
// translated into ordinary for and if statements around one of these.
type compAction struct {
	run func(b *strings.Builder) error
}

func (a *compAction) Line() int { return 0 }
