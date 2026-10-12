package runtime

// Generators run as coroutines: the body is a goroutine, and control passes
// between it and the caller through two channels, so exactly one of them runs
// at a time, as in Python. A generator that is abandoned leaves its goroutine
// parked until the program ends.

type genMsg[S any] struct {
	val S
	exc *PyError
}

type genOut[Y, R any] struct {
	yielded Y
	ret     R
	done    bool
	err     *PyError
	fatal   any
}

const (
	genNew = iota
	genSuspended
	genRunning
	genDone
)

// Gen is a generator with yield type Y, send type S and return type R.
type Gen[Y, S, R any] struct {
	body  func(*Yielder[Y, S]) R
	in    chan genMsg[S]
	out   chan genOut[Y, R]
	state int
}

// Yielder is what the body of a generator yields through.
type Yielder[Y, S any] struct {
	yield func(Y) genMsg[S]
}

// NewGen makes a generator that has not started.
func NewGen[Y, S, R any](body func(*Yielder[Y, S]) R) *Gen[Y, S, R] {
	return &Gen[Y, S, R]{body: body}
}

// Yield is the expression "yield v": it hands v to the caller and returns the
// value sent back, or raises the exception thrown in.
func (y *Yielder[Y, S]) Yield(v Y) S {
	msg := y.yield(v)
	if msg.exc != nil {
		panic(msg.exc)
	}
	return msg.val
}

func (g *Gen[Y, S, R]) start() {
	g.in = make(chan genMsg[S])
	g.out = make(chan genOut[Y, R])
	go func() {
		<-g.in
		var out genOut[Y, R]
		func() {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(*PyError); ok {
						out.err = e
					} else {
						out.fatal = r
					}
				}
			}()
			out.ret = g.body(&Yielder[Y, S]{yield: func(v Y) genMsg[S] {
				g.out <- genOut[Y, R]{yielded: v}
				return <-g.in
			}})
		}()
		if out.err != nil && out.err.Type == "StopIteration" {
			// PEP 479: a StopIteration escaping a generator is a bug.
			wrapped := newError("RuntimeError", "generator raised StopIteration")
			wrapped.Cause, wrapped.SuppressContext = out.err, true
			out.err = wrapped
		}
		out.done = true
		g.out <- out
	}()
}

// step resumes the generator with msg. It returns the next yielded value, or
// the return value and done when the body finished.
func (g *Gen[Y, S, R]) step(msg genMsg[S]) (y Y, ret R, done bool) {
	switch g.state {
	case genRunning:
		panic(newError("ValueError", "generator already executing"))
	case genDone:
		if msg.exc != nil {
			panic(msg.exc)
		}
		return y, ret, true
	case genNew:
		if msg.exc != nil {
			g.state = genDone
			panic(msg.exc)
		}
		g.start()
	}
	g.state = genRunning
	g.in <- msg
	res := <-g.out
	if res.fatal != nil {
		g.state = genDone
		panic(res.fatal)
	}
	if res.done {
		g.state = genDone
		if res.err != nil {
			panic(res.err)
		}
		return y, res.ret, true
	}
	g.state = genSuspended
	return res.yielded, ret, false
}

// stopIteration is the exception that ends next(): its value is the return
// value of the generator, if it returned one.
func stopIteration[R any](ret R) *PyError {
	e := newError("StopIteration", "")
	e.Args = nil
	if _, none := any(ret).(NoneT); !none {
		e.Args = []any{ret}
	}
	return e
}

// Send is g.send(v); it is next(g) with the zero value. A generator that
// finishes raises StopIteration.
func (g *Gen[Y, S, R]) Send(v S) Y {
	y, ret, done := g.step(genMsg[S]{val: v})
	if done {
		panic(stopIteration(ret))
	}
	return y
}

// NextValue is next(g).
func (g *Gen[Y, S, R]) NextValue() Y {
	var zero S
	return g.Send(zero)
}

// Throw is g.throw(exc).
func (g *Gen[Y, S, R]) Throw(exc *PyError) Y {
	y, ret, done := g.step(genMsg[S]{exc: exc})
	if done {
		panic(stopIteration(ret))
	}
	return y
}

// Close is g.close(): the generator gets GeneratorExit at its yield, so its
// finally blocks run.
func (g *Gen[Y, S, R]) Close() {
	switch g.state {
	case genNew:
		g.state = genDone
		return
	case genDone:
		return
	}
	e := Catch(func() {
		_, _, done := g.step(genMsg[S]{exc: newError("GeneratorExit", "")})
		if !done {
			panic(newError("RuntimeError", "generator ignored GeneratorExit"))
		}
	})
	if e != nil && e.Type != "GeneratorExit" && e.Type != "StopIteration" {
		panic(e)
	}
}

type genCursor[Y, S, R any] struct {
	g   *Gen[Y, S, R]
	cur Y
}

func (c *genCursor[Y, S, R]) Next() bool {
	var zero S
	y, _, done := c.g.step(genMsg[S]{val: zero})
	if done {
		var z Y
		c.cur = z
		return false
	}
	c.cur = y
	return true
}

func (c *genCursor[Y, S, R]) Value() Y { return c.cur }

// Iter lets a generator be walked by a for loop, which drives it with next().
func (g *Gen[Y, S, R]) Iter() Cursor[Y] { return &genCursor[Y, S, R]{g: g} }

// Iterator views the generator as the one-shot iterator type.
func (g *Gen[Y, S, R]) Iterator() *Iterator[Y] {
	var zero S
	it := NewIterator(func() (Y, bool) {
		y, _, done := g.step(genMsg[S]{val: zero})
		return y, !done
	})
	it.closer, it.thrower = g.Close, g.Throw
	return it
}

// YieldFromGen is "yield from sub" for a generator: values, sends and thrown
// exceptions pass through, and the result is what sub returned.
func YieldFromGen[Y, S, R any](y *Yielder[Y, S], sub *Gen[Y, S, R]) R {
	var msg genMsg[S]
	for {
		v, ret, done := sub.step(msg)
		if done {
			return ret
		}
		msg = y.yield(v)
	}
}

// YieldFromIter is "yield from" an iterable that is not a generator.
func YieldFromIter[Y, S any](y *Yielder[Y, S], src Iterable[Y]) NoneT {
	// When the delegating generator is closed or fails at the yield, the
	// iterator it was walking is closed too, as CPython does.
	done := false
	defer func() {
		if !done {
			if c, ok := src.(interface{ Close() }); ok {
				c.Close()
			}
		}
	}()
	for c := src.Iter(); c.Next(); {
		y.Yield(c.Value())
	}
	done = true
	return NoneV
}

// NextValue is next(it) for an iterator: StopIteration at the end.
func (it *Iterator[T]) NextValue() T {
	if !it.Next() {
		panic(newError("StopIteration", ""))
	}
	return it.cur
}

// NextOr is next(it, default).
func NextOr[T any](it interface{ Iter() Cursor[T] }, def T) T {
	c := it.Iter()
	if c.Next() {
		return c.Value()
	}
	return def
}

// NextOpt is next(it, None).
func NextOpt[T any](it interface{ Iter() Cursor[T] }) Opt[T] {
	c := it.Iter()
	if c.Next() {
		return Some(c.Value())
	}
	return NoneOf[T]()
}

// IterOf is iter(x) for any iterable.
func IterOf[T any](src Iterable[T]) *Iterator[T] {
	c := src.Iter()
	return NewIterator(func() (T, bool) {
		if !c.Next() {
			var zero T
			return zero, false
		}
		return c.Value(), true
	})
}
