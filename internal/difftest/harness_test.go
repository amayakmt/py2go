package difftest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amayakmt/py2go/internal/codegen"
	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/internal/pyrun"
)

// transpile runs the whole pipeline on one Python file.
func transpile(pyFile string) ([]byte, error) {
	js, err := pyrun.DumpAST(pyFile)
	if err != nil {
		return nil, err
	}
	root, err := pyast.Decode(js)
	if err != nil {
		return nil, err
	}
	return codegen.Generate(root)
}

// goRun builds generated source in a throwaway module, runs it, and returns
// what it wrote and how it exited. Generated code imports
// github.com/amayakmt/py2go/runtime, so the module gets a replace directive
// that points at this checkout. A build failure is an error; a program that
// exits non-zero is not.
func goRun(t *testing.T, src []byte) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	// The throwaway module must declare at least the Go version of the
	// module it requires, or go refuses to build.
	rootMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	goVersion := ""
	for _, line := range strings.Split(string(rootMod), "\n") {
		if v, ok := strings.CutPrefix(line, "go "); ok {
			goVersion = strings.TrimSpace(v)
		}
	}

	tmp := t.TempDir()
	gomod := fmt.Sprintf(`module py2goprog

go %s

require github.com/amayakmt/py2go v0.0.0

replace github.com/amayakmt/py2go => %s
`, goVersion, root)
	for name, content := range map[string]string{"go.mod": gomod, "main.go": string(src)} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Build first and run the binary, so that stderr and the exit status
	// are the program's own and not "go run"'s.
	var buildErr bytes.Buffer
	build := exec.Command("go", "build", "-o", "prog", ".")
	build.Dir = tmp
	build.Stderr = &buildErr
	if err := build.Run(); err != nil {
		return "", buildErr.String(), -1, fmt.Errorf("go build failed: %w\n%s", err, buildErr.String())
	}

	var out, errOut bytes.Buffer
	run := exec.Command(filepath.Join(tmp, "prog"))
	run.Stdout = &out
	run.Stderr = &errOut
	runErr := run.Run()
	code := 0
	if exitErr, ok := runErr.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if runErr != nil {
		return "", "", -1, runErr
	}
	return out.String(), errOut.String(), code, nil
}

// readOptional returns the trimmed file content and whether the file exists.
func readOptional(t *testing.T, path string) (string, bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimRight(string(data), "\n"), true
}

// runDiffCase checks one test directory:
//
//	input.py      the program
//	expected.txt  stdout of CPython 3.13
//	expected.err  only if CPython exits with an error: its last stderr line
//
// The transpiled program must print the same stdout. When expected.err
// exists it must also exit with status 1 and end its stderr with that line,
// which is how CPython reports an uncaught exception. Otherwise it must exit 0.
func runDiffCase(t *testing.T, dir string) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join(dir, "expected.txt"))
	if err != nil {
		t.Fatal(err)
	}
	wantErr, expectFailure := readOptional(t, filepath.Join(dir, "expected.err"))

	src, err := transpile(filepath.Join(dir, "input.py"))
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code, err := goRun(t, src)
	if err != nil {
		t.Fatalf("%v\n--- generated ---\n%s", err, src)
	}

	if expectFailure {
		lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
		if last := lines[len(lines)-1]; last != wantErr {
			t.Errorf("last line of stderr: got %q, want %q\n--- stderr ---\n%s", last, wantErr, stderr)
		}
		if code != 1 {
			t.Errorf("exit status: got %d, want 1 (CPython fails with %q)", code, wantErr)
		}
	} else if code != 0 {
		t.Fatalf("exit status %d, want 0\n%s\n--- generated ---\n%s", code, stderr, src)
	}
	if stdout != string(want) {
		t.Errorf("stdout mismatch\n--- go ---\n%q\n--- python ---\n%q\n--- generated ---\n%s", stdout, want, src)
	}
}

// runDiffCases runs every <pattern>/input.py as a parallel subtest.
func runDiffCases(t *testing.T, pattern string, minCases int) {
	t.Helper()
	inputs, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) < minCases {
		t.Fatalf("want at least %d test cases, found %d", minCases, len(inputs))
	}
	for _, in := range inputs {
		dir := filepath.Dir(in)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			runDiffCase(t, dir)
		})
	}
}
