package difftest

import "testing"

// TestPhase4 runs every testdata/phase4 case under CPython (the committed
// expected.txt, and expected.err for an uncaught exception) and as
// transpiled Go, and compares stdout, exit status and the last stderr line.
func TestPhase4(t *testing.T) {
	runDiffCases(t, "../../testdata/phase4/*/input.py", 40)
}
