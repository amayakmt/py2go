package difftest

import "testing"

// TestPhase8 runs every testdata/phase8 case under CPython (the committed
// expected.txt, and expected.err for an uncaught exception) and as
// transpiled Go, and compares stdout, exit status and the last stderr line.
func TestPhase8(t *testing.T) {
	runDiffCases(t, "../../testdata/phase8/*/input.py", 80)
}
