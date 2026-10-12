package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/internal/pyrun"
)

func TestGenerateHello(t *testing.T) {
	data, err := os.ReadFile("../../testdata/phase1/hello/ast.json")
	if err != nil {
		t.Fatal(err)
	}
	root, err := pyast.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Generate(root)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../testdata/phase1/hello/expected.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("output mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestGenerateUnsupported(t *testing.T) {
	data := []byte(`{"_type": "Module", "body": [{"_type": "Match", "lineno": 12}]}`)
	root, err := pyast.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Generate(root)
	if err == nil || !strings.Contains(err.Error(), "unsupported: Match at line 12") {
		t.Fatalf("want \"unsupported: Match at line 12\", got %v", err)
	}
}

// generate transpiles Python source with the real dumper.
func generate(t *testing.T, src string) (string, error) {
	t.Helper()
	if _, err := exec.LookPath("uv"); err != nil {
		t.Skip("uv not installed")
	}
	path := filepath.Join(t.TempDir(), "input.py")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	js, err := pyrun.DumpAST(path)
	if err != nil {
		t.Fatal(err)
	}
	root, err := pyast.Decode(js)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Generate(root)
	return string(out), err
}

func TestGenerateSignature(t *testing.T) {
	got, err := generate(t, "def add(a: int, b: int) -> int:\n    return a + b\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "func add(a int64, b int64) int64 {") {
		t.Errorf("signature not translated:\n%s", got)
	}
	if strings.Contains(got, "runtime") {
		t.Errorf("runtime must not be imported when unused:\n%s", got)
	}
}

func TestGenerateTypedOutput(t *testing.T) {
	tests := []struct {
		name, src, want string
	}{
		{"declaration", "x: int = 5\n", "var x int64 = 5\n\t_ = x\n"},
		{"floor division", "print(-7 // 2)\n", "runtime.Print(runtime.FloorDiv(-7, 2))"},
		{"modulo", "print(-7 % 3)\n", "runtime.Print(runtime.Mod(-7, 3))"},
		{"true division of ints", "print(7 / 2)\n", "runtime.DivInt(7, 2)"},
		{"mixed operands convert", "i: int = 1\nprint(i + 2.5)\n", "float64(i) + 2.5"},
		{"constant floats are typed", "print(0.1 + 0.2)\n", "float64(0.1) + float64(0.2)"},
		{"parentheses only where needed", "a: int = 1\nprint((a + 2) * 3 - (a - 1))\n", "(a+2)*3 - (a - 1)"},
		{"augmented assignment", "x: int = 1\nx += 2\n", "x += 2"},
		{"elif", "x: int = 1\nif x == 1:\n    pass\nelif x == 2:\n    pass\nelse:\n    pass\n", "} else if x == 2 {"},
		{"while True", "while True:\n    break\n", "for {"},
		{"range with literal bounds", "for i in range(3):\n    pass\n", "for _py1_i := int64(0); _py1_i < 3; _py1_i++ {"},
		{"range with negative step", "for i in range(10, 0, -2):\n    pass\n", "_py1_i > 0; _py1_i += -2 {"},
		{"missing return guard", "def f(n: int) -> int:\n    if n > 0:\n        return 1\n", `panic(runtime.MissingReturn("f"))`},
		{"no guard after return", "def f() -> int:\n    return 1\n", "return 1\n}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := generate(t, tc.src)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("output should contain %q\n%s", tc.want, got)
			}
		})
	}
}

func TestGenerateContainerOutput(t *testing.T) {
	tests := []struct {
		name, src, want string
	}{
		{"list literal", "xs: list[int] = [1, 2, 3]\n", "var xs *runtime.List[int64] = runtime.NewList[int64](1, 2, 3)"},
		{"empty list takes its type", "xs: list[str] = []\n", "runtime.NewList[string]()"},
		{"nested list", "g: list[list[int]] = [[1], []]\n", "runtime.NewList[*runtime.List[int64]](runtime.NewList[int64](1), runtime.NewList[int64]())"},
		{"tuple literal", "t: tuple[int, str] = (1, 'a')\n", `runtime.Tup2[int64, string](1, "a")`},
		{"list index", "xs: list[int] = [1]\nprint(xs[-1])\n", "xs.Get(-1)"},
		{"list item assignment", "xs: list[int] = [1]\nxs[0] = 2\n", "xs.Set(0, 2)"},
		{"tuple field", "t: tuple[int, str] = (1, 'a')\nprint(t[1], t[-2])\n", "t.F1, t.F0"},
		{"string index", "s: str = 'ab'\nprint(s[0])\n", `runtime.StrIndex(s, 0)`},
		{"slice", "xs: list[int] = [1]\nprint(xs[::-1])\n", "xs.Slice(runtime.Bounds{Step: -1, HasStep: true})"},
		{"len", "xs: list[int] = [1]\nprint(len(xs), len('s'))\n", `xs.Len(), runtime.StrLen("s")`},
		{"list equality", "a: list[int] = [1]\nprint(a == [1], a != [])\n", "runtime.Eq(a, runtime.NewList[int64](1)), !runtime.Eq(a, runtime.NewList[int64]())"},
		{"list in place", "a: list[int] = [1]\na += [2]\n", "a.Extend(runtime.NewList[int64](2))"},
		{"membership", "a: list[int] = [1]\nprint(1 in a, 'x' in 'xy')\n", `a.Contains(1), runtime.StrContains("xy", "x")`},
		{"swap is a parallel assignment", "a: int = 1\nb: int = 2\na, b = b, a\n", "a, b = b, a"},
		{"unpack declares", "x, y = 1, 'a'\n", "var x int64 = _py1_0"},
		{"unpack a call", "u, v = divmod(7, 2)\n", "runtime.DivModInt(7, 2)"},
		{"list unpack checks length", "xs: list[int] = [1, 2]\na, b = xs\n", "runtime.CheckUnpack(_py1_t.Len(), 2)"},
		{"swap list items", "xs: list[int] = [1, 2]\nxs[0], xs[1] = xs[1], xs[0]\n", "xs.Set(0, _py1_0)"},
		{"for over list", "for x in [1, 2]:\n    pass\n", "_py2_l.Len(); _py1_i++ {"},
		{"for over string", "for c in 'ab':\n    pass\n", `runtime.Chars("ab")`},
		{"enumerate", "for i, x in enumerate([1]):\n    pass\n", "i = _py1_i + 0"},
		{"zip stops at the shortest", "for a, b in zip([1], [2]):\n    pass\n", "_py1_i < _py2_l.Len() && _py1_i < _py3_l.Len()"},
		{"power with literal exponent", "x: int = 3\nprint(x ** 2, x ** -1, 2.0 ** x)\n", "runtime.PowInt(x, 2), runtime.PowFloat(float64(x), float64(-1)), runtime.PowFloat(2.0, float64(x))"},
		{"conversions", "s: str = '1'\nprint(int(s), float(s), str(1), bool(s))\n", `runtime.IntFromStr(s), runtime.FloatFromStr(s), runtime.Str(1), s != ""`},
		{"min with a type argument", "print(min(1, 2), max([1]))\n", "runtime.Min[int64](1, 2), runtime.MaxList(runtime.NewList[int64](1))"},
		{"f-string spec", "x: float = 1.5\nprint(f'{x:.2f} and {x}')\n", `runtime.FormatFloat(x, ".2f") + " and " + runtime.Str(x)`},
		{"f-string int as float", "n: int = 3\nprint(f'{n:.1f}')\n", `runtime.FormatFloat(float64(n), ".1f")`},
		{"f-string conversion", "s: str = 'a'\nprint(f'{s!r:>6}')\n", `runtime.FormatStr(runtime.Repr(s), ">6")`},
		{"print keywords", "print(1, sep='-', end='')\n", `runtime.PrintOpts("-", "", 1)`},
		{"method call", "s: str = 'a b'\nprint(s.split()[0].upper())\n", "runtime.StrUpper(runtime.StrSplit(s).Get(0))"},
		{"list method", "xs: list[int] = []\nxs.append(1)\nxs.sort(reverse=True)\n", "xs.Append(1)\n\txs.Sort(true)"},
		{"value discarded", "s: str = 'a'\nlen(s)\n", `_ = runtime.StrLen(s)`},
		{"container parameters", "def f(xs: list[int], t: tuple[int, str]) -> list[int]:\n    return xs\n", "func f(xs *runtime.List[int64], t runtime.Tuple2[int64, string]) *runtime.List[int64] {"},
		{"container truthiness", "xs: list[int] = [1]\nif xs:\n    pass\n", "if xs.Len() != 0 {"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := generate(t, tc.src)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("output should contain %q\n%s", tc.want, got)
			}
		})
	}
}
