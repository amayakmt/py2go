package runtime

import (
	"reflect"
	"testing"
)

func TestRange(t *testing.T) {
	// Expected values come from CPython 3.13.
	tests := []struct {
		r    *Range
		repr string
		want []int64
	}{
		{NewRange(0, 5, 1), "range(0, 5)", []int64{0, 1, 2, 3, 4}},
		{NewRange(10, 0, -3), "range(10, 0, -3)", []int64{10, 7, 4, 1}},
		{NewRange(2, 9, 3), "range(2, 9, 3)", []int64{2, 5, 8}},
		{NewRange(5, 5, 1), "range(5, 5)", []int64{}},
		{NewRange(5, 0, 1), "range(5, 0)", []int64{}},
	}
	for _, tc := range tests {
		if got := collect[int64](tc.r); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s iterates %v, want %v", tc.repr, got, tc.want)
		}
		if got := Repr(tc.r); got != tc.repr {
			t.Errorf("repr = %s, want %s", got, tc.repr)
		}
		if tc.r.Len() != int64(len(tc.want)) {
			t.Errorf("%s: Len = %d", tc.repr, tc.r.Len())
		}
		for i, v := range tc.want {
			if tc.r.Get(int64(i)) != v || tc.r.Get(int64(i)-int64(len(tc.want))) != v {
				t.Errorf("%s: Get(%d) is wrong", tc.repr, i)
			}
			if !tc.r.Has(v) {
				t.Errorf("%s: %d should be in the range", tc.repr, v)
			}
		}
	}
	r := NewRange(10, 0, -3)
	if r.Has(9) || r.Has(13) || r.Has(-2) || !r.Has(1) {
		t.Error("Has on a descending range")
	}
	if got := collect[int64](ReversedRange(r)); !reflect.DeepEqual(got, []int64{1, 4, 7, 10}) {
		t.Errorf("reversed = %v", got)
	}
	if !Eq(NewRange(0, 0, 1), NewRange(5, 5, 3)) || !Eq(NewRange(0, 3, 1), NewRange(0, 3, 1)) || Eq(NewRange(0, 3, 1), NewRange(0, 3, 2)) || !Eq(NewRange(0, 1, 1), NewRange(0, 5, 7)) {
		t.Error("range equality compares the sequences")
	}
	if got := panicMsg(t, func() { NewRange(0, 1, 0) }); got != "ValueError: range() arg 3 must not be zero" {
		t.Errorf("zero step: %q", got)
	}
	if got := panicMsg(t, func() { NewRange(0, 3, 1).Get(3) }); got != "IndexError: range object index out of range" {
		t.Errorf("index: %q", got)
	}
	if got := RangeList(1, 4, 1).slice(); !reflect.DeepEqual(got, []int64{1, 2, 3}) {
		t.Errorf("RangeList = %v", got)
	}
}

func TestIterators(t *testing.T) {
	l := NewList[int64](1, 2, 3)
	if got := collect[int64](Reversed(l)); !reflect.DeepEqual(got, []int64{3, 2, 1}) {
		t.Errorf("reversed = %v", got)
	}
	// A reversed list iterator reads the list live and stops if it shrinks.
	shrinking := NewList[int64](1, 2, 3, 4)
	it := Reversed(shrinking)
	it.Next()
	shrinking.Clear()
	if it.Next() {
		t.Error("reversed iterator should stop when the list is cleared")
	}
	if got := collect[string](ReversedStr("héy")); !reflect.DeepEqual(got, []string{"y", "é", "h"}) {
		t.Errorf("reversed str = %v", got)
	}
	enum := collect[Tuple2[int64, string]](Enumerate[string](NewList("a", "b"), 5))
	if len(enum) != 2 || enum[0] != Tup2(int64(5), "a") || enum[1] != Tup2(int64(6), "b") {
		t.Errorf("enumerate = %v", enum)
	}
	zip := collect[Tuple2[int64, string]](Zip2[int64, string](NewList[int64](1, 2, 3), NewList("x", "y")))
	if len(zip) != 2 || zip[1] != Tup2(int64(2), "y") {
		t.Errorf("zip stops at the shortest: %v", zip)
	}
	// zip pulls from its inputs in order, so the first input loses one item.
	first, second := NewList[int64](1, 2, 3), NewList[int64](9)
	zipped := Zip2[int64, int64](first, second)
	zipped.Next()
	zipped.Next()
	if got := collect[int64](ListFrom[int64](first)); len(got) != 3 {
		t.Errorf("zip must not change its inputs: %v", got)
	}

	// An Iterator is one-shot.
	n := 0
	once := NewIterator(func() (int64, bool) {
		n++
		return int64(n), n <= 2
	})
	if got := collect[int64](once); !reflect.DeepEqual(got, []int64{1, 2}) {
		t.Errorf("iterator = %v", got)
	}
	if got := collect[int64](once); len(got) != 0 {
		t.Errorf("a finished iterator stays finished: %v", got)
	}
	if n != 3 {
		t.Errorf("a finished iterator must not call its function again, called %d times", n)
	}
}

func TestAnyAllTruth(t *testing.T) {
	pulled := 0
	gen := NewIterator(func() (bool, bool) { pulled++; return pulled == 2, pulled < 10 })
	if !Any[bool](gen) || pulled != 2 {
		t.Errorf("any stops at the first true item, pulled %d", pulled)
	}
	if !All[int64](NewList[int64](1, 2)) || All[int64](NewList[int64](1, 0)) || Any[string](NewList("", "")) || !All[bool](NewList[bool]()) || Any[bool](NewList[bool]()) {
		t.Error("any/all")
	}
	for _, v := range []any{int64(0), 0.0, "", NewList[int64](), NewSet[int64](), NewDict[int64, int64](), NewRange(0, 0, 1), false} {
		if Truth(v) {
			t.Errorf("%s should be falsy", Repr(v))
		}
	}
	for _, v := range []any{int64(2), 0.5, "a", NewList[int64](0), NewSet[int64](0), NewDict(KV[int64, int64]{0, 0}), NewRange(0, 1, 1), Tup1[int64](0)} {
		if !Truth(v) {
			t.Errorf("%s should be truthy", Repr(v))
		}
	}
}

func TestSetFromDict(t *testing.T) {
	// Expected order from CPython 3.13: set({8: 1, 16: 1, ...}).
	d := NewDict[int64, bool]()
	for _, k := range []int64{8, 16, 0, 24, 40, 32, 4, 12} {
		d.Set(k, true)
	}
	if got := collect[int64](SetFrom[int64](d)); !reflect.DeepEqual(got, []int64{0, 32, 4, 8, 40, 12, 16, 24}) {
		t.Errorf("set(d) = %v", got)
	}
}
