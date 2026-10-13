package runtime

// Coroutines are generators that an event loop drives: "await x" delegates to
// x like "yield from", and what reaches the loop are the futures the
// coroutine waits for.

// AsCoro marks the generator as the coroutine of an "async def" called name.
func (g *Gen[Y, S, R]) AsCoro(name string) *Gen[Y, S, R] {
	g.coro, g.name = true, name
	return g
}

// PyClassName is the type name: coroutine or generator.
func (g *Gen[Y, S, R]) PyClassName() string {
	if g.coro {
		return "coroutine"
	}
	return "generator"
}

// awaitInto runs the generator to its end inside the generator y belongs to,
// passing along what it yields and what is sent or thrown back.
func (g *Gen[Y, S, R]) awaitInto(y *Yielder[any, any]) any {
	var msg genMsg[S]
	for {
		v, ret, done := g.step(msg)
		if done {
			return ret
		}
		m := y.yield(v)
		msg = genMsg[S]{exc: m.exc}
		if m.exc == nil {
			msg.val = conv[S](m.val)
		}
	}
}

type awaitIntoer interface {
	awaitInto(y *Yielder[any, any]) any
}

// AwaitObj is "await x" for an x that is not known to be a coroutine.
func AwaitObj(y *Yielder[any, any], v any) any {
	v = Unbox(v)
	if a, ok := v.(awaitIntoer); ok {
		if c, isGen := v.(interface{ isCoro() bool }); !isGen || c.isCoro() {
			return a.awaitInto(y)
		}
	}
	if !hasAttr(v, "__await__") {
		panic(newError("TypeError", "object "+TypeName(v)+" can't be used in 'await' expression"))
	}
	it := Unbox(DynCallMethod(v, "__await__", nil, nil))
	if a, ok := it.(awaitIntoer); ok {
		return a.awaitInto(y)
	}
	var send any
	var exc *PyError
	started := false
	for {
		var out any
		stop := Catch(func() {
			switch {
			case exc != nil:
				e := exc
				exc = nil
				out = DynCallMethod(it, "throw", []any{e}, nil)
			case !started:
				started = true
				out = DynNext(it, nil, false)
			default:
				out = DynCallMethod(it, "send", []any{send}, nil)
			}
		})
		if stop != nil {
			if IsSubclass(stop.Type, "StopIteration") {
				return StopValue(stop)
			}
			panic(stop)
		}
		m := y.yield(out)
		send, exc = m.val, m.exc
	}
}

func (g *Gen[Y, S, R]) isCoro() bool { return g.coro }

func hasAttr(v any, name string) bool {
	ok := true
	if e := Catch(func() { DynGetAttr(v, name) }); e != nil {
		ok = false
	}
	return ok
}

// ---- async generators ----

type agWrapped struct{ v any }

// AsyncGen is the object an "async def" with a yield returns. Its body is
// a generator whose own yields are wrapped, so they can be told from the
// futures it awaits, which pass up to the event loop untouched.
type AsyncGen struct {
	g      *Gen[any, any, any]
	closed bool
}

func NewAsyncGen(g *Gen[any, any, any]) *AsyncGen { return &AsyncGen{g: g} }

// AGYield is "yield v" in an async generator.
func AGYield(y *Yielder[any, any], v any) any { return y.Yield(agWrapped{v}) }

func (a *AsyncGen) PyClassName() string { return "async_generator" }
func (a *AsyncGen) PyRepr() string      { return "<async_generator object at " + addr(a) + ">" }

// step is the awaitable of __anext__, asend and athrow.
func (a *AsyncGen) step(val any, exc *PyError, closing bool) *Gen[any, any, any] {
	return NewGen(func(y *Yielder[any, any]) any {
		if a.closed {
			if exc != nil && !closing {
				panic(exc)
			}
			if closing {
				return NoneV
			}
			panic(newError("StopAsyncIteration", ""))
		}
		msg := genMsg[any]{val: val, exc: exc}
		for {
			var v, ret any
			var done bool
			if e := Catch(func() { v, ret, done = a.g.step(msg) }); e != nil {
				a.closed = true
				switch {
				case IsSubclass(e.Type, "StopIteration"):
					w := newError("RuntimeError", "async generator raised StopIteration")
					w.Cause, w.SuppressContext = e, true
					panic(w)
				case closing && (e.Type == "GeneratorExit" || IsSubclass(e.Type, "StopAsyncIteration")):
					return NoneV
				}
				panic(e)
			}
			if done {
				a.closed = true
				_ = ret
				if closing {
					return NoneV
				}
				panic(newError("StopAsyncIteration", ""))
			}
			if w, ok := v.(agWrapped); ok {
				if closing {
					panic(newError("RuntimeError", "async generator ignored GeneratorExit"))
				}
				return w.v
			}
			m := y.yield(v)
			msg = genMsg[any]{val: m.val, exc: m.exc}
		}
	}).AsCoro("anext")
}

func (a *AsyncGen) PyDynTable() map[string]bool { return asyncGenMethodNames }

var asyncGenMethodNames = names("__aiter__ __anext__ asend athrow aclose")

func (a *AsyncGen) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "__aiter__":
		return a, true
	case "__anext__":
		return a.step(nil, nil, false), true
	case "asend":
		countArgs("asend", args, 1, 1)
		return a.step(args[0], nil, false), true
	case "athrow":
		countArgs("athrow", args, 1, 3)
		e, ok := Unbox(args[0]).(*PyError)
		if !ok {
			if x, isInst := Unbox(args[0]).(interface{ PyExc() *PyError }); isInst {
				e = x.PyExc()
			} else {
				panic(newError("TypeError", "exceptions must derive from BaseException"))
			}
		}
		return a.step(nil, e, false), true
	case "aclose":
		if a.g.state == genNew {
			a.closed = true
		}
		return a.step(nil, newError("GeneratorExit", ""), true), true
	}
	return nil, false
}

// ExcItems turns a list of objects that are exceptions into the exceptions
// of an exception group.
func ExcItems(l *List[any]) []*PyError {
	out := make([]*PyError, 0, l.Len())
	for _, v := range l.Items() {
		out = append(out, AsException(v))
	}
	return out
}

func isNoneVal(v any) bool {
	switch v.(type) {
	case nil, NoneT:
		return true
	}
	return false
}

// agenHook is called with an async generator that was dropped before it
// ended; the event loop installs it to close the generator.
var agenHook any

func init() {
	RegisterNative("asyncio.set_agen_hook", func(args []any, kw []KW) any {
		nativeArgs("set_agen_hook", args, 1, 1)
		agenHook = args[0]
		return nil
	})
	RegisterNative("asyncio.agen_done", func(args []any, kw []KW) any {
		nativeArgs("agen_done", args, 1, 1)
		if a, ok := Unbox(args[0]).(*AsyncGen); ok && !a.closed && agenHook != nil {
			DynCall(agenHook, a)
		}
		return nil
	})
}
