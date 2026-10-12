package runtime

import "fmt"

// Cursor walks a sequence one item at a time: Next moves to the next item
// and reports whether there is one, and Value returns it. Every for loop
// and every consumer of an iterable in generated code goes through one.
type Cursor[T any] interface {
	Next() bool
	Value() T
}

// Iterable is anything a for loop can walk.
type Iterable[T any] interface {
	Iter() Cursor[T]
}

// Iterator is a one-shot lazy sequence, the value of a generator expression,
// reversed(), zip() or enumerate(). Walking it a second time finds it empty.
type Iterator[T any] struct {
	next func() (T, bool)
	cur  T
	done bool

	// A generator viewed as an Iterator keeps its close and throw.
	closer  func()
	thrower func(*PyError) T
}

// Close is it.close() of a generator: its finally blocks run.
func (it *Iterator[T]) Close() {
	if it.closer != nil {
		it.closer()
	}
	it.done = true
}

// Throw is it.throw(exc) of a generator.
func (it *Iterator[T]) Throw(e *PyError) T {
	if it.thrower == nil {
		panic(newError("AttributeError", "'iterator' object has no attribute 'throw'"))
	}
	return it.thrower(e)
}

// NewIterator wraps a function that returns the next item, or false at the end.
func NewIterator[T any](next func() (T, bool)) *Iterator[T] {
	return &Iterator[T]{next: next}
}

func (it *Iterator[T]) Next() bool {
	if it.done {
		return false
	}
	v, ok := it.next()
	if !ok {
		it.done = true
		var zero T
		it.cur = zero
		return false
	}
	it.cur = v
	return true
}

func (it *Iterator[T]) Value() T        { return it.cur }
func (it *Iterator[T]) Iter() Cursor[T] { return it }

// ListFrom is list(iterable).
func ListFrom[T any](src Iterable[T]) *List[T] {
	out := NewList[T]()
	for c := src.Iter(); c.Next(); {
		out.Append(c.Value())
	}
	return out
}

// Any is any(iterable). It stops at the first true item, so a generator
// behind it runs only as far as needed.
func Any[T any](src Iterable[T]) bool {
	for c := src.Iter(); c.Next(); {
		if Truth(c.Value()) {
			return true
		}
	}
	return false
}

// All is all(iterable), stopping at the first false item.
func All[T any](src Iterable[T]) bool {
	for c := src.Iter(); c.Next(); {
		if !Truth(c.Value()) {
			return false
		}
	}
	return true
}

// truther is implemented by the container types, whose truth is "not empty".
type truther interface{ pyTruth() bool }

// Truth is Python's truth test for the supported types.
func Truth(v any) bool {
	switch v := v.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case int:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case Seq:
		return v.Len() != 0
	case truther:
		return v.pyTruth()
	}
	return true
}

// Enumerate is enumerate(iterable, start).
func Enumerate[T any](src Iterable[T], start int64) *Iterator[Tuple2[int64, T]] {
	c := src.Iter()
	i := start
	return NewIterator(func() (Tuple2[int64, T], bool) {
		if !c.Next() {
			return Tuple2[int64, T]{}, false
		}
		t := Tup2(i, c.Value())
		i++
		return t, true
	})
}

// Reversed is reversed(sequence). A list is read live, as CPython does: if
// it shrinks while the iterator runs, the iterator stops.
func Reversed[T any](l *List[T]) *Iterator[T] {
	i := l.Len() - 1
	return NewIterator(func() (T, bool) {
		if i < 0 || i >= l.Len() {
			i = -1
			var zero T
			return zero, false
		}
		v := l.items[i]
		i--
		return v, true
	})
}

// ReversedStr is reversed(str).
func ReversedStr(s string) *Iterator[string] { return Reversed(Chars(s)) }

// ListCursor walks a list. It reads the length at every step, so appending
// to the list while looping over it behaves as in Python.
type ListCursor[T any] struct {
	l   *List[T]
	i   int
	cur T
}

func (c *ListCursor[T]) Next() bool {
	if c.i < len(c.l.items) {
		c.cur = c.l.items[c.i]
		c.i++
		return true
	}
	return false
}

func (c *ListCursor[T]) Value() T { return c.cur }

// Iter and Cursor start a walk over the list.
func (l *List[T]) Iter() Cursor[T]   { return &ListCursor[T]{l: l} }
func (l *List[T]) Cursor() Cursor[T] { return &ListCursor[T]{l: l} }

func (l *List[T]) pyTruth() bool { return len(l.items) > 0 }

// Zip2 to Zip6 are zip(a, b, ...). They stop at the shortest input and, as
// in CPython, advance the inputs in order, so a longer input before the end
// loses one item.
func Zip2[A, B any](a Iterable[A], b Iterable[B]) *Iterator[Tuple2[A, B]] {
	ca, cb := a.Iter(), b.Iter()
	return NewIterator(func() (Tuple2[A, B], bool) {
		if ca.Next() && cb.Next() {
			return Tup2(ca.Value(), cb.Value()), true
		}
		return Tuple2[A, B]{}, false
	})
}

func Zip3[A, B, C any](a Iterable[A], b Iterable[B], c Iterable[C]) *Iterator[Tuple3[A, B, C]] {
	c0, c1, c2 := a.Iter(), b.Iter(), c.Iter()
	return NewIterator(func() (Tuple3[A, B, C], bool) {
		if c0.Next() && c1.Next() && c2.Next() {
			return Tup3(c0.Value(), c1.Value(), c2.Value()), true
		}
		return Tuple3[A, B, C]{}, false
	})
}

func Zip4[A, B, C, D any](a Iterable[A], b Iterable[B], c Iterable[C], d Iterable[D]) *Iterator[Tuple4[A, B, C, D]] {
	c0, c1, c2, c3 := a.Iter(), b.Iter(), c.Iter(), d.Iter()
	return NewIterator(func() (Tuple4[A, B, C, D], bool) {
		if c0.Next() && c1.Next() && c2.Next() && c3.Next() {
			return Tup4(c0.Value(), c1.Value(), c2.Value(), c3.Value()), true
		}
		return Tuple4[A, B, C, D]{}, false
	})
}

func Zip5[A, B, C, D, E any](a Iterable[A], b Iterable[B], c Iterable[C], d Iterable[D], e Iterable[E]) *Iterator[Tuple5[A, B, C, D, E]] {
	c0, c1, c2, c3, c4 := a.Iter(), b.Iter(), c.Iter(), d.Iter(), e.Iter()
	return NewIterator(func() (Tuple5[A, B, C, D, E], bool) {
		if c0.Next() && c1.Next() && c2.Next() && c3.Next() && c4.Next() {
			return Tup5(c0.Value(), c1.Value(), c2.Value(), c3.Value(), c4.Value()), true
		}
		return Tuple5[A, B, C, D, E]{}, false
	})
}

func Zip6[A, B, C, D, E, F any](a Iterable[A], b Iterable[B], c Iterable[C], d Iterable[D], e Iterable[E], f Iterable[F]) *Iterator[Tuple6[A, B, C, D, E, F]] {
	c0, c1, c2, c3, c4, c5 := a.Iter(), b.Iter(), c.Iter(), d.Iter(), e.Iter(), f.Iter()
	return NewIterator(func() (Tuple6[A, B, C, D, E, F], bool) {
		if c0.Next() && c1.Next() && c2.Next() && c3.Next() && c4.Next() && c5.Next() {
			return Tup6(c0.Value(), c1.Value(), c2.Value(), c3.Value(), c4.Value(), c5.Value()), true
		}
		return Tuple6[A, B, C, D, E, F]{}, false
	})
}

// Zip1 is zip(a), which yields one-element tuples.
func Zip1[A any](a Iterable[A]) *Iterator[Tuple1[A]] {
	c := a.Iter()
	return NewIterator(func() (Tuple1[A], bool) {
		if c.Next() {
			return Tup1(c.Value()), true
		}
		return Tuple1[A]{}, false
	})
}

// Spread is the items of an iterable as a []any, for print(*items).
func Spread[T any](src Iterable[T]) []any {
	var out []any
	for c := src.Iter(); c.Next(); {
		out = append(out, c.Value())
	}
	return out
}

// Nexter is an iterator written in Python: an instance with __next__.
type Nexter[T any] interface{ M___next__() T }

// UserIterable is an instance with __iter__ that returns a Nexter. Each walk
// asks for the iterator again, as a for loop does, and ends when __next__
// raises StopIteration.
type UserIterable[T any] struct{ Start func() Nexter[T] }

func (u UserIterable[T]) Iter() Cursor[T] { return &userCursor[T]{it: u.Start()} }

type userCursor[T any] struct {
	it  Nexter[T]
	cur T
}

func (c *userCursor[T]) Next() bool {
	e := Catch(func() { c.cur = c.it.M___next__() })
	if e == nil {
		return true
	}
	if !IsSubclass(e.Type, "StopIteration") {
		panic(e)
	}
	var zero T
	c.cur = zero
	return false
}

func (c *userCursor[T]) Value() T { return c.cur }

func (it *Iterator[T]) PyRepr() string { return "<iterator object at " + addr(it) + ">" }

func (g *Gen[Y, S, R]) PyRepr() string { return "<generator object at " + addr(g) + ">" }

func addr(p any) string { return fmt.Sprintf("%p", p) }
