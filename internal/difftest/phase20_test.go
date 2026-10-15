package difftest

import "testing"

// TestPhase20 runs the programs added while measuring the project against
// CPython's Lib/: typing and dataclasses by qualified name, argparse, graphlib,
// reprlib, modules used as values, dict(mapping, **keywords), and the order of
// the attributes of an instance.
func TestPhase20(t *testing.T) {
	runDiffCases(t, "../../testdata/phase20/*/input.py", 1)
}
