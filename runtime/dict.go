package runtime

import "strings"

// Dict is Python's dict: it iterates in insertion order. Entries live in a
// slice, so order survives deletes and re-inserts the way CPython's compact
// dict does, and a map finds an entry's position from its key.
type Dict[K comparable, V any] struct {
	entries []dictEntry[K, V]
	index   map[K]int
	live    int // entries not deleted
	epoch   int // changes when the entry slice is compacted
}

type dictEntry[K comparable, V any] struct {
	key  K
	val  V
	dead bool
}

// KV is one key: value pair of a dict display.
type KV[K comparable, V any] struct {
	Key K
	Val V
}

// NewDict is {k1: v1, k2: v2}. A repeated key keeps its first position and
// takes the last value.
func NewDict[K comparable, V any](pairs ...KV[K, V]) *Dict[K, V] {
	d := &Dict[K, V]{index: make(map[K]int, len(pairs))}
	for _, p := range pairs {
		d.Set(p.Key, p.Val)
	}
	return d
}

func (d *Dict[K, V]) Len() int64 { return int64(d.live) }

func (d *Dict[K, V]) keyCount() int { return d.live }

func (d *Dict[K, V]) pyTruth() bool { return d.live > 0 }

func (d *Dict[K, V]) Has(k K) bool {
	_, ok := d.index[normKey(k)]
	return ok
}

// Get is d[k].
func (d *Dict[K, V]) Get(k K) V {
	i, ok := d.index[normKey(k)]
	if !ok {
		panic(NewError("KeyError", k))
	}
	return d.entries[i].val
}

// GetDefault is d.get(k, default).
func (d *Dict[K, V]) GetDefault(k K, def V) V {
	if i, ok := d.index[normKey(k)]; ok {
		return d.entries[i].val
	}
	return def
}

// Set is d[k] = v.
func (d *Dict[K, V]) Set(k K, v V) {
	if i, ok := d.index[normKey(k)]; ok {
		d.entries[i].val = v
		return
	}
	d.compact()
	d.index[normKey(k)] = len(d.entries)
	d.entries = append(d.entries, dictEntry[K, V]{key: k, val: v})
	d.live++
}

// compact squeezes out deleted entries when they outnumber the live ones.
// Cursors notice through the epoch.
func (d *Dict[K, V]) compact() {
	dead := len(d.entries) - d.live
	if dead < 32 || dead <= d.live {
		return
	}
	kept := d.entries[:0]
	for _, e := range d.entries {
		if !e.dead {
			d.index[normKey(e.key)] = len(kept)
			kept = append(kept, e)
		}
	}
	for i := len(kept); i < len(d.entries); i++ {
		d.entries[i] = dictEntry[K, V]{}
	}
	d.entries = kept
	d.epoch++
}

func (d *Dict[K, V]) remove(i int) {
	e := &d.entries[i]
	delete(d.index, e.key)
	*e = dictEntry[K, V]{dead: true}
	d.live--
}

// Delete is del d[k].
func (d *Dict[K, V]) Delete(k K) {
	i, ok := d.index[normKey(k)]
	if !ok {
		panic(NewError("KeyError", k))
	}
	d.remove(i)
}

// Pop is d.pop(k).
func (d *Dict[K, V]) Pop(k K) V {
	i, ok := d.index[normKey(k)]
	if !ok {
		panic(NewError("KeyError", k))
	}
	v := d.entries[i].val
	d.remove(i)
	return v
}

// PopDefault is d.pop(k, default).
func (d *Dict[K, V]) PopDefault(k K, def V) V {
	if !d.Has(k) {
		return def
	}
	return d.Pop(k)
}

// PopItem is d.popitem(): the last item inserted.
func (d *Dict[K, V]) PopItem() Tuple2[K, V] {
	for i := len(d.entries) - 1; i >= 0; i-- {
		if !d.entries[i].dead {
			e := d.entries[i]
			d.remove(i)
			return Tup2(e.key, e.val)
		}
	}
	panic(NewError("KeyError", "popitem(): dictionary is empty"))
}

// SetDefault is d.setdefault(k, default).
func (d *Dict[K, V]) SetDefault(k K, def V) V {
	if i, ok := d.index[normKey(k)]; ok {
		return d.entries[i].val
	}
	d.Set(k, def)
	return def
}

// Update is d.update(other) and d |= other for a dict.
func (d *Dict[K, V]) Update(other *Dict[K, V]) {
	if other == d {
		return
	}
	for _, e := range other.entries {
		if !e.dead {
			d.Set(e.key, e.val)
		}
	}
}

// UpdatePairs is d.update(iterable of pairs).
func (d *Dict[K, V]) UpdatePairs(pairs Iterable[Tuple2[K, V]]) {
	for c := pairs.Iter(); c.Next(); {
		p := c.Value()
		d.Set(p.F0, p.F1)
	}
}

// Union is d | other.
func (d *Dict[K, V]) Union(other *Dict[K, V]) *Dict[K, V] {
	out := d.Copy()
	out.Update(other)
	return out
}

// Copy is d.copy().
func (d *Dict[K, V]) Copy() *Dict[K, V] {
	out := &Dict[K, V]{index: make(map[K]int, d.live)}
	for _, e := range d.entries {
		if !e.dead {
			out.index[normKey(e.key)] = len(out.entries)
			out.entries = append(out.entries, e)
		}
	}
	out.live = len(out.entries)
	return out
}

// Clear is d.clear().
func (d *Dict[K, V]) Clear() {
	d.entries, d.index, d.live = nil, map[K]int{}, 0
	d.epoch++
}

// DictFrom is dict(pairs).
func DictFrom[K comparable, V any](pairs Iterable[Tuple2[K, V]]) *Dict[K, V] {
	d := NewDict[K, V]()
	d.UpdatePairs(pairs)
	return d
}

// FromKeys is dict.fromkeys(keys, value).
func FromKeys[K comparable, V any](keys Iterable[K], v V) *Dict[K, V] {
	d := NewDict[K, V]()
	for c := keys.Iter(); c.Next(); {
		d.Set(c.Value(), v)
	}
	return d
}

// dictCursor walks a dict's entries in insertion order. Like CPython's
// iterator it raises RuntimeError if the dict changes size mid-walk, or if
// it yields more items than it started with.
type dictCursor[K comparable, V any] struct {
	d         *Dict[K, V]
	pos       int
	used      int
	remaining int
	epoch     int
	key       K
	val       V
}

func (d *Dict[K, V]) newCursor() *dictCursor[K, V] {
	return &dictCursor[K, V]{d: d, used: d.live, remaining: d.live, epoch: d.epoch}
}

func (c *dictCursor[K, V]) advance() bool {
	if c.d == nil {
		return false
	}
	if c.used != c.d.live {
		c.used = -1
		panic(NewError("RuntimeError", "dictionary changed size during iteration"))
	}
	if c.epoch != c.d.epoch {
		panic(NewError("RuntimeError", "dictionary keys changed during iteration"))
	}
	for c.pos < len(c.d.entries) {
		e := &c.d.entries[c.pos]
		c.pos++
		if e.dead {
			continue
		}
		if c.remaining == 0 {
			panic(NewError("RuntimeError", "dictionary keys changed during iteration"))
		}
		c.remaining--
		c.key, c.val = e.key, e.val
		return true
	}
	c.d = nil
	return false
}

type keysCursor[K comparable, V any] struct{ *dictCursor[K, V] }
type valuesCursor[K comparable, V any] struct{ *dictCursor[K, V] }
type itemsCursor[K comparable, V any] struct{ *dictCursor[K, V] }

func (c keysCursor[K, V]) Next() bool           { return c.advance() }
func (c keysCursor[K, V]) Value() K             { return c.key }
func (c valuesCursor[K, V]) Next() bool         { return c.advance() }
func (c valuesCursor[K, V]) Value() V           { return c.val }
func (c itemsCursor[K, V]) Next() bool          { return c.advance() }
func (c itemsCursor[K, V]) Value() Tuple2[K, V] { return Tup2(c.key, c.val) }

// Iter walks the keys, which is what "for k in d" does.
func (d *Dict[K, V]) Iter() Cursor[K]   { return keysCursor[K, V]{d.newCursor()} }
func (d *Dict[K, V]) Cursor() Cursor[K] { return d.Iter() }

// KeysView, ValuesView and ItemsView are the live views d.keys(),
// d.values() and d.items().
type KeysView[K comparable, V any] struct{ d *Dict[K, V] }
type ValuesView[K comparable, V any] struct{ d *Dict[K, V] }
type ItemsView[K comparable, V any] struct{ d *Dict[K, V] }

func (d *Dict[K, V]) Keys() *KeysView[K, V]     { return &KeysView[K, V]{d} }
func (d *Dict[K, V]) Values() *ValuesView[K, V] { return &ValuesView[K, V]{d} }
func (d *Dict[K, V]) Items() *ItemsView[K, V]   { return &ItemsView[K, V]{d} }

func (v *KeysView[K, V]) Len() int64   { return v.d.Len() }
func (v *ValuesView[K, V]) Len() int64 { return v.d.Len() }
func (v *ItemsView[K, V]) Len() int64  { return v.d.Len() }

func (v *KeysView[K, V]) pyTruth() bool   { return v.d.live > 0 }
func (v *ValuesView[K, V]) pyTruth() bool { return v.d.live > 0 }
func (v *ItemsView[K, V]) pyTruth() bool  { return v.d.live > 0 }

func (v *KeysView[K, V]) Iter() Cursor[K]     { return v.d.Iter() }
func (v *KeysView[K, V]) Cursor() Cursor[K]   { return v.d.Iter() }
func (v *KeysView[K, V]) keyCount() int       { return v.d.live }
func (v *ValuesView[K, V]) Iter() Cursor[V]   { return valuesCursor[K, V]{v.d.newCursor()} }
func (v *ValuesView[K, V]) Cursor() Cursor[V] { return v.Iter() }
func (v *ItemsView[K, V]) Iter() Cursor[Tuple2[K, V]] {
	return itemsCursor[K, V]{v.d.newCursor()}
}
func (v *ItemsView[K, V]) Cursor() Cursor[Tuple2[K, V]] { return v.Iter() }

// Has is key in d.keys().
func (v *KeysView[K, V]) Has(k K) bool { return v.d.Has(k) }

// Has is value in d.values(), a linear search.
func (v *ValuesView[K, V]) Has(x V) bool {
	for _, e := range v.d.entries {
		if !e.dead && Eq(e.val, x) {
			return true
		}
	}
	return false
}

// Has is (key, value) in d.items().
func (v *ItemsView[K, V]) Has(item Tuple2[K, V]) bool {
	i, ok := v.d.index[normKey(item.F0)]
	return ok && Eq(v.d.entries[i].val, item.F1)
}

// Set operations on keys() work on a set built from the view, as CPython's
// do, so the result iterates in the same order.
func (v *KeysView[K, V]) set() *Set[K] { return SetFrom[K](v) }

func (v *KeysView[K, V]) asSet() *Set[K] { return v.set() }

func (v *KeysView[K, V]) Or(other Iterable[K]) *Set[K]  { return v.set().Union(other) }
func (v *KeysView[K, V]) And(other Iterable[K]) *Set[K] { return v.set().Intersection(other) }
func (v *KeysView[K, V]) Sub(other Iterable[K]) *Set[K] {
	s := v.set()
	s.DifferenceUpdate(other)
	return s
}
func (v *KeysView[K, V]) Xor(other Iterable[K]) *Set[K] { return v.set().SymmetricDifference(other) }

// pyRepr and pyEq make dicts and views print and compare like Python's.
func (d *Dict[K, V]) pyRepr() string {
	if d.live == 0 {
		return "{}"
	}
	parts := make([]string, 0, d.live)
	for _, e := range d.entries {
		if !e.dead {
			parts = append(parts, Repr(e.key)+": "+Repr(e.val))
		}
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (d *Dict[K, V]) pyEq(other any) bool {
	o, ok := other.(*Dict[K, V])
	if !ok {
		// A dict of other key or value types: compare through the objects.
		od, isDict := other.(interface {
			Len() int64
			ContainsAny(any) bool
			GetItemAny(any) any
			IterAny() Cursor[any]
		})
		if _, isSeq := other.(Seq); !isDict || isSeq || od.Len() != int64(d.live) {
			return false
		}
		if _, isSet := other.(interface{ Add(any) }); isSet {
			return false
		}
		for _, e := range d.entries {
			if e.dead {
				continue
			}
			if !od.ContainsAny(e.key) || !Eq(e.val, od.GetItemAny(e.key)) {
				return false
			}
		}
		return true
	}
	if d.live != o.live {
		return false
	}
	for _, e := range d.entries {
		if e.dead {
			continue
		}
		i, ok := o.index[normKey(e.key)]
		if !ok || !Eq(e.val, o.entries[i].val) {
			return false
		}
	}
	return true
}

func viewRepr(name string, items []string) string {
	return name + "([" + strings.Join(items, ", ") + "])"
}

func (v *KeysView[K, V]) pyRepr() string {
	var items []string
	for c := v.Iter(); c.Next(); {
		items = append(items, Repr(c.Value()))
	}
	return viewRepr("dict_keys", items)
}

func (v *ValuesView[K, V]) pyRepr() string {
	var items []string
	for c := v.Iter(); c.Next(); {
		items = append(items, Repr(c.Value()))
	}
	return viewRepr("dict_values", items)
}

func (v *ItemsView[K, V]) pyRepr() string {
	var items []string
	for c := v.Iter(); c.Next(); {
		items = append(items, Repr(c.Value()))
	}
	return viewRepr("dict_items", items)
}

// A keys view compares equal to another keys view or a set with the same items.
func (v *KeysView[K, V]) pyEq(other any) bool {
	switch o := other.(type) {
	case *KeysView[K, V]:
		return v.Len() == o.Len() && v.set().IsSubset(o)
	case *Set[K]:
		return int64(o.used) == v.Len() && v.set().IsSubset(o)
	}
	return false
}

// GetOpt is d.get(k): the value, or None when the key is missing.
func (d *Dict[K, V]) GetOpt(k K) Opt[V] {
	if i, ok := d.index[normKey(k)]; ok {
		v := d.entries[i].val
		if IsNone(any(v)) {
			// A dict of objects holds None as nil.
			return NoneOf[V]()
		}
		return Some(v)
	}
	return NoneOf[V]()
}

// DictWithout is a copy of d without the given keys, which is what the
// "**rest" of a mapping pattern captures.
func DictWithout[K comparable, V any](d *Dict[K, V], keys ...K) *Dict[K, V] {
	out := NewDict[K, V]()
	skip := make(map[K]bool, len(keys))
	for _, k := range keys {
		skip[k] = true
	}
	for _, e := range d.entries {
		if !e.dead && !skip[e.key] {
			out.Set(e.key, e.val)
		}
	}
	return out
}
