package difftest

import "testing"

// TestPhase10 runs the cases that show the numbers, text and bytes of phase
// 10: ints of any size, floats, complex, bytes and bytearray, the full str
// and its codecs, frozenset, slice, range and the hash and identity
// protocols. Same comparison as every other phase: stdout, and for a crash
// the last line of stderr.
func TestPhase10(t *testing.T) {
	runDiffCases(t, "../../testdata/phase10/*/input.py", 150)
}
