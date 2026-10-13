package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// Slice is the object a[lo:hi:step] indexes with when a is not statically
// known. A missing part is nil.
type Slice struct{ Lo, Hi, Step any }

func (s Slice) PyRepr() string {
	return fmt.Sprintf("slice(%s, %s, %s)", Repr(s.Lo), Repr(s.Hi), Repr(s.Step))
}

// Bounds converts the slice to the bounds the typed containers take.
func (s Slice) Bounds() Bounds {
	var b Bounds
	read := func(v any, dst *int64, has *bool) {
		v = Unbox(v)
		if v == nil {
			return
		}
		n, ok := intOf(v)
		if !ok {
			if bi, isBig := v.(*big.Int); isBig {
				n, ok = 1<<62, true
				if bi.Sign() < 0 {
					n = -(1 << 62)
				}
			}
		}
		if !ok {
			panic(newError("TypeError", "slice indices must be integers or None or have an __index__ method"))
		}
		*dst, *has = n, true
	}
	read(s.Lo, &b.Lo, &b.HasLo)
	read(s.Hi, &b.Hi, &b.HasHi)
	read(s.Step, &b.Step, &b.HasStep)
	return b
}

// Cast reads a dynamic value as a value of the static type T, raising
// TypeError when it is something else.
func Cast[T any](v any) T { return conv[T](v) }

// conv reads a dynamic value as the element type of a typed container.
func conv[T any](v any) T {
	v = Unbox(v)
	if t, ok := v.(T); ok {
		return t
	}
	if v == nil {
		if o, isOpt := any(new(T)).(anyOpt); isOpt {
			o.setFromAny(nil)
		}
		return *new(T)
	}
	if c, ok := v.(*Callable); ok {
		if t, ok := c.Fn.(T); ok {
			return t
		}
	}
	var zero T
	if tp, ok := any(&zero).(anyTuple); ok {
		if s, isSeq := v.(Seq); isSeq && tp.setFromSeq(s) {
			return zero
		}
	}
	if o, ok := any(&zero).(anyOpt); ok {
		if o.setFromAny(v) {
			return zero
		}
		panic(newError("TypeError", fmt.Sprintf("expected %s, got %s", TypeName(zero), TypeName(v))))
	}
	switch any(zero).(type) {
	case float64:
		if k, i, _ := numeric(v); k == isInt {
			return any(float64(i)).(T)
		}
	case int64:
		if b, ok := v.(bool); ok {
			return any(b2i(b)).(T)
		}
		if _, big := v.(*big.Int); big {
			panic(overflowError("Python int too large to convert to C long"))
		}
	case any:
		return any(v).(T)
	}
	if v == nil {
		return zero
	}
	panic(newError("TypeError", fmt.Sprintf("expected %s, got %s", TypeName(zero), TypeName(v))))
}

// convOpt is conv that reports a mismatch instead of raising it.
func convOpt[T any](v any) (T, bool) {
	var zero T
	ok := true
	// Numbers are equal across types when they are looked up: 1 == 1.0 == True.
	switch any(zero).(type) {
	case int64:
		switch x := Unbox(v).(type) {
		case float64:
			if x == math.Trunc(x) && math.Abs(x) < 1<<62 {
				return any(int64(x)).(T), true
			}
			return zero, false
		case bool:
			return any(b2i(x)).(T), true
		}
	case bool:
		switch x := Unbox(v).(type) {
		case int64:
			if x == 0 || x == 1 {
				return any(x == 1).(T), true
			}
			return zero, false
		case float64:
			if x == 0 || x == 1 {
				return any(x == 1).(T), true
			}
			return zero, false
		}
	case float64:
		switch x := Unbox(v).(type) {
		case *big.Int:
			return any(bigToFloatLoose(x)).(T), true
		}
	}
	func() {
		defer func() {
			if recover() != nil {
				ok = false
			}
		}()
		zero = conv[T](v)
	}()
	return zero, ok
}

// indexArg reads the index of a subscript.
func indexArg(key any) (int64, bool) {
	key = Unbox(key)
	if isBig(key) {
		panic(indexError("cannot fit 'int' into an index-sized integer"))
	}
	n, ok := intOf(key)
	return n, ok
}

// DynIter is iter(x) as a source of items for a for loop or a consumer.
func DynIter(v any) Iterable[any] {
	v = Unbox(v)
	switch x := v.(type) {
	case Iterable[any]:
		return x
	case interface{ IterAny() Cursor[any] }:
		return dynIterable{x}
	case string:
		return dynIterable{stringIter(x)}
	case Seq:
		return dynIterable{seqIter{x}}
	case interface{ PyIterAny() (Cursor[any], bool) }:
		return funcIterable(func() Cursor[any] { c2, _ := x.PyIterAny(); return c2 })
	case interface{ PyNext() any }:
		return funcIterable(func() Cursor[any] { return &nextCursor{it: x} })
	}
	panic(newError("TypeError", "'"+TypeName(v)+"' object is not iterable"))
}

type dynIterable struct {
	src interface{ IterAny() Cursor[any] }
}

func (d dynIterable) Iter() Cursor[any] { return d.src.IterAny() }

type funcIterable func() Cursor[any]

func (f funcIterable) Iter() Cursor[any] { return f() }

type stringIter string

func (s stringIter) IterAny() Cursor[any] {
	return &sliceCursor[any]{items: stringChars(string(s))}
}

func stringChars(s string) []any {
	var out []any
	for _, r := range s {
		out = append(out, string(r))
	}
	return out
}

type seqIter struct{ s Seq }

func (s seqIter) IterAny() Cursor[any] {
	items := make([]any, s.s.Len())
	for i := range items {
		items[i] = s.s.At(int64(i))
	}
	return &sliceCursor[any]{items: items}
}

type sliceCursor[T any] struct {
	items []T
	i     int
	cur   T
}

func (c *sliceCursor[T]) Next() bool {
	if c.i >= len(c.items) {
		return false
	}
	c.cur = c.items[c.i]
	c.i++
	return true
}
func (c *sliceCursor[T]) Value() T { return c.cur }

// DynList is list(x).
func DynList(v any) *List[any] {
	out := NewList[any]()
	for c := DynIter(v).Iter(); c.Next(); {
		out.Append(c.Value())
	}
	return out
}

// DynGetItem is obj[key].
func DynGetItem(obj, key any) any {
	obj, key = Unbox(obj), Unbox(key)
	switch x := obj.(type) {
	case string:
		if s, ok := key.(Slice); ok {
			return StrSlice(x, s.Bounds())
		}
		if i, ok := indexArg(key); ok {
			return StrIndex(x, i)
		}
		panic(newError("TypeError", "string indices must be integers, not '"+TypeName(key)+"'"))
	case interface{ GetItemAny(key any) any }:
		return x.GetItemAny(key)
	case interface{ PyGetItem(key any) (any, bool) }:
		if r, ok := x.PyGetItem(key); ok {
			return r
		}
	case Seq:
		if s, ok := key.(Slice); ok {
			items := make([]any, 0)
			for i := int64(0); i < x.Len(); i++ {
				items = append(items, x.At(i))
			}
			return NewVarTuple(items...).Slice(s.Bounds())
		}
		if i, ok := indexArg(key); ok {
			if i < 0 {
				i += x.Len()
			}
			if i < 0 || i >= x.Len() {
				panic(indexError("tuple index out of range"))
			}
			return x.At(i)
		}
		panic(newError("TypeError", "tuple indices must be integers or slices, not "+TypeName(key)))
	}
	panic(newError("TypeError", "'"+TypeName(obj)+"' object is not subscriptable"))
}

// DynSetItem is obj[key] = v.
func DynSetItem(obj, key, v any) {
	obj, key = Unbox(obj), Unbox(key)
	switch x := obj.(type) {
	case interface{ SetItemAny(key, v any) }:
		x.SetItemAny(key, v)
		return
	case interface{ PySetItem(key, v any) bool }:
		if x.PySetItem(key, v) {
			return
		}
	}
	panic(newError("TypeError", "'"+TypeName(obj)+"' object does not support item assignment"))
}

// DynDelItem is del obj[key].
func DynDelItem(obj, key any) {
	obj, key = Unbox(obj), Unbox(key)
	switch x := obj.(type) {
	case interface{ DelItemAny(key any) }:
		x.DelItemAny(key)
		return
	case interface{ PyDelItem(key any) bool }:
		if x.PyDelItem(key) {
			return
		}
	}
	panic(newError("TypeError", "'"+TypeName(obj)+"' object doesn't support item deletion"))
}

// ---- lists ----

func (l *List[T]) IterAny() Cursor[any] {
	c := l.Iter()
	return anyCursor[T]{c}
}

type anyCursor[T any] struct{ c Cursor[T] }

func (a anyCursor[T]) Next() bool { return a.c.Next() }
func (a anyCursor[T]) Value() any { return a.c.Value() }
func (l *List[T]) ContainsAny(item any) bool {
	for _, x := range l.items {
		if Eq(x, item) {
			return true
		}
	}
	return false
}

func (l *List[T]) GetItemAny(key any) any {
	if s, ok := key.(Slice); ok {
		return l.Slice(s.Bounds())
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "list indices must be integers or slices, not "+TypeName(key)))
	}
	return l.Get(i)
}

func (l *List[T]) SetItemAny(key, v any) {
	i, ok := indexArg(key)
	if !ok {
		if s, isSlice := key.(Slice); isSlice {
			l.setSlice(s.Bounds(), v)
			return
		}
		panic(newError("TypeError", "list indices must be integers or slices, not "+TypeName(key)))
	}
	l.Set(i, conv[T](v))
}

func (l *List[T]) DelItemAny(key any) {
	if s, ok := key.(Slice); ok {
		l.DelSlice(s.Bounds())
		return
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "list indices must be integers or slices, not "+TypeName(key)))
	}
	l.DelAt(i)
}

func (l *List[T]) DynBinary(op string, other any, reflected bool) (any, bool) {
	switch op {
	case "Add":
		if o, ok := other.(*List[T]); ok && !reflected {
			return l.Concat(o), true
		}
		if o, ok := other.(*List[T]); ok && reflected {
			return o.Concat(l), true
		}
		if _, isList := other.(Seq); isList {
			if s, ok := other.(Seq); ok && s.Kind() == "list" {
				out := NewList[any]()
				for _, x := range l.items {
					out.Append(x)
				}
				for i := int64(0); i < s.Len(); i++ {
					out.Append(s.At(i))
				}
				if reflected {
					swapped := NewList[any]()
					for i := int64(0); i < s.Len(); i++ {
						swapped.Append(s.At(i))
					}
					for _, x := range l.items {
						swapped.Append(x)
					}
					return swapped, true
				}
				return out, true
			}
		}
	case "Mult":
		if n, ok := intOf(other); ok {
			return l.Repeat(n), true
		}
	}
	return nil, false
}

// ---- dicts ----

func (d *Dict[K, V]) IterAny() Cursor[any] { return anyCursor[K]{d.Iter()} }

func (d *Dict[K, V]) ContainsAny(item any) bool {
	k, ok := convOpt[K](item)
	return ok && d.Has(k)
}

func (d *Dict[K, V]) GetItemAny(key any) any {
	k, ok := convOpt[K](key)
	if !ok {
		panic(NewError("KeyError", key))
	}
	return d.Get(k)
}

func (d *Dict[K, V]) SetItemAny(key, v any) { d.Set(conv[K](key), conv[V](v)) }

func (d *Dict[K, V]) DelItemAny(key any) {
	k, ok := convOpt[K](key)
	if !ok {
		panic(NewError("KeyError", key))
	}
	d.Delete(k)
}

func (d *Dict[K, V]) DynBinary(op string, other any, reflected bool) (any, bool) {
	if o, ok := other.(*Dict[K, V]); ok && op == "BitOr" {
		if reflected {
			return o.Union(d), true
		}
		return d.Union(o), true
	}
	return nil, false
}

// ---- sets ----

func (s *Set[T]) IterAny() Cursor[any] { return anyCursor[T]{s.Iter()} }

func (s *Set[T]) ContainsAny(item any) bool {
	k, ok := convOpt[T](item)
	return ok && s.Has(k)
}

// anySet copies the set as a set of objects.
func (s *Set[T]) anySet() *Set[any] {
	out := NewSet[any]()
	for c := s.Iter(); c.Next(); {
		out.Add(c.Value())
	}
	return out
}

func (s *Set[T]) DynBinary(op string, other any, reflected bool) (any, bool) {
	if fs, isFrozen := other.(*FrozenSet); isFrozen && !reflected {
		// set op frozenset is a set, of the kind of the left operand.
		r, ok := (&FrozenSet{s.anySet()}).PyBinop(op, fs, false)
		if !ok {
			return nil, false
		}
		res := r.(*FrozenSet).s
		out := NewSet[T]()
		for c := res.Iter(); c.Next(); {
			k, ok := convOpt[T](c.Value())
			if !ok {
				return res, true
			}
			out.Add(k)
		}
		return out, true
	}
	o, ok := other.(*Set[T])
	if !ok {
		return nil, false
	}
	a, b := s, o
	if reflected {
		a, b = o, s
	}
	switch op {
	case "BitOr":
		return a.Union(b), true
	case "BitAnd":
		return a.Intersection(b), true
	case "Sub":
		return a.Difference(b), true
	case "BitXor":
		return a.SymmetricDifference(b), true
	}
	return nil, false
}

// ---- tuples, ranges, views and iterators ----

func (t VarTuple[T]) IterAny() Cursor[any] { return anyCursor[T]{t.Iter()} }

func (t VarTuple[T]) ContainsAny(item any) bool {
	for _, x := range t.items {
		if Eq(x, item) {
			return true
		}
	}
	return false
}

func (t VarTuple[T]) GetItemAny(key any) any {
	if s, ok := key.(Slice); ok {
		return t.Slice(s.Bounds())
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "tuple indices must be integers or slices, not "+TypeName(key)))
	}
	return t.Get(i)
}

func (t VarTuple[T]) DynBinary(op string, other any, reflected bool) (any, bool) {
	switch op {
	case "Add":
		if o, ok := other.(VarTuple[T]); ok {
			if reflected {
				return o.Concat(t), true
			}
			return t.Concat(o), true
		}
	case "Mult":
		if n, ok := intOf(other); ok {
			return t.Repeat(n), true
		}
	}
	return nil, false
}

func (r *Range) IterAny() Cursor[any] { return anyCursor[int64]{r.Iter()} }
func (r *Range) ContainsAny(item any) bool {
	v := Unbox(item)
	if f, isFloat := v.(float64); isFloat {
		if math.IsInf(f, 0) || math.IsNaN(f) || f != math.Trunc(f) || math.Abs(f) >= 1<<62 {
			return false
		}
		v = int64(f)
	}
	n, ok := intOf(v)
	return ok && !isBig(v) && r.Has(n)
}
func (r *Range) GetItemAny(key any) any {
	if s, isSlice := key.(Slice); isSlice {
		start, step, count := s.Bounds().indices(r.Len())
		first := r.start + start*r.step
		return NewRange(first, first+count*step*r.step, step*r.step)
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "range indices must be integers or slices, not "+TypeName(key)))
	}
	return r.Get(i)
}

func (it *Iterator[T]) IterAny() Cursor[any] { return anyCursor[T]{it} }
func (g *Gen[Y, S, R]) IterAny() Cursor[any] { return anyCursor[Y]{g.Iter()} }

func (v *KeysView[K, V]) IterAny() Cursor[any]   { return anyCursor[K]{v.Iter()} }
func (v *ValuesView[K, V]) IterAny() Cursor[any] { return anyCursor[V]{v.Iter()} }
func (v *ItemsView[K, V]) IterAny() Cursor[any]  { return anyCursor[Tuple2[K, V]]{v.Iter()} }

func (v *KeysView[K, V]) ContainsAny(item any) bool {
	k, ok := convOpt[K](item)
	return ok && v.d.Has(k)
}

// ---- tuples built at run time ----

// DynTuple is tuple(x).
func DynTuple(v any) VarTuple[any] {
	var items []any
	for c := DynIter(v).Iter(); c.Next(); {
		items = append(items, c.Value())
	}
	return NewVarTuple(items...)
}

// DynSet is set(x), over items of any type.
func DynSet(v any) *Set[any] {
	s := NewSet[any]()
	for c := DynIter(v).Iter(); c.Next(); {
		s.Add(c.Value())
	}
	return s
}

// DynDict is dict(x): from another dict or from pairs.
func DynDict(v any) *Dict[any, any] {
	v = Unbox(v)
	out := NewDict[any, any]()
	if d, ok := v.(interface{ IterAny() Cursor[any] }); ok {
		if dd, isDict := v.(interface{ GetItemAny(any) any }); isDict {
			if _, isList := v.(Seq); !isList {
				for c := d.IterAny(); c.Next(); {
					k := c.Value()
					out.Set(k, dd.GetItemAny(k))
				}
				return out
			}
		}
	}
	if keys, isMap := userMappingKeys(v); isMap {
		for _, k := range keys {
			out.Set(k, DynGetItem(v, k))
		}
		return out
	}
	idx := 0
	for c := DynIter(v).Iter(); c.Next(); idx++ {
		pair := Unbox(c.Value())
		s, ok := pair.(Seq)
		if !ok {
			if _, isStr := pair.(string); isStr {
				s = NewVarTuple[any](stringChars(pair.(string))...)
			} else {
				panic(newError("TypeError", fmt.Sprintf("cannot convert dictionary update sequence element #%d to a sequence", idx)))
			}
		}
		if s.Len() != 2 {
			panic(newError("ValueError", fmt.Sprintf("dictionary update sequence element #%d has length %d; 2 is required", idx, s.Len())))
		}
		out.Set(s.At(0), s.At(1))
	}
	return out
}

// DynJoin is sep.join(items) for items of any type.
func DynJoin(sep string, items any) string {
	if Catch(func() { DynIter(items) }) != nil {
		panic(newError("TypeError", "can only join an iterable"))
	}
	var parts []string
	for c := DynIter(items).Iter(); c.Next(); {
		s, ok := Unbox(c.Value()).(string)
		if !ok {
			panic(newError("TypeError", fmt.Sprintf("sequence item %d: expected str instance, %s found", len(parts), TypeName(c.Value()))))
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, sep)
}

// DynSum is sum(items, start).
func DynSum(items any, start any) any {
	acc := start
	if acc == nil {
		acc = int64(0)
	}
	switch Unbox(acc).(type) {
	case string:
		panic(newError("TypeError", "sum() can't sum strings [use ''.join(seq) instead]"))
	case Bytes:
		panic(newError("TypeError", "sum() can't sum bytes [use b''.join(seq) instead]"))
	case *ByteArray:
		panic(newError("TypeError", "sum() can't sum bytearray [use b''.join(seq) instead]"))
	}
	// Floats are added with Neumaier's compensation, as CPython 3.12+ does.
	var f, comp float64
	fmode := false
	flush := func() {
		if comp != 0 && !math.IsInf(comp, 0) && !math.IsNaN(comp) {
			f += comp
		}
		acc, comp, fmode = f, 0, false
	}
	for c := DynIter(items).Iter(); c.Next(); {
		v := Unbox(c.Value())
		if !fmode {
			if x, ok := Unbox(acc).(float64); ok {
				f, comp, fmode = x, 0, true
			}
		}
		if fmode {
			switch x := v.(type) {
			case float64:
				t := f + x
				if math.Abs(f) >= math.Abs(x) {
					comp += (f - t) + x
				} else {
					comp += (x - t) + f
				}
				f = t
				continue
			case int64:
				f += float64(x)
				continue
			case bool:
				f += float64(b2i(x))
				continue
			}
			flush()
		}
		acc = DynBinary("Add", acc, v)
	}
	if fmode {
		flush()
	}
	return acc
}

// DynMinMax is min(items) and max(items); isMax picks which.
func DynMinMax(isMax bool, items any, key any, def any, hasDef bool) any {
	var best, bestKey any
	first := true
	for c := DynIter(items).Iter(); c.Next(); {
		v := c.Value()
		k := v
		if key != nil {
			k = DynCall(key, v)
		}
		if first {
			best, bestKey, first = v, k, false
			continue
		}
		if (isMax && Gt(k, bestKey)) || (!isMax && Lt(k, bestKey)) {
			best, bestKey = v, k
		}
	}
	if first {
		if hasDef {
			return def
		}
		name := "min"
		if isMax {
			name = "max"
		}
		panic(valueError(name + "() iterable argument is empty"))
	}
	return best
}

// DynSorted is sorted(items, key=, reverse=).
func DynSorted(items any, key any, reverse bool) *List[any] {
	l := DynList(items)
	if key == nil {
		l.Sort(reverse)
		return l
	}
	SortKey(l, func(v any) any { return DynCall(key, v) }, reverse)
	return l
}

// DynReversed is reversed(x).
func DynReversed(v any) *Iterator[any] {
	v = Unbox(v)
	l := DynList(v)
	if n := TypeName(v); n == "set" || n == "frozenset" {
		panic(newError("TypeError", "'"+n+"' object is not reversible"))
	}
	if _, ok := v.(*Dict[any, any]); !ok {
		switch v.(type) {
		case Seq, string, *Range:
		default:
			if _, isIter := v.(*Iterator[any]); isIter {
				panic(newError("TypeError", "'iterator' object is not reversible"))
			}
		}
	}
	items := l.items
	i := len(items)
	return NewIterator(func() (any, bool) {
		if i == 0 {
			return nil, false
		}
		i--
		return items[i], true
	})
}

// DynEnumerate is enumerate(x, start).
func DynEnumerate(v any, start int64) *Iterator[Tuple2[int64, any]] {
	return Enumerate[any](DynIter(v), start)
}

// DynAny and DynAll are any(x) and all(x).
func DynAny(v any) bool {
	for c := DynIter(v).Iter(); c.Next(); {
		if Truth(Unbox(c.Value())) {
			return true
		}
	}
	return false
}

func DynAll(v any) bool {
	for c := DynIter(v).Iter(); c.Next(); {
		if !Truth(Unbox(c.Value())) {
			return false
		}
	}
	return true
}

// SetFromAny fills a tuple from dynamic values, for calls that pass *args.
func (t *VarTuple[T]) SetFromAny(items []any) {
	t.items = make([]T, len(items))
	for i, a := range items {
		t.items[i] = conv[T](a)
	}
}

// InitAny prepares an empty dict made by reflection.
func (d *Dict[K, V]) InitAny() { d.index = map[K]int{} }

// IterResult is the cursor for the value an __iter__ method returned: an
// iterator class instance walks through its __next__, anything else is
// iterated as usual.
func IterResult(it any) Cursor[any] {
	it = Unbox(it)
	if n, ok := it.(interface{ PyNext() any }); ok {
		return &nextCursor{it: n}
	}
	return DynIter(it).Iter()
}

// nextCursor calls __next__ until it raises StopIteration.
type nextCursor struct {
	it  interface{ PyNext() any }
	cur any
}

func (c *nextCursor) Next() bool {
	e := Catch(func() { c.cur = c.it.PyNext() })
	if e == nil {
		return true
	}
	if !IsSubclass(e.Type, "StopIteration") {
		panic(e)
	}
	c.cur = nil
	return false
}

func (c *nextCursor) Value() any { return c.cur }

// DynMap is map(f, iterables...).
func DynMap(f any, its ...any) *Iterator[any] {
	cursors := make([]Cursor[any], len(its))
	for i, it := range its {
		cursors[i] = DynIter(it).Iter()
	}
	return NewIterator(func() (any, bool) {
		args := make([]any, len(cursors))
		for i, c := range cursors {
			if !c.Next() {
				return nil, false
			}
			args[i] = c.Value()
		}
		return DynCall(f, args...), true
	})
}

// DynFilter is filter(f, iterable); f may be None.
func DynFilter(f any, it any) *Iterator[any] {
	c := DynIter(it).Iter()
	return NewIterator(func() (any, bool) {
		for c.Next() {
			v := c.Value()
			var keep bool
			if IsNone(f) {
				keep = DynBool(v)
			} else {
				keep = DynBool(DynCall(f, v))
			}
			if keep {
				return v, true
			}
		}
		return nil, false
	})
}

// DynZip is zip(iterables...).
func DynZip(its ...any) *Iterator[VarTuple[any]] {
	cursors := make([]Cursor[any], len(its))
	for i, it := range its {
		cursors[i] = DynIter(it).Iter()
	}
	return NewIterator(func() (VarTuple[any], bool) {
		if len(cursors) == 0 {
			return VarTuple[any]{}, false
		}
		vals := make([]any, len(cursors))
		for i, c := range cursors {
			if !c.Next() {
				return VarTuple[any]{}, false
			}
			vals[i] = c.Value()
		}
		return NewVarTuple(vals...), true
	})
}

// CastIterable reads an iterable of objects as an iterable of T, checking each
// item as it is produced.
func CastIterable[T any](src Iterable[any]) Iterable[T] { return castIterable[T]{src} }

type castIterable[T any] struct{ src Iterable[any] }

func (c castIterable[T]) Iter() Cursor[T] { return castCursor[T]{c.src.Iter()} }

type castCursor[T any] struct{ c Cursor[any] }

func (c castCursor[T]) Next() bool { return c.c.Next() }
func (c castCursor[T]) Value() T   { return conv[T](c.c.Value()) }

// StrDict copies a dict with str keys into one with object values.
func StrDict(d any) *Dict[string, any] {
	out := NewDict[string, any]()
	src := DynDict(d)
	for c := src.Iter(); c.Next(); {
		k := c.Value()
		out.Set(k.(string), src.Get(k))
	}
	return out
}

// IsSequence is true for the values a sequence pattern matches: lists, tuples
// and ranges, but not strings.
func IsSequence(v any) bool {
	v = Unbox(v)
	if _, isStr := v.(string); isStr {
		return false
	}
	_, ok := v.(Seq)
	return ok
}

// IsMapping is true for the values a mapping pattern matches.
func IsMapping(v any) bool {
	v = Unbox(v)
	return TypeName(v) == "dict"
}

// anyOpt is implemented by *Opt[T]: it reads an object as a value or None.
type anyOpt interface{ setFromAny(v any) bool }

func (o *Opt[T]) setFromAny(v any) bool {
	if IsNone(v) {
		*o = Opt[T]{}
		return true
	}
	x, ok := convOpt[T](v)
	if !ok {
		return false
	}
	*o = Some(x)
	return true
}

// SeqIterable walks the items of a fixed-size tuple whose items all have the
// type T.
func SeqIterable[T any](s Seq) Iterable[T] {
	items := make([]T, s.Len())
	for i := range items {
		items[i] = conv[T](s.At(int64(i)))
	}
	return &List[T]{items: items}
}

// setSlice is l[a:b:c] = iterable. Without a step the slice is replaced by
// the items, whatever their number; with one the counts must agree.
func (l *List[T]) setSlice(b Bounds, v any) {
	var items []T
	for c := DynIter(v).Iter(); c.Next(); {
		items = append(items, conv[T](c.Value()))
	}
	start, step, count := b.indices(int64(len(l.items)))
	if !b.HasStep || b.Step == 1 {
		if step == 1 && count == 0 {
			// An empty slice inserts at the clamped position.
		}
		end := start + count
		if end < start {
			end = start
		}
		out := append([]T{}, l.items[:start]...)
		out = append(out, items...)
		out = append(out, l.items[end:]...)
		l.items = out
		return
	}
	if int64(len(items)) != count {
		panic(valueError("attempt to assign sequence of size " + itoa(len(items)) + " to extended slice of size " + itoa(int(count))))
	}
	for i := int64(0); i < count; i++ {
		l.items[start+i*step] = items[i]
	}
}

// NextAny advances an iterator for next().
func (it *Iterator[T]) NextAny() (any, bool) {
	if it.Next() {
		return it.Value(), true
	}
	return nil, false
}

// NextAny advances a generator for next(): the value it yields, or false when
// it is finished.
func (g *Gen[Y, S, R]) NextAny() (v any, ok bool) {
	if e := Catch(func() { v, ok = g.NextValue(), true }); e != nil {
		if IsSubclass(e.Type, "StopIteration") {
			return nil, false
		}
		panic(e)
	}
	return v, ok
}

// DynDeep copies a freshly built container so that its items, and the items
// of the containers inside it, are objects: a list of lists of ints becomes a
// list of lists of anything. A container that appears twice in the value is
// copied once.
func DynDeep(v any) any { return deep(v, map[any]any{}) }

func deep(v any, seen map[any]any) any {
	v = Unbox(v)
	if isPointerKey(v) {
		if done, ok := seen[v]; ok && done != nil {
			return done
		}
	}
	switch x := v.(type) {
	case interface{ deepCopy(seen map[any]any) any }:
		return x.deepCopy(seen)
	}
	return v
}

func isPointerKey(v any) bool {
	switch v.(type) {
	case interface{ deepCopy(seen map[any]any) any }:
		return true
	}
	return false
}

func (l *List[T]) deepCopy(seen map[any]any) any {
	out := &List[any]{items: make([]any, len(l.items))}
	seen[l] = out
	for i, x := range l.items {
		out.items[i] = deep(x, seen)
	}
	return out
}

func (d *Dict[K, V]) deepCopy(seen map[any]any) any {
	out := NewDict[any, any]()
	seen[d] = out
	for _, e := range d.entries {
		if !e.dead {
			out.Set(deep(e.key, seen), deep(e.val, seen))
		}
	}
	return out
}

func (s *Set[T]) deepCopy(seen map[any]any) any {
	out := NewSet[any]()
	seen[s] = out
	for c := s.Iter(); c.Next(); {
		out.Add(deep(c.Value(), seen))
	}
	return out
}

// DynDeepAs is DynDeep with the result type spelled out.
func DynDeepAs[T any](v any) T { return conv[T](DynDeep(v)) }

// userMappingKeys is the keys of an instance of a class that has keys() and
// __getitem__, which dict() and dict.update() read as a mapping.
func userMappingKeys(v any) ([]any, bool) {
	n, ok := Unbox(v).(Named)
	if !ok || !DynHasAttr(n, "keys") {
		return nil, false
	}
	var keys []any
	for c := DynIter(DynCallMethod(n, "keys", nil, nil)).Iter(); c.Next(); {
		keys = append(keys, c.Value())
	}
	return keys, true
}
