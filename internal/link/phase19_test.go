package link

import "testing"

func TestUnavailableModulesRaiseModuleNotFound(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "try:\n    import msvcrt\nexcept ImportError:\n    pass\ntry:\n    from tkinter import ttk\nexcept ImportError:\n    pass\ntry:\n    import ctypes.util\nexcept ImportError:\n    pass\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !names(mod)["ModuleNotFoundError"] {
		t.Errorf("the imports of unavailable modules should raise ModuleNotFoundError")
	}
}

func TestUnavailableName(t *testing.T) {
	for in, want := range map[string]string{"msvcrt": "msvcrt", "tkinter.ttk": "_tkinter", "turtle": "_tkinter", "curses.panel": "_curses", "ctypes.util": "_ctypes", "nt": "nt"} {
		got, ok := unavailableModule(in)
		if !ok || got != want {
			t.Errorf("unavailableModule(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if _, ok := unavailableModule("os.path"); ok {
		t.Errorf("os.path is available")
	}
}

func TestAProjectModuleWinsOverAnUnavailableOne(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import nt\nprint(nt.X)\n", "nt.py": "X = 1\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !names(mod)["nt_pm_X"] {
		t.Errorf("the project's nt module was not used")
	}
}

func TestLinkJoinsTheWindowsPathModules(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import ntpath\nimport nturl2path\n\nprint(ntpath.join('a', 'b'), nturl2path.pathname2url('C:\\\\x'))\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"ntpath_pm_join", "nturl2path_pm_pathname2url"} {
		if !got[want] {
			t.Errorf("%s is not in the joined module", want)
		}
	}
}
