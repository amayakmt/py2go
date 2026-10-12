package runtime

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Seq is what lists and tuples have in common: the runtime can print,
// compare and order them without knowing their element type.
type Seq interface {
	Len() int64
	At(i int64) any
	Kind() string // "list" or "tuple"
}

func indexError(msg string) *PyError { return &PyError{"IndexError", msg} }
func valueError(msg string) *PyError { return &PyError{"ValueError", msg} }

// Eq is Python's == for two values of the same static type. Lists and
// tuples compare element by element.
func Eq(a, b any) bool {
	if x, ok := a.(Seq); ok {
		y, ok := b.(Seq)
		if !ok || x.Len() != y.Len() {
			return false
		}
		for i := int64(0); i < x.Len(); i++ {
			if !Eq(x.At(i), y.At(i)) {
				return false
			}
		}
		return true
	}
	return a == b
}

// Lt, Le, Gt and Ge are Python's ordering operators. Lists and tuples
// order lexicographically: the first position that differs decides, and
// if there is none the shorter sequence is smaller.
func Lt(a, b any) bool { return order(a, b, "<") }
func Le(a, b any) bool { return order(a, b, "<=") }
func Gt(a, b any) bool { return order(b, a, "<") }
func Ge(a, b any) bool { return order(b, a, "<=") }

func order(a, b any, op string) bool {
	switch x := a.(type) {
	case Seq:
		y := b.(Seq)
		n := min(x.Len(), y.Len())
		for i := int64(0); i < n; i++ {
			if !Eq(x.At(i), y.At(i)) {
				return order(x.At(i), y.At(i), op)
			}
		}
		return compareScalar(x.Len(), y.Len(), op)
	case int64:
		return compareScalar(x, b.(int64), op)
	case float64:
		return compareScalar(x, b.(float64), op)
	case string:
		return compareScalar(x, b.(string), op)
	case bool:
		return compareScalar(b2i(x), b2i(b.(bool)), op)
	}
	panic(fmt.Sprintf("runtime: cannot order %T", a))
}

func compareScalar[T int64 | float64 | string](a, b T, op string) bool {
	if op == "<" {
		return a < b
	}
	return a <= b
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Bounds is the "lo:hi:step" of a slice, with a flag for each part that
// was written. Any part can be left out in Python.
type Bounds struct {
	Lo, Hi, Step          int64
	HasLo, HasHi, HasStep bool
}

// indices resolves the bounds against a length the way CPython's
// PySlice_AdjustIndices does. The slice is start, start+step, ... for count items.
func (b Bounds) indices(n int64) (start, step, count int64) {
	step = 1
	if b.HasStep {
		step = b.Step
	}
	if step == 0 {
		panic(valueError("slice step cannot be zero"))
	}
	// Out-of-range bounds clamp to lowest..highest, which depend on the
	// direction: [-1, n-1] for a negative step, [0, n] for a positive one.
	lowest, highest := int64(0), n
	if step < 0 {
		lowest, highest = -1, n-1
	}
	clamp := func(v int64, has bool, def int64) int64 {
		if !has {
			return def
		}
		if v < 0 {
			v += n
			if v < 0 {
				return lowest
			}
		} else if v >= n {
			return highest
		}
		return v
	}
	startDefault, stopDefault := lowest, highest
	if step < 0 {
		startDefault, stopDefault = highest, lowest
	}
	start = clamp(b.Lo, b.HasLo, startDefault)
	stop := clamp(b.Hi, b.HasHi, stopDefault)
	if step < 0 {
		if stop < start {
			count = (start-stop-1)/(-step) + 1
		}
	} else if start < stop {
		count = (stop-start-1)/step + 1
	}
	return start, step, count
}

// List is Python's list. It is a pointer type in generated code because
// Python lists alias: "b = a" shares one list, and append through either
// name shows through the other.
type List[T any] struct {
	items []T
}

func NewList[T any](items ...T) *List[T] { return &List[T]{items: items} }

func (l *List[T]) Len() int64     { return int64(len(l.items)) }
func (l *List[T]) At(i int64) any { return l.items[i] }
func (l *List[T]) Kind() string   { return "list" }

// norm turns a Python index (negative counts from the end) into a slice
// index, or reports false if it is out of range.
func (l *List[T]) norm(i int64) (int, bool) {
	n := int64(len(l.items))
	if i < 0 {
		i += n
	}
	return int(i), i >= 0 && i < n
}

func (l *List[T]) Get(i int64) T {
	k, ok := l.norm(i)
	if !ok {
		panic(indexError("list index out of range"))
	}
	return l.items[k]
}

func (l *List[T]) Set(i int64, v T) {
	k, ok := l.norm(i)
	if !ok {
		panic(indexError("list assignment index out of range"))
	}
	l.items[k] = v
}

func (l *List[T]) Append(v T) { l.items = append(l.items, v) }

// Pop removes the last item.
func (l *List[T]) Pop() T {
	if len(l.items) == 0 {
		panic(indexError("pop from empty list"))
	}
	return l.PopAt(-1)
}

func (l *List[T]) PopAt(i int64) T {
	if len(l.items) == 0 {
		panic(indexError("pop from empty list"))
	}
	k, ok := l.norm(i)
	if !ok {
		panic(indexError("pop index out of range"))
	}
	v := l.items[k]
	l.items = append(l.items[:k], l.items[k+1:]...)
	return v
}

// Insert clamps the position like Python: a position past either end
// inserts at that end.
func (l *List[T]) Insert(i int64, v T) {
	n := int64(len(l.items))
	if i < 0 {
		i = max(i+n, 0)
	}
	i = min(i, n)
	var zero T
	l.items = append(l.items, zero)
	copy(l.items[i+1:], l.items[i:])
	l.items[i] = v
}

func (l *List[T]) Extend(o *List[T]) {
	l.items = append(l.items, append([]T(nil), o.items...)...) // o may be l
}

func (l *List[T]) Reverse() {
	for i, j := 0, len(l.items)-1; i < j; i, j = i+1, j-1 {
		l.items[i], l.items[j] = l.items[j], l.items[i]
	}
}

// Sort is a stable sort. With reverse set it reverses, sorts, and reverses
// again, as CPython does, so equal items keep their original order.
func (l *List[T]) Sort(reverse bool) { sortItems(l.items, reverse) }

func sortItems[T any](items []T, reverse bool) {
	if reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return Lt(items[i], items[j]) })
	if reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}
}

// Sorted is sorted(list): a new list, the original untouched.
func Sorted[T any](l *List[T], reverse bool) *List[T] {
	out := l.Copy()
	out.Sort(reverse)
	return out
}

func (l *List[T]) find(v T) int {
	for i, x := range l.items {
		if Eq(x, v) {
			return i
		}
	}
	return -1
}

func (l *List[T]) Contains(v T) bool { return l.find(v) >= 0 }

func (l *List[T]) Index(v T) int64 {
	i := l.find(v)
	if i < 0 {
		panic(valueError(Repr(v) + " is not in list"))
	}
	return int64(i)
}

func (l *List[T]) Count(v T) int64 {
	var n int64
	for _, x := range l.items {
		if Eq(x, v) {
			n++
		}
	}
	return n
}

func (l *List[T]) Remove(v T) {
	i := l.find(v)
	if i < 0 {
		panic(valueError("list.remove(x): x not in list"))
	}
	l.items = append(l.items[:i], l.items[i+1:]...)
}

func (l *List[T]) Clear() { l.items = nil }

func (l *List[T]) Copy() *List[T] { return &List[T]{items: append([]T(nil), l.items...)} }

// Concat is list + list. It builds a new list.
func (l *List[T]) Concat(o *List[T]) *List[T] {
	items := make([]T, 0, len(l.items)+len(o.items))
	return &List[T]{items: append(append(items, l.items...), o.items...)}
}

// Repeat is list * n. The new list holds the same items n times, so nested
// lists end up shared, just as in Python.
func (l *List[T]) Repeat(n int64) *List[T] {
	var items []T
	for i := int64(0); i < n; i++ {
		items = append(items, l.items...)
	}
	return &List[T]{items: items}
}

// RepeatInPlace is list *= n: the same list object, now repeated.
func (l *List[T]) RepeatInPlace(n int64) { l.items = l.Repeat(n).items }

func (l *List[T]) Slice(b Bounds) *List[T] {
	start, step, count := b.indices(int64(len(l.items)))
	items := make([]T, 0, count)
	for i := int64(0); i < count; i++ {
		items = append(items, l.items[start+i*step])
	}
	return &List[T]{items: items}
}

// CheckUnpack is the length test of "a, b = some_list".
func CheckUnpack(got, want int64) {
	switch {
	case got > want:
		panic(valueError(fmt.Sprintf("too many values to unpack (expected %d)", want)))
	case got < want:
		panic(valueError(fmt.Sprintf("not enough values to unpack (expected %d, got %d)", want, got)))
	}
}

// Contains is "x in t" for a tuple t.
func SeqContains(s Seq, v any) bool {
	for i := int64(0); i < s.Len(); i++ {
		if Eq(s.At(i), v) {
			return true
		}
	}
	return false
}

// Str is Python's str(). Strings are themselves, everything else is
// its repr.
func Str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return Repr(v)
}

// Repr is Python's repr() for the supported types.
func Repr(v any) string {
	switch v := v.(type) {
	case bool:
		if v {
			return "True"
		}
		return "False"
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case float64:
		return FloatRepr(v)
	case string:
		return quote(v)
	case Seq:
		parts := make([]string, v.Len())
		for i := range parts {
			parts[i] = Repr(v.At(int64(i)))
		}
		if v.Kind() == "tuple" {
			if len(parts) == 1 {
				return "(" + parts[0] + ",)"
			}
			return "(" + strings.Join(parts, ", ") + ")"
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	panic(fmt.Sprintf("runtime: cannot repr %T", v))
}

// quote is repr() of a str: single quotes unless the text holds a single
// quote and no double quote, with Python's escapes.
func quote(s string) string {
	q := byte('\'')
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		q = '"'
	}
	var b strings.Builder
	b.WriteByte(q)
	for _, r := range s {
		switch {
		case r == rune(q) || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r < 0x7f:
			b.WriteRune(r)
		case r == utf8.RuneError || !unicode.IsPrint(r):
			switch {
			case r <= 0xff:
				fmt.Fprintf(&b, `\x%02x`, r)
			case r <= 0xffff:
				fmt.Fprintf(&b, `\u%04x`, r)
			default:
				fmt.Fprintf(&b, `\U%08x`, r)
			}
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte(q)
	return b.String()
}
