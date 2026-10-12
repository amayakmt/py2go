package difftest

import "testing"

// TestPhase2 runs every testdata/phase2 case under CPython (the committed
// expected.txt) and as transpiled Go, and compares stdout.
func TestPhase2(t *testing.T) {
	runDiffCases(t, "../../testdata/phase2/*/input.py", 30)
}
