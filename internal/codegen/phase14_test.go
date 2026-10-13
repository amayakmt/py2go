package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase14Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"a method that rebinds self gets a variable of the class", "class A:\n    def f(self):\n        self = A()\n        return self\n", []string{"self_ *", "var self A = self_"}, nil},
		{"a method that does not rebind self keeps the receiver", "class A:\n    def f(self):\n        return 1\n", []string{"func (self *"}, []string{"self_"}},
		{"type of an object is a class object", "def f(x):\n    return type(x)\n", []string{"runtime.DynType("}, nil},
		{"type of an int is a class object", "x = 5\ny = type(x)\n", []string{"runtime.DynType("}, nil},
		{"an int attribute that is given an object holds any", "class A:\n    def __init__(self):\n        self.n = 1\n\n    def grow(self, k):\n        self.n = self.n * k\n", []string{"Get_n() any"}, nil},
		{"a str attribute assigned an object keeps its type", "class A:\n    def __init__(self, d):\n        self.s = 'a'\n        self.s = d['k']\n", []string{"Get_s() string"}, nil},
		{"and with an object on the right of a narrowed operand", "def g(x):\n    return x\n\n\ndef f(x):\n    return isinstance(x, float) and (g(x) or g(x))\n", []string{"func f(x any) any", "runtime.InstanceOf[float64](x)"}, []string{"func() bool"}},
		{"an exception class held in a variable is registered", "class E(Exception):\n    pass\n\n\ndef make(cls):\n    return cls('x')\n\n\nmake(E)\n", []string{`RegisterCtor("E"`}, nil},
		{"a class of a module records its module", "class E(Exception):\n    pass\n", nil, []string{"ClassModule"}},
		{"main is renamed like a Go keyword", "def main():\n    print(1)\n\n\nmain()\n", []string{"func main_py()"}, nil},
		{"a function's __name__ is the Python name", "def default():\n    pass\n\n\nprint(default.__name__)\n", []string{`"default"`}, nil},
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
