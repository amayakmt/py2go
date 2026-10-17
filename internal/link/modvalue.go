package link

import (
	"fmt"
	"sort"

	"github.com/amayakmt/py2go/internal/pyast"
)

// moduleValue is a module used as a value (hasattr(os, "fork"), f(sys),
// x = os.path): a call to the native that returns the module object, whose
// attributes the program registers (see moduleRegistrations).
func (l *linker) moduleValue(m *module, line int) (pyast.Node, error) {
	if m.entry || m.builtins {
		return nil, fmt.Errorf("unsupported: module '%s' used as a value at line %d", m.name, line)
	}
	known := false
	for _, v := range l.valueMods {
		if v == m {
			known = true
		}
	}
	if !known {
		l.valueMods = append(l.valueMods, m)
	}
	return &pyast.Call{
		Func: &pyast.Name{ID: "_native", Lineno: line},
		Args: []pyast.Node{
			&pyast.Constant{PyType: "str", Value: "module.object", Lineno: line},
			&pyast.Constant{PyType: "str", Value: m.name, Lineno: line},
		},
		Lineno: line,
	}, nil
}

// moduleRegistrations are the statements, at the start of the program, that
// tell the runtime how to read each top-level name of the modules used as
// values. They read the names when asked, not now.
func (l *linker) moduleRegistrations() []pyast.Node {
	var out []pyast.Node
	for _, m := range l.valueMods {
		names := make([]string, 0, len(m.bound))
		for n := range m.bound {
			names = append(names, n)
		}
		sort.Strings(names)
		// Imported names are globals too: modules as module objects, names
		// from "from m import x" as what they stand for.
		var aliased []string
		for n := range m.aliases {
			if !m.bound[n] {
				aliased = append(aliased, n)
			}
		}
		sort.Strings(aliased)
		for _, n := range aliased {
			a := m.aliases[n]
			if a.mod == nil || a.mod.builtins {
				continue
			}
			var val pyast.Node
			if a.name == "" {
				if a.mod.entry {
					continue
				}
				val = &pyast.Call{Func: &pyast.Name{ID: "_native"}, Args: []pyast.Node{
					&pyast.Constant{PyType: "str", Value: "module.object"},
					&pyast.Constant{PyType: "str", Value: a.mod.name},
				}}
			} else {
				val = &pyast.Name{ID: a.mod.prefixOf() + a.name}
			}
			out = append(out, &pyast.Expr{Value: &pyast.Call{
				Func: &pyast.Name{ID: "_native"},
				Args: []pyast.Node{
					&pyast.Constant{PyType: "str", Value: "module.register"},
					&pyast.Constant{PyType: "str", Value: m.name},
					&pyast.Constant{PyType: "str", Value: n},
					&pyast.Lambda{Args: &pyast.Arguments{}, Body: val},
				},
			}})
		}
		for _, n := range names {
			out = append(out, &pyast.Expr{Value: &pyast.Call{
				Func: &pyast.Name{ID: "_native"},
				Args: []pyast.Node{
					&pyast.Constant{PyType: "str", Value: "module.register"},
					&pyast.Constant{PyType: "str", Value: m.name},
					&pyast.Constant{PyType: "str", Value: n},
					&pyast.Lambda{Args: &pyast.Arguments{}, Body: &pyast.Name{ID: m.prefixOf() + n}},
				},
			}})
		}
	}
	return out
}

// globalsCall is globals() in module m: a dict made, when it runs, from the
// top-level names of m (a copy: assigning to it does not create globals).
func (l *linker) globalsCall(m *module, line int) pyast.Node {
	known := false
	for _, v := range l.valueMods {
		if v == m {
			known = true
		}
	}
	if !known {
		l.valueMods = append(l.valueMods, m)
	}
	return &pyast.Call{
		Func: &pyast.Name{ID: "_native", Lineno: line},
		Args: []pyast.Node{
			&pyast.Constant{PyType: "str", Value: "module.globals", Lineno: line},
			&pyast.Constant{PyType: "str", Value: m.name, Lineno: line},
		},
		Lineno: line,
	}
}
