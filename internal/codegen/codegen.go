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
	defined map[string]bool // functions whose def has been passed so far
	mainEnv *scope          // the scope of main(), for better error messages

	// State of the function (or main) being generated.
	env       *scope
	fnName    string // "" while generating main()
	ret       pyType // return type of the current function
	loopDepth int    // enclosing loops, for break and continue
	nextID    int    // counter for generated names
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
	}
	g.collectSignatures(mod)

	g.mainEnv = newScope(nil)
	g.env = g.mainEnv
	var funcs, body strings.Builder

	// Function definitions go to package level.
	// Everything else goes into main(), in source order.
	for _, stmt := range mod.Body {
		var err error
		if fn, ok := stmt.(*pyast.FunctionDef); ok {
			err = g.funcDef(&funcs, fn)
		} else {
			err = g.stmt(&body, stmt)
		}
		if err != nil {
			return nil, err
		}
	}

	var out strings.Builder
	out.WriteString("package main\n\n")
	if g.usesRuntime {
		fmt.Fprintf(&out, "import %q\n\n", runtimeImport)
	}
	out.WriteString(funcs.String())
	out.WriteString("func main() {\n")
	out.WriteString(body.String())
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

// collectSignatures records every module-level function's signature.
// A bad signature is stored, not returned, so errors still surface in
// source order when the offending def or call is reached.
func (g *generator) collectSignatures(mod *pyast.Module) {
	for _, stmt := range mod.Body {
		fn, ok := stmt.(*pyast.FunctionDef)
		if !ok {
			continue
		}
		if _, dup := g.sigs[fn.Name]; dup {
			continue
		}
		if _, dup := g.sigErrs[fn.Name]; dup {
			continue
		}
		sig, err := signature(fn)
		if err != nil {
			g.sigErrs[fn.Name] = err
			continue
		}
		g.sigs[fn.Name] = sig
	}
}

// signature validates a def line and returns its parameter and return types.
func signature(fn *pyast.FunctionDef) (*funcSig, error) {
	if fn.Name == "main" {
		return nil, unsupported("function named main (collides with generated main)", fn.Line())
	}
	if err := checkName(fn.Name, fn.Line()); err != nil {
		return nil, err
	}
	if fn.HasDecorators {
		return nil, unsupported("decorators", fn.Line())
	}
	if fn.HasTypeParams {
		return nil, unsupported("generic functions", fn.Line())
	}

	a := fn.Args
	if a.Vararg != nil {
		return nil, unsupported("*args parameter", a.Vararg.Line())
	}
	if len(a.Kwonly) > 0 {
		return nil, unsupported("keyword-only parameters", a.Kwonly[0].Line())
	}
	if a.Kwarg != nil {
		return nil, unsupported("**kwargs parameter", a.Kwarg.Line())
	}
	if len(a.Defaults) > 0 {
		return nil, unsupported("default parameter values", a.Defaults[0].Line())
	}

	sig := &funcSig{}
	seen := map[string]bool{}
	for _, p := range a.Positional() {
		if err := checkName(p.Name, p.Line()); err != nil {
			return nil, err
		}
		if seen[p.Name] {
			return nil, unsupported(fmt.Sprintf("duplicate parameter '%s'", p.Name), p.Line())
		}
		seen[p.Name] = true
		if p.Annotation == nil {
			return nil, unsupported(fmt.Sprintf("parameter '%s' without type annotation", p.Name), p.Line())
		}
		t, err := annotationType(p.Annotation, false)
		if err != nil {
			return nil, err
		}
		sig.params = append(sig.params, t)
	}

	if fn.Returns == nil {
		return nil, unsupported("function without return annotation", fn.Line())
	}
	ret, err := annotationType(fn.Returns, true)
	if err != nil {
		return nil, err
	}
	sig.ret = ret
	return sig, nil
}

func (g *generator) funcDef(b *strings.Builder, fn *pyast.FunctionDef) error {
	if g.defined[fn.Name] {
		return unsupported(fmt.Sprintf("redefinition of function '%s'", fn.Name), fn.Line())
	}
	if err := g.sigErrs[fn.Name]; err != nil {
		return err
	}
	sig := g.sigs[fn.Name]
	g.defined[fn.Name] = true

	// A function body is its own scope chain. Its names come from the
	// parameters and its own declarations only.
	g.env = newScope(nil)
	g.fnName, g.ret = fn.Name, sig.ret
	defer func() { g.env, g.fnName, g.ret = g.mainEnv, "", "" }()

	params := make([]string, 0, len(sig.params))
	for i, p := range fn.Args.Positional() {
		g.env.declare(p.Name, sig.params[i])
		params = append(params, p.Name+" "+g.goType(sig.params[i]))
	}
	result := ""
	if sig.ret != tNone {
		result = " " + g.goType(sig.ret)
	}

	fmt.Fprintf(b, "func %s(%s)%s {\n", fn.Name, strings.Join(params, ", "), result)
	for _, s := range fn.Body {
		if err := g.stmt(b, s); err != nil {
			return err
		}
	}
	// Go insists that a function with a result ends in a terminating
	// statement. Python would return None here, which breaks the annotation.
	if _, endsInReturn := fn.Body[len(fn.Body)-1].(*pyast.Return); sig.ret != tNone && !endsInReturn {
		fmt.Fprintf(b, "panic(%s(%q))\n", g.rt("MissingReturn"), fn.Name)
	}
	b.WriteString("}\n\n")
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
