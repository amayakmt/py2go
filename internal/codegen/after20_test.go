package codegen

import (
	"strings"
	"testing"
)

// Code generation for the language features added after phase 20.
func TestGenerateAfterPhase20Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"type() of None and of a container is looked up at run time", "print(type(None), type([1]))\n", []string{"DynType(runtime.NoneV)"}, nil},
		{"a docstring travels with the function value", "def f():\n    \"\"\"doc\"\"\"\n\n\ng = [f]\nprint(g[0].__doc__)\n", []string{`Doc: "doc"`}, nil},
		{"a display with a shared container is copied shallow", "def f(x):\n    x.append(1)\n\n\nshared = ['a']\nf([shared])\nprint(shared)\n", []string{"DynShallow("}, []string{"DynDeep("}},
		{"a display of constants is copied deep", "def f(x):\n    x.append(1)\n\n\nf([[1, 2], [3]])\n", []string{"DynDeep("}, nil},
		{"unpacking an iterator goes through a list", "a, b = (int(x) for x in '1-2'.split('-'))\nprint(a + b)\n", []string{"CheckUnpack"}, nil},
		{"item assignment on an Optional dict unwraps it", "class A:\n    def __init__(self):\n        self.d = None\n\n    def fill(self):\n        self.d = {}\n        self.d['k'] = 1\n\n\na = A()\na.fill()\nprint(a.d)\n", nil, []string{"item assignment on"}},
		{"an override that adds a parameter with a default keeps both signatures", "class A:\n    def f(self, a=1):\n        return a\n\n\nclass B(A):\n    def f(self, a=1, b=2):\n        return a + b\n\n\nx: A = B()\nprint(x.f(), B().f(1, 2))\n", []string{"M_f__ext(", ".M_f__ext(int64(1), int64(2))"}, nil},
		{"except with a class in a variable", "E = KeyError\ntry:\n    {}['x']\nexcept E:\n    print('caught')\n", []string{"ExcNamesOf("}, nil},
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
					t.Errorf("missing %q in:\n%s", w, got)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(got, w) {
					t.Errorf("unexpected %q in:\n%s", w, got)
				}
			}
		})
	}
}

func TestMethodArgumentShapeErrorsStayCompileErrorsOutsideMethods(t *testing.T) {
	// Inside a method the same call is a TypeError raised when it runs: the
	// method is compiled for each class, and the branch may never run.
	_, err := generate(t, "class C:\n    def f(self, a: int) -> int:\n        return a\n\n\nc: C = C()\nprint(c.f(1, 2))\n")
	if err == nil || !strings.Contains(err.Error(), "C.f() takes 2 positional arguments but 3 were given") {
		t.Errorf("got %v, want the TypeError message as a compile error", err)
	}
	if _, err := generate(t, "class C:\n    def f(self, a: int) -> int:\n        return a\n\n    def g(self) -> int:\n        return self.f(1, 2)\n\n\nprint(C().f(1))\n"); err != nil {
		t.Errorf("inside a method the bad call should compile to a raise, got %v", err)
	}
}
