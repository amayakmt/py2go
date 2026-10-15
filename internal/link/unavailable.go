package link

import "strings"

// unavailable lists the modules that cannot be imported, with the name
// CPython reports when its build lacks them. The Windows-only modules are
// missing everywhere else, as in CPython; the graphical, terminal and foreign
// function modules need C libraries (Tcl/Tk, ncurses, libffi), which a
// program compiled to Go without cgo does not have.
var unavailable = map[string]string{
	"msvcrt": "msvcrt", "winreg": "winreg", "winsound": "winsound", "_winapi": "_winapi",
	"_overlapped": "_overlapped", "_wmi": "_wmi", "nt": "nt", "msilib": "msilib", "_winreg": "_winreg",
	"tkinter": "_tkinter", "_tkinter": "_tkinter", "turtle": "_tkinter", "turtledemo": "_tkinter", "idlelib": "_tkinter",
	"curses": "_curses", "_curses": "_curses", "_curses_panel": "_curses",
	"ctypes": "_ctypes", "_ctypes": "_ctypes",
}

// unavailableModule tells whether importing the dotted name fails, and the
// module its ModuleNotFoundError names.
func unavailableModule(name string) (string, bool) {
	top := name
	if i := strings.IndexByte(name, '.'); i >= 0 {
		top = name[:i]
	}
	missing, ok := unavailable[top]
	return missing, ok
}

// unavailable is unavailableModule unless the project has a module of that name.
func (l *linker) unavailable(name string) (string, bool) {
	missing, ok := unavailableModule(name)
	if !ok {
		return "", false
	}
	top := strings.Split(name, ".")[0]
	if m, err := l.find(top); err == nil && m != nil {
		return "", false
	}
	return missing, true
}
