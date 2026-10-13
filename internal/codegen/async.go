package codegen

import (
	"fmt"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Coroutines are generators: an "async def" compiles to a goroutine-backed
// Gen[any, any, R] marked as a coroutine, and "await x" delegates to x the
// way "yield from" does. What a coroutine yields to its driver are the
// futures it waits for. An async generator is the same body with its own
// yields wrapped, so a driver can tell them from the awaits.

func nameNode(id string, line int) *pyast.Name { return &pyast.Name{ID: id, Lineno: line} }

func boolConst(v bool, line int) *pyast.Constant {
	return &pyast.Constant{PyType: "bool", Value: v, Lineno: line}
}

func methodCall(recv pyast.Node, method string, line int, args ...pyast.Node) *pyast.Call {
	return &pyast.Call{Func: &pyast.Attribute{Value: recv, Attr: method, Lineno: line}, Args: args, Lineno: line}
}

func assignTo(target, value pyast.Node, line int) *pyast.Assign {
	return &pyast.Assign{Targets: []pyast.Node{target}, Value: value, Lineno: line}
}

// awaitExpr is "await x".
func (g *generator) awaitExpr(n *pyast.Await) (value, error) {
	if g.gen == nil || !(g.gen.coro || g.gen.asyncGen) {
		return value{}, unsupported("'await' outside async function", n.Line())
	}
	v, err := g.typedExpr(n.Value)
	if err != nil {
		return value{}, err
	}
	if v.typ.isGen() {
		a := v.typ.typeArgs()
		if a[0] == tObject && a[1] == tObject {
			return value{code: g.rt("YieldFromGen") + "(" + yielderVar + ", " + v.code + ")", typ: a[2], prec: precAtom}, nil
		}
	}
	return value{code: g.rt("AwaitObj") + "(" + yielderVar + ", " + g.boxed(v) + ")", typ: tObject, prec: precAtom}, nil
}

// asyncFor rewrites "async for t in it: body else: orelse" as a loop that
// awaits __anext__ until StopAsyncIteration.
func (g *generator) asyncFor(b *strings.Builder, n *pyast.For) error {
	if g.gen == nil || !(g.gen.coro || g.gen.asyncGen) {
		return unsupported("'async for' outside async function", n.Line())
	}
	line := n.Line()
	it, more, item := fmt.Sprintf("_pyzai%d", line), fmt.Sprintf("_pyzam%d", line), fmt.Sprintf("_pyzax%d", line)
	next := &pyast.Await{Value: methodCall(nameNode(it, line), "__anext__", line), Lineno: line}
	loop := &pyast.While{
		Test: nameNode(more, line),
		Body: []pyast.Node{&pyast.Try{
			Body: []pyast.Node{assignTo(nameNode(item, line), next, line)},
			Handlers: []*pyast.ExceptHandler{{
				Type: nameNode("StopAsyncIteration", line), Lineno: line,
				Body: []pyast.Node{assignTo(nameNode(more, line), boolConst(false, line), line)},
			}},
			Orelse: append([]pyast.Node{assignTo(n.Target, nameNode(item, line), line)}, n.Body...),
			Lineno: line,
		}},
		Orelse: n.Orelse,
		Lineno: line,
	}
	var run pyast.Node = loop
	if _, fresh := n.Iter.(*pyast.Call); fresh {
		// An async generator that is dropped before it ends is closed by
		// the event loop, as CPython does when the last reference goes.
		run = &pyast.Try{
			Body: []pyast.Node{loop},
			Finalbody: []pyast.Node{&pyast.If{
				Test: nameNode(more, line),
				Body: []pyast.Node{&pyast.Expr{Value: &pyast.Call{Func: nameNode("_native", line),
					Args: []pyast.Node{&pyast.Constant{PyType: "str", Value: "asyncio.agen_done", Lineno: line}, nameNode(it, line)}, Lineno: line}, Lineno: line}},
				Lineno: line,
			}},
			Lineno: line,
		}
	}
	return g.stmtList(b, []pyast.Node{
		assignTo(nameNode(it, line), methodCall(n.Iter, "__aiter__", line), line),
		assignTo(nameNode(more, line), boolConst(true, line), line),
		&pyast.AnnAssign{Target: nameNode(item, line), Annotation: nameNode("object", line), Value: &pyast.Constant{PyType: "none", Lineno: line}, Simple: true, Lineno: line},
		run,
	})
}

// asyncWith rewrites "async with m as v: body" with __aenter__ and
// __aexit__ awaited, as PEP 492 spells it out.
func (g *generator) asyncWith(b *strings.Builder, n *pyast.With) error {
	if g.gen == nil || !(g.gen.coro || g.gen.asyncGen) {
		return unsupported("'async with' outside async function", n.Line())
	}
	line := n.Line()
	item := n.Items[0]
	body := n.Body
	if len(n.Items) > 1 {
		body = []pyast.Node{&pyast.With{Items: n.Items[1:], Body: n.Body, IsAsync: true, Lineno: line}}
	}
	depth := len(n.Items)
	m, exc, e := fmt.Sprintf("_pyzaw%d_%d", line, depth), fmt.Sprintf("_pyzae%d_%d", line, depth), fmt.Sprintf("_pyzax%d_%d", line, depth)
	enter := &pyast.Await{Value: methodCall(nameNode(m, line), "__aenter__", line), Lineno: line}
	var first pyast.Node = &pyast.Expr{Value: enter, Lineno: line}
	if item.Vars != nil {
		first = assignTo(item.Vars, enter, line)
	}
	none := &pyast.Constant{PyType: "none", Lineno: line}
	exitErr := &pyast.Await{Value: methodCall(nameNode(m, line), "__aexit__", line,
		&pyast.Call{Func: nameNode("type", line), Args: []pyast.Node{nameNode(e, line)}, Lineno: line}, nameNode(e, line), none), Lineno: line}
	exitOK := &pyast.Await{Value: methodCall(nameNode(m, line), "__aexit__", line, none, none, none), Lineno: line}
	try := &pyast.Try{
		Body: body,
		Handlers: []*pyast.ExceptHandler{{
			Type: nameNode("BaseException", line), Name: e, Lineno: line,
			Body: []pyast.Node{
				assignTo(nameNode(exc, line), boolConst(true, line), line),
				&pyast.If{Test: &pyast.UnaryOp{Op: "Not", Operand: exitErr, Lineno: line}, Body: []pyast.Node{&pyast.Raise{Lineno: line}}, Lineno: line},
			},
		}},
		Finalbody: []pyast.Node{&pyast.If{
			Test:   &pyast.UnaryOp{Op: "Not", Operand: nameNode(exc, line), Lineno: line},
			Body:   []pyast.Node{&pyast.Expr{Value: exitOK, Lineno: line}},
			Lineno: line,
		}},
		Lineno: line,
	}
	return g.stmtList(b, []pyast.Node{
		assignTo(nameNode(m, line), item.Context, line),
		first,
		assignTo(nameNode(exc, line), boolConst(false, line), line),
		try,
	})
}
