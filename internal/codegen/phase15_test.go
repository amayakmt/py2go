package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase15Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"vars of an instance reads the attributes", "class A:\n    def __init__(self):\n        self.x = 1\n\n\nprint(vars(A()))\n", []string{"runtime.DynVars("}, nil},
		{"dir reads the attribute names", "class A:\n    pass\n\n\nprint(dir(A()))\n", []string{"runtime.DynDir("}, nil},
		{"class attributes are registered for getattr", "class A:\n    k = 'v'\n\n\nprint(getattr(A, 'k'))\n", []string{"RegisterClassAttr"}, nil},
		{"instance attribute names keep their definition order", "class A:\n    def __init__(self):\n        self.zeta = 1\n        self.alpha = 2\n", []string{`[]string{"zeta", "alpha"}`}, nil},
		{"pickle registers module functions", "def f():\n    pass\n\n\nprint(_native(\"pickle.is_class\", f))\n", []string{"RegisterGlobal("}, nil},
		{"pickle registers an allocator for each class", "class A:\n    pass\n\n\nprint(_native(\"pickle.is_class\", A()))\n", []string{`RegisterAlloc("A"`}, nil},
		{"no pickle, no allocator", "class A:\n    pass\n\n\nprint(A())\n", nil, []string{"RegisterAlloc"}},
		{"a pickle protocol method is reachable from hasattr", "class A:\n    def __getstate__(self):\n        return 1\n\n\nprint(hasattr(A(), '__getstate__'))\n", []string{`"__getstate__"`}, nil},
		{"an assignment to a narrowed parameter drops the narrowing", "def f(x):\n    if isinstance(x, str):\n        x = x.encode()\n    return x\n", []string{"func f(x any)"}, nil},
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
