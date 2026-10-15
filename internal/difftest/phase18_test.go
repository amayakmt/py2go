package difftest

import "testing"

// TestPhase18 runs programs that use the developer modules: unittest with
// its mock, difflib, pprint, inspect and token.
func TestPhase18(t *testing.T) {
	runDiffCases(t, "../../testdata/phase18/*/input.py", 1)
}
