package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase18Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"isinstance(x, type) asks the runtime", "def f(x):\n    return isinstance(x, type)\n\n\nprint(f(int))\n", []string{`"type"`}, nil},
		{"the members of a class are registered for dir", "class A:\n    def m(self):\n        pass\n\n\nprint(dir(A))\n", []string{"RegisterClassMembers"}, nil},
		{"classmethods are registered as values of the class", "class A:\n    @classmethod\n    def make(cls):\n        return 1\n\n\nprint(A.make())\n", []string{"RegisterClassCallable"}, nil},
		{"a name bound by with is usable after it", "class C:\n    def __enter__(self):\n        return self\n\n    def __exit__(self, a, b, c):\n        return False\n\n\ndef f():\n    with C() as cm:\n        pass\n    return cm\n\n\nprint(f())\n", []string{"var cm"}, nil},
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
