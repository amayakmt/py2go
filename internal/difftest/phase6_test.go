package difftest

import "testing"

// TestPhase6 runs every testdata/phase6 case under CPython (the committed
// expected.txt, and expected.err for an uncaught exception) and as
// transpiled Go, and compares stdout, exit status and the last stderr line.
func TestPhase6(t *testing.T) {
	runDiffCases(t, "../../testdata/phase6/*/input.py", 60)
}
