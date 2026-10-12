package difftest

import "testing"

// TestPhase7 runs every testdata/phase7 case under CPython (the committed
// expected.txt, and expected.err for an uncaught exception) and as
// transpiled Go, and compares stdout, exit status and the last stderr line.
func TestPhase7(t *testing.T) {
	runDiffCases(t, "../../testdata/phase7/*/input.py", 80)
}
