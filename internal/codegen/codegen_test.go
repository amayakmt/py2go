package codegen

import (
	"os"
	"strings"
	"testing"

	"github.com/amayakmt/py2go/internal/pyast"
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
