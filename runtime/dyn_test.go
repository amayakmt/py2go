package runtime

import (
	"reflect"
	"testing"
)

func TestDynBinaryNumbers(t *testing.T) {
	for _, tc := range []struct {
		op   string
		a, b any
		want any
	}{
		{"Add", int64(1), int64(2), int64(3)},
		{"Add", int64(1), 2.5, 3.5},
		{"Add", true, true, int64(2)},
		{"Mult", "ab", int64(2), "abab"},
		{"Mult", int64(2), "ab", "abab"},
		{"Add", "a", "b", "ab"},
		{"FloorDiv", int64(-7), int64(2), int64(-4)},
		{"Mod", int64(-7), int64(3), int64(2)},
		{"Div", int64(7), int64(2), 3.5},
		{"Pow", int64(2), int64(10), int64(1024)},
		{"Pow", int64(2), int64(-1), 0.5},
		{"LShift", int64(1), int64(4), int64(16)},
		{"BitAnd", int64(12), int64(10), int64(8)},
		{"Mod", "%d|%s", NewVarTuple[any](int64(7), "x"), "7|x"},
	} {
		if got := DynBinary(tc.op, tc.a, tc.b); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%v %s %v = %v (%T), want %v", tc.a, tc.op, tc.b, got, got, tc.want)
		}
	}
}

func TestDynBinaryErrors(t *testing.T) {
	for _, tc := range []struct {
		op   string
		a, b any
		want string
	}{
		{"Add", int64(1), "x", "TypeError: unsupported operand type(s) for +: 'int' and 'str'"},
		{"Add", "x", int64(1), `TypeError: can only concatenate str (not "int") to str`},
		{"Add", NewList[int64](), NewVarTuple[int64](), `TypeError: can only concatenate list (not "tuple") to list`},
		{"Mult", "a", 1.5, "TypeError: can't multiply sequence by non-int of type 'float'"},
		{"LShift", 1.5, int64(2), "TypeError: unsupported operand type(s) for <<: 'float' and 'int'"},
		{"Div", int64(1), int64(0), "ZeroDivisionError: division by zero"},
		{"Mod", int64(1), int64(0), "ZeroDivisionError: integer modulo by zero"},
	} {
		got := catch(func() { DynBinary(tc.op, tc.a, tc.b) })
		if got != tc.want {
			t.Errorf("%v %s %v: got %q, want %q", tc.a, tc.op, tc.b, got, tc.want)
		}
	}
}

func TestDynCompareAcrossTypes(t *testing.T) {
	if !DynCompare("Eq", int64(1), 1.0) || !DynCompare("Eq", true, int64(1)) || DynCompare("Eq", NewList[int64](), NewVarTuple[int64]()) {
		t.Error("equality across number types and sequence kinds")
	}
	if !DynCompare("Lt", int64(1), 2.5) || !DynCompare("GtE", "b", "a") || !DynCompare("Lt", NewVarTuple[any](int64(1)), NewVarTuple[any](int64(1), int64(0))) {
		t.Error("ordering")
	}
	if got := catch(func() { DynCompare("Lt", "a", int64(1)) }); got != "TypeError: '<' not supported between instances of 'str' and 'int'" {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { DynCompare("Lt", nil, int64(1)) }); got != "TypeError: '<' not supported between instances of 'NoneType' and 'int'" {
		t.Errorf("got %q", got)
	}
}

func TestDynContainersAndKeys(t *testing.T) {
	d := NewDict[any, any]()
	DynSetItem(d, int64(1), "int")
	DynSetItem(d, 1.0, "float")
	DynSetItem(d, true, "bool")
	if d.Len() != 1 || DynGetItem(d, int64(1)) != "bool" {
		t.Errorf("1, 1.0 and True are one key: %v", d)
	}
	DynSetItem(d, NewVarTuple[any](int64(1), "a"), "tuple")
	if got := DynGetItem(d, NewVarTuple[any](1.0, "a")); got != "tuple" {
		t.Errorf("tuple key: %v", got)
	}
	if got := catch(func() { DynSetItem(d, NewList[any](), 0) }); got != "TypeError: unhashable type: 'list'" {
		t.Errorf("got %q", got)
	}
	s := NewSet[any]()
	s.Add(int64(2))
	s.Add(2.0)
	if s.Len() != 1 {
		t.Error("2 and 2.0 are one element")
	}
	l := NewList[int64](1, 2, 3)
	if got := catch(func() { DynSetItem(l, int64(0), "a") }); got != "TypeError: expected int, got str" {
		t.Errorf("got %q", got)
	}
	DynSetItem(l, Slice{Lo: int64(1)}, NewList[int64](9))
	if Repr(l) != "[1, 9]" {
		t.Errorf("slice assignment: %s", Repr(l))
	}
	if got := DynGetItem(NewVarTuple[any](int64(1), int64(2), int64(3)), Slice{Step: int64(-1)}); Repr(got) != "(3, 2, 1)" {
		t.Errorf("tuple slice: %s", Repr(got))
	}
	if !DynContains(NewVarTuple[any](int64(1), "a"), "a") || DynContains("abc", "d") {
		t.Error("in")
	}
}

func TestDynCallsAndMethods(t *testing.T) {
	f := func(a int64, b int64) int64 { return a*10 + b }
	RegisterFunc(f, &FuncInfo{Name: "f", Params: []string{"a", "b"}, Defaults: map[string]func() any{"b": func() any { return int64(5) }}})
	if DynCall(f, int64(1)) != int64(15) || DynCallKw(f, []any{int64(1)}, []KW{{"b", int64(2)}}) != int64(12) || DynCallKw(f, nil, []KW{{"b", int64(2)}, {"a", int64(3)}}) != int64(32) {
		t.Error("defaults and keywords")
	}
	if got := catch(func() { DynCall(f) }); got != "TypeError: f() missing 1 required positional argument: 'a'" {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { DynCallKw(f, []any{int64(1)}, []KW{{"c", 1}}) }); got != "TypeError: f() got an unexpected keyword argument 'c'" {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { DynCall(int64(3)) }); got != "TypeError: 'int' object is not callable" {
		t.Errorf("got %q", got)
	}
	if got := DynCallMethod("a,b", "split", []any{","}, nil); Repr(got) != "['a', 'b']" {
		t.Errorf("split: %s", Repr(got))
	}
	l := NewList[any](int64(3), int64(1), int64(2))
	DynCallMethod(l, "sort", nil, []KW{{"reverse", true}})
	DynCallMethod(l, "append", []any{"x"}, nil)
	if Repr(l) != "[3, 2, 1, 'x']" {
		t.Errorf("list methods: %s", Repr(l))
	}
	if got := catch(func() { DynCallMethod(l, "append", []any{1, 2}, nil) }); got != "TypeError: list.append() takes exactly one argument (2 given)" {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { DynCallMethod("s", "nope", nil, nil) }); got != "AttributeError: 'str' object has no attribute 'nope'" {
		t.Errorf("got %q", got)
	}
	if got := DynCall(BuiltinFunc("len"), "abc"); got != int64(3) {
		t.Errorf("len as a value: %v", got)
	}
	if got := DynCallMethod(BuiltinClassOf("dict"), "fromkeys", []any{NewList[any](int64(1))}, nil); Repr(got) != "{1: None}" {
		t.Errorf("fromkeys: %s", Repr(got))
	}
}

func TestStrFormat(t *testing.T) {
	for _, tc := range []struct {
		format string
		args   []any
		kw     []KW
		want   string
	}{
		{"{} {}", []any{int64(1), "a"}, nil, "1 a"},
		{"{1}{0}", []any{"a", "b"}, nil, "ba"},
		{"{x:>5}|{y:.2f}", nil, []KW{{"x", "ab"}, {"y", 3.14159}}, "   ab|3.14"},
		{"{0!r} {0!s}", []any{"q"}, nil, "'q' q"},
		{"{{}} {:05d}", []any{int64(42)}, nil, "{} 00042"},
		{"{0[1]} {0[0]}", []any{NewList[any]("a", "b")}, nil, "b a"},
		{"{:{}}", []any{int64(7), int64(4)}, nil, "   7"},
	} {
		if got := strFormat(tc.format, tc.args, tc.kw); got != tc.want {
			t.Errorf("%q = %q, want %q", tc.format, got, tc.want)
		}
	}
	if got := catch(func() { strFormat("{}", nil, nil) }); got != "IndexError: Replacement index 0 out of range for positional args tuple" {
		t.Errorf("got %q", got)
	}
}

func TestDynBuiltins(t *testing.T) {
	if DynBin(int64(5)) != "0b101" || DynHex(int64(-255)) != "-0xff" || DynOct(int64(8)) != "0o10" || DynChr(int64(65)) != "A" || DynOrd("a") != 97 {
		t.Error("bin hex oct chr ord")
	}
	if DynPow(int64(3), int64(4), int64(5)) != int64(1) || DynIntBase("ff", int64(16)) != 255 || DynIntBase("0b101", int64(0)) != 5 {
		t.Error("pow and int with a base")
	}
	if Repr(DynDivmod(int64(-17), int64(5))) != "(-4, 3)" {
		t.Error("divmod")
	}
	if got := catch(func() { DynDivmod("a", int64(1)) }); got != "TypeError: unsupported operand type(s) for divmod(): 'str' and 'int'" {
		t.Errorf("got %q", got)
	}
	if got := DynSorted(NewList[any](int64(3), 1.5, true), nil, false); Repr(got) != "[True, 1.5, 3]" {
		t.Errorf("sorted: %s", Repr(got))
	}
	if got := DynMinMax(true, NewList[any]("a", "bb"), BuiltinFunc("len"), nil, false); got != "bb" {
		t.Errorf("max with key: %v", got)
	}
	if got := DynSum(NewList[any](int64(1), 2.5), nil); got != 3.5 {
		t.Errorf("sum: %v", got)
	}
	if got := catch(func() { DynMinMax(false, NewList[any](), nil, nil, false) }); got != "ValueError: min() iterable argument is empty" {
		t.Errorf("got %q", got)
	}
}

func TestDynDeepKeepsSharing(t *testing.T) {
	inner := NewList[int64](0, 0)
	outer := NewList[*List[int64]](inner, inner)
	got := DynDeep(outer).(*List[any])
	DynSetItem(got.items[0], int64(0), 1.5)
	if Repr(got) != "[[1.5, 0], [1.5, 0]]" {
		t.Errorf("deep copy lost the sharing or the types: %s", Repr(got))
	}
	if Repr(outer) != "[[0, 0], [0, 0]]" {
		t.Error("the original must not change")
	}
}

func TestIterationAndNext(t *testing.T) {
	var out []any
	for c := DynIter("ab").Iter(); c.Next(); {
		out = append(out, c.Value())
	}
	if !reflect.DeepEqual(out, []any{"a", "b"}) {
		t.Errorf("iterating a string: %v", out)
	}
	it := DynIterator(NewList[any](int64(1)))
	if DynNext(it, nil, false) != int64(1) || DynNext(it, "done", true) != "done" {
		t.Error("next with a default")
	}
	if got := catch(func() { DynNext(it, nil, false) }); got != "StopIteration" && got != "StopIteration: " {
		t.Errorf("got %q", got)
	}
	if got := catch(func() { DynNext(int64(1), nil, false) }); got != "TypeError: 'int' object is not an iterator" {
		t.Errorf("got %q", got)
	}
}
