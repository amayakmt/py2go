package link

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amayakmt/py2go/internal/pyast"
)

// project writes files (path -> source) into a temporary directory and links
// the entry file.
func project(t *testing.T, files map[string]string) (*pyast.Module, error) {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return LoadFile(filepath.Join(dir, "main.py"))
}

// names collects every identifier the joined module mentions.
func names(mod *pyast.Module) map[string]bool {
	out := map[string]bool{}
	var visit func(n pyast.Node)
	visit = func(n pyast.Node) {
		switch x := n.(type) {
		case *pyast.Name:
			out[x.ID] = true
		case *pyast.FunctionDef:
			out[x.Name] = true
		case *pyast.ClassDef:
			out[x.Name] = true
		}
		walkChildren(n, visit)
	}
	for _, st := range mod.Body {
		visit(st)
	}
	return out
}

func TestLinkMangling(t *testing.T) {
	mod, err := project(t, map[string]string{
		"main.py": "import util\nfrom util import f as g\nprint(util.X, g(1), util.f(2), __name__)\n",
		"util.py": "X = 1\ndef f(a):\n    b = X\n    return a + b\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"util_pm_X", "util_pm_f", "print", "a", "b"} {
		if !got[want] {
			t.Errorf("joined module should mention %q: %v", want, got)
		}
	}
	for _, bad := range []string{"util", "g", "X", "f"} {
		if got[bad] {
			t.Errorf("joined module should not mention %q", bad)
		}
	}
	if len(mod.Body) == 0 {
		t.Fatal("empty module")
	}
}

func TestLinkRunsAModuleOnce(t *testing.T) {
	mod, err := project(t, map[string]string{
		"main.py": "import a\nimport b\nimport a\n",
		"a.py":    "import b\nA = 1\n",
		"b.py":    "B = 1\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, st := range mod.Body {
		if as, ok := st.(*pyast.Assign); ok {
			if nm, ok := as.Targets[0].(*pyast.Name); ok && nm.ID == "b_pm_B" {
				count++
			}
		}
	}
	if count != 1 {
		t.Errorf("module b ran %d times, want 1", count)
	}
}

func TestLinkPackagesAndRelativeImports(t *testing.T) {
	mod, err := project(t, map[string]string{
		"main.py":         "import pkg.sub\nfrom pkg import helper\nprint(pkg.sub.V, helper())\n",
		"pkg/__init__.py": "from .util import helper\n",
		"pkg/util.py":     "def helper():\n    return 1\n",
		"pkg/sub.py":      "from . import util\nV = util.helper()\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"pkg__util_pm_helper", "pkg__sub_pm_V"} {
		if !got[want] {
			t.Errorf("joined module should mention %q: %v", want, got)
		}
	}
}

func TestLinkMissingModuleRaisesWhenRun(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "try:\n    import nothere\nexcept ImportError:\n    pass\n"})
	if err != nil {
		t.Fatal(err)
	}
	tr := mod.Body[0].(*pyast.Try)
	raise, ok := tr.Body[0].(*pyast.Raise)
	if !ok {
		t.Fatalf("the import should become a raise, got %T", tr.Body[0])
	}
	msg := raise.Exc.(*pyast.Call).Args[0].(*pyast.Constant).Value
	if msg != "No module named 'nothere'" {
		t.Errorf("message = %v", msg)
	}
}

func TestLinkLeavesTheStandardLibraryToTheCompiler(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import pdb\nfrom typing import Any\n"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := mod.Body[0].(*pyast.Import); !ok {
		t.Errorf("import pdb should stay, got %T", mod.Body[0])
	}
	if _, ok := mod.Body[1].(*pyast.ImportFrom); !ok {
		t.Errorf("from typing import should stay, got %T", mod.Body[1])
	}
}

func TestLinkModuleAsValueIsAModuleObject(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import util\nprint(util)\n", "util.py": "X = 1\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !names(mod)["_native"] {
		t.Errorf("a module used as a value should be a call to _native")
	}
}

func TestLinkAssignmentToAModuleAttributeDeclaresGlobal(t *testing.T) {
	mod, err := project(t, map[string]string{
		"main.py": "import cfg\ndef f():\n    cfg.level = 2\n",
		"cfg.py":  "level = 1\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	var fn *pyast.FunctionDef
	for _, st := range mod.Body {
		if f, ok := st.(*pyast.FunctionDef); ok {
			fn = f
		}
	}
	g, ok := fn.Body[0].(*pyast.Global)
	if !ok || g.Names[0] != "cfg_pm_level" {
		t.Errorf("f should start with global cfg_pm_level, got %#v", fn.Body[0])
	}
}
