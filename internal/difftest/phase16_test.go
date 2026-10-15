package difftest

import "testing"

func TestPhase16(t *testing.T) {
	runDiffCases(t, "../../testdata/phase16/*/input.py", 1)
}
