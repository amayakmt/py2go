package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amayakmt/py2go/internal/codegen"
	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/internal/pyrun"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "py2go:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: py2go <file.py>")
	}
	in := args[0]
	if filepath.Ext(in) != ".py" {
		return fmt.Errorf("input must be a .py file, got %q", in)
	}

	js, err := pyrun.DumpAST(in)
	if err != nil {
		return err
	}
	root, err := pyast.Decode(js)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	src, err := codegen.Generate(root)
	if err != nil {
		// Prefix with the file so errors read like compiler output.
		return fmt.Errorf("%s: %w", in, err)
	}

	out := strings.TrimSuffix(in, ".py") + ".go"
	if err := os.WriteFile(out, src, 0o644); err != nil {
		return err
	}
	fmt.Println("wrote", out)
	return nil
}
