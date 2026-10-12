// Package codegen turns a pyast tree into Go source.
package codegen

import (
	"fmt"
	"go/format"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

const runtimeImport = "github.com/amayakmt/py2go/runtime"

// generator holds state that accumulates while walking the tree.
type generator struct {
	usesRuntime bool // set when generated code calls runtime.X, so we import it

	// Module-level function signatures, collected before any body is
	// generated so functions can call each other in any order.
	sigs    map[string]*funcSig
	sigErrs map[string]error
	defined map[string]bool            // functions whose def has been passed so far
	mainEnv *scope                     // the scope of main(), for better error messages
	calls   map[string]map[string]bool // which module-level function calls which

	// State of the function (or main) being generated.
	env       *scope
	fnName    string   // "" while generating main()
	ret       pyType   // return type of the current function
	loopDepth int      // enclosing loops, for break and continue
	nextID    int      // counter for generated names
	handlers  []string // Go variables of the exceptions being handled, innermost last
	flow      *flowCtx // the closure being generated for a try body, or nil

	// Functions.
	topFn        string          // the top-level function being generated, for the call graph
	fnNonlocals  map[string]bool // names the current function declared nonlocal
	callsUnknown bool            // the code being generated calls a function it cannot resolve
	defaultVars  []defaultVar    // package variables holding top-level defaults

	// Module-level variables. They are package-level Go variables, so that
	// functions can read them and, after "global", assign them.
	globalTypes map[string]pyType // annotated module-level names found before generation
	globalLines map[string]int    // where each was first annotated
	globalDecl  map[string]pyType // those declared so far, in declaration order
	globalOrder []string
	globalReads map[string]bool // read by some function: they need a "was it set" flag
	fnGlobals   map[string]bool // names the current function declared global
	compRoot    *scope          // the scope of the comprehension being generated, or nil
}

// Generate converts a decoded Python module into formatted Go source.
func Generate(root pyast.Node) ([]byte, error) {
	mod, ok := root.(*pyast.Module)
	if !ok {
		return nil, fmt.Errorf("codegen: root must be Module, got %T", root)
	}

	g := &generator{
		sigs:    map[string]*funcSig{},
		sigErrs: map[string]error{},
		defined: map[string]bool{},
		calls:   map[string]map[string]bool{},

		globalTypes: map[string]pyType{},
		globalLines: map[string]int{},
		globalDecl:  map[string]pyType{},
		globalReads: map[string]bool{},
		fnGlobals:   map[string]bool{},
		fnNonlocals: map[string]bool{},
	}
	g.collectSignatures(mod)
	g.decoratedTypes()
	if err := g.collectGlobals(mod); err != nil {
		return nil, err
	}

	g.mainEnv = newScope(nil)
	g.env = g.mainEnv
	var funcs, body strings.Builder

	// Function definitions go to package level.
	// Everything else goes into main(), in source order.
	for _, stmt := range mod.Body {
		var err error
		if fn, ok := stmt.(*pyast.FunctionDef); ok {
			err = g.funcDef(&funcs, &body, fn)
		} else {
			err = g.stmt(&body, stmt)
		}
		if err != nil {
			return nil, err
		}
	}

	funcText := g.placeRecursionGuards(funcs.String())
	funcText, bodyText := g.placeGlobalFlags(funcText), g.placeGlobalFlags(body.String())
	globals := g.globalDeclarations()

	var out strings.Builder
	out.WriteString("package main\n\n")
	if g.usesRuntime {
		fmt.Fprintf(&out, "import %q\n\n", runtimeImport)
	}
	out.WriteString(globals)
	out.WriteString(funcText)
	out.WriteString("func main() {\n")
	if g.usesRuntime {
		// An exception that nothing caught ends the program the way CPython
		// does: a message on stderr and exit status 1.
		out.WriteString("defer " + g.rt("Report") + "()\n")
	}
	out.WriteString(bodyText)
	out.WriteString("}\n")

	// go/format fixes indentation and fails on invalid syntax,
	// so a generator bug surfaces here instead of at go build.
	src, err := format.Source([]byte(out.String()))
	if err != nil {
		return nil, fmt.Errorf("codegen produced invalid Go: %w\n%s", err, out.String())
	}
	return src, nil
}

// rt returns the qualified name of a runtime helper and marks the import used.
func (g *generator) rt(name string) string {
	g.usesRuntime = true
	return "runtime." + name
}

// goType spells a type in Go and notes when that needs the runtime import.
func (g *generator) goType(t pyType) string {
	s := t.goType()
	if strings.Contains(s, "runtime.") {
		g.usesRuntime = true
	}
	return s
}

// tmp returns a fresh name for generated variables. The _py prefix is
// rejected for user identifiers, so these never collide with them.
func (g *generator) tmp(kind string) string {
	g.nextID++
	return fmt.Sprintf("_py%d_%s", g.nextID, kind)
}

// goReserved holds names that cannot be used for Python identifiers: Go
// keywords that are legal in Python, Go names the generated code relies on,
// and the Python builtins this transpiler special-cases.
var goReserved = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(`
		chan const default defer fallthrough func go goto interface map
		package range select struct switch type var
		bool float64 int64 string true false nil panic runtime init _
		print`) {
		m[w] = true
	}
	return m
}()

// checkName rejects identifiers that would break or hijack the Go output.
// Renaming them is future work, so for now they fail loudly.
func checkName(name string, line int) error {
	if goReserved[name] {
		return unsupported(fmt.Sprintf("identifier '%s' (reserved in Go output)", name), line)
	}
	if strings.HasPrefix(name, "_py") {
		return unsupported(fmt.Sprintf("identifier '%s' (the _py prefix is reserved)", name), line)
	}
	return nil
}

func unsupported(what string, line int) error {
	return fmt.Errorf("unsupported: %s at line %d", what, line)
}

// mismatch reports a value whose type differs from what its destination
// requires. Python would accept an int where a float is annotated and keep
// it an int, which Go cannot do, so no implicit conversion happens.
func mismatch(dest string, want, got pyType, line int) error {
	msg := fmt.Sprintf("%s expects %s, got %s", dest, want, got)
	if want == tFloat && got == tInt {
		msg += " (no implicit int to float conversion, write a float literal)"
	}
	return unsupported(msg, line)
}

// nodeName turns "*pyast.FunctionDef" into "FunctionDef" for error messages.
func nodeName(n pyast.Node) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", n), "*pyast.")
}

// guardMarker stands in for a function's recursion guard until every body
// has been generated and the call graph is known.
func guardMarker(name string) string { return "/*recursion-guard:" + name + "*/" }

// placeRecursionGuards fills in the markers. A function that can reach
// itself through calls counts its depth, so runaway recursion raises
// RecursionError as in Python instead of overflowing the Go stack. Other
// functions pay nothing.
func (g *generator) placeRecursionGuards(text string) string {
	reaches := func(from, target string) bool {
		seen := map[string]bool{}
		var visit func(f string) bool
		visit = func(f string) bool {
			for callee := range g.calls[f] {
				if callee == target {
					return true
				}
				if !seen[callee] {
					seen[callee] = true
					if visit(callee) {
						return true
					}
				}
			}
			return false
		}
		return visit(from)
	}
	for name := range g.defined {
		marker := guardMarker(name) + "\n"
		guard := ""
		if g.calls[name]["*"] || reaches(name, name) {
			guard = g.rt("Enter") + "()\ndefer " + g.rt("Leave") + "()\n"
		}
		text = strings.Replace(text, marker, guard, 1)
	}
	return text
}
