package difftest

import "testing"

// TestPhase11 runs small projects: each case directory holds an input.py and
// the modules and packages it imports. Same comparison as every other phase.
func TestPhase11(t *testing.T) {
	runDiffCases(t, "../../testdata/phase11/*/input.py", 30)
}
