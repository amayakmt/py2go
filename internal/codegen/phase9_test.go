package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase9Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"unannotated parameters and result are objects", "def add(a, b):\n    return a + b\n", []string{"func add(a any, b any) any", `runtime.DynBinary("Add", a, b)`}, nil},
		{"a function that returns nothing returns nothing", "def f(a):\n    print(a)\n", []string{"func f(a any) {"}, []string{"func f(a any) any"}},
		{"an unannotated variable takes the type of its value", "x = 1\ny = 'a'\nprint(x + 1, y + 'b')\n", []string{"var x int64", "var y string"}, []string{"DynBinary"}},
		{"a variable assigned two types is an object", "x = 1\nx = 'a'\n", []string{"var x any"}, nil},
		{"None and a value join to an Optional", "def f(a):\n    x = None\n    if a:\n        x = 1\n    return x\n", []string{"runtime.Opt[int64]"}, nil},
		{"an empty list holds objects", "xs = []\nxs.append(1)\nxs.append('a')\n", []string{"runtime.NewList[any]()"}, nil},
		{"appending another type widens the list", "xs = [1, 2]\nxs.append('a')\n", []string{"var xs *runtime.List[any]"}, []string{"var xs *runtime.List[int64]"}},
		{"a list that is returned holds objects", "def make():\n    return [[0] * 2]\n", []string{"runtime.DynDeep("}, nil},
		{"a method of an object is called by name", "def f(x):\n    return x.upper()\n", []string{`runtime.DynCallMethod(x, "upper"`}, nil},
		{"an attribute of an object is read by name", "def f(x):\n    return x.name\n", []string{`runtime.DynGetAttr(x, "name")`}, nil},
		{"a call of an object keeps the keywords", "def f(g):\n    return g(1, k=2)\n", []string{`runtime.DynCallKw(g, []any{int64(1)}, []runtime.KW{runtime.KW{Name: "k"`}, nil},
		{"unannotated comparison of objects", "def f(a, b):\n    return a < b\n", []string{`runtime.DynCompare("Lt", a, b)`}, nil},
		{"truth of an object", "def f(a):\n    if a:\n        return 1\n    return 2\n", []string{"runtime.DynBool(a)"}, nil},
		{"and of non-booleans returns an operand", "def f(a, b):\n    return a or b\n", []string{"if runtime.DynBool(_py"}, nil},
		{"a class gets run-time hooks", "class A:\n    def __init__(self, x):\n        self.x = x\n\n    def __add__(self, o):\n        return A(self.x + o.x)\n", []string{"PyGetAttr(name string)", "PySetAttr(name string, v any) bool", "PyBinop(op string, o any, reflected bool)", "map[string]any"}, nil},
		{"a reserved Go word is renamed", "default = 1\nprint(default)\n", []string{"default_py"}, []string{"var default "}},
		{"a Python builtin that is also a Go word stays", "print(list(map(str, [1])))\n", []string{"runtime.NewList"}, []string{"map_py"}},
		{"a def of a Go builtin name is renamed", "def copy(x):\n    return x\n\nprint(copy(1))\n", []string{"func copy_py("}, nil},
		{"registration is left out when the function is not used as a value", "def f(a: int) -> int:\n    return a\n\nprint(f(1))\n", nil, []string{"RegisterFunc"}},
		{"registration is there when it is", "def f(a, b=2):\n    return a\n\nprint(f)\nx = f\n", []string{"runtime.Callable{", "Defaults:"}, nil},
		{"a tuple with many items is a variable-length tuple", "t = (1, 2, 3, 4, 5, 6, 7)\n", []string{"runtime.NewVarTuple[int64](1, 2, 3, 4, 5, 6, 7)"}, nil},
		{"a lambda with defaults is the def it stands for", "fs = [lambda x, n=n: x * n for n in range(3)]\n", []string{"runtime.Callable{", "Name: \"<lambda>\""}, nil},
		{"unpacking an object goes through a list", "def f(p):\n    a, b = p\n    return a\n", []string{"runtime.DynList("}, nil},
		{"module variables are typed before the functions", "log = []\n\n\ndef note(s):\n    log.append(s)\n", []string{"var log *runtime.List[any]"}, nil},
		{"a global assigned another type joins the module variable", "config = None\n\n\ndef setup():\n    global config\n    config = {'a': 1}\n", []string{"runtime.Opt[*runtime.Dict[string, int64]]"}, nil},
		{"an annotation keeps the typed path", "def add(a: int, b: int) -> int:\n    return a + b\n", []string{"return a + b"}, []string{"DynBinary"}},
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

func TestPhase9Errors(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"an annotation is a contract", "def f(x: int) -> int:\n    return x\n\n\nf('a')\n", "argument 1 of 'f' expects int, got str"},
		{"an annotated variable keeps its type", "x: int = 'a'\n", "variable 'x' expects int, got str"},
		{"a declared attribute cannot be deleted", "class A:\n    def __init__(self):\n        self.x = 1\n\n\ndel A().x\n", "which the class declares"},
		{"functions have no attributes", "def f():\n    pass\n\n\nf.calls = 1\n", "assignment to attribute 'calls'"},
		{"del of a name", "x = 1\ndel x\n", "del of a variable"},
		{"nested containers keep their first types", "g = [[0]]\ng[0].append('x')\n", "expects int, got str"},
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

func TestPyName(t *testing.T) {
	for in, want := range map[string]string{
		"default_py": "default", "var_py": "var", "string_py": "string", "map": "map", "len": "len", "other": "other", "my_py": "my_py",
	} {
		if got := pyName(in); got != want {
			t.Errorf("pyName(%q) = %q, want %q", in, got, want)
		}
	}
}
