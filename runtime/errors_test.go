package runtime

import "testing"

func TestExceptionStrAndRepr(t *testing.T) {
	// Expected values come from CPython 3.13.
	tests := []struct {
		e         *PyError
		str, repr string
		last      string // the last line of an uncaught traceback
	}{
		{NewError("ValueError"), "", "ValueError()", "ValueError"},
		{NewError("ValueError", "x"), "x", "ValueError('x')", "ValueError: x"},
		{NewError("ValueError", "x", int64(2)), "('x', 2)", "ValueError('x', 2)", "ValueError: ('x', 2)"},
		{NewError("ValueError", 5), "5", "ValueError(5)", "ValueError: 5"},
		{NewError("ValueError", 1.5), "1.5", "ValueError(1.5)", "ValueError: 1.5"},
		{NewError("ValueError", "it's"), "it's", `ValueError("it's")`, "ValueError: it's"},
		{NewError("KeyError", "k"), "'k'", "KeyError('k')", "KeyError: 'k'"},
		{NewError("ValueError", NewList("a")), "['a']", "ValueError(['a'])", "ValueError: ['a']"},
		{NewError("AssertionError"), "", "AssertionError()", "AssertionError"},
	}
	for _, tc := range tests {
		if got := Str(tc.e); got != tc.str {
			t.Errorf("str(%s) = %q, want %q", tc.repr, got, tc.str)
		}
		if got := Repr(tc.e); got != tc.repr {
			t.Errorf("repr = %q, want %q", got, tc.repr)
		}
		if got := tc.e.Error(); got != tc.last {
			t.Errorf("Error() = %q, want %q", got, tc.last)
		}
	}
}

func TestExceptionHierarchy(t *testing.T) {
	// Expected values come from the __mro__ of each class in CPython 3.13.
	tests := []struct {
		class, parent string
		want          bool
	}{
		{"ZeroDivisionError", "ZeroDivisionError", true},
		{"ZeroDivisionError", "ArithmeticError", true},
		{"ZeroDivisionError", "Exception", true},
		{"ZeroDivisionError", "BaseException", true},
		{"ZeroDivisionError", "LookupError", false},
		{"IndexError", "LookupError", true},
		{"KeyError", "LookupError", true},
		{"IndexError", "KeyError", false},
		{"RecursionError", "RuntimeError", true},
		{"NotImplementedError", "RuntimeError", true},
		{"UnicodeError", "ValueError", true},
		{"ValueError", "TypeError", false},
		{"Exception", "ValueError", false},
		{"BaseException", "Exception", false},
	}
	for _, tc := range tests {
		if got := IsSubclass(tc.class, tc.parent); got != tc.want {
			t.Errorf("IsSubclass(%s, %s) = %v, want %v", tc.class, tc.parent, got, tc.want)
		}
	}
	e := NewError("IndexError", "x")
	if !Match(e, "ValueError", "LookupError") || Match(e, "ValueError", "TypeError") || Match(e) {
		t.Error("Match tests every listed class")
	}
	for _, name := range []string{"Exception", "OSError", "RecursionError", "TimeoutError", "SyntaxError"} {
		if !IsExceptionClass(name) {
			t.Errorf("%s must be a known class", name)
		}
	}
	for _, name := range []string{"Foo", "int", "CustomThing", ""} {
		if IsExceptionClass(name) {
			t.Errorf("%q must not be a known class", name)
		}
	}
}

func TestCatch(t *testing.T) {
	if got := Catch(func() {}); got != nil {
		t.Errorf("no panic: got %v", got)
	}
	got := Catch(func() { panic(NewError("ValueError", "boom")) })
	if got == nil || got.Type != "ValueError" || Str(got) != "boom" {
		t.Errorf("Catch must return the exception, got %v", got)
	}
	got = Catch(func() { FloorDiv(1, 0) })
	if got == nil || !Match(got, "ArithmeticError") {
		t.Errorf("runtime errors are catchable by class, got %v", got)
	}

	// A panic that is not a Python exception is a bug and must go on.
	defer func() {
		if r := recover(); r != "not python" {
			t.Errorf("foreign panic must propagate, got %v", r)
		}
	}()
	Catch(func() { panic("not python") })
	t.Error("unreachable: Catch swallowed a foreign panic")
}

func TestRecursionGuard(t *testing.T) {
	var reached int
	var deep func(n int)
	deep = func(n int) {
		Enter()
		defer Leave()
		reached = n
		deep(n + 1)
	}
	err := Catch(func() { deep(1) })
	if err == nil || err.Error() != "RecursionError: maximum recursion depth exceeded" {
		t.Fatalf("got %v", err)
	}
	if reached != 999 {
		t.Errorf("CPython reaches depth 999 before failing, got %d", reached)
	}
	if depth != 0 {
		t.Errorf("the counter must unwind with the stack, got %d", depth)
	}

	// Recursion that stays under the limit is fine.
	var down func(n int) int
	down = func(n int) int {
		Enter()
		defer Leave()
		if n == 0 {
			return 0
		}
		return 1 + down(n-1)
	}
	if got := down(900); got != 900 {
		t.Errorf("down(900) = %d", got)
	}
}
