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
	data := []byte(`{"_type": "Module", "body": [{"_type": "Await", "lineno": 12}]}`)
	root, err := pyast.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Generate(root)
	if err == nil || !strings.Contains(err.Error(), "unsupported: Await at line 12") {
		t.Fatalf("want \"unsupported: Await at line 12\", got %v", err)
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
		{"module variable", "x: int = 5\n", "var x int64\n"},
		{"local variable", "def f() -> None:\n    x: int = 5\n", "var x int64 = 5\n\t_ = x\n"},
		{"floor division", "print(-7 // 2)\n", "runtime.Print(runtime.FloorDiv(-7, 2))"},
		{"modulo", "print(-7 % 3)\n", "runtime.Print(runtime.Mod(-7, 3))"},
		{"true division of ints", "print(7 / 2)\n", "runtime.DivInt(7, 2)"},
		{"mixed operands convert", "i: int = 1\nprint(i + 2.5)\n", "float64(i) + 2.5"},
		{"constant floats are typed", "print(0.1 + 0.2)\n", "float64(0.1) + float64(0.2)"},
		{"parentheses only where needed", "a: int = 1\nprint((a + 2) * 3 - (a - 1))\n", "MulInt(a+2, 3) - (a - 1)"},
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
		{"list literal", "xs: list[int] = [1, 2, 3]\n", "var xs *runtime.List[int64]\n"},
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
		{"unpack declares", "x, y = 1, 'a'\n", "var x int64\n"},
		{"unpack a call", "u, v = divmod(7, 2)\n", "runtime.DivModInt(7, 2)"},
		{"list unpack checks length", "xs: list[int] = [1, 2]\na, b = xs\n", "runtime.CheckUnpack(_py1_t.Len(), 2)"},
		{"swap list items", "xs: list[int] = [1, 2]\nxs[0], xs[1] = xs[1], xs[0]\n", "xs.Set(0, _py1_0)"},
		{"for over list", "for x in [1, 2]:\n    pass\n", "var _py1_c runtime.Cursor[int64] = runtime.NewList[int64](1, 2).Iter()\n\tfor _py1_c.Next() {"},
		{"for over string", "for c in 'ab':\n    pass\n", `runtime.Chars("ab")`},
		{"enumerate", "for i, x in enumerate([1]):\n    pass\n", "runtime.Enumerate[int64](runtime.NewList[int64](1), 0)"},
		{"zip", "for a, b in zip([1], [2]):\n    pass\n", "runtime.Zip2[int64, int64](runtime.NewList[int64](1), runtime.NewList[int64](2))"},
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

func TestGenerateExceptionOutput(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"raise with arguments", "raise ValueError('x', 2)\n", []string{`panic(runtime.NewError("ValueError", "x", 2))`}, nil},
		{"raise a class", "raise KeyError\n", []string{`panic(runtime.NewError("KeyError"))`}, nil},
		{"assert", "x: int = 1\nassert x < 3, 'big'\n", []string{"if !(x < 3) {", `panic(runtime.NewError("AssertionError", "big"))`}, nil},
		{"uncaught exceptions are reported", "print(1)\n", []string{"defer runtime.Report()"}, nil},
		{"no runtime, no report", "x: int = 1\n", nil, []string{"Report", "import"}},
		{"try body is a closure", "try:\n    print(1)\nexcept ValueError:\n    pass\n", []string{"runtime.Catch(func() {", `runtime.Match(_py1_exc, "ValueError")`, "panic(_py1_exc)"}, nil},
		{"handler binds a name", "try:\n    pass\nexcept (KeyError, IndexError) as e:\n    print(e, type(e).__name__)\n", []string{`runtime.Match(_py1_exc, "KeyError", "IndexError")`, "var e *runtime.PyError = _py1_exc", "runtime.Print(e, e.Type)"}, nil},
		{"bare raise re-raises the caught exception", "try:\n    pass\nexcept ValueError:\n    raise\n", []string{"\"ValueError\") {\npanic(_py1_exc)"}, nil},
		{"bare except has no match test", "try:\n    pass\nexcept:\n    pass\n", []string{"if _py1_exc != nil {\n{"}, []string{"Match"}},
		{"finally runs after the closure", "try:\n    print(1)\nfinally:\n    print(2)\n", []string{"_py2_pending := runtime.Catch(func() {", "if _py2_pending != nil {\npanic(_py2_pending)"}, nil},
		{"break leaves the closure by status", "for i in range(2):\n    try:\n        break\n    finally:\n        pass\n", []string{"_py2_flow = 1\nreturn", "if _py2_flow == 1 {\nbreak"}, nil},
		{"return leaves the closure by status", "def f() -> int:\n    try:\n        return 1\n    finally:\n        pass\n", []string{"_py2_ret = 1", "_py1_flow = 3\nreturn", "return _py2_ret"}, nil},
		{"else waits for a clean body", "for i in range(2):\n    try:\n        continue\n    except ValueError:\n        pass\n    else:\n        print(i)\n", []string{"} else if _py3_flow == 0 {"}, nil},
		{"try declarations are hoisted", "try:\n    n: int = int('1')\nexcept ValueError:\n    n = 0\nprint(n)\n", []string{"var n int64\n", "n = runtime.IntFromStr(\"1\")"}, []string{"var n int64 = "}},
		{"recursive functions count depth", "def f(n: int) -> int:\n    return f(n)\n", []string{"runtime.Enter()\ndefer runtime.Leave()"}, nil},
		{"mutually recursive functions count depth", "def a() -> None:\n    b()\n\n\ndef b() -> None:\n    a()\n", []string{"func a() {\nruntime.Enter()", "func b() {\nruntime.Enter()"}, nil},
		{"other functions do not", "def g(n: int) -> int:\n    return n\n\n\ndef f(n: int) -> int:\n    return g(n)\n", nil, []string{"Enter"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := generate(t, tc.src)
			if err != nil {
				t.Fatal(err)
			}
			got = strings.ReplaceAll(got, "\t", "") // compare without gofmt's indentation
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

func TestGenerateDictSetOutput(t *testing.T) {
	tests := []struct {
		name, src, want string
	}{
		{"dict literal", "d: dict[str, int] = {'a': 1, 'b': 2}\n", `d = runtime.NewDict[string, int64](runtime.KV[string, int64]{Key: "a", Val: 1}, runtime.KV[string, int64]{Key: "b", Val: 2})`},
		{"empty dict takes its type", "d: dict[int, list[str]] = {}\n", "runtime.NewDict[int64, *runtime.List[string]]()"},
		{"dict()", "d: dict[int, int] = dict()\n", "d = runtime.NewDict[int64, int64]()"},
		{"dict item", "d: dict[str, int] = {}\nd['a'] = 1\nprint(d['a'])\n", `d.Set("a", 1)`},
		{"dict augmented item", "d: dict[str, int] = {}\nd['a'] += 1\n", "d.Set(_py1_i, d.Get(_py1_i)+1)"},
		{"del item", "d: dict[str, int] = {}\ndel d['a']\n", `d.Delete("a")`},
		{"dict get", "d: dict[str, int] = {}\nprint(d.get('a', 0))\n", `d.GetDefault("a", 0)`},
		{"dict views", "d: dict[str, int] = {}\nprint(d.keys(), d.values(), d.items())\n", "d.Keys(), d.Values(), d.Items()"},
		{"dict loop", "d: dict[str, int] = {}\nfor k, v in d.items():\n    print(k, v)\n", "runtime.Cursor[runtime.Tuple2[string, int64]] = d.Items().Iter()"},
		{"dict membership", "d: dict[str, int] = {}\nprint('a' in d, 'b' not in d)\n", `d.Has("a"), !d.Has("b")`},
		{"dict merge", "a: dict[int, int] = {}\nb: dict[int, int] = {}\nprint(a | b)\na |= b\n", "a.Union(b)"},
		{"dict in place merge", "a: dict[int, int] = {}\nb: dict[int, int] = {}\na |= b\n", "a.Update(b)"},
		{"set literal", "s: set[int] = {3, 1}\n", "s = runtime.NewSet[int64](3, 1)"},
		{"set operators", "a: set[int] = {1}\nb: set[int] = {2}\nprint(a | b, a & b, a - b, a ^ b)\n", "a.Union(b), a.Intersection(b), a.Difference(b), a.SymmetricDifference(b)"},
		{"set in place", "a: set[int] = {1}\nb: set[int] = {2}\na |= b\na &= b\na -= b\na ^= b\n", "a.Update(b)\n\ta.IntersectionUpdate(b)\n\ta.DifferenceUpdate(b)\n\ta.SymmetricDifferenceUpdate(b)"},
		{"set subset", "a: set[int] = {1}\nb: set[int] = {2}\nprint(a <= b, a < b)\n", "runtime.Le(a, b), runtime.Lt(a, b)"},
		{"set methods", "a: set[int] = set()\na.add(1)\nprint(a.union([2], {3}), a.issubset([1]))\n", "a.Union(runtime.NewList[int64](2), runtime.NewSet[int64](3)), a.IsSubset(runtime.NewList[int64](1))"},
		{"set from list", "print(set([1, 2]))\n", "runtime.SetFrom[int64](runtime.NewList[int64](1, 2))"},
		{"range value", "r: range = range(2, 9, 3)\nprint(len(r), r[0], 4 in r)\n", "runtime.NewRange(2, 9, 3)"},
		{"list of range", "print(list(range(3)))\n", "runtime.ListFrom[int64](runtime.NewRange(0, 3, 1))"},
		{"list comprehension", "n: int = 3\nprint([x * 2 for x in range(n) if x > 0])\n", "func() *runtime.List[int64] {"},
		{"comprehension variables are local", "x: int = 7\nprint([x for x in range(2)], x)\n", "var x int64\n"},
		{"dict comprehension", "print({x: str(x) for x in range(2)})\n", "runtime.NewDict[int64, string]()"},
		{"set comprehension", "print({x % 2 for x in range(4)})\n", "runtime.NewSet[int64]()"},
		{"generator expression is lazy", "print(any(x > 1 for x in range(3)))\n", "runtime.NewIterator(func() (bool, bool) {"},
		{"any on a generator", "print(any(x > 1 for x in range(3)))\n", "runtime.Any[bool](func() *runtime.Iterator[bool]"},
		{"sum of a generator", "print(sum(x for x in range(3)))\n", "runtime.SumInt(runtime.ListFrom[int64]("},
		{"sorted dict", "d: dict[str, int] = {}\nprint(sorted(d))\n", "runtime.Sorted(runtime.ListFrom[string](d), false)"},
		{"reversed", "print(list(reversed([1, 2])))\n", "runtime.Reversed[int64](runtime.NewList[int64](1, 2))"},
		{"dict.fromkeys", "print(dict.fromkeys([1, 2], 0))\n", "runtime.FromKeys[int64, int64](runtime.NewList[int64](1, 2), 0)"},
		{"dict from pairs", "print(dict(zip([1], ['a'])))\n", "runtime.DictFrom[int64, string](runtime.Zip2[int64, string]("},
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

func TestGenerateGlobals(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"module variables are package variables", "LIMIT: int = 3\n\n\ndef f() -> int:\n    return LIMIT\n",
			[]string{"var LIMIT int64\n", "var _pyg_LIMIT bool\n", `runtime.Bound(LIMIT, _pyg_LIMIT, "LIMIT")`, "LIMIT = 3\n\t_pyg_LIMIT = true"}, nil},
		{"unread variables get no flag", "X: int = 1\nprint(X)\n", []string{"var X int64\n"}, []string{"_pyg_"}},
		{"global assignment", "n: int = 0\n\n\ndef bump() -> None:\n    global n\n    n += 1\n", []string{"n += 1", "_pyg_n = true"}, nil},
		{"names declared in a block are module variables", "if True:\n    y: int = 1\nprint(y)\n", []string{"var y int64\n"}, nil},
		{"loop variables at module level", "for i in range(2):\n    pass\n", []string{"var i int64\n"}, nil},
		{"a local hides the global", "x: int = 1\n\n\ndef f() -> int:\n    x: int = 2\n    return x\n", []string{"var x int64 = 2"}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := generate(t, tc.src)
			if err != nil {
				t.Fatal(err)
			}
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
