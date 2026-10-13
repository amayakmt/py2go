package runtime

import (
	"math/big"
	"strings"
	"testing"
)

func bigLit(s string) any { return BigLit(s) }

func TestBigIntArithmetic(t *testing.T) {
	two70 := bigLit("1180591620717411303424")
	tests := []struct {
		name string
		got  any
		want string
	}{
		{"add overflow promotes", DynBinary("Add", int64(1<<62), int64(1<<62)), "9223372036854775808"},
		{"sub underflow promotes", DynBinary("Sub", int64(-1<<63), int64(1)), "-9223372036854775809"},
		{"mult promotes", DynBinary("Mult", int64(1<<40), int64(1<<40)), "1208925819614629174706176"},
		{"result that fits is a machine int", DynBinary("Sub", two70, bigLit("1180591620717411303423")), "1"},
		{"floor division rounds down", DynBinary("FloorDiv", DynUnary("USub", two70), int64(3)), "-393530540239137101142"},
		{"modulo takes the sign of the divisor", DynBinary("Mod", DynUnary("USub", two70), int64(3)), "2"},
		{"pow", DynBinary("Pow", int64(3), int64(50)), "717897987691852588770249"},
		{"left shift", DynBinary("LShift", int64(1), int64(100)), "1267650600228229401496703205376"},
		{"right shift of a negative rounds down", DynBinary("RShift", DynUnary("USub", two70), int64(69)), "-2"},
		{"and with a negative", DynBinary("BitAnd", DynUnary("USub", two70), int64(255)), "0"},
		{"invert", DynUnary("Invert", two70), "-1180591620717411303425"},
	}
	for _, tc := range tests {
		if got := Repr(tc.got); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestBigIntErrors(t *testing.T) {
	two70 := bigLit("1180591620717411303424")
	tests := []struct {
		name string
		f    func()
		want string
	}{
		{"division by zero", func() { DynBinary("FloorDiv", two70, int64(0)) }, "ZeroDivisionError: integer division or modulo by zero"},
		{"true division by zero", func() { DynBinary("Div", two70, int64(0)) }, "ZeroDivisionError: division by zero"},
		{"negative shift", func() { DynBinary("LShift", int64(1), int64(-1)) }, "ValueError: negative shift count"},
		{"huge shift count", func() { DynBinary("LShift", int64(1), bigLit("1000000000000000000000000000000")) }, "OverflowError: too many digits in integer"},
		{"int to float overflow", func() { DynFloat(bigLit("1" + strings.Repeat("0", 400))) }, "OverflowError: int too large to convert to float"},
		{"digit limit", func() { Str(BigLit("1" + strings.Repeat("0", 5000))) }, "ValueError: Exceeds the limit (4300 digits) for integer string conversion; use sys.set_int_max_str_digits() to increase the limit"},
	}
	for _, tc := range tests {
		if got := catch(tc.f); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBigIntHashMatchesCPython(t *testing.T) {
	for _, tc := range []struct {
		v    any
		want int64
	}{
		{int64(1<<61 - 1), 0},
		{int64(1 << 61), 1},
		{bigLit("1180591620717411303424"), 512},
		{DynUnary("USub", bigLit("1180591620717411303424")), -512},
		{int64(-1), -2},
	} {
		if got := Hash(tc.v); got != tc.want {
			t.Errorf("hash(%s) = %d, want %d", Repr(tc.v), got, tc.want)
		}
	}
	if Hash(bigLit("1180591620717411303424")) != Hash(float64(1<<62)*float64(1<<8)) {
		t.Error("an int and the float equal to it must hash alike")
	}
}

func TestExactIntFloatComparison(t *testing.T) {
	if !Lt(int64(1<<53), bigLit("9007199254740993")) {
		t.Error("2**53 < 2**53 + 1")
	}
	if Eq(bigLit("9007199254740993"), float64(1<<53)) {
		t.Error("2**53 + 1 must not equal 2.0**53")
	}
	if !Eq(bigLit("1180591620717411303424"), float64(1<<62)*float64(1<<8)) {
		t.Error("2**70 == 2.0**70")
	}
	if !Gt(bigLit("1"+strings.Repeat("0", 400)), float64(1e308)) {
		t.Error("a huge int is greater than any finite float")
	}
}

func TestBigTrueDivision(t *testing.T) {
	a := bigLit("1" + strings.Repeat("0", 400))
	b := bigLit("1" + strings.Repeat("0", 399))
	if got := Repr(DynBinary("Div", a, b)); got != "10.0" {
		t.Errorf("10**400 / 10**399 = %s, want 10.0", got)
	}
	if got := Repr(DynBinary("Div", int64(1), int64(3))); got != "0.3333333333333333" {
		t.Errorf("1/3 = %s", got)
	}
}

func TestIntFromTextBases(t *testing.T) {
	for _, tc := range []struct {
		s    string
		base any
		want string
	}{
		{"ff", int64(16), "255"},
		{"0xFF", int64(0), "255"},
		{"-0b101", int64(0), "-5"},
		{"0x_ff", int64(0), "255"},
		{"zz", int64(36), "1295"},
		{strings.Repeat("f", 30), int64(16), "1329227995784915872903807060280344575"},
	} {
		if got := Repr(DynIntBase(tc.s, tc.base)); got != tc.want {
			t.Errorf("int(%q, %v) = %s, want %s", tc.s, tc.base, got, tc.want)
		}
	}
	if got := catch(func() { DynIntBase("12", int64(2)) }); got != "ValueError: invalid literal for int() with base 2: '12'" {
		t.Errorf("got %q", got)
	}
	if got := Repr(DynIntAny("٣٤")); got != "34" {
		t.Errorf("int of Arabic-Indic digits = %s", got)
	}
}

func TestBytesBasics(t *testing.T) {
	b := Bytes("hello\x00\xff'")
	if got := Repr(b); got != `b"hello\x00\xff'"` {
		t.Errorf("repr = %s", got)
	}
	if got := Repr(DynCallMethod(b, "hex", []any{"-"}, nil)); got != "'68-65-6c-6c-6f-00-ff-27'" {
		t.Errorf("hex = %s", got)
	}
	if got := Repr(DynCallMethod(Bytes("a,b,,c"), "split", []any{Bytes(",")}, nil)); got != "[b'a', b'b', b'', b'c']" {
		t.Errorf("split = %s", got)
	}
	if got := catch(func() { DynBinary("Add", Bytes("a"), "b") }); got != `TypeError: can't concat str to bytes` {
		t.Errorf("concat: %q", got)
	}
	if got := catch(func() { NewBytes([]any{NewList[int64](256)}, nil) }); got == "" {
		t.Error("bytes(256) must not be silent")
	}
	ba := NewByteArray([]any{Bytes("abc")}, nil)
	DynCallMethod(ba, "append", []any{int64(100)}, nil)
	DynCallMethod(ba, "extend", []any{Bytes("ef")}, nil)
	if got := Repr(ba); got != "bytearray(b'abcdef')" {
		t.Errorf("bytearray = %s", got)
	}
	if got := Repr(DynCallMethod(ba, "pop", nil, nil)); got != "102" {
		t.Errorf("pop = %s", got)
	}
}

func TestBytesFormatting(t *testing.T) {
	for _, tc := range []struct {
		format string
		args   any
		want   string
	}{
		{"%d items", int64(5), "5 items"},
		{"%s-%s", NewVarTuple[any](Bytes("a"), Bytes("b")), "a-b"},
		{"%5d|%-5d|", NewVarTuple[any](int64(1), int64(2)), "    1|2    |"},
		{"%r", Bytes("x"), "b'x'"},
		{"%c%c", NewVarTuple[any](int64(65), Bytes("z")), "Az"},
	} {
		if got := string(bytesToTextless(bytesPercent(tc.format, tc.args))); got != tc.want {
			t.Errorf("%q %% ... = %q, want %q", tc.format, got, tc.want)
		}
	}
	if got := catch(func() { bytesPercent("%s", "str") }); !strings.HasPrefix(got, "TypeError: %b requires a bytes-like object") {
		t.Errorf("got %q", got)
	}
}

func bytesToTextless(s string) string { return s }

func TestCodecs(t *testing.T) {
	if got := string(EncodeStr("é€😀", "utf-16-le", "strict")); got != "\xe9\x00\xac \x3d\xd8\x00\xde" {
		t.Errorf("utf-16-le = %q", got)
	}
	if got := DecodeBytes("\xff\xfeA\x00", "utf-16", "strict"); got != "A" {
		t.Errorf("utf-16 with BOM = %q", got)
	}
	if got := string(EncodeStr("é€", "ascii", "backslashreplace")); got != `\xe9\u20ac` {
		t.Errorf("backslashreplace = %q", got)
	}
	if got := string(EncodeStr("é€", "ascii", "xmlcharrefreplace")); got != "&#233;&#8364;" {
		t.Errorf("xmlcharrefreplace = %q", got)
	}
	if got := DecodeBytes("a\xffb", "utf-8", "replace"); got != "a�b" {
		t.Errorf("replace = %q", got)
	}
	for _, tc := range []struct {
		f    func()
		want string
	}{
		{func() { DecodeBytes("caf\xe9", "utf-8", "strict") }, "UnicodeDecodeError: 'utf-8' codec can't decode byte 0xe9 in position 3: unexpected end of data"},
		{func() { DecodeBytes("\xe2\x82", "utf-8", "strict") }, "UnicodeDecodeError: 'utf-8' codec can't decode bytes in position 0-1: unexpected end of data"},
		{func() { DecodeBytes("\xff", "ascii", "strict") }, "UnicodeDecodeError: 'ascii' codec can't decode byte 0xff in position 0: ordinal not in range(128)"},
		{func() { EncodeStr("é", "ascii", "strict") }, "UnicodeEncodeError: 'ascii' codec can't encode character '\\xe9' in position 0: ordinal not in range(128)"},
		{func() { EncodeStr("x", "no-such-codec", "strict") }, "LookupError: unknown encoding: no-such-codec"},
		{func() { DecodeBytes("\x00\xd8", "utf-16-le", "strict") }, "UnicodeDecodeError: 'utf-16-le' codec can't decode bytes in position 0-1: unexpected end of data"},
		{func() { DecodeBytes("\\x4", "unicode-escape", "strict") }, "UnicodeDecodeError: 'unicodeescape' codec can't decode bytes in position 0-2: truncated \\xXX escape"},
	} {
		if got := catch(tc.f); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestFullCaseMapping(t *testing.T) {
	for _, tc := range []struct{ name, got, want string }{
		{"upper sharp s", FullUpper("Straße"), "STRASSE"},
		{"lower dotted I", FullLower("İ"), "i̇"},
		{"final sigma", FullLower("ΑΣ"), "ας"},
		{"inner sigma", FullLower("ΑΣΑ"), "ασα"},
		{"casefold", FullCasefold("Straße"), "strasse"},
		{"title digraph", FullTitle("ǆungla"), "ǅungla"},
		{"title apostrophe", FullTitle("they're"), "They'Re"},
		{"swapcase", FullSwapcase("Ǆß"), "ǆSS"},
		{"capitalize", FullCapitalize("ﬁnal"), "Final"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
	if !StrIsLower("ª") || StrIsUpper("ª") || !StrIsTitle("Ab Cd") || StrIsTitle("AB") {
		t.Error("case predicates")
	}
}

func TestFrozenSet(t *testing.T) {
	f := NewFrozenSet([]any{NewList[int64](1, 2, 3, 2)})
	if f.Len() != 3 || Repr(NewFrozenSet(nil)) != "frozenset()" {
		t.Errorf("len = %d", f.Len())
	}
	g := NewFrozenSet([]any{NewList[int64](3, 2, 1)})
	if !Eq(f, g) || Hash(f) != Hash(g) {
		t.Error("equal frozensets must hash alike")
	}
	u, _ := f.PyBinop("BitOr", NewFrozenSet([]any{NewList[int64](9)}), false)
	if Repr(u) != "frozenset({1, 2, 3, 9})" {
		t.Errorf("union = %s", Repr(u))
	}
	d := NewDict[any, any]()
	d.Set(f, "a")
	d.Set(g, "b")
	if d.Len() != 1 {
		t.Errorf("equal frozensets are one key, got %d", d.Len())
	}
}

func TestSliceAndRange(t *testing.T) {
	if got := Repr(Slice{Lo: int64(1), Hi: nil, Step: int64(-2)}.indicesFor(10)); got != "(1, -1, -2)" {
		t.Errorf("indices = %s", got)
	}
	if got := catch(func() { Slice{Step: int64(0)}.indicesFor(5) }); got != "ValueError: slice step cannot be zero" {
		t.Errorf("got %q", got)
	}
	r := rangeFrom([]any{bigLit("100000000000000000000"), bigLit("100000000000000000003")})
	if Repr(r) != "range(100000000000000000000, 100000000000000000003)" {
		t.Errorf("repr = %s", Repr(r))
	}
	if n := DynLen(r); n != 3 {
		t.Errorf("len = %d", n)
	}
	if !DynContains(r, bigLit("100000000000000000001")) || DynContains(r, bigLit("100000000000000000003")) {
		t.Error("membership")
	}
	if got := Repr(DynGetItem(r, int64(-1))); got != "100000000000000000002" {
		t.Errorf("r[-1] = %s", got)
	}
}

func TestComplexFormatting(t *testing.T) {
	for _, tc := range []struct {
		z    complex128
		spec string
		want string
	}{
		{3 + 4i, "", "(3+4j)"},
		{3i, "", "3j"},
		{3 + 4i, ".2f", "3.00+4.00j"},
		{3i, ">6", "    3j"},
		{1.5 - 2i, "g", "1.5-2j"},
	} {
		if got := FormatDyn(tc.z, tc.spec); got != tc.want {
			t.Errorf("format(%v, %q) = %q, want %q", tc.z, tc.spec, got, tc.want)
		}
	}
	if got := Repr(DynBinary("Div", complex(1, 2), complex(3, -4))); got != "(-0.2+0.4j)" {
		t.Errorf("division = %s", got)
	}
	if got := catch(func() { DynBinary("Div", complex(1, 0), complex(0, 0)) }); got != "ZeroDivisionError: complex division by zero" {
		t.Errorf("got %q", got)
	}
}

func TestFloatStringFormats(t *testing.T) {
	for _, tc := range []struct {
		f    float64
		spec string
		want string
	}{
		{1e16, "g", "1e+16"},
		{123456.0, "g", "123456"},
		{1234567.0, "g", "1.23457e+06"},
		{2.5, ".0f", "2"},
		{3.5, ".0f", "4"},
		{0.25, ".1%", "25.0%"},
		{1234567.891, ",.2f", "1,234,567.89"},
	} {
		if got := FormatFloat(tc.f, tc.spec); got != tc.want {
			t.Errorf("format(%v, %q) = %q, want %q", tc.f, tc.spec, got, tc.want)
		}
	}
	if got := PercentFormat("%g|%5.1f|%e", NewVarTuple[any](float64(3.14159265358979), float64Inf(), 12345.678)); got != "3.14159|  inf|1.234568e+04" {
		t.Errorf("percent format = %q", got)
	}
}

func float64Inf() float64 { f, _ := new(big.Float).SetInf(false).Float64(); return f }

func TestIdentity(t *testing.T) {
	l := NewList[int64](1)
	if Id(l) != Id(l) || Id(l) == Id(NewList[int64](1)) {
		t.Error("id of a list is its address")
	}
	if Id(int64(7)) != Id(int64(7)) || Id(int64(7)) == Id(int64(8)) || Id("a") != Id("a") {
		t.Error("id of a value depends on the value")
	}
	nan := float64Inf() - float64Inf()
	if !Is(nan, nan) {
		t.Error("a nan is itself")
	}
}
