package runtime

import (
	"math"
	"reflect"
	"testing"
)

// Expected values in this file come from CPython 3.13.

func TestStringOps(t *testing.T) {
	s := "héllo wörld"
	if StrLen(s) != 11 || StrLen("") != 0 {
		t.Errorf("StrLen counts code points, got %d", StrLen(s))
	}
	checks := []struct{ got, want string }{
		{StrIndex(s, 1), "é"},
		{StrIndex(s, -1), "d"},
		{StrSlice(s, Bounds{Lo: 2, Hi: 5, HasLo: true, HasHi: true}), "llo"},
		{StrSlice(s, Bounds{Step: -1, HasStep: true}), "dlröw olléh"},
		{StrUpper(s), "HÉLLO WÖRLD"},
		{StrStrip("  x y "), "x y"},
		{StrStrip("\x1c a\x1f"), "a"},
		{StrStripChars("xxhixx", "x"), "hi"},
		{StrJoin(",", NewList("a", "b")), "a,b"},
		{StrReplace("aaa", "a", "bb"), "bbbbbb"},
		{StrRepeat("ab", 3), "ababab"},
		{StrRepeat("ab", 0), ""},
		{StrRepeat("ab", -2), ""},
	}
	for i, c := range checks {
		if c.got != c.want {
			t.Errorf("check %d: got %q, want %q", i, c.got, c.want)
		}
	}
	if StrFind(s, "w") != 6 || StrFind(s, "z") != -1 || StrFind(s, "") != 0 {
		t.Error("StrFind returns code point indexes")
	}
	if !StrStartsWith("abc", "ab") || !StrEndsWith("abc", "bc") || StrStartsWith("abc", "bc") {
		t.Error("StartsWith/EndsWith")
	}
	if StrCount("aaaa", "aa") != 2 || StrCount("abc", "") != 4 {
		t.Error("StrCount: non-overlapping, and 'abc'.count('') is 4")
	}
	if !StrContains(s, "é") || StrContains(s, "xyz") || !StrContains(s, "") {
		t.Error("StrContains")
	}
	if got := Chars("hé").slice(); !reflect.DeepEqual(got, []string{"h", "é"}) {
		t.Errorf("Chars: %q", got)
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		got  *List[string]
		want []string
	}{
		{StrSplitSep("a,b,,c", ","), []string{"a", "b", "", "c"}},
		{StrSplitSep("", ","), []string{""}},
		{StrSplit(" a  b\tc\n"), []string{"a", "b", "c"}},
		{StrSplit(""), []string{}},
	}
	for i, tc := range tests {
		if len(tc.want) == 0 && tc.got.Len() == 0 {
			continue
		}
		if !reflect.DeepEqual(tc.got.slice(), tc.want) {
			t.Errorf("split %d: got %q, want %q", i, tc.got.slice(), tc.want)
		}
	}
}

func TestStringErrors(t *testing.T) {
	tests := []struct {
		name string
		f    func()
		want string
	}{
		{"index", func() { StrIndex("abc", 5) }, "IndexError: string index out of range"},
		{"negative index", func() { StrIndex("abc", -4) }, "IndexError: string index out of range"},
		{"empty separator", func() { StrSplitSep("a", "") }, "ValueError: empty separator"},
		{"slice step", func() { StrSlice("a", Bounds{HasStep: true}) }, "ValueError: slice step cannot be zero"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := panicMsg(t, tc.f); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIntFromStr(t *testing.T) {
	ok := map[string]int64{
		"  -12 ": -12, "+7": 7, "1_000": 1000, "007": 7, " \t5\n": 5,
		"9223372036854775807": math.MaxInt64, "-9223372036854775808": math.MinInt64,
	}
	for in, want := range ok {
		if got := IntFromStr(in); got != want {
			t.Errorf("int(%q) = %d, want %d", in, got, want)
		}
	}
	for _, in := range []string{"1__0", "_1", "1_", "", "-", "abc", "12a", "1.5", "0x10"} {
		want := "ValueError: invalid literal for int() with base 10: " + quote(in)
		if got := panicMsg(t, func() { IntFromStr(in) }); got != want {
			t.Errorf("int(%q): got %q, want %q", in, got, want)
		}
	}
	if got := panicMsg(t, func() { IntFromStr("9223372036854775808") }); got != "OverflowError: int too large to convert to int64" {
		t.Errorf("overflow: %q", got)
	}
}

func TestFloatFromStr(t *testing.T) {
	ok := map[string]float64{
		"1.5": 1.5, " 1e3 ": 1000, "-.5": -0.5, "5.": 5, "1_0.5": 10.5, "inf": math.Inf(1),
		"-Infinity": math.Inf(-1), "1e+5": 1e5, "1.5e-3": 0.0015, "+1": 1, "1e999": math.Inf(1),
	}
	for in, want := range ok {
		if got := FloatFromStr(in); got != want {
			t.Errorf("float(%q) = %v, want %v", in, got, want)
		}
	}
	if !math.IsNaN(FloatFromStr("NaN")) {
		t.Error(`float("NaN") must be nan`)
	}
	for _, in := range []string{"1e", "e5", ".", "0x10", "--1", "1 2", "", "1_e5", "1e_5", "abc"} {
		want := "ValueError: could not convert string to float: " + quote(in)
		if got := panicMsg(t, func() { FloatFromStr(in) }); got != want {
			t.Errorf("float(%q): got %q, want %q", in, got, want)
		}
	}
}

func TestNumericConversions(t *testing.T) {
	for in, want := range map[float64]int64{2.7: 2, -2.7: -2, -0.5: 0, 1e10: 10000000000} {
		if got := IntFromFloat(in); got != want {
			t.Errorf("int(%v) = %d, want %d", in, got, want)
		}
	}
	for in, want := range map[float64]int64{0.5: 0, 1.5: 2, 2.5: 2, -2.5: -2, 2.4999: 2, 3.5: 4} {
		if got := RoundFloat(in); got != want {
			t.Errorf("round(%v) = %d, want %d", in, got, want)
		}
	}
	if got := panicMsg(t, func() { IntFromFloat(math.NaN()) }); got != "ValueError: cannot convert float NaN to integer" {
		t.Errorf("nan: %q", got)
	}
	if got := panicMsg(t, func() { RoundFloat(math.Inf(1)) }); got != "OverflowError: cannot convert float infinity to integer" {
		t.Errorf("inf: %q", got)
	}
	if got := panicMsg(t, func() { IntFromFloat(1e30) }); got != "OverflowError: int too large to convert to int64" {
		t.Errorf("1e30: %q", got)
	}
}

func TestSumMinMax(t *testing.T) {
	floats := func(n int, v float64) *List[float64] {
		l := NewList[float64]()
		for i := 0; i < n; i++ {
			l.Append(v)
		}
		return l
	}
	if got := SumFloat(floats(10, 0.1)); got != 1.0 {
		t.Errorf("sum([0.1] * 10) = %v, CPython 3.12+ gives 1.0", got)
	}
	if got := SumFloat(NewList(1e100, 1.0, -1e100)); got != 1.0 {
		t.Errorf("compensated sum = %v, want 1.0", got)
	}
	if got := SumFloat(NewList(0.1, 0.2, 0.3)); got != 0.6 {
		t.Errorf("sum([0.1, 0.2, 0.3]) = %v, want 0.6", got)
	}
	if got := SumFloat(NewList(math.Inf(1), 1.0)); !math.IsInf(got, 1) {
		t.Errorf("sum with inf = %v", got)
	}
	if got := SumFloat(NewList(math.Inf(1), math.Inf(-1))); !math.IsNaN(got) {
		t.Errorf("inf + -inf = %v, want nan", got)
	}
	if SumInt(NewList[int64](1, 2, 3)) != 6 {
		t.Error("SumInt")
	}
	if Min(3, 1, 2) != 1 || Max(3, 1, 2) != 3 || Min("b", "a") != "a" {
		t.Error("Min/Max")
	}
	if nan := math.NaN(); Min(1, nan) != 1 || !math.IsNaN(Min(nan, 1)) {
		t.Error("Min must keep the first item when comparisons are false, like Python")
	}
	if got := panicMsg(t, func() { MaxList(NewList[int64]()) }); got != "ValueError: max() iterable argument is empty" {
		t.Errorf("MaxList: %q", got)
	}
	if got := panicMsg(t, func() { MinList(NewList[int64]()) }); got != "ValueError: min() iterable argument is empty" {
		t.Errorf("MinList: %q", got)
	}
}

func TestPow(t *testing.T) {
	if PowInt(2, 10) != 1024 || PowInt(-2, 3) != -8 || PowInt(5, 0) != 1 || PowInt(2, 62) != 1<<62 {
		t.Error("PowInt")
	}
	floats := []struct{ a, b, want float64 }{
		{2, -2, 0.25}, {0.5, 3, 0.125}, {2, 0.5, 1.4142135623730951}, {-2, 3, -8}, {-8, 3, -512},
		{0, 0, 1}, {math.NaN(), 0, 1}, {1, math.NaN(), 1},
	}
	for _, f := range floats {
		if got := PowFloat(f.a, f.b); got != f.want {
			t.Errorf("%v ** %v = %v, want %v", f.a, f.b, got, f.want)
		}
	}
	if got := panicMsg(t, func() { PowFloat(0, -1) }); got != "ZeroDivisionError: 0.0 cannot be raised to a negative power" {
		t.Errorf("0 ** -1: %q", got)
	}
	if got := panicMsg(t, func() { PowFloat(10, 400) }); got != "OverflowError: (34, 'Numerical result out of range')" {
		t.Errorf("10.0 ** 400: %q", got)
	}
}

func TestDivMod(t *testing.T) {
	if d := DivModInt(-7, 2); d.F0 != -4 || d.F1 != 1 {
		t.Errorf("divmod(-7, 2) = %v", d)
	}
	if d := DivModInt(1, -3); d.F0 != -1 || d.F1 != -2 {
		t.Errorf("divmod(1, -3) = %v", d)
	}
	if d := DivModFloat(-7.5, 2); d.F0 != -4 || d.F1 != 0.5 {
		t.Errorf("divmod(-7.5, 2.0) = %v", d)
	}
	if got := panicMsg(t, func() { DivModFloat(1, 0) }); got != "ZeroDivisionError: float divmod()" {
		t.Errorf("float divmod by zero: %q", got)
	}
	if got := panicMsg(t, func() { DivModInt(1, 0) }); got != "ZeroDivisionError: integer division or modulo by zero" {
		t.Errorf("int divmod by zero: %q", got)
	}
}

func TestFormat(t *testing.T) {
	ints := []struct {
		v    int64
		spec string
		want string
	}{
		{42, "", "42"}, {42, "5", "   42"}, {42, "<5", "42   "}, {42, "^6", "  42  "}, {42, "*^7", "**42***"},
		{-42, "06", "-00042"}, {42, "+", "+42"}, {42, " ", " 42"}, {42, "+06", "+00042"}, {-42, "=6", "-   42"},
		{1234567, ",", "1,234,567"}, {-1234567, "12,", "  -1,234,567"}, {42, "05d", "00042"}, {0, "3d", "  0"},
		{math.MinInt64, "d", "-9223372036854775808"}, {7, "<05", "70000"}, {5, "x<4", "5xxx"},
	}
	for _, tc := range ints {
		if got := FormatInt(tc.v, tc.spec); got != tc.want {
			t.Errorf("format(%d, %q) = %q, want %q", tc.v, tc.spec, got, tc.want)
		}
	}
	floats := []struct {
		v    float64
		spec string
		want string
	}{
		{3.14159, ".2f", "3.14"}, {3.14159, "8.2f", "    3.14"}, {-3.14159, "08.2f", "-0003.14"},
		{2.5, ".0f", "2"}, {0.5, ".0f", "0"}, {1.5, ".0f", "2"}, {1234567.891, ",.2f", "1,234,567.89"},
		{1.5, "", "1.5"}, {1e16, "", "1e+16"}, {1, "", "1.0"}, {math.Copysign(0, -1), ".1f", "-0.0"},
		{-0.001, ".2f", "-0.00"}, {math.Inf(1), "6.1f", "   inf"}, {math.NaN(), ".1f", "nan"},
		{math.Inf(1), "08.1f", "00000inf"}, {2, "+.1f", "+2.0"}, {float64(0.1) + float64(0.2), ".20f", "0.30000000000000004441"},
		{1e22, ".1f", "10000000000000000000000.0"}, {5, "<8", "5.0     "}, {1.5, "^9.1f", "   1.5   "},
		{123.456, "10.1f", "     123.5"}, {123.456, "<10.1f", "123.5     "}, {42, ".2f", "42.00"}, {42, "8.3f", "  42.000"},
	}
	for _, tc := range floats {
		if got := FormatFloat(tc.v, tc.spec); got != tc.want {
			t.Errorf("format(%v, %q) = %q, want %q", tc.v, tc.spec, got, tc.want)
		}
	}
	strs := []struct{ v, spec, want string }{
		{"abc", "", "abc"}, {"abc", "6", "abc   "}, {"abc", ">6", "   abc"}, {"abc", "^7", "  abc  "},
		{"abc", "*^7", "**abc**"}, {"abcdef", ".3", "abc"}, {"héllo", "8", "héllo   "}, {"héllo", ">8.2", "      hé"}, {"", "3", "   "},
	}
	for _, tc := range strs {
		if got := FormatStr(tc.v, tc.spec); got != tc.want {
			t.Errorf("format(%q, %q) = %q, want %q", tc.v, tc.spec, got, tc.want)
		}
	}
}

func TestValidateSpec(t *testing.T) {
	good := []struct{ kind, spec string }{
		{"int", ""}, {"int", "05d"}, {"int", ">8"}, {"int", ".2f"}, {"int", ",d"},
		{"float", ""}, {"float", ".2f"}, {"float", "08.3f"}, {"float", "+,.1f"},
		{"str", ""}, {"str", ">10"}, {"str", "<8.3s"}, {"str", "*^9"},
	}
	for _, g := range good {
		if err := ValidateSpec(g.kind, g.spec); err != nil {
			t.Errorf("%s %q should be valid: %v", g.kind, g.spec, err)
		}
	}
	bad := []struct{ kind, spec string }{
		{"int", "x"}, {"int", "e"}, {"int", ".2"}, {"int", "s"}, {"float", "d"}, {"float", ".2"}, {"float", "g"},
		{"float", "%"}, {"str", "d"}, {"str", "+"}, {"str", "05"}, {"str", ","}, {"str", "=5"},
		{"int", "05,d"}, {"int", "."}, {"int", "5.2.1"},
	}
	for _, b := range bad {
		if err := ValidateSpec(b.kind, b.spec); err == nil {
			t.Errorf("%s %q should be rejected", b.kind, b.spec)
		}
	}
}
