package difftest

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestPhase9 runs the cases that show the dynamic rules: objects, operators
// on values whose types are known when the program runs, and the errors
// Python raises there. Same comparison as every other phase.
func TestPhase9(t *testing.T) {
	runDiffCases(t, "../../testdata/phase9/*/input.py", 150)
}

// TestPhase9Stripped is the exit test of phase 9: every differential test of
// phases 2 to 8 that does not depend on its annotations gives the same result
// with all of them removed.
func TestPhase9Stripped(t *testing.T) {
	out := t.TempDir()
	script, err := filepath.Abs("../../scripts/strip_annotations.py")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("uv", "run", "--python", "3.13", "python", script, out)
	if msg, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("stripping the annotations: %v\n%s", err, msg)
	}
	runDiffCases(t, filepath.Join(out, "*", "input.py"), 450)
}
