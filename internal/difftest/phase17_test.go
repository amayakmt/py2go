package difftest

import "testing"

// TestPhase17 runs programs that use the network modules on the loopback
// interface: sockets, http, urllib, email, smtplib and the rest.
func TestPhase17(t *testing.T) {
	runDiffCases(t, "../../testdata/phase17/*/input.py", 1)
}
