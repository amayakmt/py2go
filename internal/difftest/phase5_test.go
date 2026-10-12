package difftest

import "testing"

// TestPhase5 runs every testdata/phase5 case under CPython (the committed
// expected.txt, and expected.err for an uncaught exception) and as
// transpiled Go, and compares stdout, exit status and the last stderr line.
func TestPhase5(t *testing.T) {
	runDiffCases(t, "../../testdata/phase5/*/input.py", 60)
}
