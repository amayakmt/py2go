package runtime

import (
	"math"
	"math/big"
	"strings"
)

// BigRange is range(start, stop, step) when an end does not fit a machine
// word. It only exists for objects; a typed range is a Range.
type BigRange struct{ start, stop, step *big.Int }

func newBigRange(start, stop, step *big.Int) *BigRange {
	if step.Sign() == 0 {
		panic(valueError("range() arg 3 must not be zero"))
	}
	return &BigRange{start, stop, step}
}

// bigLen is the number of items.
func (r *BigRange) bigLen() *big.Int {
	d := new(big.Int)
	switch {
	case r.step.Sign() > 0 && r.start.Cmp(r.stop) < 0:
		d.Sub(r.stop, r.start)
		d.Sub(d, bigOne)
		d.Quo(d, r.step)
		return d.Add(d, bigOne)
	case r.step.Sign() < 0 && r.start.Cmp(r.stop) > 0:
		d.Sub(r.start, r.stop)
		d.Sub(d, bigOne)
		d.Quo(d, new(big.Int).Neg(r.step))
		return d.Add(d, bigOne)
	}
	return d
}

func (r *BigRange) PyClassName() string { return "range" }

func (r *BigRange) PyLen() (int64, bool) {
	n := r.bigLen()
	if !n.IsInt64() {
		panic(overflowError("Python int too large to convert to C ssize_t"))
	}
	return n.Int64(), true
}

func (r *BigRange) PyTruth() bool { return r.bigLen().Sign() > 0 }

func (r *BigRange) at(i *big.Int) *big.Int {
	return new(big.Int).Add(r.start, new(big.Int).Mul(i, r.step))
}

type bigRangeCursor struct {
	r   *BigRange
	i   *big.Int
	n   *big.Int
	cur any
}

func (c *bigRangeCursor) Next() bool {
	if c.i.Cmp(c.n) >= 0 {
		return false
	}
	c.cur = NormBig(c.r.at(c.i))
	c.i = new(big.Int).Add(c.i, bigOne)
	return true
}

func (c *bigRangeCursor) Value() any { return c.cur }

func (r *BigRange) IterAny() Cursor[any] {
	return &bigRangeCursor{r: r, i: new(big.Int), n: r.bigLen()}
}

func (r *BigRange) has(x *big.Int) bool {
	if r.step.Sign() > 0 {
		if x.Cmp(r.start) < 0 || x.Cmp(r.stop) >= 0 {
			return false
		}
	} else if x.Cmp(r.start) > 0 || x.Cmp(r.stop) <= 0 {
		return false
	}
	return new(big.Int).Rem(new(big.Int).Sub(x, r.start), r.step).Sign() == 0
}

func (r *BigRange) ContainsAny(item any) bool {
	v := Unbox(item)
	if f, ok := v.(float64); ok {
		if math.IsInf(f, 0) || math.IsNaN(f) || f != math.Trunc(f) {
			return false
		}
		v = bigFromFloat(f)
	}
	if b, ok := bigOf(v); ok {
		return r.has(b)
	}
	return false
}

func (r *BigRange) GetItemAny(key any) any {
	if s, isSlice := key.(Slice); isSlice {
		return r.slice(s)
	}
	b, ok := bigOf(Unbox(key))
	if !ok {
		panic(newError("TypeError", "range indices must be integers or slices, not "+TypeName(key)))
	}
	n := r.bigLen()
	i := new(big.Int).Set(b)
	if i.Sign() < 0 {
		i.Add(i, n)
	}
	if i.Sign() < 0 || i.Cmp(n) >= 0 {
		panic(indexError("range object index out of range"))
	}
	return NormBig(r.at(i))
}

// slice is r[a:b:c], a range again.
func (r *BigRange) slice(s Slice) any {
	n := r.bigLen()
	step := big.NewInt(1)
	if st := Unbox(s.Step); st != nil {
		b, ok := bigOf(st)
		if !ok {
			panic(newError("TypeError", "slice indices must be integers or None or have an __index__ method"))
		}
		step = b
	}
	if step.Sign() == 0 {
		panic(valueError("slice step cannot be zero"))
	}
	lower, upper := big.NewInt(0), new(big.Int).Set(n)
	if step.Sign() < 0 {
		lower, upper = big.NewInt(-1), new(big.Int).Sub(n, bigOne)
	}
	adjust := func(v any, def *big.Int) *big.Int {
		v = Unbox(v)
		if v == nil {
			return def
		}
		b, ok := bigOf(v)
		if !ok {
			panic(newError("TypeError", "slice indices must be integers or None or have an __index__ method"))
		}
		x := new(big.Int).Set(b)
		if x.Sign() < 0 {
			x.Add(x, n)
			if x.Cmp(lower) < 0 {
				return lower
			}
			return x
		}
		if x.Cmp(upper) > 0 {
			return upper
		}
		return x
	}
	startDef, stopDef := lower, upper
	if step.Sign() < 0 {
		startDef, stopDef = upper, lower
	}
	a, z := adjust(s.Lo, startDef), adjust(s.Hi, stopDef)
	newStep := new(big.Int).Mul(r.step, step)
	return rangeOfBig(r.at(a), r.at(z), newStep)
}

// rangeOfBig makes the range object for the given ends, a plain Range when
// they fit.
func rangeOfBig(start, stop, step *big.Int) any {
	if start.IsInt64() && stop.IsInt64() && step.IsInt64() {
		return NewRange(start.Int64(), stop.Int64(), step.Int64())
	}
	return newBigRange(start, stop, step)
}

func (r *BigRange) PyRepr() string {
	s := "range(" + r.start.String() + ", " + r.stop.String()
	if r.step.Cmp(bigOne) != 0 {
		s += ", " + r.step.String()
	}
	return s + ")"
}

func (r *BigRange) PyStr() string { return r.PyRepr() }

func (r *BigRange) PyEq(other any) (bool, bool) {
	o, ok := Unbox(other).(*BigRange)
	if !ok {
		return false, true
	}
	n := r.bigLen()
	if n.Cmp(o.bigLen()) != 0 {
		return false, true
	}
	if n.Sign() == 0 {
		return true, true
	}
	if r.start.Cmp(o.start) != 0 {
		return false, true
	}
	return n.Cmp(bigOne) == 0 || r.step.Cmp(o.step) == 0, true
}

func (r *BigRange) PyDynTable() map[string]bool { return rangeMethodNames }

var rangeMethodNames = names("index count")

func (r *BigRange) PyGetAttr(name string) (any, bool) {
	switch name {
	case "start":
		return NormBig(r.start), true
	case "stop":
		return NormBig(r.stop), true
	case "step":
		return NormBig(r.step), true
	case "index", "count":
		return rangeMethodFn(r, name), true
	}
	return nil, false
}

func (r *BigRange) DynMethod(name string, args []any, kw []KW) (any, bool) {
	return rangeMethod(r, name, args)
}

// rangeMethod is r.index(x) and r.count(x) for either kind of range.
func rangeMethod(r any, name string, args []any) (any, bool) {
	if name != "index" && name != "count" {
		return nil, false
	}
	countArgs(name, args, 1, 1)
	x := Unbox(args[0])
	var contains bool
	var pos *big.Int
	if b, ok := bigOf(x); ok {
		switch rr := r.(type) {
		case *Range:
			br := newBigRange(big.NewInt(rr.start), big.NewInt(rr.stop), big.NewInt(rr.step))
			if contains = br.has(b); contains {
				pos = new(big.Int).Quo(new(big.Int).Sub(b, br.start), br.step)
			}
		case *BigRange:
			if contains = rr.has(b); contains {
				pos = new(big.Int).Quo(new(big.Int).Sub(b, rr.start), rr.step)
			}
		}
	} else if _, isNum := x.(float64); isNum {
		// A float equal to a member counts as a member.
		var items Cursor[any]
		switch rr := r.(type) {
		case *Range:
			items = rr.IterAny()
		case *BigRange:
			items = rr.IterAny()
		}
		i := int64(0)
		for items.Next() {
			if Eq(items.Value(), x) {
				contains, pos = true, big.NewInt(i)
				break
			}
			i++
		}
	}
	if name == "count" {
		if contains {
			return int64(1), true
		}
		return int64(0), true
	}
	if !contains {
		panic(valueError(strings.TrimSpace(Repr(x)) + " is not in range"))
	}
	return NormBig(pos), true
}

func (r *Range) PyDynTable() map[string]bool { return rangeMethodNames }

func (r *Range) DynMethod(name string, args []any, kw []KW) (any, bool) {
	return rangeMethod(r, name, args)
}

func (r *Range) PyGetAttr(name string) (any, bool) {
	switch name {
	case "start":
		return r.start, true
	case "stop":
		return r.stop, true
	case "step":
		return r.step, true
	case "index", "count":
		return rangeMethodFn(r, name), true
	}
	return nil, false
}

func rangeMethodFn(r any, name string) func(args ...any) any {
	return func(args ...any) any {
		v, _ := rangeMethod(r, name, args)
		return v
	}
}
