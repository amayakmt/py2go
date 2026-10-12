package runtime

import (
	"reflect"
	"testing"
)

func TestMessages(t *testing.T) {
	// Expected texts come from CPython 3.13.
	tests := []struct{ got, want string }{
		{MsgTooManyPositional("f", 2, 1, 1, 0), "f() takes 1 positional argument but 2 were given"},
		{MsgTooManyPositional("f", 3, 1, 2, 0), "f() takes from 1 to 2 positional arguments but 3 were given"},
		{MsgTooManyPositional("f", 1, 0, 0, 0), "f() takes 0 positional arguments but 1 was given"},
		{MsgTooManyPositional("f", 2, 1, 1, 1), "f() takes 1 positional argument but 2 positional arguments (and 1 keyword-only argument) were given"},
		{MsgMissing("f", "positional", "a"), "f() missing 1 required positional argument: 'a'"},
		{MsgMissing("f", "positional", "a", "b"), "f() missing 2 required positional arguments: 'a' and 'b'"},
		{MsgMissing("f", "positional", "a", "b", "c"), "f() missing 3 required positional arguments: 'a', 'b', and 'c'"},
		{MsgMissing("f", "keyword-only", "k"), "f() missing 1 required keyword-only argument: 'k'"},
		{MsgUnexpectedKeyword("f", "c", ""), "f() got an unexpected keyword argument 'c'"},
		{MsgUnexpectedKeyword("f", "nme", "name"), "f() got an unexpected keyword argument 'nme'. Did you mean 'name'?"},
		{MsgMultipleValues("f", "a"), "f() got multiple values for argument 'a'"},
		{MsgPositionalOnlyAsKeyword("f", "a", "b"), "f() got some positional-only arguments passed as keyword arguments: 'a', 'b'"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

func TestSuggest(t *testing.T) {
	tests := []struct {
		name  string
		cands []string
		want  string
	}{
		{"nme", []string{"name", "age"}, "name"},
		{"vlaue", []string{"value", "key"}, "value"},
		{"zzz", []string{"name", "age"}, ""},
		{"x", nil, ""},
	}
	for _, tc := range tests {
		if got := Suggest(tc.name, tc.cands); got != tc.want {
			t.Errorf("Suggest(%q, %v) = %q, want %q", tc.name, tc.cands, got, tc.want)
		}
	}
}

func catch(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg = r.(*PyError).Error()
		}
	}()
	f()
	return ""
}

func TestSplitKeys(t *testing.T) {
	d := NewDict[string, int64]()
	d.Set("a", 1)
	d.Set("z", 2)
	extra := SplitKeys(d, "f", []string{"b"}, nil, true)
	if extra.Len() != 2 {
		t.Errorf("collected %d keys, want 2", extra.Len())
	}
	if got := catch(func() { SplitKeys(d, "f", []string{"b"}, nil, false) }); got != "TypeError: f() got an unexpected keyword argument 'a'" {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { SplitKeys(d, "f", []string{"b"}, []string{"a"}, false) }); got != "TypeError: f() got some positional-only arguments passed as keyword arguments: 'a'" {
		t.Errorf("got %q", got)
	}
}

func TestVarTuple(t *testing.T) {
	v := NewVarTuple[int64](3, 1, 4, 1)
	if v.Len() != 4 || v.Get(-1) != 1 || v.Count(1) != 2 || v.Index(4) != 2 || !v.Contains(3) || v.Contains(9) {
		t.Errorf("basic operations wrong: %v", v)
	}
	if got := Repr(v); got != "(3, 1, 4, 1)" {
		t.Errorf("Repr = %s", got)
	}
	if got := Repr(NewVarTuple[int64](7)); got != "(7,)" {
		t.Errorf("Repr of one item = %s", got)
	}
	if got := Repr(NewVarTuple[int64]()); got != "()" {
		t.Errorf("Repr of empty = %s", got)
	}
	if got := v.Concat(NewVarTuple[int64](5)); got.Len() != 5 {
		t.Errorf("concat length %d", got.Len())
	}
	if got := v.Repeat(2); got.Len() != 8 || v.Repeat(0).Len() != 0 {
		t.Errorf("repeat length %d", got.Len())
	}
	if got := catch(func() { v.Get(4) }); got != "IndexError: tuple index out of range" {
		t.Errorf("got %q", got)
	}
	if !reflect.DeepEqual(v.ToList().items, []int64{3, 1, 4, 1}) {
		t.Errorf("ToList wrong")
	}
	if got := TupleFrom[int64](NewList[int64](1, 2)); got.Len() != 2 {
		t.Errorf("TupleFrom length %d", got.Len())
	}
}

func drain[T any](it *Iterator[T]) []T {
	var out []T
	for it.Next() {
		out = append(out, it.Value())
	}
	return out
}

func TestSortKey(t *testing.T) {
	l := NewList[string]("bb", "a", "ccc", "dd")
	byLen := func(s string) int64 { return int64(len(s)) }
	if got := SortedKey(l, byLen, false).items; !reflect.DeepEqual(got, []string{"a", "bb", "dd", "ccc"}) {
		t.Errorf("stable ascending: %v", got)
	}
	// reverse keeps equal items in their original order, as CPython does.
	if got := SortedKey(l, byLen, true).items; !reflect.DeepEqual(got, []string{"ccc", "bb", "dd", "a"}) {
		t.Errorf("stable descending: %v", got)
	}
	if !reflect.DeepEqual(l.items, []string{"bb", "a", "ccc", "dd"}) {
		t.Errorf("SortedKey changed its input: %v", l.items)
	}
	SortKey(l, byLen, false)
	if !reflect.DeepEqual(l.items, []string{"a", "bb", "dd", "ccc"}) {
		t.Errorf("SortKey: %v", l.items)
	}
	if MinKey(l, byLen) != "a" || MaxKey(NewList("aa", "bb"), byLen) != "aa" {
		t.Errorf("min/max must return the first of equal keys")
	}
}

func TestMapFilter(t *testing.T) {
	src := NewList[int64](1, 2, 3, 4)
	if got := drain(Map(func(x int64) int64 { return x * x }, Iterable[int64](src))); !reflect.DeepEqual(got, []int64{1, 4, 9, 16}) {
		t.Errorf("Map: %v", got)
	}
	if got := drain(Filter(func(x int64) bool { return x%2 == 0 }, Iterable[int64](src))); !reflect.DeepEqual(got, []int64{2, 4}) {
		t.Errorf("Filter: %v", got)
	}
	calls := 0
	it := Map(func(x int64) int64 { calls++; return x }, Iterable[int64](src))
	if calls != 0 {
		t.Errorf("Map ran before it was consumed")
	}
	it.Next()
	if calls != 1 {
		t.Errorf("Map ran %d times for one item", calls)
	}
}
