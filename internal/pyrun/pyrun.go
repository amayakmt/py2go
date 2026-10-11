// Package pyrun runs the embedded AST dumper under Python 3.13.
package pyrun

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"strings"
)

//go:embed dump_ast.py
var dumpScript string

// DumpAST runs dump_ast.py on the given file and returns the JSON tree.
func DumpAST(pyFile string) ([]byte, error) {
	// uv resolves the interpreter pinned to 3.13, so a bare python3
	// on PATH never decides which version produces the tree.
	cmd := exec.Command("uv", "run", "--python", "3.13", "python", "-", pyFile)
	cmd.Stdin = strings.NewReader(dumpScript)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("python AST dump failed: %s", msg)
	}
	return stdout.Bytes(), nil
}
