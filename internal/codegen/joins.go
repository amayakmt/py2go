package codegen

import (
	"fmt"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Variables without an annotation take the type of what they are first
// assigned. A name that is assigned values of different types in a function
// gets the join of them, which is object when nothing closer fits. The join is
// found by compiling the function body once or more with the output thrown
// away (recordJoins), noting every assignment, until the types stop changing.

// scopeKey names the function or module being compiled.
func (g *generator) scopeKey() string { return g.joinScope }

// joinType is the type that holds values of both types.
func (g *generator) joinType(a, b pyType) pyType {
	if a == b {
		return a
	}
	if a == "" {
		return b
	}
	// A container that was found to hold other items than it was given.
	if w, ok := widen(a); ok && w == b {
		return b
	}
	if w, ok := widen(b); ok && w == a {
		return a
	}
	if j, ok := g.joinTypes(a, b); ok {
		return j
	}
	return tObject
}

// bindingScope is the scope whose variable name is: the module's for a name a
// function declared global.
func (g *generator) bindingScope(name string) string {
	if g.fnName != "" && g.fnGlobals[name] {
		return moduleScope
	}
	return g.joinScope
}

// noteBinding records that name is assigned a value of type t.
func (g *generator) noteBinding(name string, t pyType) {
	scope := g.bindingScope(name)
	if !g.recordJoins || g.annotatedNames[scope][name] {
		return
	}
	m := g.joins[scope]
	if m == nil {
		m = map[string]pyType{}
		g.joins[scope] = m
	}
	cur := m[name]
	next := g.joinType(cur, t)
	if next != cur {
		m[name] = next
		g.joinsChanged = true
		if scope == moduleScope {
			g.moduleJoinsVersion++
		}
	}
}

// declType is the type a variable is declared with when it is first assigned
// a value of type t.
func (g *generator) declType(name string, t pyType) pyType {
	if g.annotatedNames[g.joinScope][name] {
		return t
	}
	if j, ok := g.joins[g.joinScope][name]; ok && j != "" {
		return j
	}
	return t
}

// markAnnotated says that name has a declared type in the current scope.
func (g *generator) markAnnotated(name string) {
	m := g.annotatedNames[g.joinScope]
	if m == nil {
		m = map[string]bool{}
		g.annotatedNames[g.joinScope] = m
	}
	m[name] = true
}

// convertInit turns the code of a value of type from into one of type to.
func (g *generator) convertInit(code string, from, to pyType) string {
	if from == to || code == "" {
		return code
	}
	switch {
	case to == tObject && (from == tInt || from == tFloat || from == tBool || from == tStr):
		// A constant gets its Go type before it is boxed.
		return g.goType(from) + "(" + code + ")"
	case from.isList() && to.isList(), from.isSet() && to.isSet(), from.isDict() && to.isDict():
		return g.rt("DynDeepAs") + "[" + g.goType(to) + "](" + code + ")"
	}
	return g.relabel(value{code: code, typ: from, prec: precAtom}, to).code
}

// mismatchOK is called where an assignment gives a variable a value of
// another type than it has. While joins are being recorded that is how the
// join grows; otherwise it is an error the caller reports.
func (g *generator) mismatchOK(name string, got pyType) bool {
	if !g.recordJoins || g.annotatedNames[g.bindingScope(name)][name] {
		return false
	}
	g.noteBinding(name, got)
	return true
}

// hasPlainBindings reports whether a function body binds a name without a
// type hint, which is when its variables need their types worked out.
func hasPlainBindings(body []pyast.Node) bool {
	found := false
	for _, st := range body {
		walkPrune(st, func(n pyast.Node) bool {
			switch x := n.(type) {
			case *pyast.Assign, *pyast.For, *pyast.With, *pyast.NamedExpr, *pyast.Match, *pyast.AugAssign:
				found = true
			case *pyast.Try:
				for _, h := range x.Handlers {
					if h.Name != "" {
						found = true
					}
				}
			case *pyast.FunctionDef, *pyast.Lambda, *pyast.ClassDef:
				return false
			}
			return !found
		})
	}
	return found
}

// withJoins compiles a body until the types of its variables settle, and then
// once for real. gen compiles the body and returns its code.
func (g *generator) withJoins(key string, plain bool, gen func() (string, error)) (string, error) {
	savedScope := g.joinScope
	g.joinScope = key
	defer func() { g.joinScope = savedScope }()
	if plain && !g.recordJoins {
		for round := 0; round < 5; round++ {
			g.joinsChanged = false
			g.discard(func() { _, _ = gen() })
			if !g.joinsChanged {
				break
			}
		}
	}
	return gen()
}

// discard runs f in recording mode, keeping nothing it does to the generator
// that a second run would repeat.
func (g *generator) discard(f func()) {
	usesRuntime := g.usesRuntime
	savedInfer, savedRecord, savedNext := g.infer, g.recordJoins, g.nextID
	savedDefaults := len(g.defaultVars)
	g.infer, g.recordJoins = true, true
	defer func() {
		g.infer, g.recordJoins = savedInfer, savedRecord
		g.usesRuntime = usesRuntime
		g.nextID = savedNext
		g.defaultVars = g.defaultVars[:savedDefaults]
	}()
	f()
}

func (g *generator) joinReport() string {
	var b strings.Builder
	for k, m := range g.joins {
		for n, t := range m {
			fmt.Fprintf(&b, "%s.%s: %s\n", k, n, t)
		}
	}
	return b.String()
}

// moduleScope is the join scope of the statements at module level.
const moduleScope = "<module>"

// inferModuleVars types the module variables that have no annotation, so that
// functions can read them. The statements outside defs and classes are
// compiled with the output thrown away, as for a function's variables, and the
// variables they declare are added to the module's. The functions are compiled
// the same way, since they may assign a module variable through "global" a
// value of another type than the one it first had (config = None, then a dict).
func (g *generator) inferModuleVars(mod *pyast.Module) {
	var stmts []pyast.Node
	for _, st := range mod.Body {
		switch st.(type) {
		case *pyast.FunctionDef, *pyast.ClassDef:
		default:
			stmts = append(stmts, st)
		}
	}
	g.joinScope = moduleScope
	if !hasPlainBindings(stmts) {
		return
	}
	inferred := map[string]bool{}
	for round := 0; round < 4; round++ {
		for name, t := range g.moduleVarsPass(mod) {
			if _, typed := g.globalTypes[name]; !typed || inferred[name] {
				g.globalTypes[name] = t
				g.globalLines[name] = 0
				inferred[name] = true
			}
		}
		before := g.moduleJoinsVersion
		g.functionsPass(mod)
		if g.moduleJoinsVersion == before {
			break
		}
	}
}

// moduleVarsPass compiles the module-level statements until the types of
// the variables settle, and returns the variables they declare.
func (g *generator) moduleVarsPass(mod *pyast.Module) map[string]pyType {
	var stmts []pyast.Node
	for _, st := range mod.Body {
		switch st.(type) {
		case *pyast.FunctionDef, *pyast.ClassDef:
		default:
			stmts = append(stmts, st)
		}
	}
	savedEnv, savedMain := g.env, g.mainEnv
	savedDecl, savedOrder := g.globalDecl, g.globalOrder
	found := map[string]pyType{}
	for round := 0; round < 5; round++ {
		g.joinsChanged = false
		g.mainEnv = newScope(nil)
		g.env = g.mainEnv
		g.globalDecl, g.globalOrder = map[string]pyType{}, nil
		g.discard(func() {
			var out strings.Builder
			var marked []string
			for _, st := range mod.Body {
				switch fn := st.(type) {
				case *pyast.FunctionDef:
					// The function is defined from here on.
					if !g.defined[fn.Name] {
						g.defined[fn.Name] = true
						marked = append(marked, fn.Name)
					}
				case *pyast.ClassDef:
				default:
					_ = g.stmtList(&out, []pyast.Node{st})
				}
			}
			for _, name := range marked {
				delete(g.defined, name)
			}
		})
		for name, t := range g.globalDecl {
			found[name] = t
		}
		if !g.joinsChanged {
			break
		}
	}
	g.env, g.mainEnv = savedEnv, savedMain
	g.globalDecl, g.globalOrder = savedDecl, savedOrder
	return found
}

// functionsPass compiles the top-level defs with the output thrown away, to
// find the module variables they assign through "global".
func (g *generator) functionsPass(mod *pyast.Module) {
	savedEnv, savedMain := g.env, g.mainEnv
	g.mainEnv = newScope(nil)
	g.env = g.mainEnv
	defer func() { g.env, g.mainEnv = savedEnv, savedMain }()
	for _, st := range mod.Body {
		fn, ok := st.(*pyast.FunctionDef)
		if !ok || g.genericFns[fn.Name] != nil || len(fn.Decorators) > 0 {
			continue
		}
		sig := g.sigs[fn.Name]
		if sig == nil || !hasGlobalStmt(fn) {
			continue
		}
		g.discard(func() {
			saved := g.saveFunctionState()
			params, result := g.goSignature(sig)
			_, _ = g.funcBody(fn, sig, fmt.Sprintf("func %s(%s)%s", sig.goName, params, result), nil, true, nil)
			g.restoreFunctionState(saved)
		})
	}
}

// hasGlobalStmt reports whether a def declares a name global.
func hasGlobalStmt(fn *pyast.FunctionDef) bool {
	found := false
	for _, st := range fn.Body {
		walkPrune(st, func(n pyast.Node) bool {
			switch n.(type) {
			case *pyast.Global:
				found = true
			case *pyast.FunctionDef, *pyast.ClassDef, *pyast.Lambda:
				return false
			}
			return !found
		})
	}
	return found
}

// widen is the type a container takes when something of another type is put
// into it: the items become objects.
func widen(t pyType) (pyType, bool) {
	switch {
	case t.isList():
		return listOf(tObject), t.elem() != tObject
	case t.isSet():
		return setOf(tObject), t.elem() != tObject
	case t.isDict():
		return dictOf(tObject, tObject), t.keyType() != tObject || t.valType() != tObject
	}
	return "", false
}

// widenVar records that the variable a node names holds items of other types
// than the ones it was first given. It only applies while the types of a
// function's variables are being worked out, and only to variables without an
// annotation; the next round declares the variable with the wider type.
func (g *generator) widenVar(node pyast.Node) bool {
	name, ok := node.(*pyast.Name)
	if !ok || !g.recordJoins {
		return false
	}
	t, declared, _ := g.lookupBinding(name.ID)
	if !declared || g.annotatedNames[g.joinScope][name.ID] {
		return false
	}
	w, changed := widen(t)
	if !changed {
		return false
	}
	g.noteBinding(name.ID, w)
	return true
}

// isMutableContainer reports whether t is a list, set or dict.
func isMutableContainer(t pyType) bool { return t.isList() || t.isSet() || t.isDict() }

// isObjectContainer reports whether the items of a container are objects
// already.
func (t pyType) isObjectContainer() bool {
	switch {
	case t.isList(), t.isSet():
		return t.elem() == tObject
	case t.isDict():
		return t.keyType() == tObject && t.valType() == tObject
	}
	return false
}

// widenName widens the variable of that name, which holds a container that
// has just been handed to something that may put anything into it.
func (g *generator) widenName(name string) {
	t, declared, _ := g.lookupBinding(name)
	if !declared || g.annotatedNames[g.bindingScope(name)][name] {
		return
	}
	if w, changed := widen(t); changed {
		g.noteBinding(name, w)
	}
}

// escape prepares a value for a place where anything may be put into it, as
// an object can be. A container that was just built gets objects for items
// (a copy, since nothing else refers to it); the variable a container was
// read from is declared with objects for items in the next round.
func (g *generator) escape(v value) value {
	if !isMutableContainer(v.typ) || v.typ.isObjectContainer() {
		return v
	}
	switch {
	case v.fresh:
		return value{code: g.rt("DynDeep") + "(" + v.code + ")", typ: tObject, prec: precAtom}
	case v.varName != "" && g.recordJoins:
		g.widenName(v.varName)
	}
	return v
}
