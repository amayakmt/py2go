package runtime

import (
	"reflect"
	"testing"
)

func TestDictOrderAndBasics(t *testing.T) {
	// Expected values come from CPython 3.13.
	d := NewDict(KV[string, int64]{"b", 2}, KV[string, int64]{"a", 1}, KV[string, int64]{"b", 3})
	if got := Repr(d); got != "{'b': 3, 'a': 1}" {
		t.Errorf("a repeated key keeps its place and takes the last value: %s", got)
	}
	d.Delete("b")
	d.Set("b", 9)
	if got := Repr(d); got != "{'a': 1, 'b': 9}" {
		t.Errorf("re-inserting after a delete goes last: %s", got)
	}
	if d.Len() != 2 || !d.Has("a") || d.Has("z") || d.Get("a") != 1 || d.GetDefault("z", 5) != 5 {
		t.Error("Len/Has/Get/GetDefault")
	}
	if got := Repr(NewDict[int64, int64]()); got != "{}" {
		t.Errorf("empty dict repr = %s", got)
	}
	if got := d.SetDefault("c", 4); got != 4 || d.SetDefault("a", 100) != 1 {
		t.Error("SetDefault")
	}
	if got := d.PopItem(); got.F0 != "c" || got.F1 != 4 {
		t.Errorf("popitem takes the last item, got %v", got)
	}
	if d.Pop("a") != 1 || d.PopDefault("a", -1) != -1 || d.Len() != 1 {
		t.Error("Pop/PopDefault")
	}
	c := d.Copy()
	c.Set("x", 0)
	if d.Has("x") {
		t.Error("a copy must not share storage")
	}
	d.Clear()
	if d.Len() != 0 || Repr(d) != "{}" {
		t.Error("Clear")
	}
}

func TestDictUpdateUnionEq(t *testing.T) {
	a := NewDict(KV[int64, string]{1, "x"}, KV[int64, string]{2, "y"})
	b := NewDict(KV[int64, string]{2, "Y"}, KV[int64, string]{3, "z"})
	if got := Repr(a.Union(b)); got != "{1: 'x', 2: 'Y', 3: 'z'}" {
		t.Errorf("a | b = %s", got)
	}
	if Repr(a) != "{1: 'x', 2: 'y'}" {
		t.Error("| must not change its left operand")
	}
	a.Update(b)
	if got := Repr(a); got != "{1: 'x', 2: 'Y', 3: 'z'}" {
		t.Errorf("update = %s", got)
	}
	a.UpdatePairs(NewList(Tup2[int64, string](9, "n"), Tup2[int64, string](1, "X")))
	if got := Repr(a); got != "{1: 'X', 2: 'Y', 3: 'z', 9: 'n'}" {
		t.Errorf("update with pairs = %s", got)
	}
	if !Eq(NewDict(KV[int64, int64]{1, 1}, KV[int64, int64]{2, 2}), NewDict(KV[int64, int64]{2, 2}, KV[int64, int64]{1, 1})) {
		t.Error("dict equality ignores order")
	}
	if Eq(NewDict(KV[int64, int64]{1, 1}), NewDict(KV[int64, int64]{1, 2})) {
		t.Error("dict equality compares values")
	}
	nested := func(v int64) *Dict[string, *List[int64]] {
		return NewDict(KV[string, *List[int64]]{"k", NewList(v)})
	}
	if !Eq(nested(1), nested(1)) || Eq(nested(1), nested(2)) {
		t.Error("nested values compare by content")
	}
	if got := Repr(FromKeys[int64, int64](NewList[int64](1, 2, 1), 0)); got != "{1: 0, 2: 0}" {
		t.Errorf("fromkeys = %s", got)
	}
	if got := Repr(DictFrom[string, int64](NewList(Tup2[string, int64]("a", 1)))); got != "{'a': 1}" {
		t.Errorf("dict(pairs) = %s", got)
	}
}

func TestDictViews(t *testing.T) {
	d := NewDict(KV[string, int64]{"a", 1}, KV[string, int64]{"b", 2})
	if got := Repr(d.Keys()); got != "dict_keys(['a', 'b'])" {
		t.Errorf("keys repr = %s", got)
	}
	if got := Repr(d.Values()); got != "dict_values([1, 2])" {
		t.Errorf("values repr = %s", got)
	}
	if got := Repr(d.Items()); got != "dict_items([('a', 1), ('b', 2)])" {
		t.Errorf("items repr = %s", got)
	}
	keys := d.Keys()
	d.Set("c", 3)
	if keys.Len() != 3 || !keys.Has("c") {
		t.Error("a view is live")
	}
	if !d.Values().Has(3) || d.Values().Has(9) || !d.Items().Has(Tup2("a", int64(1))) || d.Items().Has(Tup2("a", int64(2))) {
		t.Error("membership in values and items")
	}
	if got := collect[Tuple2[string, int64]](d.Items()); len(got) != 3 || got[2].F0 != "c" {
		t.Errorf("items iteration: %v", got)
	}
	// Set operations on keys() produce sets that iterate like CPython's.
	e := NewDict(KV[int64, bool]{1, true}, KV[int64, bool]{2, true}, KV[int64, bool]{3, true})
	if got := collect[int64](e.Keys().And(NewSet[int64](3, 2, 9))); !reflect.DeepEqual(got, []int64{2, 3}) {
		t.Errorf("keys & set = %v", got)
	}
	if got := collect[int64](e.Keys().Or(NewList[int64](7, 1))); !reflect.DeepEqual(got, []int64{1, 2, 3, 7}) {
		t.Errorf("keys | list = %v", got)
	}
	if got := collect[int64](e.Keys().Sub(NewSet[int64](2))); !reflect.DeepEqual(got, []int64{1, 3}) {
		t.Errorf("keys - set = %v", got)
	}
	if got := collect[int64](e.Keys().Xor(NewSet[int64](3, 4))); !reflect.DeepEqual(got, []int64{1, 2, 4}) {
		t.Errorf("keys ^ set = %v", got)
	}
}

func TestDictErrors(t *testing.T) {
	d := NewDict(KV[string, int64]{"a", 1})
	tests := []struct {
		name string
		f    func()
		want string
	}{
		{"Get", func() { d.Get("zz") }, "KeyError: 'zz'"},
		{"Get int", func() { NewDict[int64, int64]().Get(5) }, "KeyError: 5"},
		{"Get tuple", func() { NewDict[Tuple2[int64, int64], int64]().Get(Tup2[int64, int64](1, 2)) }, "KeyError: (1, 2)"},
		{"Delete", func() { d.Delete("zz") }, "KeyError: 'zz'"},
		{"Pop", func() { d.Pop("zz") }, "KeyError: 'zz'"},
		{"PopItem", func() { NewDict[int64, int64]().PopItem() }, "KeyError: 'popitem(): dictionary is empty'"},
		{"grow during iteration", func() {
			m := NewDict(KV[int64, int64]{1, 1}, KV[int64, int64]{2, 2})
			for c := m.Iter(); c.Next(); {
				m.Set(c.Value()+10, 0)
			}
		}, "RuntimeError: dictionary changed size during iteration"},
		{"delete during iteration", func() {
			m := NewDict(KV[int64, int64]{1, 1}, KV[int64, int64]{2, 2})
			for c := m.Iter(); c.Next(); {
				m.Delete(c.Value())
			}
		}, "RuntimeError: dictionary changed size during iteration"},
		{"swap during iteration", func() {
			m := NewDict(KV[int64, int64]{1, 1}, KV[int64, int64]{2, 2})
			for c := m.Iter(); c.Next(); {
				m.Delete(c.Value())
				m.Set(c.Value()+10, 0)
			}
		}, "RuntimeError: dictionary keys changed during iteration"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := panicMsg(t, tc.f); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
	// Changing values while iterating is fine.
	m := NewDict(KV[int64, int64]{1, 1}, KV[int64, int64]{2, 2})
	for c := m.Iter(); c.Next(); {
		m.Set(c.Value(), 0)
	}
}

func TestDictCompaction(t *testing.T) {
	d := NewDict[int64, int64]()
	for round := int64(0); round < 50; round++ {
		for i := int64(0); i < 20; i++ {
			d.Set(round*100+i, i)
		}
		for i := int64(0); i < 19; i++ {
			d.Delete(round*100 + i)
		}
	}
	if d.Len() != 50 || len(d.entries) > 200 {
		t.Errorf("deleted entries must be squeezed out: len %d, %d entries", d.Len(), len(d.entries))
	}
	keys := collect[int64](d)
	for i, k := range keys {
		if k != int64(i)*100+19 {
			t.Fatalf("order after compaction: %v", keys)
		}
	}
}
