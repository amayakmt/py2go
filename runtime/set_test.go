package runtime

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func intsOf(v []any) []int64 {
	out := []int64{}
	for _, x := range v {
		out = append(out, int64(x.(float64)))
	}
	return out
}

func collect[T any](src Iterable[T]) []T {
	out := []T{}
	for c := src.Iter(); c.Next(); {
		out = append(out, c.Value())
	}
	return out
}

// TestSetOrderMatchesCPython replays random sequences of set operations and
// compares the iteration order of every set with CPython 3.13's. The order is
// the layout of the hash table, so it only matches if the probing, resizing,
// dummy handling and operator algorithms all follow setobject.c.
// testdata/set_orders.json was recorded from CPython 3.13.
func TestSetOrderMatchesCPython(t *testing.T) {
	data, err := os.ReadFile("testdata/set_orders.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Ops   [][]any          `json:"ops"`
		Final map[string][]any `json:"final"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 100 {
		t.Fatalf("only %d recorded cases", len(cases))
	}
	for ci, c := range cases {
		sets := map[string]*Set[int64]{"a": NewSet[int64](), "b": NewSet[int64](), "c": NewSet[int64]()}
		for _, op := range c.Ops {
			name, n := op[0].(string), op[1].(string)
			s := sets[n]
			list := func() *List[int64] { return NewList(intsOf(op[2].([]any))...) }
			item := func() int64 { return int64(op[2].(float64)) }
			other := func() *Set[int64] { return sets[op[2].(string)] }
			switch name {
			case "add":
				s.Add(item())
			case "discard":
				s.Discard(item())
			case "remove":
				s.Remove(item())
			case "pop":
				s.Pop()
			case "clear":
				s.Clear()
			case "update":
				s.Update(list())
			case "union":
				sets[n] = s.Union(other())
			case "inter":
				sets[n] = s.Intersection(other())
			case "diff":
				sets[n] = s.Difference(other())
			case "sym":
				sets[n] = s.SymmetricDifference(other())
			case "copy":
				sets[n] = other().Copy()
			case "ior":
				s.Update(other())
			case "iand":
				s.IntersectionUpdate(other())
			case "isub":
				s.DifferenceUpdate(other())
			case "ixor":
				s.SymmetricDifferenceUpdate(other())
			case "fromlist":
				sets[n] = SetFrom[int64](list())
			case "inter_update":
				s.IntersectionUpdate(list())
			case "diff_update":
				s.DifferenceUpdate(list())
			case "sym_update":
				s.SymmetricDifferenceUpdate(list())
			default:
				t.Fatalf("unknown operation %q", name)
			}
		}
		for k, want := range c.Final {
			got := collect[int64](sets[k])
			if w := intsOf(want); !reflect.DeepEqual(got, w) {
				t.Fatalf("case %d, set %s: got %v, want %v\nops: %v", ci, k, got, w, c.Ops)
			}
		}
	}
}

func TestSetBasics(t *testing.T) {
	// Expected values come from CPython 3.13.
	s := NewSet[int64](3, 1, 2, 1)
	if s.Len() != 3 || !s.Has(2) || s.Has(9) {
		t.Fatalf("basic membership: %s", Repr(s))
	}
	if got := Repr(s); got != "{1, 2, 3}" {
		t.Errorf("repr = %s", got)
	}
	if got := Repr(NewSet[int64]()); got != "set()" {
		t.Errorf("empty repr = %s", got)
	}
	if got := Repr(NewSet(Tup2[int64, string](1, "a"))); got != "{(1, 'a')}" {
		t.Errorf("tuple set repr = %s", got)
	}
	big := NewSet[int64](5, 9, 1)
	if got := collect[int64](big.Union(NewSet[int64](7, 8))); !reflect.DeepEqual(got, []int64{1, 5, 7, 8, 9}) {
		t.Errorf("union order = %v", got)
	}
	a, b := NewSet[int64](1, 2, 3), NewSet[int64](2, 3, 4)
	if got := collect[int64](a.Intersection(b)); !reflect.DeepEqual(got, []int64{2, 3}) {
		t.Errorf("intersection = %v", got)
	}
	if got := collect[int64](a.Difference(b)); !reflect.DeepEqual(got, []int64{1}) {
		t.Errorf("difference = %v", got)
	}
	if got := collect[int64](a.SymmetricDifference(b)); !reflect.DeepEqual(got, []int64{1, 4}) {
		t.Errorf("symmetric difference = %v", got)
	}
	if !NewSet[int64](1, 2).IsSubset(a) || a.IsSubset(NewSet[int64](1, 2)) || !a.IsSuperset(NewList[int64](1, 3)) {
		t.Error("subset tests")
	}
	if !a.IsDisjoint(NewList[int64](7, 8)) || a.IsDisjoint(NewList[int64](3)) {
		t.Error("isdisjoint")
	}
	if !Eq(NewSet[int64](1, 2), NewSet[int64](2, 1)) || Eq(NewSet[int64](1), NewSet[int64](1, 2)) {
		t.Error("set equality ignores order")
	}
	if !Le(NewSet[int64](1), NewSet[int64](1, 2)) || !Lt(NewSet[int64](1), NewSet[int64](1, 2)) || Lt(NewSet[int64](1), NewSet[int64](1)) || !Ge(a, NewSet[int64](1)) || !Gt(a, NewSet[int64](1)) {
		t.Error("ordering means subset")
	}
}

func TestSetErrorsAndMutation(t *testing.T) {
	tests := []struct {
		name string
		f    func()
		want string
	}{
		{"remove", func() { NewSet[int64](1).Remove(5) }, "KeyError: 5"},
		{"remove str", func() { NewSet("a").Remove("b") }, "KeyError: 'b'"},
		{"pop empty", func() { NewSet[int64]().Pop() }, "KeyError: 'pop from an empty set'"},
		{"change during iteration", func() {
			s := NewSet[int64](1, 2, 3)
			for c := s.Iter(); c.Next(); {
				s.Add(c.Value() + 10)
			}
		}, "RuntimeError: Set changed size during iteration"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := panicMsg(t, tc.f); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
	s := NewSet[int64](1)
	s.Discard(9) // no error
	if s.Pop() != 1 || s.Len() != 0 {
		t.Error("pop")
	}
}
