package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// attribute handles the one attribute the transpiler knows:
// type(e).__name__ on an exception, which is the name of its class.
func (g *generator) attribute(n *pyast.Attribute) (value, error) {
	// __name__, __qualname__ and __doc__ of a function whose def is known.
	if name, ok := n.Value.(*pyast.Name); ok {
		if sig, found := g.env.lookupSig(name.ID); found && sig != nil {
			return g.functionAttribute(sig, n.Attr, n.Line())
		}
		if _, isVar := g.env.lookup(name.ID); !isVar {
			if sig, ok := g.sigs[name.ID]; ok && len(sig.decorators) == 0 {
				return g.functionAttribute(sig, n.Attr, n.Line())
			}
		}
	}
	if call, ok := n.Value.(*pyast.Call); ok && n.Attr == "__name__" {
		if name, _ := callName(call); name == "type" && len(call.Args) == 1 && !call.HasKeywords {
			v, err := g.typedExpr(call.Args[0])
			if err != nil {
				return value{}, err
			}
			if v.typ == tExc {
				return value{code: v.at(precAtom) + ".Type", typ: tStr, prec: precAtom}, nil
			}
			return value{}, unsupported(fmt.Sprintf("type() of %s", v.typ), n.Line())
		}
	}
	return value{}, unsupported("attribute access '."+n.Attr+"'", n.Line())
}

// isExceptionClass reports whether a name in an expression refers to a
// built-in exception class, rather than to a variable or a function that
// happens to share its name.
func (g *generator) isExceptionClass(name string) bool {
	if g.isVariable(name) {
		return false
	}
	if _, isFunc := g.sigs[name]; isFunc {
		return false
	}
	return runtime.IsExceptionClass(name)
}

// isUnknownName reports a name that is no variable, function or builtin
// exception class. In a raise it is most likely a class the program defines.
func (g *generator) isUnknownName(name string) bool {
	if g.isVariable(name) {
		return false
	}
	if _, isFunc := g.sigs[name]; isFunc {
		return false
	}
	return !runtime.IsExceptionClass(name)
}

// excValue generates an expression that evaluates to an exception: a class
// (raised with no arguments), a call of a class, or a name bound by except.
func (g *generator) excValue(n pyast.Node) (string, error) {
	switch n := n.(type) {
	case *pyast.Name:
		if g.isExceptionClass(n.ID) {
			return call(g.rt("NewError"), strconv.Quote(n.ID)), nil
		}
		if t, ok := g.env.lookup(n.ID); ok && t == tExc {
			return n.ID, nil
		}
		if g.isUnknownName(n.ID) {
			return "", unsupported(fmt.Sprintf("unknown exception class '%s'", n.ID), n.Line())
		}
	case *pyast.Call:
		name, isName := callName(n)
		if isName && g.isUnknownName(name) {
			return "", unsupported(fmt.Sprintf("unknown exception class '%s'", name), n.Line())
		}
		if !isName || !g.isExceptionClass(name) {
			break
		}
		if n.HasKeywords {
			return "", unsupported("keyword arguments", n.Line())
		}
		args := []string{strconv.Quote(name)}
		for _, a := range n.Args {
			v, err := g.typedExpr(a)
			if err != nil {
				return "", err
			}
			if v.typ == tExc {
				return "", unsupported("an exception as the argument of an exception", a.Line())
			}
			args = append(args, v.code)
		}
		return call(g.rt("NewError"), args...), nil
	}
	return "", unsupported("raise of something that is not an exception", n.Line())
}

// raiseStmt is "raise", "raise Exc" and "raise Exc from Cause".
func (g *generator) raiseStmt(b *strings.Builder, n *pyast.Raise) error {
	if n.Exc == nil {
		// A bare raise re-raises the exception the innermost handler caught.
		if len(g.handlers) == 0 {
			return unsupported("'raise' with no active exception", n.Line())
		}
		fmt.Fprintf(b, "panic(%s)\n", g.handlers[len(g.handlers)-1])
		return nil
	}
	code, err := g.excValue(n.Exc)
	if err != nil {
		return err
	}
	// The cause only appears in the traceback on stderr. Check that it is
	// well formed, then drop it.
	if c, isConst := n.Cause.(*pyast.Constant); n.Cause != nil && !(isConst && c.PyType == "none") {
		if _, err := g.excValue(n.Cause); err != nil {
			return err
		}
	}
	fmt.Fprintf(b, "panic(%s)\n", code)
	return nil
}

// assertStmt is "assert test" and "assert test, message".
func (g *generator) assertStmt(b *strings.Builder, n *pyast.Assert) error {
	cond, err := g.condition(n.Test)
	if err != nil {
		return err
	}
	args := []string{strconv.Quote("AssertionError")}
	if n.Msg != nil {
		v, err := g.typedExpr(n.Msg)
		if err != nil {
			return err
		}
		if v.typ == tExc {
			return unsupported("an exception as the argument of an exception", n.Msg.Line())
		}
		args = append(args, v.code)
	}
	fmt.Fprintf(b, "if !%s {\npanic(%s)\n}\n", cond.at(precUnary), call(g.rt("NewError"), args...))
	return nil
}

// Go has no exceptions, so the body of a try becomes a closure run by
// runtime.Catch. A break, continue or return inside the closure cannot act
// on the loop or function around it, so it records what it wanted in a
// status variable and returns from the closure. The code after the closure
// reads the status and does the real jump.
const (
	flowNone = iota
	flowBreak
	flowContinue
	flowReturn
)

// flowCtx describes one closure being generated.
type flowCtx struct {
	flowVar string // Go variable holding the pending jump
	retVar  string // Go variable holding the return value, empty if there is none
	// outerLoop is whether a loop around the closure, in the same function,
	// can be the target of a break or continue.
	outerLoop                           bool
	usedBreak, usedContinue, usedReturn bool
}

// withFlow generates statements for the inside of a closure. Loops inside
// it jump natively; jumps that leave it are rewritten into status codes.
func (g *generator) withFlow(body func(b *strings.Builder) error) (string, *flowCtx, error) {
	ctx := &flowCtx{flowVar: g.tmp("flow")}
	if g.fnName != "" && g.ret != tNone {
		ctx.retVar = g.tmp("ret")
	}
	ctx.outerLoop = g.loopDepth > 0 || (g.flow != nil && g.flow.outerLoop)

	savedFlow, savedDepth := g.flow, g.loopDepth
	g.flow, g.loopDepth = ctx, 0
	var b strings.Builder
	err := body(&b)
	g.flow, g.loopDepth = savedFlow, savedDepth
	return b.String(), ctx, err
}

// declareFlow declares the status variables a closure may set.
func (g *generator) declareFlow(b *strings.Builder, ctx *flowCtx) {
	fmt.Fprintf(b, "var %s int\n_ = %s\n", ctx.flowVar, ctx.flowVar)
	if ctx.retVar != "" {
		fmt.Fprintf(b, "var %s %s\n_ = %s\n", ctx.retVar, g.goType(g.ret), ctx.retVar)
	}
}

// dispatchFlow acts on the jump a closure asked for, in the code around it.
// That code may itself be inside another closure, in which case the jump is
// passed on one level further.
func (g *generator) dispatchFlow(b *strings.Builder, ctx *flowCtx) error {
	emit := func(used bool, code int, act func() error) error {
		if !used {
			return nil
		}
		fmt.Fprintf(b, "if %s == %d {\n", ctx.flowVar, code)
		if err := act(); err != nil {
			return err
		}
		b.WriteString("}\n")
		return nil
	}
	if err := emit(ctx.usedBreak, flowBreak, func() error { return g.jump(b, "break", 0) }); err != nil {
		return err
	}
	if err := emit(ctx.usedContinue, flowContinue, func() error { return g.jump(b, "continue", 0) }); err != nil {
		return err
	}
	return emit(ctx.usedReturn, flowReturn, func() error {
		g.emitReturn(b, ctx.retVar)
		return nil
	})
}

// jump is break or continue.
func (g *generator) jump(b *strings.Builder, kind string, line int) error {
	if g.loopDepth > 0 {
		b.WriteString(kind + "\n")
		return nil
	}
	if g.flow != nil && g.flow.outerLoop {
		code := flowBreak
		g.flow.usedBreak = true
		if kind == "continue" {
			code = flowContinue
			g.flow.usedContinue = true
		}
		fmt.Fprintf(b, "%s = %d\nreturn\n", g.flow.flowVar, code)
		return nil
	}
	if kind == "break" {
		return unsupported("'break' outside loop", line)
	}
	return unsupported("'continue' not properly in loop", line)
}

// emitReturn is "return" or "return value" in the current closure or function.
func (g *generator) emitReturn(b *strings.Builder, value string) {
	if g.flow == nil {
		if value == "" {
			b.WriteString("return\n")
		} else {
			fmt.Fprintf(b, "return %s\n", value)
		}
		return
	}
	g.flow.usedReturn = true
	if value != "" {
		fmt.Fprintf(b, "%s = %s\n", g.flow.retVar, value)
	}
	fmt.Fprintf(b, "%s = %d\nreturn\n", g.flow.flowVar, flowReturn)
}

// hoistDecls declares, before a try statement, the variables that the top
// level of its body annotates. Python keeps them alive after the try, and
// the common pattern is to assign them in the body and again in a handler.
// The closure would otherwise hide them.
func (g *generator) hoistDecls(b *strings.Builder, body []pyast.Node) {
	for _, stmt := range body {
		a, ok := stmt.(*pyast.AnnAssign)
		if !ok || a.Value == nil || !a.Simple {
			continue
		}
		name, ok := a.Target.(*pyast.Name)
		if !ok || name.ID == "_" || checkName(name.ID, a.Line()) != nil {
			continue
		}
		if _, declared := g.env.lookup(name.ID); declared {
			continue
		}
		if t, err := annotationType(a.Annotation, false); err == nil {
			g.declare(b, name.ID, t, "")
		}
	}
}

func (g *generator) tryStmt(b *strings.Builder, n *pyast.Try) error {
	g.hoistDecls(b, n.Body)
	// A block keeps the generated variables local.
	b.WriteString("{\n")
	var err error
	if len(n.Finalbody) > 0 {
		err = g.tryFinally(b, n)
	} else {
		err = g.tryExcept(b, n)
	}
	b.WriteString("}\n")
	return err
}

// tryFinally wraps the whole try/except in a closure, runs the finally
// body after it whatever happened, and then lets a pending exception or
// jump continue. A return or break inside the finally body goes first and
// so replaces them, as in Python.
func (g *generator) tryFinally(b *strings.Builder, n *pyast.Try) error {
	code, ctx, err := g.withFlow(func(cb *strings.Builder) error {
		if len(n.Handlers) > 0 {
			return g.tryExcept(cb, n)
		}
		return g.block(cb, n.Body)
	})
	if err != nil {
		return err
	}
	pending := g.tmp("pending")
	g.declareFlow(b, ctx)
	fmt.Fprintf(b, "%s := %s(func() {\n%s})\n", pending, g.rt("Catch"), code)
	if err := g.block(b, n.Finalbody); err != nil {
		return err
	}
	fmt.Fprintf(b, "if %s != nil {\npanic(%s)\n}\n", pending, pending)
	return g.dispatchFlow(b, ctx)
}

// tryExcept generates try/except/else.
func (g *generator) tryExcept(b *strings.Builder, n *pyast.Try) error {
	for i, h := range n.Handlers {
		if h.Type == nil && i != len(n.Handlers)-1 {
			return unsupported("'except:' that is not last", h.Line())
		}
	}

	exc := g.tmp("exc")
	code, ctx, err := g.withFlow(func(cb *strings.Builder) error { return g.block(cb, n.Body) })
	if err != nil {
		return err
	}
	g.declareFlow(b, ctx)
	fmt.Fprintf(b, "%s := %s(func() {\n%s})\n", exc, g.rt("Catch"), code)

	fmt.Fprintf(b, "if %s != nil {\n", exc)
	caught := false
	for i, h := range n.Handlers {
		if h.Type != nil {
			classes, err := g.exceptClasses(h.Type)
			if err != nil {
				return err
			}
			cond := call(g.rt("Match"), append([]string{exc}, classes...)...)
			if i == 0 {
				fmt.Fprintf(b, "if %s {\n", cond)
			} else {
				fmt.Fprintf(b, "} else if %s {\n", cond)
			}
		} else {
			caught = true
			if i == 0 {
				b.WriteString("{\n")
			} else {
				b.WriteString("} else {\n")
			}
		}
		if err := g.handlerBody(b, h, exc); err != nil {
			return err
		}
	}
	if caught {
		b.WriteString("}\n")
	} else {
		fmt.Fprintf(b, "} else {\npanic(%s)\n}\n", exc)
	}

	if len(n.Orelse) > 0 {
		// else runs when the body finished without an exception or a jump.
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

// exceptClasses reads the class or tuple of classes of an except clause.
func (g *generator) exceptClasses(n pyast.Node) ([]string, error) {
	nodes := []pyast.Node{n}
	if t, ok := n.(*pyast.Tuple); ok {
		nodes = t.Elts
	}
	classes := make([]string, 0, len(nodes))
	for _, node := range nodes {
		name, ok := node.(*pyast.Name)
		if !ok {
			return nil, unsupported("except clause of kind "+targetKind(node), node.Line())
		}
		if !g.isExceptionClass(name.ID) {
			return nil, unsupported(fmt.Sprintf("unknown exception class '%s'", name.ID), name.Line())
		}
		classes = append(classes, strconv.Quote(name.ID))
	}
	return classes, nil
}

// handlerBody generates the body of one except clause. The name after "as"
// lives only inside it.
func (g *generator) handlerBody(b *strings.Builder, h *pyast.ExceptHandler, exc string) error {
	outer := g.env
	g.env = newScope(outer)
	g.handlers = append(g.handlers, exc)
	defer func() {
		g.env = outer
		g.handlers = g.handlers[:len(g.handlers)-1]
	}()

	if h.Name != "" {
		if err := checkName(h.Name, h.Line()); err != nil {
			return err
		}
		g.declareLocal(b, h.Name, tExc, exc)
	}
	if err := g.predeclareDefs(b, h.Body); err != nil {
		return err
	}
	for _, s := range h.Body {
		if err := g.stmt(b, s); err != nil {
			return err
		}
	}
	return nil
}
