package difftest

import "testing"

// TestPhase15 runs programs that read and write data formats: base64,
// binascii, csv, configparser, tomllib, html, pickle, zlib, hashlib and the
// rest of the phase.
func TestPhase15(t *testing.T) {
	runDiffCases(t, "../../testdata/phase15/*/input.py", 1)
}
