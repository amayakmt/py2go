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
	l.addValueMod(m)
	return &pyast.Call{
		Func: &pyast.Name{ID: "_native", Lineno: line},
		Args: []pyast.Node{
			&pyast.Constant{PyType: "str", Value: "module.object", Lineno: line},
			&pyast.Constant{PyType: "str", Value: m.name, Lineno: line},
		},
		Lineno: line,
	}, nil
}

// addValueMod records that the program uses a module as a value.
func (l *linker) addValueMod(m *module) {
	for _, v := range l.valueMods {
		if v == m {
			return
		}
	}
	l.valueMods = append(l.valueMods, m)
}

// moduleRegistrations are the statements, at the start of the program, that
// tell the runtime how to read each top-level name of the modules used as
// values. They read the names when asked, not now.
func (l *linker) moduleRegistrations() []pyast.Node {
	var out []pyast.Node
	if _, ok := l.mods["importlib"]; ok {
		// importlib.import_module(name) can name any module the program
		// links: all of them are values.
		var names []string
		for n, m := range l.mods {
			// Modules only looked at (by the pre-pass of import fallbacks)
			// are not in the program.
			if !m.entry && !m.builtins && m.started {
				names = append(names, n)
			}
		}
		sort.Strings(names)
		for _, n := range names {
			m := l.mods[n]
			if !m.namespace {
				l.addValueMod(m)
			}
			out = append(out, &pyast.Expr{Value: &pyast.Call{
				Func: &pyast.Name{ID: "_native"},
				Args: []pyast.Node{&pyast.Constant{PyType: "str", Value: "module.declare"}, &pyast.Constant{PyType: "str", Value: n}},
			}})
		}
	}
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
	l.addValueMod(m)
	return &pyast.Call{
		Func: &pyast.Name{ID: "_native", Lineno: line},
		Args: []pyast.Node{
			&pyast.Constant{PyType: "str", Value: "module.globals", Lineno: line},
			&pyast.Constant{PyType: "str", Value: m.name, Lineno: line},
		},
		Lineno: line,
	}
}
