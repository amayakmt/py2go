package difftest

import "testing"

// TestPhase13 runs programs with coroutines, asyncio, threads and the
// modules built on them. Same comparison as every other phase.
func TestPhase13(t *testing.T) {
	runDiffCases(t, "../../testdata/phase13/*/input.py", 1)
}
