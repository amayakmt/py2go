package runtime

import (
	"math"
	"testing"
)

func TestCorrectlyRoundedFunctions(t *testing.T) {
	// Values that Go's math package gets wrong in the last place (checked
	// against CPython 3.13 on glibc).
	tests := []struct {
		name string
		got  float64
		want float64
	}{
		{"exp(1)", CrExp(1), 2.718281828459045},
		{"exp(-745)", CrExp(-745), 5e-324},
		{"exp(709.7)", CrExp(709.7), 1.6549840276802644e+308},
		{"log(10)", CrLog(10), 2.302585092994046},
		{"log(2**1000)", CrLog(math.Ldexp(1, 1000)), 693.1471805599453},
		{"log2(8)", CrLog2(8), 3},
		{"log10(1000)", glibcLog10(1000), 3},
		{"sin(1e22)", CrSin(1e22), -0.8522008497671888},
		{"cos(1e22)", CrCos(1e22), 0.523214785395139},
		{"tan(1e100)", CrTan(1e100), -0.4116229628832498},
		{"atan(1)", CrAtan(1), math.Pi / 4},
		{"pow(2,0.5)", CrPow(2, 0.5), 1.4142135623730951},
		{"pow(10,-3)", CrPow(10, -3), 0.001},
		{"pow(-2,3)", CrPow(-2, 3), -8},
		{"pow(3,34)", CrPow(3, 34), 16677181699666568},
		{"sinh(1)", glibcSinh(1), 1.1752011936438014},
		{"tanh(0.5)", glibcTanh(0.5), 0.46211715726000974},
		{"cosh(1)", glibcCosh(1), 1.5430806348152437},
		{"hypot", vectorNorm([]float64{3, 4}, 4), 5},
		{"cbrt(27)", glibcCbrt(27), 3.0000000000000004},
		{"cbrt(-8)", glibcCbrt(-8), -2},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
	if !math.IsNaN(CrLog(math.NaN())) || !math.IsInf(CrExp(1000), 1) || CrExp(math.Inf(-1)) != 0 {
		t.Error("special values")
	}
}

func TestGammaAndLgamma(t *testing.T) {
	for _, tc := range []struct{ x, want float64 }{{5, 24}, {0.5, 1.7724538509055159}, {-1.5, 2.3632718012073544}, {171, 7.257415615307998e+306}} {
		got, e := pyGamma(tc.x)
		if e != 0 || got != tc.want {
			t.Errorf("gamma(%v) = %v (%d), want %v", tc.x, got, e, tc.want)
		}
	}
	if _, e := pyGamma(0); e != errDomain {
		t.Error("gamma(0)")
	}
	if _, e := pyGamma(172); e != errRange {
		t.Error("gamma(172)")
	}
	if got, _ := pyLgamma(100); got != 359.1342053695754 {
		t.Errorf("lgamma(100) = %v", got)
	}
}

func TestStructPackUnpack(t *testing.T) {
	sf := parseStructFmt("<hhl")
	b := sf.pack([]any{int64(1), int64(-2), int64(3)})
	if string(b) != "\x01\x00\xfe\xff\x03\x00\x00\x00" {
		t.Errorf("pack = %q", b)
	}
	got := sf.unpack(string(b))
	if got.Len() != 3 || got.At(1) != int64(-2) {
		t.Errorf("unpack = %v", got)
	}
	if parseStructFmt("@bhiq").size != 16 || parseStructFmt("<bhiq").size != 15 {
		t.Error("alignment")
	}
	if got := catch(func() { parseStructFmt("<hh").pack([]any{int64(1)}) }); got != "struct.error: pack expected 2 items for packing (got 1)" {
		t.Errorf("got %s", got)
	}
	if got := catch(func() { parseStructFmt("<b").pack([]any{int64(200)}) }); got != "struct.error: 'b' format requires -128 <= number <= 127" {
		t.Errorf("got %s", got)
	}
	if float16from(float16bits(3.14159)) != 3.140625 {
		t.Error("half floats")
	}
}

func TestUnicodeDataNormalization(t *testing.T) {
	d := loadUD()
	for _, tc := range []struct{ form, in, want string }{
		{"NFC", "é", "é"},
		{"NFD", "é", "é"},
		{"NFKC", "ﬁ", "fi"},
		{"NFKD", "①", "1"},
		{"NFC", "한", "한"},
		{"NFD", "한", "한"},
		{"NFD", "ạ̇", "ạ̇"},
	} {
		if got := d.normalize(tc.form, tc.in); got != tc.want {
			t.Errorf("%s(%q) = %q, want %q", tc.form, tc.in, got, tc.want)
		}
	}
	if n, ok := d.name('a'); !ok || n != "LATIN SMALL LETTER A" {
		t.Errorf("name(a) = %q", n)
	}
	if r, ok := d.lookup("GREEK SMALL LETTER ALPHA"); !ok || r != 0x3b1 {
		t.Error("lookup")
	}
	if n, _ := d.name(0x4e2d); n != "CJK UNIFIED IDEOGRAPH-4E2D" {
		t.Errorf("name(CJK) = %q", n)
	}
}

func TestSameClassesAreHashable(t *testing.T) {
	a, b := BuiltinClassOf("int"), BuiltinClassOf("int")
	if a != b || Hash(a) != Hash(b) {
		t.Error("builtin class refs are shared and hashable")
	}
	if got := Repr(ClassOf("Zed")); got != "<class '__main__.Zed'>" {
		t.Errorf("repr = %s", got)
	}
	ClassModule("Zed2", "pkg.mod")
	if got := Repr(ClassOf("Zed2")); got != "<class 'pkg.mod.Zed2'>" {
		t.Errorf("repr = %s", got)
	}
}

func TestDivmodOfUserClasses(t *testing.T) {
	if r := DynDivmod(int64(7), int64(2)); r.At(0) != int64(3) || r.At(1) != int64(1) {
		t.Errorf("divmod = %v", r)
	}
}
