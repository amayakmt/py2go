package codegen

import (
	"fmt"
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
	case *pyast.Break:
		if g.loopDepth == 0 {
			return unsupported("'break' outside loop", n.Line())
		}
		b.WriteString("break\n")
		return nil
	case *pyast.Continue:
		if g.loopDepth == 0 {
			return unsupported("'continue' not properly in loop", n.Line())
		}
		b.WriteString("continue\n")
		return nil
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
	b.WriteString(v.code + "\n")
	return nil
}

// declare introduces a variable in the current block and emits its Go
// declaration. The "_ = x" keeps Go from rejecting a variable that is only
// ever assigned, which Python allows.
func (g *generator) declare(b *strings.Builder, name string, t pyType, init string) {
	g.env.declare(name, t)
	if init == "" {
		fmt.Fprintf(b, "var %s %s\n", name, t.goType())
	} else {
		fmt.Fprintf(b, "var %s %s = %s\n", name, t.goType(), init)
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
	v, err := g.typedExpr(n.Value)
	if err != nil {
		return err
	}
	if v.typ != t {
		return mismatch(fmt.Sprintf("variable '%s'", target.ID), t, v.typ, n.Line())
	}

	// Python has one variable per name per function. A second annotation
	// of the same name assigns to it, even inside a nested block.
	if existing, declared := g.env.lookup(target.ID); declared {
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
	target, ok := n.Targets[0].(*pyast.Name)
	if !ok {
		return unsupported("assignment to "+targetKind(n.Targets[0]), n.Line())
	}
	t, declared := g.env.lookup(target.ID)
	if !declared {
		return unsupported(fmt.Sprintf("assignment to undeclared name '%s' (needs a type annotation)", target.ID), n.Line())
	}
	v, err := g.typedExpr(n.Value)
	if err != nil {
		return err
	}
	if v.typ != t {
		return mismatch(fmt.Sprintf("variable '%s'", target.ID), t, v.typ, n.Line())
	}
	fmt.Fprintf(b, "%s = %s\n", target.ID, v.code)
	return nil
}

func (g *generator) augAssign(b *strings.Builder, n *pyast.AugAssign) error {
	target, ok := n.Target.(*pyast.Name)
	if !ok {
		return unsupported("augmented assignment to "+targetKind(n.Target), n.Line())
	}
	t, declared := g.env.lookup(target.ID)
	if !declared {
		return unsupported(fmt.Sprintf("augmented assignment to undeclared name '%s'", target.ID), n.Line())
	}
	right, err := g.typedExpr(n.Value)
	if err != nil {
		return err
	}
	cur := value{code: target.ID, typ: t, prec: precAtom}
	res, err := g.binary(n.Op, cur, right, n.Line())
	if err != nil {
		return err
	}
	if res.typ != t {
		return mismatch(fmt.Sprintf("variable '%s'", target.ID), t, res.typ, n.Line())
	}

	// Plain "x += y" when both sides already have the same type.
	if goOp, direct := map[string]string{"Add": "+", "Sub": "-", "Mult": "*", "BitAnd": "&", "BitOr": "|", "BitXor": "^"}[n.Op]; direct && right.typ == t {
		fmt.Fprintf(b, "%s %s= %s\n", target.ID, goOp, right.code)
		return nil
	}
	fmt.Fprintf(b, "%s = %s\n", target.ID, res.code)
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
		b.WriteString("return\n")
		return nil
	}

	v, err := g.typedExpr(n.Value)
	if err != nil {
		return err
	}
	if v.typ != g.ret {
		return mismatch(dest, g.ret, v.typ, n.Line())
	}
	fmt.Fprintf(b, "return %s\n", v.code)
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
	target, ok := n.Target.(*pyast.Name)
	if !ok {
		return unsupported("for loop target of kind "+targetKind(n.Target), n.Line())
	}
	call, ok := n.Iter.(*pyast.Call)
	if name, isName := callName(call); !ok || !isName || name != "range" {
		return unsupported("for loop over something other than range()", n.Line())
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
		if t, declared := g.env.lookup(target.ID); declared {
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
