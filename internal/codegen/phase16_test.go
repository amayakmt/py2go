package codegen

import (
	"strings"
	"testing"
)

func TestGeneratePhase16Output(t *testing.T) {
	tests := []struct {
		name, src string
		want      []string
		notWant   []string
	}{
		{"an attribute assigned after the class is a class attribute", "class A:\n    pass\n\n\nA.top = A()\nprint(A.top)\n", []string{"RegisterClassAttr"}, nil},
		{"a subclass can shadow an inherited class attribute", "class A:\n    x: int = 1\n\n\nclass B(A):\n    pass\n\n\nB.x = 2\n", []string{"RegisterClassAttr"}, nil},
		{"a class body reads an attribute it set before", "class A:\n    a = [1, 2]\n    b = a\n", []string{"_pyca_A_a"}, nil},
		{"open(...).write(...) closes the file", "open('x', 'w').write('s')\n", []string{"WithTempFile"}, nil},
		{"print with flush flushes the file", "print('a', flush=True)\n", []string{"FlushFile"}, nil},
		{"print without flush does not", "print('a')\n", nil, []string{"FlushFile"}},
		{"an except clause registers the exception for sys.exc_info", "try:\n    pass\nexcept ValueError:\n    print(1)\n", []string{"PushHandled", "PopHandled"}, nil},
		{"an except clause that returns does not", "def f():\n    try:\n        pass\n    except ValueError:\n        return 1\n    return 2\n", nil, []string{"PushHandled"}},
		{"__fspath__ is visible to hasattr", "class A:\n    def __fspath__(self):\n        return 'p'\n\n\nprint(hasattr(A(), '__fspath__'))\n", []string{`"__fspath__"`}, nil},
		{"__index__ makes an object usable as an index", "class A:\n    def __index__(self):\n        return 1\n\n\ndef f(i):\n    return [1, 2][i]\n\n\nprint(f(A()))\n", []string{"PyIndex"}, nil},
		{"chr of a literal is typed in module variables", "A = chr(65)\n\n\ndef f():\n    return A\n\n\nprint(f())\n", []string{"DynChr"}, []string{"UndefinedName"}},
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
