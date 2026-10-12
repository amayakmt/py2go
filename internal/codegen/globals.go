package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Module-level variables live at package level in the generated Go, so that
// functions can read them, and rebind them after a "global" statement. Their
// types come from annotations, which are found before any code is generated,
// since a function may be defined above the variable it reads.

// collectGlobals finds the annotated names of the module, at any depth of
// if, loops and try, but not inside functions.
func (g *generator) collectGlobals(mod *pyast.Module) error {
	var visit func(stmts []pyast.Node)
	visit = func(stmts []pyast.Node) {
		for _, stmt := range stmts {
			switch n := stmt.(type) {
			case *pyast.AnnAssign:
				name, ok := n.Target.(*pyast.Name)
				if !ok || !n.Simple {
					continue
				}
				if _, seen := g.globalTypes[name.ID]; seen {
					continue
				}
				if t, err := g.annotationType(n.Annotation, false); err == nil {
					g.globalTypes[name.ID] = t
					g.globalLines[name.ID] = n.Line()
				}
			case *pyast.If:
				visit(n.Body)
				visit(n.Orelse)
			case *pyast.While:
				visit(n.Body)
			case *pyast.For:
				// The variable of "for i in range(...)" is an int, which is
				// the one loop variable whose type is known without running
				// the code generator.
				if name, ok := n.Target.(*pyast.Name); ok {
					if call, isCall := n.Iter.(*pyast.Call); isCall {
						if fn, _ := callName(call); fn == "range" {
							if _, seen := g.globalTypes[name.ID]; !seen {
								g.globalTypes[name.ID] = tInt
								g.globalLines[name.ID] = n.Line()
							}
						}
					}
				}
				visit(n.Body)
			case *pyast.Try:
				visit(n.Body)
				for _, h := range n.Handlers {
					visit(h.Body)
				}
				visit(n.Orelse)
				visit(n.Finalbody)
			}
		}
	}
	visit(mod.Body)

	// A Go package has one namespace for functions and variables.
	for name, line := range g.globalLines {
		if _, isClass := g.classes[name]; isClass {
			return unsupported(fmt.Sprintf("'%s' is both a class and a module-level variable", name), line)
		}
		if sig, isFunc := g.sigs[name]; isFunc {
			if len(sig.decorators) > 0 {
				continue // the decorated name is itself a variable
			}
			return unsupported(fmt.Sprintf("'%s' is both a function and a module-level variable", name), line)
		}
		if _, isFunc := g.sigErrs[name]; isFunc {
			return unsupported(fmt.Sprintf("'%s' is both a function and a module-level variable", name), line)
		}
	}
	return nil
}

// atModuleLevel reports whether declarations now create module variables:
// outside functions, and outside the scope of a comprehension.
func (g *generator) atModuleLevel() bool { return g.fnName == "" && g.compRoot == nil }

// flagName is the Go variable that records that a module variable was set.
func flagName(name string) string { return "_pyg_" + name }

// flagMarker stands in for "flag = true" until it is known whether any
// function reads the variable, which needs the flag.
func flagMarker(name string) string { return "/*gset:" + name + "*/" }

// markGlobalSet notes that a module variable has just been given a value.
func (g *generator) markGlobalSet(b *strings.Builder, name string) {
	b.WriteString(flagMarker(name) + "\n")
}

// placeGlobalFlags turns the markers into assignments for the variables
// that functions read, and drops the rest.
func (g *generator) placeGlobalFlags(text string) string {
	for _, name := range g.globalOrder {
		repl := ""
		if g.globalReads[name] {
			repl = flagName(name) + " = true\n"
		}
		text = strings.ReplaceAll(text, flagMarker(name)+"\n", repl)
	}
	return text
}

// globalDeclarations emits the package-level variables.
func (g *generator) globalDeclarations() string {
	var b strings.Builder
	for _, d := range g.defaultVars {
		fmt.Fprintf(&b, "var %s %s\n", d.name, g.goType(d.typ))
	}
	for _, name := range g.globalOrder {
		fmt.Fprintf(&b, "var %s %s\n", name, g.goType(g.globalDecl[name]))
		if g.globalReads[name] {
			fmt.Fprintf(&b, "var %s bool\n", flagName(name))
		}
	}
	if b.Len() > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

// declareGlobal makes name a module variable. The first declaration emits
// the assignment; later ones in other branches share the variable.
func (g *generator) declareGlobal(b *strings.Builder, name string, t pyType, init string) {
	g.mainEnv.declare(name, t)
	if _, done := g.globalDecl[name]; !done {
		g.globalDecl[name] = t
		g.globalOrder = append(g.globalOrder, name)
	}
	if init != "" {
		fmt.Fprintf(b, "%s = %s\n", name, init)
	}
	g.markGlobalSet(b, name)
}

// lookupBinding finds the variable that assigning to name would change, in
// the current scopes and, inside a function, among the names it declared
// global. Inside a comprehension only the comprehension's own scopes count,
// since its loop variables shadow outer names instead of assigning to them.
func (g *generator) lookupBinding(name string) (t pyType, found, global bool) {
	if g.compRoot != nil {
		for s := g.env; s != nil; s = s.parent {
			if t, ok := s.vars[name]; ok {
				return t, true, false
			}
			if s == g.compRoot {
				break
			}
		}
		return "", false, false
	}
	if t, ok := g.env.lookupWithinFunction(name); ok {
		return t, true, false
	}
	if g.fnNonlocals[name] {
		if root := g.env.functionRoot(); root != nil && root.parent != nil {
			if t, ok := root.parent.lookup(name); ok {
				return t, true, false
			}
		}
	}
	if g.fnName != "" && g.fnGlobals[name] {
		if t, ok := g.globalTypes[name]; ok {
			return t, true, true
		}
	}
	return "", false, false
}

// afterAssign records, for an assignment inside a function to a variable it
// declared global, that the variable now has a value.
func (g *generator) afterAssign(b *strings.Builder, name string, global bool) {
	g.env.dropAlias(name)
	if global {
		g.markGlobalSet(b, name)
	}
}

// readGlobal generates a read of a module variable from inside a function.
// The function may run before the module-level assignment has, which Python
// reports as a NameError, so the read checks the variable's flag.
func (g *generator) readGlobal(name string) (value, bool) {
	t, ok := g.globalTypes[name]
	if !ok || g.fnName == "" {
		return value{}, false
	}
	g.globalReads[name] = true
	code := g.rt("Bound") + "(" + name + ", " + flagName(name) + ", " + strconv.Quote(name) + ")"
	return value{code: code, typ: t, prec: precAtom}, true
}

// globalStmt is "global a, b". At module level it does nothing.
func (g *generator) globalStmt(n *pyast.Global) error {
	if g.fnName == "" {
		return nil
	}
	for _, name := range n.Names {
		if _, ok := g.globalTypes[name]; !ok {
			return unsupported(fmt.Sprintf("global '%s', which has no type annotation at module level", name), n.Line())
		}
		if _, local := g.env.lookup(name); local {
			return unsupported(fmt.Sprintf("name '%s' is assigned to before global declaration", name), n.Line())
		}
		g.fnGlobals[name] = true
	}
	return nil
}
