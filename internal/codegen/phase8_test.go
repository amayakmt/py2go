package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase8Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"conditional expression is lazy", "x: int = 1\ny: int = 2 if x > 0 else 3\n", []string{"func() int64 {\nif x > 0 {\nreturn 2\n}\nreturn 3\n}()"}, nil},
		{"chained comparison evaluates each operand once", "a: int = 1\nprint(0 < a < 5)\n", []string{"var _py", "if !(", "return false"}, nil},
		{"walrus declares before the statement", "if (n := 3) > 2:\n    print(n)\n", []string{"var n int64\n", "n = 3\nreturn n"}, nil},
		{"optional is an Opt", "from typing import Optional\nx: Optional[int] = None\ny: int | None = 3\n", []string{"runtime.Opt[int64]", "runtime.NoneOf[int64]()", "runtime.Some[int64](3)"}, nil},
		{"is None tests IsSome", "from typing import Optional\nx: Optional[int] = None\nprint(x is None, x is not None)\n", []string{"!x.IsSome()", "x.IsSome()"}, nil},
		{"narrowing reads the value without a new variable", "from typing import Optional\n\n\ndef f(x: Optional[int]) -> int:\n    if x is not None:\n        return x + 1\n    return 0\n", []string{"x.Get() + 1"}, []string{"var x int64"}},
		{"falling off an optional function returns None", "from typing import Optional\n\n\ndef f(x: int) -> Optional[int]:\n    if x > 0:\n        return x\n", []string{"return runtime.NoneOf[int64]()"}, []string{"MissingReturn"}},
		{"generator function makes a coroutine", "from typing import Iterator\n\n\ndef g() -> Iterator[int]:\n    yield 1\n", []string{"runtime.NewGen[int64, runtime.NoneT, runtime.NoneT](func(_pyy *runtime.Yielder[int64, runtime.NoneT]) runtime.NoneT {", "_pyy.Yield(1)", ".Iterator()"}, nil},
		{"generator with send keeps the Gen", "from typing import Generator\n\n\ndef g() -> Generator[int, str, bool]:\n    s: str = yield 1\n    return True\n", []string{"*runtime.Gen[int64, string, bool]", "return true"}, []string{".Iterator()"}},
		{"yield from a generator delegates", "from typing import Generator\n\n\ndef a() -> Generator[int, None, None]:\n    yield 1\n\n\ndef b() -> Generator[int, None, None]:\n    yield from a()\n", []string{"runtime.YieldFromGen(_pyy, a())"}, nil},
		{"a loop over a generator call closes it", "from typing import Iterator\n\n\ndef g() -> Iterator[int]:\n    yield 1\n\n\nfor v in g():\n    print(v)\n", []string{".Close()", "runtime.Catch(func() {"}, nil},
		{"with calls enter and exit", "class C:\n    def __enter__(self) -> int:\n        return 1\n\n    def __exit__(self, a: object, b: object, c: object) -> None:\n        pass\n\n\nwith C() as x:\n    print(x)\n", []string{".M___enter__()", "runtime.ExitCall(", "M___exit__(nil, nil, nil)"}, nil},
		{"match is a chain of closures", "def f(x: int) -> str:\n    match x:\n        case 1 | 2:\n            return 'a'\n        case n:\n            return 'b'\n    return 'c'\n", []string{"matched := false", "if !_py", "func() bool {"}, nil},
		{"loop else uses a flag", "for i in range(3):\n    if i == 1:\n        break\nelse:\n    print('x')\n", []string{"broke bool", "= true\nbreak", "if !_py"}, nil},
		{"star unpacking checks the length", "xs: list[int] = [1, 2]\na, *b = xs\n", []string{"runtime.CheckUnpackStar(", ".Slice(runtime.Bounds{Lo: 1"}, nil},
		{"generic function instances have their own names", "def f[T](x: T) -> T:\n    return x\n\n\nprint(f(1), f('s'))\n", []string{"func f_int(x int64) int64", "func f_str(x string) string"}, nil},
		{"generic class instances have their own names", "class Box[T]:\n    def __init__(self, v: T) -> None:\n        self.v = v\n\n\nb: Box[int] = Box(1)\n", []string{"type Box_int interface", "func _pyNew_Box_int(v int64) Box_int", `PyClassName() string { return "Box" }`}, nil},
		{"except star splits the group", "try:\n    pass\nexcept* ValueError as e:\n    print(e)\n", []string{"runtime.SplitStar(rem, \"ValueError\")"}, nil},
		{"nested format spec is built at run time", "w: int = 5\nprint(f'{w:{w}}')\n", []string{"runtime.FormatIntDyn(w,"}, nil},
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

func TestPhase8Errors(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"generator annotation", "def g() -> int:\n    yield 1\n", "generator function annotated int"},
		{"yield type", "from typing import Iterator\n\n\ndef g() -> Iterator[int]:\n    yield 's'\n", "yielded value expects int, got str"},
		{"with a plain value", "with 1 as x:\n    pass\n", "context managers are classes"},
		{"union", "x: int | str = 1\n", "union of several types"},
		{"inference", "def f[T]() -> T:\n    pass\n\n\nf()\n", "cannot infer type parameter 'T'"},
		{"bound", "def f[T: str](x: T) -> T:\n    return x\n\n\nf(1)\n", "violates the bound"},
		{"two stars in a pattern", "xs: list[int] = []\nmatch xs:\n    case [*a, *b]:\n        pass\n", "two starred names"},
		{"or pattern with a capture", "x: int = 1\nmatch x:\n    case 1 | y:\n        pass\n", "or-pattern that captures names"},
		{"next of a list", "print(next([1]))\n", "not an iterator"},
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

func TestParseAnnotation(t *testing.T) {
	g := &generator{classes: map[string]*classInfo{"Node": {name: "Node"}}, genericClasses: map[string]*genericClass{}, typeVars: map[string]typeVarInfo{}, typeAliases: map[string]*typeAlias{}}
	tests := []struct {
		src  string
		want pyType
	}{
		{"int", tInt},
		{"list[int]", listOf(tInt)},
		{"Node | None", optionalOf("Node")},
		{"dict[str, list[Node]]", dictOf(tStr, listOf("Node"))},
		{"tuple[int, ...]", varTupleOf(tInt)},
		{"Optional[list['Node']]", optionalOf(listOf("Node"))},
		{"Callable[[int, str], bool]", funcOf([]pyType{tInt, tStr}, tBool)},
	}
	for _, tc := range tests {
		node, err := parseAnnotation(tc.src, 1)
		if err != nil {
			t.Fatalf("%s: %v", tc.src, err)
		}
		got, err := g.annotationType(node, false)
		if err != nil || got != tc.want {
			t.Errorf("%s: got %q (%v), want %q", tc.src, got, err, tc.want)
		}
	}
	for _, bad := range []string{"list[", "int int", "[", ""} {
		if _, err := parseAnnotation(bad, 1); err == nil {
			t.Errorf("%q should not parse", bad)
		}
	}
}
