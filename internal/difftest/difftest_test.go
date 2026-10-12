package difftest

import "testing"

func TestPhase1(t *testing.T) {
	runDiffCases(t, "../../testdata/phase1/*/input.py", 1)
}
