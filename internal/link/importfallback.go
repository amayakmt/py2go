package link

import (
	"path/filepath"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// dropImportFallbacks rewrites
//
//	try:
//	    import gzip
//	except ImportError:
//	    gzip = None
//
// to the imports alone when every module they name exists, and every name a
// "from m import a" takes is defined at the top of m: the handler cannot
// run, and the fallback would rebind the names to other things (a module
// to None, an imported function to a def of the module).
func dropImportFallbacks(mod *pyast.Module, exists func(string) bool, defines func(mod, name string) bool) {
	var inBody func(body []pyast.Node) []pyast.Node
	inBody = func(body []pyast.Node) []pyast.Node {
		var out []pyast.Node
		for _, st := range body {
			switch x := st.(type) {
			case *pyast.Try:
				if importFallback(x, exists, defines) {
					out = append(out, x.Body...)
					out = append(out, inBody(x.Orelse)...)
					out = append(out, inBody(x.Finalbody)...)
					continue
				}
			case *pyast.FunctionDef:
				x.Body = inBody(x.Body)
			case *pyast.If:
				x.Body = inBody(x.Body)
				x.Orelse = inBody(x.Orelse)
			}
			out = append(out, st)
		}
		return out
	}
	mod.Body = inBody(mod.Body)
}

// importFallback reports whether t is a try of imports that succeed, whose
// handlers only catch ImportError.
func importFallback(t *pyast.Try, exists func(string) bool, defines func(mod, name string) bool) bool {
	if len(t.Body) == 0 || len(t.Handlers) == 0 {
		return false
	}
	for _, st := range t.Body {
		switch imp := st.(type) {
		case *pyast.Import:
			for _, a := range imp.Names {
				if !exists(a.Name) {
					return false
				}
			}
		case *pyast.ImportFrom:
			if imp.Level != 0 || imp.Module == "" {
				return false
			}
			for _, a := range imp.Names {
				if a.Name == "*" || !defines(imp.Module, a.Name) {
					return false
				}
			}
		default:
			return false
		}
	}
	for _, h := range t.Handlers {
		names := []pyast.Node{h.Type}
		if tup, ok := h.Type.(*pyast.Tuple); ok {
			names = tup.Elts
		}
		for _, n := range names {
			nm, ok := n.(*pyast.Name)
			if !ok || nm.ID != "ImportError" && nm.ID != "ModuleNotFoundError" {
				return false
			}
		}
	}
	return true
}

// moduleDefines reports whether the library module of that dotted name
// binds name at its top level. Modules of the project are not trusted with
// it: they can import each other in a cycle, and then "from m import a"
// fails while m is half done, as CPython's does. Nor is a module whose own
// pre-passes are running, or that has started and not finished.
func (l *linker) moduleDefines(mod, name string) bool {
	top := strings.Split(mod, ".")[0]
	if l.loading[mod] || existsIn(l.root, top) || !l.moduleExists(mod) {
		return false
	}
	m, err := l.find(mod)
	if err != nil || m == nil || m.bound == nil || m.started && !m.finished {
		return false
	}
	return m.bound[name]
}

// moduleExists reports whether the project or the library has a module of
// that dotted name, without loading it.
func (l *linker) moduleExists(name string) bool {
	parts := strings.Split(name, ".")
	for _, root := range []string{l.root, stdlibRoot()} {
		if root == "" || !existsIn(root, parts[0]) {
			continue
		}
		dir := root
		for i, part := range parts {
			sub := filepath.Join(dir, part)
			if i == len(parts)-1 {
				return fileExists(sub+".py") || dirExists(sub)
			}
			if !dirExists(sub) {
				return false
			}
			dir = sub
		}
	}
	return false
}
