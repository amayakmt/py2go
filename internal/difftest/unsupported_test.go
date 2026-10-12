package difftest

import (
	"path/filepath"
	"testing"
)

// TestUnsupported asserts the exact error and line number for every
// construct the transpiler rejects.
func TestUnsupported(t *testing.T) {
	inputs, err := filepath.Glob("../../testdata/unsupported/*/input.py")
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Fatal("no test cases found")
	}
	for _, in := range inputs {
		dir := filepath.Dir(in)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			runUnsupportedCase(t, dir)
		})
	}
}

// runUnsupportedCase checks that the transpiler rejects input.py with
// exactly the message in expected_error.txt, line number included.
func runUnsupportedCase(t *testing.T, dir string) {
	t.Helper()
	want, ok := readOptional(t, filepath.Join(dir, "expected_error.txt"))
	if !ok {
		t.Fatal("missing expected_error.txt")
	}
	src, err := transpile(filepath.Join(dir, "input.py"))
	if err == nil {
		t.Fatalf("want error %q, but transpiling succeeded:\n%s", want, src)
	}
	if err.Error() != want {
		t.Errorf("error mismatch\n got: %s\nwant: %s", err, want)
	}
}
