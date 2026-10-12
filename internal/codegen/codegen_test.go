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
	data := []byte(`{"_type": "Module", "body": [{"_type": "Match", "lineno": 12}]}`)
	root, err := pyast.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Generate(root)
	if err == nil || !strings.Contains(err.Error(), "unsupported: Match at line 12") {
		t.Fatalf("want \"unsupported: Match at line 12\", got %v", err)
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
		{"declaration", "x: int = 5\n", "var x int64 = 5\n\t_ = x\n"},
		{"floor division", "print(-7 // 2)\n", "runtime.Print(runtime.FloorDiv(-7, 2))"},
		{"modulo", "print(-7 % 3)\n", "runtime.Print(runtime.Mod(-7, 3))"},
		{"true division of ints", "print(7 / 2)\n", "runtime.DivInt(7, 2)"},
		{"mixed operands convert", "i: int = 1\nprint(i + 2.5)\n", "float64(i) + 2.5"},
		{"constant floats are typed", "print(0.1 + 0.2)\n", "float64(0.1) + float64(0.2)"},
		{"parentheses only where needed", "a: int = 1\nprint((a + 2) * 3 - (a - 1))\n", "(a+2)*3 - (a - 1)"},
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
