package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase20Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"dict of a mapping and keywords", "def f(base: dict[str, int]):\n    return dict(base, a=1)\n\n\nprint(f({'x': 1}))\n", []string{"StrDictWith"}, nil},
		{"dict of keywords of different types", "d = dict(a=1, b='x')\nprint(d)\n", []string{"StrDictWith"}, nil},
		{"keyword names that are Go words stay Python names in **kwargs", "def f(**kw):\n    return sorted(kw)\n\n\nprint(f(default=1, go=2))\n", []string{`Key: "default"`}, []string{`"default_py"`}},
		{"attribute order is kept", "class A:\n    pass\n\n\na = A()\na.z = 1\na.b = 2\nprint(vars(a))\n", []string{"_pyo"}, nil},
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
