package runtime

import "sort"

// keyed pairs an item with its sort key, so each key is computed once, as
// CPython does for sorted(xs, key=f).
type keyed[T, K any] struct {
	item T
	key  K
}

func keyItems[T, K any](items []T, key func(T) K) []keyed[T, K] {
	out := make([]keyed[T, K], len(items))
	for i, x := range items {
		out[i] = keyed[T, K]{x, key(x)}
	}
	return out
}

func sortKeyed[T, K any](ks []keyed[T, K], reverse bool) {
	if reverse {
		for i, j := 0, len(ks)-1; i < j; i, j = i+1, j-1 {
			ks[i], ks[j] = ks[j], ks[i]
		}
	}
	sort.SliceStable(ks, func(i, j int) bool { return Lt(ks[i].key, ks[j].key) })
	if reverse {
		for i, j := 0, len(ks)-1; i < j; i, j = i+1, j-1 {
			ks[i], ks[j] = ks[j], ks[i]
		}
	}
}

// SortKey is xs.sort(key=f, reverse=...).
func SortKey[T, K any](l *List[T], key func(T) K, reverse bool) {
	ks := keyItems(l.items, key)
	sortKeyed(ks, reverse)
	for i, k := range ks {
		l.items[i] = k.item
	}
}

// SortedKey is sorted(xs, key=f, reverse=...).
func SortedKey[T, K any](l *List[T], key func(T) K, reverse bool) *List[T] {
	out := l.Copy()
	SortKey(out, key, reverse)
	return out
}

// MinKey is min(xs, key=f): the first item whose key is smallest.
func MinKey[T, K any](l *List[T], key func(T) K) T {
	if len(l.items) == 0 {
		panic(valueError("min() iterable argument is empty"))
	}
	best, bestKey := l.items[0], key(l.items[0])
	for _, x := range l.items[1:] {
		if k := key(x); Lt(k, bestKey) {
			best, bestKey = x, k
		}
	}
	return best
}

// MaxKey is max(xs, key=f): the first item whose key is largest.
func MaxKey[T, K any](l *List[T], key func(T) K) T {
	if len(l.items) == 0 {
		panic(valueError("max() iterable argument is empty"))
	}
	best, bestKey := l.items[0], key(l.items[0])
	for _, x := range l.items[1:] {
		if k := key(x); Gt(k, bestKey) {
			best, bestKey = x, k
		}
	}
	return best
}

// Map is map(f, iterable): a lazy iterator of f applied to each item.
func Map[T, R any](f func(T) R, src Iterable[T]) *Iterator[R] {
	c := src.Iter()
	return NewIterator(func() (R, bool) {
		if c.Next() {
			return f(c.Value()), true
		}
		var zero R
		return zero, false
	})
}

// Map2 is map(f, a, b), which stops at the shorter input.
func Map2[A, B, R any](f func(A, B) R, a Iterable[A], b Iterable[B]) *Iterator[R] {
	ca, cb := a.Iter(), b.Iter()
	return NewIterator(func() (R, bool) {
		if ca.Next() && cb.Next() {
			return f(ca.Value(), cb.Value()), true
		}
		var zero R
		return zero, false
	})
}

// Filter is filter(f, iterable).
func Filter[T any](f func(T) bool, src Iterable[T]) *Iterator[T] {
	c := src.Iter()
	return NewIterator(func() (T, bool) {
		for c.Next() {
			if v := c.Value(); f(v) {
				return v, true
			}
		}
		var zero T
		return zero, false
	})
}
