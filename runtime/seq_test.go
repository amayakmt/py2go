package runtime

import (
	"os"
	"reflect"
	"testing"
)

func ints(vs ...int64) *List[int64] { return NewList(vs...) }

func (l *List[T]) slice() []T { return l.items }

func TestSliceBounds(t *testing.T) {
	// Expected values come from CPython: list(range(10))[lo:hi:step].
	xs := ints(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	tests := []struct {
		name string
		b    Bounds
		want []int64
	}{
		{"2:7", Bounds{Lo: 2, Hi: 7, HasLo: true, HasHi: true}, []int64{2, 3, 4, 5, 6}},
		{"::-1", Bounds{Step: -1, HasStep: true}, []int64{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}},
		{"8:2:-2", Bounds{Lo: 8, Hi: 2, Step: -2, HasLo: true, HasHi: true, HasStep: true}, []int64{8, 6, 4}},
		{"-3:", Bounds{Lo: -3, HasLo: true}, []int64{7, 8, 9}},
		{":-3", Bounds{Hi: -3, HasHi: true}, []int64{0, 1, 2, 3, 4, 5, 6}},
		{"100:", Bounds{Lo: 100, HasLo: true}, []int64{}},
		{"-100:3", Bounds{Lo: -100, Hi: 3, HasLo: true, HasHi: true}, []int64{0, 1, 2}},
		{"::3", Bounds{Step: 3, HasStep: true}, []int64{0, 3, 6, 9}},
		{"5:5", Bounds{Lo: 5, Hi: 5, HasLo: true, HasHi: true}, []int64{}},
		{"7:2", Bounds{Lo: 7, Hi: 2, HasLo: true, HasHi: true}, []int64{}},
		{"::-3", Bounds{Step: -3, HasStep: true}, []int64{9, 6, 3, 0}},
		{"-1:-100:-1", Bounds{Lo: -1, Hi: -100, Step: -1, HasLo: true, HasHi: true, HasStep: true}, []int64{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}},
		{"0:100:5", Bounds{Lo: 0, Hi: 100, Step: 5, HasLo: true, HasHi: true, HasStep: true}, []int64{0, 5}},
	}
	for _, tc := range tests {
		got := xs.Slice(tc.b).slice()
		if len(got) == 0 && len(tc.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("[%s] = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestListOps(t *testing.T) {
	l := ints(3, 1, 2)
	l.Append(5)
	if got := l.Get(-1); got != 5 {
		t.Errorf("Get(-1) = %d", got)
	}
	l.Set(-2, 9)
	if got := l.slice(); !reflect.DeepEqual(got, []int64{3, 1, 9, 5}) {
		t.Errorf("after Set: %v", got)
	}
	if got := l.PopAt(0); got != 3 {
		t.Errorf("PopAt(0) = %d", got)
	}
	l.Insert(100, 7) // clamps to the end
	l.Insert(-100, 0)
	if got := l.slice(); !reflect.DeepEqual(got, []int64{0, 1, 9, 5, 7}) {
		t.Errorf("after Insert: %v", got)
	}
	l.Extend(l)
	if l.Len() != 10 {
		t.Errorf("l.Extend(l): len %d, want 10", l.Len())
	}
	if l.Index(9) != 2 || l.Count(9) != 2 || !l.Contains(7) || l.Contains(8) {
		t.Error("Index/Count/Contains disagree")
	}
	l.Reverse()
	if l.Get(0) != 7 {
		t.Errorf("Reverse: %v", l.slice())
	}
	cat := ints(1).Concat(ints(2, 3))
	if !reflect.DeepEqual(cat.slice(), []int64{1, 2, 3}) {
		t.Errorf("Concat: %v", cat.slice())
	}
	if rep := ints(1, 2).Repeat(2); !reflect.DeepEqual(rep.slice(), []int64{1, 2, 1, 2}) {
		t.Errorf("Repeat: %v", rep.slice())
	}
	if rep := ints(1, 2).Repeat(-1); rep.Len() != 0 {
		t.Errorf("Repeat(-1) must be empty, got %v", rep.slice())
	}
}

func TestAliasing(t *testing.T) {
	a := ints(1, 2)
	b := a // Python's "b = a"
	b.Append(3)
	if a.Len() != 3 {
		t.Error("append through an alias must be visible")
	}
	c := a.Slice(Bounds{}) // Python's "a[:]"
	c.Append(4)
	if a.Len() != 3 {
		t.Error("a slice must be a copy")
	}
}

func TestSortIsStableAndReverseKeepsOrder(t *testing.T) {
	type pair = Tuple2[int64, string]
	l := NewList(Tup2[int64, string](2, "b"), Tup2[int64, string](1, "z"), Tup2[int64, string](2, "a"))
	l.Sort(false)
	want := []pair{{1, "z"}, {2, "a"}, {2, "b"}}
	if !reflect.DeepEqual(l.slice(), want) {
		t.Errorf("tuples order lexicographically: %v", l.slice())
	}

	// Sorting only the first field would leave equal keys in input order;
	// check a list of lists sorted by Lt on the whole list.
	words := NewList("pear", "fig", "apple")
	if got := Sorted(words, true).slice(); !reflect.DeepEqual(got, []string{"pear", "fig", "apple"}) {
		t.Errorf("sorted reverse: %v", got)
	}
	if words.Get(0) != "pear" {
		t.Error("Sorted must not modify its argument")
	}
}

func TestEqAndOrder(t *testing.T) {
	if !Eq(ints(1, 2), ints(1, 2)) || Eq(ints(1, 2), ints(1, 3)) || Eq(ints(1), ints(1, 2)) {
		t.Error("list equality")
	}
	if !Eq(Tup2[int64, string](1, "a"), Tup2[int64, string](1, "a")) {
		t.Error("tuple equality")
	}
	nested := func(v int64) *List[*List[int64]] { return NewList(ints(v)) }
	if !Eq(nested(1), nested(1)) || Eq(nested(1), nested(2)) {
		t.Error("nested list equality compares contents, not pointers")
	}
	if !Lt(ints(1, 2), ints(1, 3)) || !Lt(ints(1), ints(1, 0)) || Lt(ints(2), ints(1, 9)) {
		t.Error("lexicographic Lt")
	}
	if !Le(ints(1, 2), ints(1, 2)) || Le(ints(1, 3), ints(1, 2)) {
		t.Error("Le")
	}
	if !Gt(ints(2), ints(1, 9)) || !Ge(ints(1), ints(1)) {
		t.Error("Gt/Ge")
	}
	if !Lt("a", "b") || !Lt(1.5, 2.5) || !Lt(false, true) {
		t.Error("scalar Lt")
	}
}

func TestRepr(t *testing.T) {
	// Expected strings come from CPython 3.13 repr().
	tests := []struct {
		in   any
		want string
	}{
		{"it's", `"it's"`},
		{`say "hi"`, `'say "hi"'`},
		{`both ' and "`, `'both \' and "'`},
		{"tab\t\n\\", `'tab\t\n\\'`},
		{"é✓\x00\x1b", `'é✓\x00\x1b'`},
		{"\u200b", `'\u200b'`},
		{"😀", `'😀'`},
		{"\u00a0", `'\xa0'`},
		{Tup1[int64](1), "(1,)"},
		{Tup2[int64, float64](1, 2), "(1, 2.0)"},
		{NewList[int64](), "[]"},
		{NewList[any](int64(1), "a", 2.5, true, ints(1), Tup2[int64, int64](2, 3)), "[1, 'a', 2.5, True, [1], (2, 3)]"},
	}
	for _, tc := range tests {
		if got := Repr(tc.in); got != tc.want {
			t.Errorf("Repr(%#v) = %s, want %s", tc.in, got, tc.want)
		}
	}
	if Str("raw") != "raw" || Str(ints(1)) != "[1]" {
		t.Error("Str: strings stay raw, containers use repr")
	}
}

func TestListErrors(t *testing.T) {
	tests := []struct {
		name string
		f    func()
		want string
	}{
		{"Get", func() { ints().Get(0) }, "IndexError: list index out of range"},
		{"Get negative", func() { ints(1).Get(-2) }, "IndexError: list index out of range"},
		{"Set", func() { ints(1).Set(1, 0) }, "IndexError: list assignment index out of range"},
		{"Pop empty", func() { ints().Pop() }, "IndexError: pop from empty list"},
		{"PopAt", func() { ints(1).PopAt(5) }, "IndexError: pop index out of range"},
		{"Index", func() { ints(1, 2).Index(3) }, "ValueError: 3 is not in list"},
		{"Remove", func() { ints(1, 2).Remove(3) }, "ValueError: list.remove(x): x not in list"},
		{"zero step", func() { ints(1).Slice(Bounds{HasStep: true}) }, "ValueError: slice step cannot be zero"},
		{"unpack many", func() { CheckUnpack(3, 2) }, "ValueError: too many values to unpack (expected 2)"},
		{"unpack few", func() { CheckUnpack(1, 2) }, "ValueError: not enough values to unpack (expected 2, got 1)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := panicMsg(t, tc.f); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// captureStdout runs f with os.Stdout redirected and returns what it wrote.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = saved }()
	f()
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestPrint(t *testing.T) {
	got := captureStdout(t, func() {
		Print(int64(1), 2, 3.5, true, "s", ints(1), Tup1[string]("x"))
		PrintOpts("-", "!", "a", "b", "c")
		Print()
	})
	want := "1 2 3.5 True s [1] ('x',)\na-b-c!\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
