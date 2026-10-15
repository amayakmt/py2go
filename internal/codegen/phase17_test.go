package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase17Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"a method alias in a class body is a method", "class A:\n    def f(self):\n        return 1\n    g = f\n\n\nprint(A().g())\n", []string{"M_g"}, nil},
		{"a property getter that only raises is typed as an object", "class A:\n    @property\n    def p(self):\n        raise ValueError('x')\n\n\nprint(1)\n", []string{"Get_p"}, nil},
		{"isinstance(self, Sub) in a base method is false", "class A:\n    def f(self):\n        if isinstance(self, B):\n            return 1\n        return 2\n\n\nclass B(A):\n    pass\n\n\nprint(A().f())\n", nil, nil},
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
