package difftest

import "testing"

// TestPhase3 runs every testdata/phase3 case under CPython (the committed
// expected.txt) and as transpiled Go, and compares stdout.
func TestPhase3(t *testing.T) {
	runDiffCases(t, "../../testdata/phase3/*/input.py", 40)
}
