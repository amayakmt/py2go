package runtime

import (
	"math"
	"testing"
)

func TestFloatRepr(t *testing.T) {
	// Expected strings come from CPython 3.13 repr().
	tests := []struct {
		in   float64
		want string
	}{
		{1, "1.0"},
		{0, "0.0"},
		{math.Copysign(0, -1), "-0.0"},
		{2.5, "2.5"},
		{-2.5, "-2.5"},
		{float64(0.1) + float64(0.2), "0.30000000000000004"},
		{100, "100.0"},
		{123456789012345.0, "123456789012345.0"},
		{1e15, "1000000000000000.0"},
		{1e16, "1e+16"},
		{1.5e16, "1.5e+16"},
		{123456789012345678.0, "1.2345678901234568e+17"},
		{1e22, "1e+22"},
		{0.0001, "0.0001"},
		{0.00012345, "0.00012345"},
		{1e-5, "1e-05"},
		{1.5e-7, "1.5e-07"},
		{1e-100, "1e-100"},
		{1e100, "1e+100"},
		{math.Inf(1), "inf"},
		{math.Inf(-1), "-inf"},
		{math.NaN(), "nan"},
		{5e-324, "5e-324"},
		{math.MaxFloat64, "1.7976931348623157e+308"},
	}
	for _, tc := range tests {
		if got := FloatRepr(tc.in); got != tc.want {
			t.Errorf("FloatRepr(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestStr(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{true, "True"},
		{false, "False"},
		{int64(-3), "-3"},
		{7, "7"}, // untyped constants arrive as int
		{1.0, "1.0"},
		{"héllo", "héllo"},
	}
	for _, tc := range tests {
		if got := Str(tc.in); got != tc.want {
			t.Errorf("Str(%#v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFloorDivAndMod(t *testing.T) {
	// Expected values come from CPython: a // b and a % b.
	tests := []struct{ a, b, div, mod int64 }{
		{7, 2, 3, 1},
		{-7, 2, -4, 1},
		{7, -2, -4, -1},
		{-7, -2, 3, -1},
		{-7, 3, -3, 2},
		{6, 3, 2, 0},
		{-6, 3, -2, 0},
		{0, 5, 0, 0},
		{0, -5, 0, 0},
	}
	for _, tc := range tests {
		if got := FloorDiv(tc.a, tc.b); got != tc.div {
			t.Errorf("FloorDiv(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.div)
		}
		if got := Mod(tc.a, tc.b); got != tc.mod {
			t.Errorf("Mod(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.mod)
		}
	}
}

func TestFloatFloorDivAndMod(t *testing.T) {
	tests := []struct{ a, b, div, mod float64 }{
		{7, 2, 3, 1},
		{-7, 2, -4, 1},
		{7, -2, -4, -1},
		{7.5, 2, 3, 1.5},
		{-7.5, 2, -4, 0.5},
		{1, 0.1, 9, 0.09999999999999995}, // 1/0.1 rounds to 10.0, Python still gives 9
	}
	for _, tc := range tests {
		if got := FloorDivFloat(tc.a, tc.b); got != tc.div {
			t.Errorf("FloorDivFloat(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.div)
		}
		if got := ModFloat(tc.a, tc.b); got != tc.mod {
			t.Errorf("ModFloat(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.mod)
		}
	}
}

// panicMsg runs f and returns the Python-style message it panicked with.
func panicMsg(t *testing.T, f func()) (msg string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		msg = r.(*PyError).Error()
	}()
	f()
	return ""
}

func TestZeroDivisionMessages(t *testing.T) {
	tests := []struct {
		name string
		f    func()
		want string
	}{
		{"FloorDiv", func() { FloorDiv(1, 0) }, "ZeroDivisionError: integer division or modulo by zero"},
		{"Mod", func() { Mod(1, 0) }, "ZeroDivisionError: integer modulo by zero"},
		{"Div", func() { Div(1, 0) }, "ZeroDivisionError: float division by zero"},
		{"DivInt", func() { DivInt(1, 0) }, "ZeroDivisionError: division by zero"},
		{"FloorDivFloat", func() { FloorDivFloat(1, 0) }, "ZeroDivisionError: float floor division by zero"},
		{"ModFloat", func() { ModFloat(1, 0) }, "ZeroDivisionError: float modulo by zero"},
		{"LShift", func() { LShift(1, -1) }, "ValueError: negative shift count"},
		{"CheckRangeStep", func() { CheckRangeStep(0) }, "ValueError: range() arg 3 must not be zero"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := panicMsg(t, tc.f); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestShifts(t *testing.T) {
	if got := LShift(1, 10); got != 1024 {
		t.Errorf("LShift(1, 10) = %d", got)
	}
	if got := RShift(-8, 1); got != -4 {
		t.Errorf("RShift(-8, 1) = %d", got)
	}
	if got := RShift(-8, 100); got != -1 {
		t.Errorf("RShift(-8, 100) = %d, Python gives -1", got)
	}
	if got := RShift(8, 100); got != 0 {
		t.Errorf("RShift(8, 100) = %d, Python gives 0", got)
	}
}
