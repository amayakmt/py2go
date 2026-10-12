package codegen

import (
	"fmt"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// withStmt is "with a as x, b: body". The body runs in a closure like the
// body of a try, so that an exception reaches __exit__ and jumps out of the
// body (return, break, continue) happen after it:
//
//	m := a; x := m.__enter__()
//	pending := Catch(body)
//	if pending != nil { if !m.__exit__(type, value, None) { panic(pending) } } else { m.__exit__(None, None, None) }
func (g *generator) withStmt(b *strings.Builder, n *pyast.With) error {
	item := n.Items[0]
	body := n.Body
	if len(n.Items) > 1 {
		body = []pyast.Node{&pyast.With{Items: n.Items[1:], Body: n.Body, Lineno: n.Line()}}
	}
	mgr, err := g.typedExpr(item.Context)
	if err != nil {
		return err
	}
	c := g.classOf(mgr.typ)
	if c == nil {
		return unsupported(fmt.Sprintf("with of a value of type %s (context managers are classes with __enter__ and __exit__)", mgr.typ), n.Line())
	}
	enter := c.lookup("__enter__")
	exit := c.lookup("__exit__")
	if enter == nil || enter.kind != mMethod || len(enter.sig.pos) != 0 {
		return unsupported(fmt.Sprintf("'%s' object does not support the context manager protocol (no __enter__)", c.name), n.Line())
	}
	if exit == nil || exit.kind != mMethod || len(exit.sig.pos) != 3 {
		return unsupported(fmt.Sprintf("'%s' object does not support the context manager protocol (__exit__ needs three parameters)", c.name), n.Line())
	}
	for _, p := range exit.sig.pos {
		if p.typ != tObject && p.typ != optionalOf(tExc) {
			return unsupported("__exit__ parameter of type "+string(p.typ)+" (annotate the parameters object, or Optional[BaseException] for the second)", n.Line())
		}
	}
	if exit.special {
		return unsupported("__exit__ that may return NotImplemented", n.Line())
	}

	b.WriteString("{\n")
	m := g.tmp("with")
	g.temp(b, m, mgr.typ, mgr.code)

	// The manager is entered outside the body: an error there does not call
	// __exit__.
	entered := value{code: m + ".M___enter__()", typ: enter.sig.ret, prec: precAtom}
	if item.Vars != nil {
		if enter.sig.ret == tNone {
			return unsupported("with ... as on __enter__ that returns nothing", n.Line())
		}
		tmp := g.tmp("v")
		g.temp(b, tmp, entered.typ, entered.code)
		if err := g.assignTarget(b, item.Vars, tmp, entered.typ, n.Line()); err != nil {
			return err
		}
	} else if enter.sig.ret == tNone {
		b.WriteString(entered.code + "\n")
	} else {
		b.WriteString("_ = " + entered.code + "\n")
	}

	code, ctx, err := g.withFlow(func(cb *strings.Builder) error { return g.block(cb, body) })
	if err != nil {
		return err
	}
	pending := g.tmp("pending")
	g.declareFlow(b, ctx)
	fmt.Fprintf(b, "%s := %s(func() {\n%s})\n", pending, g.rt("Catch"), code)

	call := func(excType, excValue string) string {
		args := make([]string, 3)
		args[0], args[2] = excType, "nil"
		if exit.sig.pos[1].typ == optionalOf(tExc) {
			args[1] = excValue
		} else {
			args[1] = excValue
		}
		return m + ".M___exit__(" + strings.Join(args, ", ") + ")"
	}
	suppressed := func(c string) string {
		switch exit.sig.ret {
		case tBool:
			return c
		case optionalOf(tBool):
			return g.rt("OptTruth") + "(" + c + ")"
		}
		return "func() bool {\n_ = " + c + "\nreturn false\n}()"
	}
	if exit.sig.ret == tNone {
		suppressed = func(c string) string { return "func() bool {\n" + c + "\nreturn false\n}()" }
	}
	var excValue, noValue string
	if exit.sig.pos[1].typ == optionalOf(tExc) {
		excValue = g.rt("Some") + "(" + pending + ")"
		noValue = g.rt("NoneOf") + "[*" + g.rt("PyError") + "]()"
	} else {
		excValue = g.rt("ExcValue") + "(" + pending + ")"
		noValue = "nil"
	}
	fmt.Fprintf(b, "if %s != nil {\nif !%s(%s, func() bool {\nreturn %s\n}) {\npanic(%s)\n}\n} else {\n%s\n}\n", pending, g.rt("ExitCall"), pending,
		suppressed(call(g.rt("ExcClass")+"("+pending+")", excValue)), pending, "_ = "+suppressed(call("nil", noValue)))
	if err := g.dispatchFlow(b, ctx); err != nil {
		return err
	}
	b.WriteString("}\n")
	return nil
}
