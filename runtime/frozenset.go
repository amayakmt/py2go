package runtime

import "strings"

// FrozenSet is an immutable set. It holds a Set, so it iterates in the same
// order a set with the same history would.
type FrozenSet struct{ s *Set[any] }

// NewFrozenSet is frozenset(iterable).
func NewFrozenSet(args []any) *FrozenSet {
	s := NewSet[any]()
	if len(args) > 0 {
		for c := DynIter(args[0]).Iter(); c.Next(); {
			s.Add(normSetItem(c.Value()))
		}
	}
	return &FrozenSet{s}
}

func normSetItem(v any) any {
	v = Unbox(v)
	Hash(v) // unhashable items raise TypeError
	return v
}

func (f *FrozenSet) PyClassName() string { return "frozenset" }
func (f *FrozenSet) Len() int64          { return f.s.Len() }
func (f *FrozenSet) PyLen() (int64, bool) {
	return f.s.Len(), true
}
func (f *FrozenSet) PyTruth() bool               { return f.s.used > 0 }
func (f *FrozenSet) IterAny() Cursor[any]        { return f.s.IterAny() }
func (f *FrozenSet) ContainsAny(x any) bool      { return f.s.ContainsAny(x) }
func (f *FrozenSet) PyDynTable() map[string]bool { return frozenMethodNames }

func (f *FrozenSet) PyRepr() string {
	if f.s.used == 0 {
		return "frozenset()"
	}
	return "frozenset(" + f.s.pyRepr() + ")"
}
func (f *FrozenSet) PyStr() string { return f.PyRepr() }

// PyHash is CPython's frozenset hash, which depends on the layout of the table.
func (f *FrozenSet) PyHash() int64 {
	shuffle := func(h uint64) uint64 { return ((h ^ 89869747) ^ (h << 16)) * 3644798167 }
	var hash uint64
	for _, e := range f.s.table {
		if e.state == slotActive {
			hash ^= shuffle(uint64(e.hash))
		}
	}
	if (len(f.s.table)-f.s.fill)&1 == 1 {
		hash ^= shuffle(0)
	}
	if (f.s.fill-f.s.used)&1 == 1 {
		hash ^= shuffle(^uint64(0))
	}
	hash ^= (uint64(f.s.used) + 1) * 1927868237
	hash ^= (hash >> 11) ^ (hash >> 25)
	hash = hash*69069 + 907133923
	if hash == ^uint64(0) {
		hash = 590923713
	}
	return int64(hash)
}

func otherSet(v any) (*Set[any], bool) {
	switch x := Unbox(v).(type) {
	case *FrozenSet:
		return x.s, true
	case *Set[any]:
		return x, true
	case interface{ anySet() *Set[any] }:
		return x.anySet(), true
	}
	return nil, false
}

func (f *FrozenSet) PyEq(other any) (bool, bool) {
	o, ok := otherSet(other)
	if !ok {
		return false, true
	}
	return f.s.used == o.used && f.s.IsSubset(o), true
}

func (f *FrozenSet) PyOrder(op string, other any) (bool, bool) {
	o, ok := otherSet(other)
	if !ok {
		return false, false
	}
	switch op {
	case "lt":
		return f.s.used < o.used && f.s.IsSubset(o), true
	case "le":
		return f.s.IsSubset(o), true
	case "gt":
		return o.used < f.s.used && o.IsSubset(f.s), true
	}
	return o.IsSubset(f.s), true
}

func (f *FrozenSet) PyBinop(op string, other any, reflected bool) (any, bool) {
	if _, frozen := Unbox(other).(*FrozenSet); reflected && !frozen {
		return nil, false // set op frozenset is the set's to do
	}
	o, ok := otherSet(other)
	if !ok {
		return nil, false
	}
	a, b := f.s, o
	if reflected {
		a, b = o, f.s
	}
	switch op {
	case "BitOr":
		return &FrozenSet{a.Union(b)}, true
	case "BitAnd":
		return &FrozenSet{a.Intersection(b)}, true
	case "Sub":
		return &FrozenSet{a.Difference(b)}, true
	case "BitXor":
		return &FrozenSet{a.SymmetricDifference(b)}, true
	}
	return nil, false
}

var frozenMethodNames = names("union intersection difference symmetric_difference issubset issuperset isdisjoint copy")

func (f *FrozenSet) DynMethod(name string, args []any, kw []KW) (any, bool) {
	others := func() []Iterable[any] {
		var out []Iterable[any]
		for _, a := range args {
			l := NewList[any]()
			for c := DynIter(a).Iter(); c.Next(); {
				l.Append(c.Value())
			}
			out = append(out, l)
		}
		return out
	}
	one := func() Iterable[any] {
		countArgs(name, args, 1, 1)
		return others()[0]
	}
	switch name {
	case "union":
		return &FrozenSet{f.s.Union(others()...)}, true
	case "intersection":
		return &FrozenSet{f.s.Intersection(others()...)}, true
	case "difference":
		return &FrozenSet{f.s.Difference(others()...)}, true
	case "symmetric_difference":
		return &FrozenSet{f.s.SymmetricDifference(one())}, true
	case "issubset":
		return f.s.IsSubset(one()), true
	case "issuperset":
		return f.s.IsSuperset(one()), true
	case "isdisjoint":
		return f.s.IsDisjoint(one()), true
	case "copy":
		return f, true
	}
	return nil, false
}

var _ = strings.Join

// ---- slice, Ellipsis and NotImplemented ----

func (s Slice) PyClassName() string { return "slice" }
func (s Slice) PyGetAttr(name string) (any, bool) {
	switch name {
	case "start":
		return nilIfNone(s.Lo), true
	case "stop":
		return nilIfNone(s.Hi), true
	case "step":
		return nilIfNone(s.Step), true
	case "indices":
		return func(args ...any) any {
			nv := Unbox(argAt(args, 0))
			n, ok := intOf(nv)
			if !ok || isBig(nv) {
				if isBig(nv) {
					panic(overflowError("cannot fit 'int' into an index-sized integer"))
				}
				panic(newError("TypeError", "'"+TypeName(nv)+"' object cannot be interpreted as an integer"))
			}
			if n < 0 {
				panic(valueError("length should not be negative"))
			}
			return s.indicesFor(n)
		}, true
	}
	return nil, false
}

// indicesFor is slice.indices(n): the start, stop and step that a slice of a
// sequence of n items stands for.
func (s Slice) indicesFor(n int64) any {
	b := s.Bounds()
	step := int64(1)
	if b.HasStep {
		step = b.Step
	}
	if step == 0 {
		panic(valueError("slice step cannot be zero"))
	}
	lower, upper := int64(0), n
	if step < 0 {
		lower, upper = -1, n-1
	}
	adjust := func(v int64, has bool, def int64) int64 {
		if !has {
			return def
		}
		if v < 0 {
			v += n
			if v < lower {
				return lower
			}
			return v
		}
		if v > upper {
			return upper
		}
		return v
	}
	var startDef, stopDef int64
	if step < 0 {
		startDef, stopDef = upper, lower
	} else {
		startDef, stopDef = lower, upper
	}
	start := adjust(b.Lo, b.HasLo, startDef)
	stop := adjust(b.Hi, b.HasHi, stopDef)
	var stepOut any = step
	if st := Unbox(s.Step); isBig(st) {
		stepOut = st
	}
	return NewVarTuple[any](start, stop, stepOut)
}

// NewSlice is slice(stop), slice(start, stop) and slice(start, stop, step).
func NewSlice(args []any) Slice {
	switch len(args) {
	case 1:
		return Slice{Hi: nilIfNone(args[0])}
	case 2:
		return Slice{Lo: nilIfNone(args[0]), Hi: nilIfNone(args[1])}
	case 3:
		return Slice{Lo: nilIfNone(args[0]), Hi: nilIfNone(args[1]), Step: nilIfNone(args[2])}
	}
	if len(args) == 0 {
		panic(newError("TypeError", "slice expected at least 1 argument, got 0"))
	}
	panic(newError("TypeError", "slice expected at most 3 arguments, got "+itoa(len(args))))
}

func (s Slice) PyEq(other any) (bool, bool) {
	o, ok := Unbox(other).(Slice)
	if !ok {
		return false, true
	}
	return Eq(s.Lo, o.Lo) && Eq(s.Hi, o.Hi) && Eq(s.Step, o.Step), true
}

// EllipsisT is the type of Ellipsis, NotImplementedT of NotImplemented.
type EllipsisT struct{}
type NotImplementedT struct{}

var (
	Ellipsis        = EllipsisT{}
	NotImplementedV = NotImplementedT{}
)

func (EllipsisT) PyClassName() string       { return "ellipsis" }
func (EllipsisT) PyRepr() string            { return "Ellipsis" }
func (EllipsisT) PyStr() string             { return "Ellipsis" }
func (NotImplementedT) PyClassName() string { return "NotImplementedType" }
func (NotImplementedT) PyRepr() string      { return "NotImplemented" }
func (NotImplementedT) PyStr() string       { return "NotImplemented" }
