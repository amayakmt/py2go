package runtime

import (
	"math"
	"testing"
)

func TestHashFloat(t *testing.T) {
	// Values from CPython 3.13.
	tests := []struct {
		x    float64
		want int64
	}{
		{0.0, 0}, {math.Copysign(0, -1), 0}, {1.0, 1}, {2.5, 1152921504606846978}, {-2.5, -1152921504606846978},
		{1e300, 1224995262755759164}, {1e-300, 482449582752280463}, {3.14159, 326484311674566659},
		{1e22, 1864712049423028464}, {0.1, 230584300921369408}, {math.Pow(2, 61), 1}, {math.Pow(2, 62), 2},
		{-7.0, -7}, {123456.789, 1819310134279660096}, {5e-324, 16777216}, {math.MaxFloat64, 2234066890152476671},
		{math.Inf(1), 314159}, {math.Inf(-1), -314159},
	}
	for _, tc := range tests {
		if got := Hash(tc.x); got != tc.want {
			t.Errorf("hash(%v) = %d, want %d", tc.x, got, tc.want)
		}
	}
}

type fakeObj struct{ name string }

func (f *fakeObj) PyClassName() string { return f.name }

func TestTypeNameAndClassRef(t *testing.T) {
	tests := []struct {
		v    any
		want string
	}{
		{int64(1), "int"}, {true, "bool"}, {2.5, "float"}, {"s", "str"}, {NewList[int64](), "list"},
		{NewDict[string, int64](), "dict"}, {NewSet[int64](), "set"}, {NewRange(0, 1, 1), "range"},
		{Tup2[int64, int64](1, 2), "tuple"}, {&fakeObj{"Thing"}, "Thing"}, {nil, "NoneType"},
	}
	for _, tc := range tests {
		if got := TypeName(tc.v); got != tc.want {
			t.Errorf("TypeName(%T) = %q, want %q", tc.v, got, tc.want)
		}
	}
	if ClassOf("A") != ClassOf("A") || ClassOf("A") == ClassOf("B") {
		t.Error("class objects must be unique per name")
	}
	if got := Repr(ClassOf("A")); got != "<class '__main__.A'>" {
		t.Errorf("repr of a class = %s", got)
	}
	if got := Repr(BuiltinClassOf("int")); got != "<class 'int'>" {
		t.Errorf("repr of int = %s", got)
	}
}

func TestMessagesOfObjectErrors(t *testing.T) {
	tests := []struct{ got, want string }{
		{UnsupportedOperand("+", "A", "int").Error(), "TypeError: unsupported operand type(s) for +: 'A' and 'int'"},
		{NotOrderable("<", &fakeObj{"A"}, &fakeObj{"B"}).Error(), "TypeError: '<' not supported between instances of 'A' and 'B'"},
		{NoAttribute("C", "x").Error(), "AttributeError: 'C' object has no attribute 'x'"},
		{NoSetter("C", "x").Error(), "AttributeError: property 'x' of 'C' object has no setter"},
		{FrozenInstance("x").Error(), "FrozenInstanceError: cannot assign to field 'x'"},
		{UnhashableType("C").Error(), "TypeError: unhashable type: 'C'"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

func TestExceptionChainReport(t *testing.T) {
	inner := newError("ValueError", "root")
	outer := newError("RuntimeError", "wrapper")
	RaiseFrom(outer, inner)
	want := "Traceback (most recent call last):\nValueError: root\n\n" +
		"The above exception was the direct cause of the following exception:\n\n" +
		"Traceback (most recent call last):\nRuntimeError: wrapper\n"
	if got := report(outer, map[*PyError]bool{}); got != want {
		t.Errorf("cause chain:\n%s\nwant:\n%s", got, want)
	}

	ctx := newError("KeyError", "k")
	handled := newError("ValueError", "first")
	WithContext(ctx, handled)
	want = "Traceback (most recent call last):\nValueError: first\n\n" +
		"During handling of the above exception, another exception occurred:\n\n" +
		"Traceback (most recent call last):\nKeyError: 'k'\n"
	if got := report(ctx, map[*PyError]bool{}); got != want {
		t.Errorf("context chain:\n%s\nwant:\n%s", got, want)
	}

	quiet := newError("RuntimeError", "quiet")
	RaiseFrom(quiet, nil)
	WithContext(quiet, handled)
	if got := report(quiet, map[*PyError]bool{}); got != "Traceback (most recent call last):\nRuntimeError: quiet\n" {
		t.Errorf("from None must hide the context, got %q", got)
	}
	if quiet.Context != handled || !quiet.SuppressContext {
		t.Error("from None keeps __context__ and sets __suppress_context__")
	}
}

func TestUserExceptionRegistry(t *testing.T) {
	RegisterClass("MyErr", "MyErr", "ValueError", "Exception", "BaseException")
	if !IsSubclass("MyErr", "ValueError") || !IsSubclass("MyErr", "Exception") || IsSubclass("MyErr", "KeyError") {
		t.Error("user exception hierarchy")
	}
	if got := BuiltinMRO("KeyError"); len(got) != 4 || got[0] != "KeyError" || got[3] != "BaseException" {
		t.Errorf("BuiltinMRO(KeyError) = %v", got)
	}
	if BuiltinMRO("Nope") != nil {
		t.Error("unknown class")
	}
}

func TestRawArgs(t *testing.T) {
	all := []any{int64(1), "x", 2.5}
	if got := RawArgs(2, all...); len(got) != 2 {
		t.Errorf("RawArgs(2) = %v", got)
	}
	if got := RawArgs(-1, all...); len(got) != 3 {
		t.Errorf("RawArgs(-1) = %v", got)
	}
	if got := RawArgs(0, all...); len(got) != 0 {
		t.Errorf("RawArgs(0) = %v", got)
	}
}

func TestAnySeq(t *testing.T) {
	a, b := AnySeq{int64(1), "x"}, AnySeq{int64(1), "x"}
	if !Eq(a, b) || Eq(a, AnySeq{int64(1), "y"}) {
		t.Error("AnySeq equality")
	}
	if !Lt(a, AnySeq{int64(2), "a"}) {
		t.Error("AnySeq ordering")
	}
	if Hash(a) != Hash(b) {
		t.Error("AnySeq hash")
	}
	if got := Repr(a); got != "(1, 'x')" {
		t.Errorf("repr %s", got)
	}
}
