package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/amayakmt/py2go/internal/codegen"
	"github.com/amayakmt/py2go/internal/link"
)

// project builds a program of several modules into an executable, and for
// "run" starts it. The input is the entry file, or a directory with a
// __main__.py as "python dir" has it. The modules it imports are found next to
// the entry file.
func project(cmd string, args []string) error {
	out := ""
	var rest []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			out = args[i+1]
			i++
			continue
		}
		rest = append(rest, args[i])
	}
	if len(rest) == 0 {
		return fmt.Errorf("usage: py2go %s <file.py|dir>%s", cmd, map[bool]string{true: " [-o out]", false: " [args]"}[cmd == "build"])
	}
	entry := rest[0]
	if st, err := os.Stat(entry); err == nil && st.IsDir() {
		entry = filepath.Join(entry, "__main__.py")
	}
	if filepath.Ext(entry) != ".py" {
		return fmt.Errorf("input must be a .py file or a directory with __main__.py, got %q", rest[0])
	}
	mod, err := link.LoadFile(entry)
	if err != nil {
		return fmt.Errorf("%s: %w", entry, err)
	}
	src, err := codegen.Generate(mod)
	if err != nil {
		return fmt.Errorf("%s: %w", entry, err)
	}
	root, err := checkoutRoot()
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "py2go-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	gomod := fmt.Sprintf("module py2goprog\n\ngo %s\n\nrequire github.com/amayakmt/py2go v0.0.0\n\nreplace github.com/amayakmt/py2go => %s\n", goVersion(root), root)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), src, 0o644); err != nil {
		return err
	}
	if out == "" {
		out = strings.TrimSuffix(filepath.Base(entry), ".py")
		if filepath.Base(entry) == "__main__.py" {
			out = filepath.Base(filepath.Dir(mustAbs(entry)))
		}
	}
	if cmd == "run" {
		out = filepath.Join(dir, "prog")
	}
	build := exec.Command("go", "build", "-o", mustAbs(out), ".")
	build.Dir = dir
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}
	if cmd == "build" {
		fmt.Println("wrote", out)
		return nil
	}
	run := exec.Command(out, rest[1:]...)
	run.Stdin, run.Stdout, run.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := run.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		return err
	}
	return nil
}

func mustAbs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return a
}

// checkoutRoot finds the directory of the py2go module that the generated
// program imports the runtime from: $PY2GO_ROOT, or the nearest ancestor of
// the current directory or of the executable that holds its go.mod.
func checkoutRoot() (string, error) {
	if r := os.Getenv("PY2GO_ROOT"); r != "" {
		return filepath.Abs(r)
	}
	var starts []string
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	for _, d := range starts {
		for ; ; d = filepath.Dir(d) {
			if b, err := os.ReadFile(filepath.Join(d, "go.mod")); err == nil && strings.HasPrefix(string(b), "module github.com/amayakmt/py2go\n") {
				return d, nil
			}
			if filepath.Dir(d) == d {
				break
			}
		}
	}
	return "", fmt.Errorf("cannot find the py2go checkout; set PY2GO_ROOT")
}

func goVersion(root string) string {
	b, _ := os.ReadFile(filepath.Join(root, "go.mod"))
	for _, line := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(line, "go "); ok {
			return strings.TrimSpace(v)
		}
	}
	return "1.22"
}
