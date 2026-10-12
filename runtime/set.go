package runtime

import (
	"strings"
)

// Set is Python's set. It is a port of CPython's open-addressing hash table
// (Objects/setobject.c), not a Go map, because the order a set iterates in
// is the order of its table, and programs print sets. With the same hash
// values, the same insertion history and the same probing and resize rules,
// this table iterates in the same order as CPython's.
type Set[T comparable] struct {
	table  []setEntry[T]
	fill   int // active entries plus dummies
	used   int // active entries
	mask   int
	finger int // where pop() resumes
}

type slot uint8

const (
	slotEmpty slot = iota
	slotActive
	slotDummy
)

type setEntry[T comparable] struct {
	key   T
	hash  int64
	state slot
}

const (
	setMinSize    = 8
	linearProbes  = 9
	perturbShift  = 5
	setSmallLimit = 50000
)

func newSetTable[T comparable](size int) []setEntry[T] { return make([]setEntry[T], size) }

// NewSet is {a, b, c}: the items are added one at a time, in order.
func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{table: newSetTable[T](setMinSize), mask: setMinSize - 1}
	for _, x := range items {
		s.Add(x)
	}
	return s
}

// NewSetOfConstants is a set display whose items are all constants, three or
// more of them. CPython's compiler builds a frozenset from the items, rebuilds
// it from that frozenset's iteration order to normalize it, and merges the
// result into an empty set at run time. That lays the table out differently
// from adding the items one at a time.
func NewSetOfConstants[T comparable](items ...T) *Set[T] {
	first := NewSet(items...)
	second := NewSet[T]()
	for c := first.Iter(); c.Next(); {
		second.Add(c.Value())
	}
	s := NewSet[T]()
	s.merge(second)
	return s
}

func (s *Set[T]) Len() int64 { return int64(s.used) }

func (s *Set[T]) pyTruth() bool { return s.used > 0 }

func hashOf[T any](key T) int64 { return Hash(any(key)) }

// lookup returns the index of key's entry, or -1.
func (s *Set[T]) lookup(key T, hash int64) int {
	mask := uint64(s.mask)
	perturb := uint64(hash)
	i := perturb & mask
	for {
		probes := 0
		if i+linearProbes <= mask {
			probes = linearProbes
		}
		for j := 0; j <= probes; j++ {
			e := &s.table[i+uint64(j)]
			if e.state == slotEmpty {
				return -1
			}
			if e.state == slotActive && e.hash == hash && e.key == key {
				return int(i) + j
			}
		}
		perturb >>= perturbShift
		i = (i*5 + 1 + perturb) & mask
	}
}

// insertClean adds a key known to be absent into a table with no dummies.
func insertClean[T comparable](table []setEntry[T], mask int, key T, hash int64) {
	m := uint64(mask)
	perturb := uint64(hash)
	i := perturb & m
	for {
		if table[i].state == slotEmpty {
			table[i] = setEntry[T]{key: key, hash: hash, state: slotActive}
			return
		}
		if i+linearProbes <= m {
			for j := uint64(1); j <= linearProbes; j++ {
				if table[i+j].state == slotEmpty {
					table[i+j] = setEntry[T]{key: key, hash: hash, state: slotActive}
					return
				}
			}
		}
		perturb >>= perturbShift
		i = (i*5 + 1 + perturb) & m
	}
}

// addEntry is set_add_entry.
func (s *Set[T]) addEntry(key T, hash int64) {
	mask := uint64(s.mask)
	perturb := uint64(hash)
	i := perturb & mask
	freeslot := -1
	for {
		probes := 0
		if i+linearProbes <= mask {
			probes = linearProbes
		}
		for j := 0; j <= probes; j++ {
			idx := int(i) + j
			e := &s.table[idx]
			if e.state == slotEmpty {
				if freeslot < 0 {
					// A virgin slot: the key is new.
					s.fill++
					s.used++
					*e = setEntry[T]{key: key, hash: hash, state: slotActive}
					if s.fill*5 >= s.mask*3 {
						if s.used > setSmallLimit {
							s.resize(s.used * 2)
						} else {
							s.resize(s.used * 4)
						}
					}
					return
				}
				// The key is new and reuses a dummy slot seen on the way.
				s.used++
				s.table[freeslot] = setEntry[T]{key: key, hash: hash, state: slotActive}
				return
			}
			if e.state == slotActive {
				if e.hash == hash && e.key == key {
					return
				}
			} else { // dummy
				freeslot = idx
			}
		}
		perturb >>= perturbShift
		i = (i*5 + 1 + perturb) & mask
	}
}

// resize is set_table_resize.
func (s *Set[T]) resize(minused int) {
	newsize := setMinSize
	for newsize <= minused {
		newsize <<= 1
	}
	old := s.table
	if newsize == setMinSize && len(old) == setMinSize && s.fill == s.used {
		return // already the smallest size and no dummies to purge
	}
	newtable := newSetTable[T](newsize)
	s.table, s.mask = newtable, newsize-1
	if s.fill == s.used {
		for _, e := range old {
			if e.state != slotEmpty {
				insertClean(newtable, s.mask, e.key, e.hash)
			}
		}
	} else {
		s.fill = s.used
		for _, e := range old {
			if e.state == slotActive {
				insertClean(newtable, s.mask, e.key, e.hash)
			}
		}
	}
}

// Add is set.add.
func (s *Set[T]) Add(key T) { s.addEntry(key, hashOf(key)) }

// Has is key in set.
func (s *Set[T]) Has(key T) bool { return s.lookup(key, hashOf(key)) >= 0 }

func (s *Set[T]) discard(key T, hash int64) bool {
	i := s.lookup(key, hash)
	if i < 0 {
		return false
	}
	s.table[i] = setEntry[T]{hash: -1, state: slotDummy}
	s.used--
	return true
}

// Discard is set.discard.
func (s *Set[T]) Discard(key T) { s.discard(key, hashOf(key)) }

// Remove is set.remove, which raises KeyError for a missing item.
func (s *Set[T]) Remove(key T) {
	if !s.discard(key, hashOf(key)) {
		panic(NewError("KeyError", key))
	}
}

// Pop is set.pop. It scans from where the last pop stopped.
func (s *Set[T]) Pop() T {
	if s.used == 0 {
		panic(NewError("KeyError", "pop from an empty set"))
	}
	i := s.finger & s.mask
	for s.table[i].state != slotActive {
		i++
		if i > s.mask {
			i = 0
		}
	}
	key := s.table[i].key
	s.table[i] = setEntry[T]{hash: -1, state: slotDummy}
	s.used--
	s.finger = i + 1
	return key
}

// Clear is set.clear: back to the small table.
func (s *Set[T]) Clear() {
	s.table, s.fill, s.used, s.mask = newSetTable[T](setMinSize), 0, 0, setMinSize-1
}

// entries returns the active entries in table order, which is iteration order.
func (s *Set[T]) entries() []setEntry[T] {
	out := make([]setEntry[T], 0, s.used)
	for _, e := range s.table {
		if e.state == slotActive {
			out = append(out, e)
		}
	}
	return out
}

// swapBodies is set_swap_bodies: the tables trade places but each set keeps
// its own finger, as in CPython.
func (s *Set[T]) swapBodies(o *Set[T]) {
	sf, of := s.finger, o.finger
	*s, *o = *o, *s
	s.finger, o.finger = sf, of
}

// merge is set_merge: add every entry of other.
func (s *Set[T]) merge(other *Set[T]) {
	if other == s || other.used == 0 {
		return
	}
	// One resize at the start, rather than growing as entries arrive.
	if (s.fill+other.used)*5 >= s.mask*3 {
		s.resize((s.used + other.used) * 2)
	}
	if s.fill == 0 && s.mask == other.mask && other.fill == other.used {
		copy(s.table, other.table)
		s.fill, s.used = other.fill, other.used
		return
	}
	if s.fill == 0 {
		for _, e := range other.table {
			if e.state == slotActive {
				insertClean(s.table, s.mask, e.key, e.hash)
				s.fill++
				s.used++
			}
		}
		return
	}
	for _, e := range other.table {
		if e.state == slotActive {
			s.addEntry(e.key, e.hash)
		}
	}
}

// updateFrom is set_update_internal for one iterable.
func (s *Set[T]) updateFrom(src Iterable[T]) {
	switch o := src.(type) {
	case *Set[T]:
		s.merge(o)
		return
	case dictKeys[T]:
		// A dict is pre-sized once, then its keys are added in order.
		n := o.keyCount()
		if (s.fill+n)*5 >= s.mask*3 {
			s.resize((s.used + n) * 2)
		}
	}
	for c := src.Iter(); c.Next(); {
		s.Add(c.Value())
	}
}

// dictKeys is a dict used as the source of a set.
type dictKeys[T any] interface {
	Iterable[T]
	keyCount() int
}

// SetFrom is set(iterable).
func SetFrom[T comparable](src Iterable[T]) *Set[T] {
	s := NewSet[T]()
	s.updateFrom(src)
	return s
}

// Copy is set.copy.
func (s *Set[T]) Copy() *Set[T] {
	out := NewSet[T]()
	out.merge(s)
	return out
}

// Update is set.update(*others).
func (s *Set[T]) Update(others ...Iterable[T]) {
	for _, o := range others {
		s.updateFrom(o)
	}
}

// Union is set.union(*others) and a | b.
func (s *Set[T]) Union(others ...Iterable[T]) *Set[T] {
	out := s.Copy()
	for _, o := range others {
		if o == Iterable[T](s) {
			continue // a | a adds nothing, and skips the resize that adding would do
		}
		out.updateFrom(o)
	}
	return out
}

func (s *Set[T]) intersection(other Iterable[T]) *Set[T] {
	if o, ok := other.(*Set[T]); ok {
		if o == s {
			return s.Copy()
		}
		result := NewSet[T]()
		big, small := s, o
		if o.used > s.used {
			big, small = o, s
		}
		for _, e := range small.table {
			if e.state == slotActive && big.lookup(e.key, e.hash) >= 0 {
				result.addEntry(e.key, e.hash)
			}
		}
		return result
	}
	result := NewSet[T]()
	for c := other.Iter(); c.Next(); {
		key := c.Value()
		hash := hashOf(key)
		if s.lookup(key, hash) >= 0 {
			result.addEntry(key, hash)
		}
	}
	return result
}

// Intersection is set.intersection(*others) and a & b.
func (s *Set[T]) Intersection(others ...Iterable[T]) *Set[T] {
	result := s
	if len(others) == 0 {
		return s.Copy()
	}
	for _, o := range others {
		result = result.intersection(o)
	}
	return result
}

// IntersectionUpdate is set.intersection_update and a &= b.
func (s *Set[T]) IntersectionUpdate(others ...Iterable[T]) {
	tmp := s.Intersection(others...)
	s.swapBodies(tmp)
}

func (s *Set[T]) differenceUpdateFrom(other Iterable[T]) {
	if o, ok := other.(*Set[T]); ok {
		if o == s {
			s.Clear()
			return
		}
		for _, e := range o.table {
			if e.state == slotActive {
				s.discard(e.key, e.hash)
			}
		}
	} else {
		for c := other.Iter(); c.Next(); {
			s.Discard(c.Value())
		}
	}
	// If more than a quarter of the table is dummies, rebuild it without them.
	if s.fill-s.used > s.mask/4 {
		if s.used > setSmallLimit {
			s.resize(s.used * 2)
		} else {
			s.resize(s.used * 4)
		}
	}
}

func (s *Set[T]) difference(other Iterable[T]) *Set[T] {
	if o, ok := other.(*Set[T]); ok {
		// With far fewer items in other, copy and remove them one by one.
		if s.used>>2 > o.used {
			out := s.Copy()
			out.differenceUpdateFrom(other)
			return out
		}
		result := NewSet[T]()
		for _, e := range s.table {
			if e.state == slotActive && o.lookup(e.key, e.hash) < 0 {
				result.addEntry(e.key, e.hash)
			}
		}
		return result
	}
	out := s.Copy()
	out.differenceUpdateFrom(other)
	return out
}

// Difference is set.difference(*others) and a - b.
func (s *Set[T]) Difference(others ...Iterable[T]) *Set[T] {
	if len(others) == 0 {
		return s.Copy()
	}
	if len(others) == 1 {
		return s.difference(others[0])
	}
	out := s.Copy()
	for _, o := range others {
		out.differenceUpdateFrom(o)
	}
	return out
}

// DifferenceUpdate is set.difference_update and a -= b.
func (s *Set[T]) DifferenceUpdate(others ...Iterable[T]) {
	for _, o := range others {
		s.differenceUpdateFrom(o)
	}
}

// SymmetricDifferenceUpdate is set.symmetric_difference_update and a ^= b.
func (s *Set[T]) SymmetricDifferenceUpdate(other Iterable[T]) {
	var items *Set[T]
	switch o := other.(type) {
	case *Set[T]:
		if o == s {
			s.Clear()
			return
		}
		items = o
	default:
		items = SetFrom(other)
	}
	for _, e := range items.table {
		if e.state == slotActive {
			if !s.discard(e.key, e.hash) {
				s.addEntry(e.key, e.hash)
			}
		}
	}
}

// SymmetricDifference is set.symmetric_difference and a ^ b.
func (s *Set[T]) SymmetricDifference(other Iterable[T]) *Set[T] {
	result := SetFrom(other)
	result.SymmetricDifferenceUpdate(s)
	return result
}

// IsSubset is set.issubset and a <= b.
func (s *Set[T]) IsSubset(other Iterable[T]) bool {
	o, ok := other.(*Set[T])
	if !ok {
		o = SetFrom(other)
	}
	if s.used > o.used {
		return false
	}
	for _, e := range s.table {
		if e.state == slotActive && o.lookup(e.key, e.hash) < 0 {
			return false
		}
	}
	return true
}

// IsSuperset is set.issuperset and a >= b.
func (s *Set[T]) IsSuperset(other Iterable[T]) bool {
	for c := other.Iter(); c.Next(); {
		if !s.Has(c.Value()) {
			return false
		}
	}
	return true
}

// IsDisjoint is set.isdisjoint.
func (s *Set[T]) IsDisjoint(other Iterable[T]) bool {
	for c := other.Iter(); c.Next(); {
		if s.Has(c.Value()) {
			return false
		}
	}
	return true
}

// SetCursor walks a set in table order and notices a change in size.
type SetCursor[T comparable] struct {
	s    *Set[T]
	pos  int
	used int
	cur  T
}

func (s *Set[T]) Iter() Cursor[T]   { return &SetCursor[T]{s: s, used: s.used} }
func (s *Set[T]) Cursor() Cursor[T] { return s.Iter() }

func (c *SetCursor[T]) Next() bool {
	if c.s == nil {
		return false
	}
	if c.used != c.s.used {
		c.used = -1
		panic(NewError("RuntimeError", "Set changed size during iteration"))
	}
	for c.pos < len(c.s.table) {
		e := &c.s.table[c.pos]
		c.pos++
		if e.state == slotActive {
			c.cur = e.key
			return true
		}
	}
	c.s = nil
	return false
}

func (c *SetCursor[T]) Value() T { return c.cur }

// pyEq is a == b for sets.
func (s *Set[T]) pyEq(other any) bool {
	o, ok := other.(*Set[T])
	if !ok {
		// A keys view equals a set with the same items.
		v, isView := other.(interface{ asSet() *Set[T] })
		if !isView {
			return false
		}
		o = v.asSet()
	}
	return s.used == o.used && s.IsSubset(o)
}

// pySubset implements the ordering operators, which mean "is a subset".
func (s *Set[T]) pySubset(other any, proper bool) bool {
	o := other.(*Set[T])
	if proper && s.used >= o.used {
		return false
	}
	return s.IsSubset(o)
}

// pyRepr is repr(set): {1, 2}, or set() when empty.
func (s *Set[T]) pyRepr() string {
	if s.used == 0 {
		return "set()"
	}
	parts := make([]string, 0, s.used)
	for _, e := range s.table {
		if e.state == slotActive {
			parts = append(parts, Repr(e.key))
		}
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
