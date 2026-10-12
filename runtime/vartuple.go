package runtime

// VarTuple is a tuple of any length whose items share one type, the value of
// tuple[T, ...]. It is what a *args parameter holds, and what tuple(x) builds.
// Like a Tuple it is immutable.
type VarTuple[T any] struct {
	items []T
}

func NewVarTuple[T any](items ...T) VarTuple[T] { return VarTuple[T]{items: items} }

// TupleFrom is tuple(iterable).
func TupleFrom[T any](src Iterable[T]) VarTuple[T] {
	return VarTuple[T]{items: ListFrom(src).items}
}

func (t VarTuple[T]) Len() int64     { return int64(len(t.items)) }
func (t VarTuple[T]) At(i int64) any { return t.items[i] }
func (t VarTuple[T]) Kind() string   { return "tuple" }
func (t VarTuple[T]) pyTruth() bool  { return len(t.items) > 0 }

// Get is t[i], with negative indexes counting from the end.
func (t VarTuple[T]) Get(i int64) T {
	n := int64(len(t.items))
	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		panic(indexError("tuple index out of range"))
	}
	return t.items[i]
}

func (t VarTuple[T]) Slice(b Bounds) VarTuple[T] {
	start, step, count := b.indices(int64(len(t.items)))
	items := make([]T, 0, count)
	for i := int64(0); i < count; i++ {
		items = append(items, t.items[start+i*step])
	}
	return VarTuple[T]{items: items}
}

func (t VarTuple[T]) Concat(o VarTuple[T]) VarTuple[T] {
	items := make([]T, 0, len(t.items)+len(o.items))
	return VarTuple[T]{items: append(append(items, t.items...), o.items...)}
}

func (t VarTuple[T]) Repeat(n int64) VarTuple[T] {
	var items []T
	for i := int64(0); i < n; i++ {
		items = append(items, t.items...)
	}
	return VarTuple[T]{items: items}
}

func (t VarTuple[T]) Contains(v T) bool { return t.Index0(v) >= 0 }

// Index0 is the position of v, or -1.
func (t VarTuple[T]) Index0(v T) int {
	for i, x := range t.items {
		if Eq(x, v) {
			return i
		}
	}
	return -1
}

func (t VarTuple[T]) Index(v T) int64 {
	i := t.Index0(v)
	if i < 0 {
		panic(valueError("tuple.index(x): x not in tuple"))
	}
	return int64(i)
}

func (t VarTuple[T]) Count(v T) int64 {
	var n int64
	for _, x := range t.items {
		if Eq(x, v) {
			n++
		}
	}
	return n
}

type varTupleCursor[T any] struct {
	t   VarTuple[T]
	i   int
	cur T
}

func (c *varTupleCursor[T]) Next() bool {
	if c.i < len(c.t.items) {
		c.cur = c.t.items[c.i]
		c.i++
		return true
	}
	return false
}

func (c *varTupleCursor[T]) Value() T { return c.cur }

func (t VarTuple[T]) Iter() Cursor[T] { return &varTupleCursor[T]{t: t} }

// ToList copies the items into a list.
func (t VarTuple[T]) ToList() *List[T] { return &List[T]{items: append([]T(nil), t.items...)} }
