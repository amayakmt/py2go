package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase10Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"unannotated int arithmetic can grow", "def f(a, b):\n    return a * b + a ** b\n", []string{`runtime.DynBinary("Mult"`, `runtime.DynBinary("Pow"`}, []string{"runtime.MulInt"}},
		{"annotated int arithmetic is checked", "def f(a: int, b: int) -> int:\n    return a * b\n", []string{"runtime.MulInt(a, b)"}, []string{"DynBinary"}},
		{"annotated power is narrowed back to an int", "def f(a: int, b: int) -> int:\n    return a ** b\n", []string{`runtime.Cast[int64](runtime.DynBinary("Pow", a, b))`}, nil},
		{"annotated shift is checked", "def f(a: int, b: int) -> int:\n    return a << b\n", []string{"runtime.LShift(a, b)"}, nil},
		{"a literal beyond 64 bits is a big int", "x = 1000000000000000000000000000000\nprint(x)\n", []string{`runtime.BigLit("1000000000000000000000000000000")`}, nil},
		{"int of a string can be any size", "def f(s):\n    return int(s)\n", []string{"runtime.DynIntAny(s)"}, nil},
		{"range of objects is an object", "def f(n):\n    return range(n, n + 2)\n", []string{"runtime.DynRange("}, nil},
		{"id is a built-in", "x = [1]\nprint(id(x) == id(x))\n", []string{"runtime.Id("}, nil},
		{"hash of an int", "print(hash(5))\n", []string{"runtime.Hash("}, nil},
		{"a str method the table lacks gives a typed result", "def f(s: str) -> str:\n    return s.capitalize()\n", []string{"runtime.Cast[string](runtime.DynCallMethod(s, \"capitalize\""}, nil},
		{"frozenset annotation is an object", "def f(x: frozenset[int]) -> frozenset[int]:\n    return x\n", []string{"func f(x any) any"}, nil},
		{"a constant float overflow is a run-time value", "print(1e308 * 10)\n", []string{"func() float64 { return float64(10) }()"}, nil},
		{"a float constant sum stays a constant", "print(0.1 + 0.2)\n", []string{"float64(0.1) + float64(0.2)"}, []string{"func() float64"}},
		{"a hash of any size is reduced", "class A:\n    def __hash__(self):\n        return 2 ** 70\n", []string{"runtime.HashResult("}, nil},
		{"a big index slices through the runtime", "xs = [1, 2, 3]\nprint(xs[10 ** 30:])\n", []string{"runtime.DynGetItem("}, nil},
		{"a class with its own __hash__ is a dict key as an object", "class A:\n    def __eq__(self, o):\n        return True\n    def __hash__(self):\n        return 1\n\n\nd = {A(): 1}\n", []string{"runtime.NewDict[any, int64]"}, nil},
		{"complex literal", "z = 3 + 4j\nprint(z)\n", []string{"complex128(complex(0.0, 4.0))"}, nil},
		{"bytes literal", "b = b'ab'\nprint(b)\n", []string{`runtime.Bytes("ab")`}, nil},
		{"ellipsis literal", "print(...)\n", []string{"runtime.Ellipsis"}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := generate(t, tc.src)
			if err != nil {
				t.Fatal(err)
			}
			got = strings.ReplaceAll(got, "\t", "")
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Errorf("output should contain %q\n%s", w, got)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(got, w) {
					t.Errorf("output should not contain %q\n%s", w, got)
				}
			}
		})
	}
}

func TestPhase10Errors(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"the operand of ~ must be an int", "print(~1.5)\n", "operator ~ on float"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := generate(t, tc.src)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want an error containing %q", err, tc.want)
			}
		})
	}
}
