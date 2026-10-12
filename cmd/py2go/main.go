package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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
	if len(args) > 0 && args[0] == "check" {
		return check(args[1:])
	}
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

// check transpiles every .py file under the given paths without writing
// anything, and reports how many were rejected and why. The number of
// rejected files is the measure of how much of Python is covered: the goal of
// the roadmap is a corpus of CPython's Lib/ with none.
func check(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("usage: py2go check <file.py|dir>...")
	}
	var files []string
	for _, p := range paths {
		err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".py" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	reasons := map[string]int{}
	rejected := 0
	for _, f := range files {
		if err := transpile(f); err != nil {
			rejected++
			reasons[reason(err)]++
			fmt.Printf("%s: %v\n", f, err)
		}
	}
	type kv struct {
		reason string
		n      int
	}
	var top []kv
	for r, n := range reasons {
		top = append(top, kv{r, n})
	}
	sort.Slice(top, func(i, j int) bool {
		return top[i].n > top[j].n || top[i].n == top[j].n && top[i].reason < top[j].reason
	})
	fmt.Printf("\n%d files, %d rejected, %d accepted\n", len(files), rejected, len(files)-rejected)
	for i, r := range top {
		if i == 15 {
			break
		}
		fmt.Printf("%6d  %s\n", r.n, r.reason)
	}
	return nil
}

func transpile(path string) error {
	js, err := pyrun.DumpAST(path)
	if err != nil {
		return err
	}
	root, err := pyast.Decode(js)
	if err != nil {
		return err
	}
	_, err = codegen.Generate(root)
	return err
}

// reason is the part of an error that says what is unsupported, without the
// names and line numbers.
func reason(err error) string {
	msg := err.Error()
	if i := strings.Index(msg, " at line "); i >= 0 {
		msg = msg[:i]
	}
	msg = strings.TrimPrefix(msg, "unsupported: ")
	if i := strings.Index(msg, "'"); i >= 0 {
		msg = msg[:i] + "..."
	}
	return msg
}
