package difftest

import "testing"

// TestPhase19 runs programs for the modules phase 19 can offer without C
// libraries: the Windows path modules, and the failing imports of the modules
// that need Windows, Tcl/Tk, ncurses or libffi.
func TestPhase19(t *testing.T) {
	runDiffCases(t, "../../testdata/phase19/*/input.py", 1)
}
