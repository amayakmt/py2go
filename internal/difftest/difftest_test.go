package difftest

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/amayakmt/py2go/internal/codegen"
	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/internal/pyrun"
)

func TestPhase1(t *testing.T) {
	inputs, err := filepath.Glob("../../testdata/phase1/*/input.py")
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Fatal("no test cases found")
	}

	for _, in := range inputs {
		dir := filepath.Dir(in)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join(dir, "expected.txt"))
			if err != nil {
				t.Fatal(err)
			}

			js, err := pyrun.DumpAST(in)
			if err != nil {
				t.Fatal(err)
			}
			root, err := pyast.Decode(js)
			if err != nil {
				t.Fatal(err)
			}
			src, err := codegen.Generate(root)
			if err != nil {
				t.Fatal(err)
			}

			// Build in a temp dir so no artifacts land in the repo.
			tmp := t.TempDir()
			goFile := filepath.Join(tmp, "main.go")
			if err := os.WriteFile(goFile, src, 0o644); err != nil {
				t.Fatal(err)
			}

			var stdout, stderr bytes.Buffer
			cmd := exec.Command("go", "run", goFile)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("go run failed: %v\n%s\n--- generated ---\n%s", err, stderr.String(), src)
			}

			if !bytes.Equal(stdout.Bytes(), want) {
				t.Errorf("stdout mismatch\n--- go ---\n%q\n--- python ---\n%q", stdout.Bytes(), want)
			}
		})
	}
}
