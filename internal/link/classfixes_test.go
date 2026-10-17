package link

import (
	"strings"
	"testing"

	"github.com/amayakmt/py2go/internal/pyast"
)

func TestBranchDefsAreRenamedOnlyWhenUsedOutside(t *testing.T) {
	src := "def f():\n    out = []\n    for n in range(3):\n        def get(n=n):\n            return n\n        out.append(get)\n    return out\n\n\n" +
		"import sys\nif sys.argv:\n    def later():\n        return 1\nprint(later())\n"
	mod, err := project(t, map[string]string{"main.py": src})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	if !got["get"] {
		t.Errorf("a def used only inside its loop should keep its name")
	}
	for n := range got {
		if len(n) > 7 && n[:7] == "get_def" {
			t.Errorf("get was renamed to %s", n)
		}
	}
	renamed := false
	for n := range got {
		if len(n) > 9 && n[:9] == "later_def" {
			renamed = true
		}
	}
	if !renamed {
		t.Errorf("a def in an if that is used after it should be renamed")
	}
}

func TestImportFallbackIsDroppedWhenTheModuleExists(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "try:\n    import helper\nexcept ImportError:\n    helper = None\nprint(helper.X)\n", "helper.py": "X = 1\n"})
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range mod.Body {
		if _, isTry := st.(*pyast.Try); isTry {
			t.Errorf("the try around an import of an existing module should be gone")
		}
	}
	mod, err = project(t, map[string]string{"main.py": "try:\n    import nothere\nexcept ImportError:\n    nothere = None\nprint(nothere)\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !names(mod)["ModuleNotFoundError"] {
		t.Errorf("the import of a missing module should still raise inside the try")
	}
}

func TestModuleFunctionAliasedInAClassBecomesAMethod(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "def _repr(self):\n    return 'C'\n\n\nclass C:\n    __repr__ = _repr\n\n\nprint(C())\n"})
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range mod.Body {
		cd, ok := st.(*pyast.ClassDef)
		if !ok || cd.Name != "C" {
			continue
		}
		for _, cst := range cd.Body {
			if fd, ok := cst.(*pyast.FunctionDef); ok && fd.Name == "__repr__" {
				return
			}
		}
	}
	t.Errorf("C should have a __repr__ method that calls _repr")
}

func TestReadlineIsUnavailable(t *testing.T) {
	if got, ok := unavailableModule("readline"); !ok || got != "readline" {
		t.Errorf("unavailableModule(readline) = %q, %v", got, ok)
	}
}

func TestFromImportFallbackIsDroppedForALibraryName(t *testing.T) {
	src := "try:\n    from gettext import gettext\nexcept ImportError:\n    def gettext(m):\n        return m\nprint(gettext('x'))\n"
	mod, err := project(t, map[string]string{"main.py": src})
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range mod.Body {
		if _, isTry := st.(*pyast.Try); isTry {
			t.Errorf("the try around an import of a name the library module defines should be gone")
		}
	}
	// A module of the project may be half done (an import cycle): the
	// fallback stays.
	mod, err = project(t, map[string]string{"main.py": strings.ReplaceAll(src, "gettext import", "helper import"), "helper.py": "def gettext(m):\n    return m\n"})
	if err != nil {
		t.Fatal(err)
	}
	kept := false
	for _, st := range mod.Body {
		if _, isTry := st.(*pyast.Try); isTry {
			kept = true
		}
	}
	if !kept {
		t.Errorf("the try around an import from a project module should stay")
	}
}
