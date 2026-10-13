package difftest

import "testing"

// TestPhase14 runs programs that use the text, numeric and algorithm
// modules: math down to the last digit, fractions, statistics, struct,
// unicodedata, glob, shlex and the rest of the phase.
func TestPhase14(t *testing.T) {
	runDiffCases(t, "../../testdata/phase14/*/input.py", 1)
}
