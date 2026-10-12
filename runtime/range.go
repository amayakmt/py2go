package runtime

import "strconv"

// Range is the object range(start, stop, step).
type Range struct {
	start, stop, step int64
}

// NewRange is range(start, stop, step). A zero step is a ValueError.
func NewRange(start, stop, step int64) *Range {
	if step == 0 {
		panic(valueError("range() arg 3 must not be zero"))
	}
	return &Range{start, stop, step}
}

// Len is len(r), by the formula CPython uses.
func (r *Range) Len() int64 {
	switch {
	case r.step > 0 && r.start < r.stop:
		return (r.stop-r.start-1)/r.step + 1
	case r.step < 0 && r.start > r.stop:
		return (r.start-r.stop-1)/(-r.step) + 1
	}
	return 0
}

func (r *Range) pyTruth() bool { return r.Len() > 0 }

// Get is r[i], with negative indexes counting from the end.
func (r *Range) Get(i int64) int64 {
	n := r.Len()
	if i < 0 {
		i += n
	}
	if i < 0 || i >= n {
		panic(indexError("range object index out of range"))
	}
	return r.start + i*r.step
}

// Has is x in r, computed without walking the range.
func (r *Range) Has(x int64) bool {
	if r.step > 0 {
		if x < r.start || x >= r.stop {
			return false
		}
		return (x-r.start)%r.step == 0
	}
	if x > r.start || x <= r.stop {
		return false
	}
	return (r.start-x)%(-r.step) == 0
}

type rangeCursor struct {
	r   *Range
	i   int64
	n   int64
	cur int64
}

func (c *rangeCursor) Next() bool {
	if c.i >= c.n {
		return false
	}
	c.cur = c.r.start + c.i*c.r.step
	c.i++
	return true
}

func (c *rangeCursor) Value() int64 { return c.cur }

func (r *Range) Iter() Cursor[int64]   { return &rangeCursor{r: r, n: r.Len()} }
func (r *Range) Cursor() Cursor[int64] { return r.Iter() }

// ReversedRange is reversed(r).
func ReversedRange(r *Range) *Range {
	n := r.Len()
	if n == 0 {
		return &Range{0, 0, 1}
	}
	return &Range{r.start + (n-1)*r.step, r.start - r.step, -r.step}
}

// pyEq: ranges are equal when they hold the same sequence.
func (r *Range) pyEq(other any) bool {
	o, ok := other.(*Range)
	if !ok {
		return false
	}
	n := r.Len()
	if n != o.Len() {
		return false
	}
	if n == 0 {
		return true
	}
	if r.start != o.start {
		return false
	}
	return n == 1 || r.step == o.step
}

// pyRepr is range(0, 5), or range(0, 10, 2) when the step is not 1.
func (r *Range) pyRepr() string {
	s := "range(" + strconv.FormatInt(r.start, 10) + ", " + strconv.FormatInt(r.stop, 10)
	if r.step != 1 {
		s += ", " + strconv.FormatInt(r.step, 10)
	}
	return s + ")"
}

// RangeList is list(range(...)), without building the Range object.
func RangeList(start, stop, step int64) *List[int64] {
	return ListFrom[int64](NewRange(start, stop, step))
}
