package difftest

import "testing"

// TestPhase12 runs programs that use the standard library modules py2go
// ships: math, itertools, functools, operator, collections and the rest of
// the core. Same comparison as every other phase.
func TestPhase12(t *testing.T) {
	runDiffCases(t, "../../testdata/phase12/*/input.py", 1)
}
