package codegen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// Generics are instantiated, not erased: every distinct list of type
// arguments of a generic function or class compiles a copy of it with the type
// parameters replaced by the arguments (g.tsubst), under a name of its own.
// The program is closed, so the instantiations are all known, and the copies
// are ordinary typed code. The type arguments of a call are inferred from the
// arguments by matching the parameter annotations against their types.

// typeVarInfo is a type parameter with its bound or constraints and default.
type typeVarInfo struct {
	name        string
	bound       pyast.Node
	constraints []pyast.Node
	def         pyast.Node
}

type genericFn struct {
	name   string
	node   *pyast.FunctionDef
	params []typeVarInfo
}

type genericClass struct {
	name   string
	node   *pyast.ClassDef
	params []typeVarInfo
	index  int
}

type typeAlias struct {
	params []typeVarInfo
	value  pyast.Node
	line   int
}

func typeVarFromParam(tp *pyast.TypeParam) (typeVarInfo, error) {
	if tp.Kind != "TypeVar" {
		return typeVarInfo{}, unsupported(tp.Kind+" '"+tp.Name+"' (arrives in phase 9)", 0)
	}
	tv := typeVarInfo{name: tp.Name, def: tp.Default}
	if tup, ok := tp.Bound.(*pyast.Tuple); ok {
		tv.constraints = tup.Elts
	} else {
		tv.bound = tp.Bound
	}
	return tv, nil
}

// scanTypeVars finds the old-style type variables, "T = TypeVar('T')", and
// the "type X = ..." aliases of a module.
func (g *generator) scanTypeVars(mod *pyast.Module) {
	for _, st := range mod.Body {
		switch n := st.(type) {
		case *pyast.Assign:
			if len(n.Targets) != 1 {
				continue
			}
			name, ok := n.Targets[0].(*pyast.Name)
			call, isCall := n.Value.(*pyast.Call)
			if !ok || !isCall {
				continue
			}
			if fn, _ := callName(call); fn != "TypeVar" {
				continue
			}
			tv := typeVarInfo{name: name.ID}
			if len(call.Args) > 1 {
				tv.constraints = call.Args[1:]
			}
			for _, kw := range call.Keywords {
				switch kw.Arg {
				case "bound":
					tv.bound = kw.Value
				case "default":
					tv.def = kw.Value
				}
			}
			g.typeVars[name.ID] = tv
		case *pyast.TypeAlias:
			ps := make([]typeVarInfo, 0, len(n.TypeParams))
			for _, tp := range n.TypeParams {
				tv, err := typeVarFromParam(tp)
				if err == nil {
					ps = append(ps, tv)
				}
			}
			g.typeAliases[n.Name] = &typeAlias{params: ps, value: n.Value, line: n.Line()}
		}
	}
}

// freeTypeVars lists the old-style type variables named in the annotations,
// in order of appearance.
func (g *generator) freeTypeVars(nodes ...pyast.Node) []typeVarInfo {
	var out []typeVarInfo
	seen := map[string]bool{}
	for _, n := range nodes {
		if n == nil {
			continue
		}
		walk(n, func(x pyast.Node) {
			if id, ok := x.(*pyast.Name); ok {
				if tv, isVar := g.typeVars[id.ID]; isVar && !seen[id.ID] {
					seen[id.ID] = true
					out = append(out, tv)
				}
			}
		})
	}
	return out
}

// functionTypeParams are the type parameters of a def, if it is generic.
func (g *generator) functionTypeParams(fn *pyast.FunctionDef) ([]typeVarInfo, error) {
	if len(fn.TypeParams) > 0 {
		var out []typeVarInfo
		for _, tp := range fn.TypeParams {
			tv, err := typeVarFromParam(tp)
			if err != nil {
				return nil, unsupported(tp.Kind+" '"+tp.Name+"' (arrives in phase 9)", fn.Line())
			}
			out = append(out, tv)
		}
		return out, nil
	}
	if len(g.typeVars) == 0 {
		return nil, nil
	}
	var anns []pyast.Node
	a := fn.Args
	for _, arg := range append(append(a.Positional(), a.Kwonly...), []*pyast.Arg{a.Vararg, a.Kwarg}...) {
		if arg != nil {
			anns = append(anns, arg.Annotation)
		}
	}
	anns = append(anns, fn.Returns)
	return g.freeTypeVars(anns...), nil
}

// classTypeParams are the type parameters of a class: the brackets of PEP 695,
// or the type variables of Generic[T] and of subscripted bases.
func (g *generator) classTypeParams(cd *pyast.ClassDef) ([]typeVarInfo, error) {
	if len(cd.TypeParams) > 0 {
		var out []typeVarInfo
		for _, tp := range cd.TypeParams {
			tv, err := typeVarFromParam(tp)
			if err != nil {
				return nil, unsupported(tp.Kind+" '"+tp.Name+"' (arrives in phase 9)", cd.Line())
			}
			out = append(out, tv)
		}
		return out, nil
	}
	if len(g.typeVars) == 0 {
		return nil, nil
	}
	for _, b := range cd.Bases {
		if sub, ok := b.(*pyast.Subscript); ok {
			if head, ok := sub.Value.(*pyast.Name); ok && head.ID == "Generic" {
				return g.freeTypeVars(sub.Index), nil
			}
		}
	}
	return g.freeTypeVars(cd.Bases...), nil
}

// withSubst runs f with the type parameters bound.
func (g *generator) withSubst(subst map[string]pyType, f func() error) error {
	saved := g.tsubst
	g.tsubst = subst
	defer func() { g.tsubst = saved }()
	return f()
}

// inModuleContext runs f as if at module level, outside any function: the
// instantiation of a generic can happen inside the body of another function.
func (g *generator) inModuleContext(f func() error) error {
	saved := g.saveFunctionState()
	savedHoist, savedPending := g.hoist, g.pendingNarrow
	defer func() {
		g.restoreFunctionState(saved)
		g.hoist, g.pendingNarrow = savedHoist, savedPending
	}()
	g.env, g.fnName, g.topFn, g.compRoot = g.mainEnv, "", "", nil
	g.method, g.gen, g.specialRet, g.genNone = nil, nil, false, false
	g.flow, g.handlers, g.loopDepth = nil, nil, 0
	g.fnGlobals, g.fnNonlocals = map[string]bool{}, map[string]bool{}
	return f()
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9]+`)

// instanceName makes the Go name of an instantiation, unique.
func (g *generator) instanceName(base string, targs []pyType) string {
	parts := make([]string, len(targs))
	for i, t := range targs {
		parts[i] = strings.Trim(unsafeName.ReplaceAllString(string(t), "_"), "_")
	}
	name := base + "_" + strings.Join(parts, "_")
	if !g.usedNames[name] {
		g.usedNames[name] = true
		return name
	}
	for i := 2; ; i++ {
		alt := fmt.Sprintf("%s_%d", name, i)
		if !g.usedNames[alt] {
			g.usedNames[alt] = true
			return alt
		}
	}
}

func typesKey(name string, targs []pyType) string {
	parts := make([]string, len(targs))
	for i, t := range targs {
		parts[i] = string(t)
	}
	return name + "[" + strings.Join(parts, ";") + "]"
}

// bindParams checks type arguments against bounds and constraints, fills in
// defaults, and returns the substitution.
func (g *generator) bindParams(owner string, params []typeVarInfo, targs []pyType, line int) (map[string]pyType, []pyType, error) {
	if len(targs) > len(params) {
		return nil, nil, unsupported(fmt.Sprintf("%s takes %d type arguments but %d were given", owner, len(params), len(targs)), line)
	}
	subst := map[string]pyType{}
	full := append([]pyType{}, targs...)
	for i, p := range params {
		if i < len(targs) {
			subst[p.name] = targs[i]
			continue
		}
		if p.def == nil {
			return nil, nil, unsupported(fmt.Sprintf("%s needs %d type arguments but %d were given", owner, len(params), len(targs)), line)
		}
		var t pyType
		if err := g.withSubst(subst, func() (err error) {
			t, err = g.annotationType(p.def, false)
			return
		}); err != nil {
			return nil, nil, err
		}
		subst[p.name] = t
		full = append(full, t)
	}
	for _, p := range params {
		t := subst[p.name]
		var bad bool
		switch {
		case p.bound != nil:
			var bt pyType
			if err := g.withSubst(subst, func() (err error) { bt, err = g.annotationType(p.bound, false); return }); err != nil {
				return nil, nil, err
			}
			bad = !g.isSubtype(t, bt)
		case len(p.constraints) > 0:
			bad = true
			for _, c := range p.constraints {
				var ct pyType
				if err := g.withSubst(subst, func() (err error) { ct, err = g.annotationType(c, false); return }); err != nil {
					return nil, nil, err
				}
				if ct == t {
					bad = false
				}
			}
		}
		if bad && t != tObject {
			return nil, nil, unsupported(fmt.Sprintf("type argument %s of %s violates the bound or constraints of %s", t, owner, p.name), line)
		}
	}
	return subst, full, nil
}

// ---- generic classes ----

// classInstance returns the instantiation of a generic class.
func (g *generator) classInstance(tmpl *genericClass, targs []pyType, line int) (*classInfo, error) {
	subst, full, err := g.bindParams(tmpl.name, tmpl.params, targs, line)
	if err != nil {
		return nil, err
	}
	key := typesKey(tmpl.name, full)
	if c := g.genericInst[key]; c != nil {
		return c, nil
	}
	c := &classInfo{
		name: g.instanceName(tmpl.name, full), display: tmpl.name, node: tmpl.node, line: tmpl.node.Line(),
		index: tmpl.index, members: map[string]*member{}, tsubst: subst, generic: tmpl, targs: full,
	}
	g.genericInst[key] = c
	g.classes[c.name] = c
	g.classOrder = append(g.classOrder, c)
	if g.phase == 0 {
		g.incomplete = append(g.incomplete, c)
		return c, nil
	}
	return c, g.completeInstance(c)
}

// completeInstance reads the bases, members and fields of an instantiation.
func (g *generator) completeInstance(c *classInfo) error {
	if c.completed {
		return nil
	}
	c.completed = true
	err := g.inModuleContext(func() error {
		return g.withSubst(c.tsubst, func() error {
			if err := g.classBases(c); err != nil {
				return err
			}
			if err := g.classBody(c); err != nil {
				return err
			}
			return g.finishClass(c)
		})
	})
	if err != nil {
		return err
	}
	if g.phase >= 2 {
		// The program is already being generated: infer now.
		if err := g.inModuleContext(func() error { return g.inferClassAttrs(c) }); err != nil {
			return err
		}
		g.inferRounds([]*classInfo{c})
	}
	g.pendingGen = append(g.pendingGen, c)
	return nil
}

// completeInstances finishes the instantiations that were asked for while the
// classes were being read.
func (g *generator) completeInstances() error {
	for len(g.incomplete) > 0 {
		c := g.incomplete[0]
		g.incomplete = g.incomplete[1:]
		if err := g.completeInstance(c); err != nil {
			return err
		}
	}
	return nil
}

// emitGenerics writes the instantiations that were not written yet: classes,
// and functions, whose code can ask for more.
func (g *generator) emitGenerics() error {
	for len(g.pendingGen) > 0 {
		c := g.pendingGen[0]
		g.pendingGen = g.pendingGen[1:]
		init := g.lateInitFor(c.generic.name)
		err := g.inModuleContext(func() error {
			return g.withSubst(c.tsubst, func() error { return g.classDecl(&g.lateFuncs, init, c) })
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) lateInitFor(name string) *strings.Builder {
	b := g.lateInit[name]
	if b == nil {
		b = &strings.Builder{}
		g.lateInit[name] = b
	}
	return b
}

// lateMarker stands in main() for the initialization of a generic's instances.
func lateMarker(name string) string { return "/*late:" + name + "*/" }

// ---- type arguments of a use ----

// typeArgs reads the arguments of Box[int, str].
func (g *generator) typeArgs(index pyast.Node) ([]pyType, error) {
	nodes := []pyast.Node{index}
	if t, ok := index.(*pyast.Tuple); ok {
		nodes = t.Elts
	}
	out := make([]pyType, len(nodes))
	for i, n := range nodes {
		t, err := g.annotationType(n, false)
		if err != nil {
			return nil, err
		}
		out[i] = t
	}
	return out, nil
}

// genericSubscript resolves Name[args] in an annotation when Name is a
// generic class or a type alias with parameters.
func (g *generator) genericSubscript(n *pyast.Subscript) (pyType, bool, error) {
	head, ok := n.Value.(*pyast.Name)
	if !ok {
		return "", false, nil
	}
	if tmpl := g.genericClasses[head.ID]; tmpl != nil {
		targs, err := g.typeArgs(n.Index)
		if err != nil {
			return "", true, err
		}
		c, err := g.classInstance(tmpl, targs, n.Line())
		if err != nil {
			return "", true, err
		}
		return c.typ(), true, nil
	}
	if al := g.typeAliases[head.ID]; al != nil && len(al.params) > 0 {
		targs, err := g.typeArgs(n.Index)
		if err != nil {
			return "", true, err
		}
		subst, _, err := g.bindParams(head.ID, al.params, targs, n.Line())
		if err != nil {
			return "", true, err
		}
		var t pyType
		err = g.withSubst(subst, func() (e error) { t, e = g.annotationType(al.value, false); return })
		return t, true, err
	}
	return "", false, nil
}

// ---- inference of the type arguments of a call ----

// unify matches an annotation against the type of an argument and records
// what the type parameters must be.
func (g *generator) unify(ann pyast.Node, t pyType, params map[string]bool, bind map[string]pyType) error {
	switch a := ann.(type) {
	case *pyast.Name:
		if !params[a.ID] {
			return nil
		}
		if cur, ok := bind[a.ID]; ok {
			if cur == t {
				return nil
			}
			j, ok := g.joinTypes(cur, t)
			if !ok {
				j = tObject
			}
			bind[a.ID] = j
			return nil
		}
		if t == tNoneType {
			return nil
		}
		bind[a.ID] = t
		return nil
	case *pyast.BinOp:
		if a.Op != "BitOr" {
			return nil
		}
		var members []pyast.Node
		var flat func(x pyast.Node)
		flat = func(x pyast.Node) {
			if b, ok := x.(*pyast.BinOp); ok && b.Op == "BitOr" {
				flat(b.Left)
				flat(b.Right)
				return
			}
			if c, ok := x.(*pyast.Constant); ok && c.PyType == "none" {
				return
			}
			members = append(members, x)
		}
		flat(a)
		if len(members) == 1 {
			if t.isOptional() {
				t = t.optElem()
			}
			return g.unify(members[0], t, params, bind)
		}
		return nil
	case *pyast.Subscript:
		head, ok := a.Value.(*pyast.Name)
		if !ok {
			return nil
		}
		args := []pyast.Node{a.Index}
		if tup, ok := a.Index.(*pyast.Tuple); ok {
			args = tup.Elts
		}
		name := head.ID
		if alias, ok := typingAliases[name]; ok {
			name = alias
		}
		switch name {
		case "list", "set":
			if (name == "list" && t.isList()) || (name == "set" && t.isSet()) {
				return g.unify(args[0], t.elem(), params, bind)
			}
		case "dict":
			if t.isDict() && len(args) == 2 {
				if err := g.unify(args[0], t.keyType(), params, bind); err != nil {
					return err
				}
				return g.unify(args[1], t.valType(), params, bind)
			}
		case "tuple":
			switch {
			case t.isVarTuple():
				return g.unify(args[0], t.elem(), params, bind)
			case t.isTuple() && len(args) == len(t.elems()):
				for i, x := range args {
					if err := g.unify(x, t.elems()[i], params, bind); err != nil {
						return err
					}
				}
			}
		case "Iterable", "Iterator", "Sequence", "Collection":
			if elem, ok := t.iterElem(); ok && !t.isTuple() {
				return g.unify(args[0], elem, params, bind)
			}
		case "Generator":
			if t.isGen() {
				for i, x := range args {
					if err := g.unify(x, t.typeArgs()[i], params, bind); err != nil {
						return err
					}
				}
			}
		case "Optional":
			if t.isOptional() {
				t = t.optElem()
			}
			return g.unify(args[0], t, params, bind)
		case "Callable":
			if t.isFunc() && len(args) == 2 {
				if l, ok := args[0].(*pyast.List); ok && len(l.Elts) == len(t.funcParams()) {
					for i, x := range l.Elts {
						if err := g.unify(x, t.funcParams()[i], params, bind); err != nil {
							return err
						}
					}
				}
				return g.unify(args[1], t.funcRet(), params, bind)
			}
		default:
			if tmpl := g.genericClasses[head.ID]; tmpl != nil {
				if c := g.classOf(t); c != nil && c.generic == tmpl {
					for i, x := range args {
						if i < len(c.targs) {
							if err := g.unify(x, c.targs[i], params, bind); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// annotatedParam is a parameter of a generic def with its annotation.
type annotatedParam struct {
	name     string
	ann      pyast.Node
	vararg   bool
	kwonly   bool
	required bool
}

func fnAnnotatedParams(a *pyast.Arguments) []annotatedParam {
	var out []annotatedParam
	pos := a.Positional()
	firstDefault := len(pos) - len(a.Defaults)
	for i, p := range pos {
		out = append(out, annotatedParam{name: p.Name, ann: p.Annotation, required: i < firstDefault})
	}
	if a.Vararg != nil {
		out = append(out, annotatedParam{name: a.Vararg.Name, ann: a.Vararg.Annotation, vararg: true})
	}
	for i, p := range a.Kwonly {
		out = append(out, annotatedParam{name: p.Name, ann: p.Annotation, kwonly: true, required: i >= len(a.KwDefaults) || a.KwDefaults[i] == nil})
	}
	return out
}

// inferTypeArgs finds the type arguments from the arguments of a call. skip
// is the number of leading parameters (self) that the call does not write.
func (g *generator) inferTypeArgs(owner string, params []typeVarInfo, ps []annotatedParam, n *pyast.Call, ret pyast.Node, want pyType) ([]pyType, error) {
	names := map[string]bool{}
	for _, p := range params {
		names[p.name] = true
	}
	bind := map[string]pyType{}

	// Match the arguments to the parameters.
	type pair struct {
		arg pyast.Node
		ann pyast.Node
	}
	var pairs []pair
	nPos := 0
	for _, p := range ps {
		if !p.vararg && !p.kwonly {
			nPos++
		}
	}
	for i, a := range n.Args {
		if _, star := a.(*pyast.Starred); star {
			return nil, unsupported("*args in a call of a generic ("+owner+")", a.Line())
		}
		switch {
		case i < nPos:
			pairs = append(pairs, pair{a, ps[i].ann})
		default:
			for _, p := range ps {
				if p.vararg {
					pairs = append(pairs, pair{a, p.ann})
				}
			}
		}
	}
	for _, kw := range n.Keywords {
		if kw.Arg == "" {
			return nil, unsupported("**mapping in a call of a generic ("+owner+")", kw.Line())
		}
		for _, p := range ps {
			if p.name == kw.Arg {
				pairs = append(pairs, pair{kw.Value, p.ann})
			}
		}
	}

	// First the arguments whose type does not depend on the parameters.
	var deferred []pair
	for _, pr := range pairs {
		if pr.ann == nil {
			continue
		}
		if _, isLambda := pr.arg.(*pyast.Lambda); isLambda {
			deferred = append(deferred, pr)
			continue
		}
		if nm, isName := pr.arg.(*pyast.Name); isName && builtinFunctions[nm.ID] && !g.isVariable(nm.ID) && g.sigs[nm.ID] == nil {
			deferred = append(deferred, pr)
			continue
		}
		v, err := g.typedExpr(pr.arg)
		if err != nil {
			// An argument that needs the parameter type to be typed at all,
			// such as an empty list, is checked by the call itself.
			continue
		}
		if v.typ.isClassRef() && v.classRef == nil {
			continue
		}
		if err := g.unify(pr.ann, v.typ, names, bind); err != nil {
			return nil, err
		}
	}
	// A lambda gets its parameter types from what is known so far, and its
	// result gives the rest.
	for _, pr := range deferred {
		sub, ok := pr.ann.(*pyast.Subscript)
		if !ok {
			continue
		}
		args := []pyast.Node{sub.Index}
		if tup, ok := sub.Index.(*pyast.Tuple); ok {
			args = tup.Elts
		}
		if len(args) != 2 {
			continue
		}
		l, ok := args[0].(*pyast.List)
		if !ok {
			continue
		}
		var ptypes []pyType
		var perr error
		_ = g.withSubst(bind, func() error {
			for _, x := range l.Elts {
				t, err := g.annotationType(x, false)
				if err != nil {
					perr = err
					return err
				}
				ptypes = append(ptypes, t)
			}
			return nil
		})
		if perr != nil {
			return nil, unsupported(fmt.Sprintf("cannot infer the parameters of a function passed to %s: %v", owner, perr), pr.arg.Line())
		}
		var lam *pyast.Lambda
		if l, ok := pr.arg.(*pyast.Lambda); ok {
			lam = l
		} else if lam = g.builtinAsLambda(pr.arg, funcOf(ptypes, tNone)); lam == nil {
			continue
		}
		v, err := g.lambda(lam, ptypes, "")
		if err != nil {
			return nil, err
		}
		if err := g.unify(pr.ann, v.typ, names, bind); err != nil {
			return nil, err
		}
	}
	// What the result will be used as can settle the rest.
	if ret != nil && want != "" {
		if err := g.unify(ret, want, names, bind); err != nil {
			return nil, err
		}
	}

	out := make([]pyType, 0, len(params))
	for _, p := range params {
		t, ok := bind[p.name]
		if !ok {
			if p.def != nil {
				break // the defaults fill in the rest
			}
			t = tObject // nothing says what it is: any value will do
		}
		out = append(out, t)
	}
	return out, nil
}

// ---- generic functions ----

// functionInstance is the instantiation of a generic function.
func (g *generator) functionInstance(gf *genericFn, targs []pyType, line int) (*funcSig, error) {
	subst, full, err := g.bindParams(gf.name, gf.params, targs, line)
	if err != nil {
		return nil, err
	}
	key := typesKey(gf.name, full)
	if s := g.fnInst[key]; s != nil {
		return s, nil
	}
	goName := g.instanceName(gf.name, full)
	cp := *gf.node
	cp.Name = goName
	cp.TypeParams = nil
	cp.HasTypeParams = false
	var sig *funcSig
	err = g.inModuleContext(func() error {
		return g.withSubst(subst, func() error {
			var e error
			sig, e = g.buildSig(&cp, gf.name, false)
			if e != nil {
				return e
			}
			sig.goName = goName
			sig.display = gf.name
			ps, _ := sig.defaultNodes()
			for i, p := range ps {
				p.defaultVar = fmt.Sprintf("_pyd_%s_%d", goName, i)
			}
			g.fnInst[key] = sig
			g.defined[goName] = true
			init := g.lateInitFor(gf.name)
			if e := g.assignDefaults(init, sig, false); e != nil {
				return e
			}
			params, result := g.goSignature(sig)
			text, e := g.funcBody(&cp, sig, fmt.Sprintf("func %s(%s)%s", goName, params, result), nil, true, nil)
			if e != nil {
				return e
			}
			g.lateFuncs.WriteString(text + "\n\n")
			return nil
		})
	})
	if err != nil {
		delete(g.fnInst, key)
		return nil, err
	}
	return sig, nil
}

// genericCall is a call of a generic function with inferred type arguments.
func (g *generator) genericCall(n *pyast.Call, gf *genericFn, explicit []pyType) (value, error) {
	targs := explicit
	if targs == nil {
		var err error
		targs, err = g.inferTypeArgs(gf.name, gf.params, fnAnnotatedParams(gf.node.Args), n, gf.node.Returns, g.callWant)
		if err != nil {
			return value{}, err
		}
	}
	sig, err := g.functionInstance(gf, targs, n.Line())
	if err != nil {
		return value{}, err
	}
	return g.userCall(n, sig)
}

// genericClassCall is Box(args), with the type arguments inferred from the
// arguments of __init__ or from the type the value is wanted as.
func (g *generator) genericClassCall(n *pyast.Call, tmpl *genericClass, explicit []pyType) (value, error) {
	targs := explicit
	if targs == nil {
		if want := g.callWant; want != "" {
			if c := g.classOf(want); c != nil && c.generic == tmpl {
				targs = c.targs
			}
		}
		if targs == nil {
			ps, err := g.constructorParams(tmpl)
			if err != nil {
				return value{}, err
			}
			targs, err = g.inferTypeArgs(tmpl.name, tmpl.params, ps, n, nil, "")
			if err != nil {
				return value{}, err
			}
		}
	}
	c, err := g.classInstance(tmpl, targs, n.Line())
	if err != nil {
		return value{}, err
	}
	return g.constructCall(n, c)
}

// constructorParams are the annotated parameters of the __init__ of a generic
// class, or the fields of a dataclass.
func (g *generator) constructorParams(tmpl *genericClass) ([]annotatedParam, error) {
	for _, st := range tmpl.node.Body {
		if fn, ok := st.(*pyast.FunctionDef); ok && fn.Name == "__init__" {
			ps := fnAnnotatedParams(fn.Args)
			if len(ps) > 0 {
				ps = ps[1:] // self
			}
			return ps, nil
		}
	}
	var ps []annotatedParam
	for _, st := range tmpl.node.Body {
		if a, ok := st.(*pyast.AnnAssign); ok {
			if name, ok := a.Target.(*pyast.Name); ok {
				ps = append(ps, annotatedParam{name: name.ID, ann: a.Annotation, required: a.Value == nil})
			}
		}
	}
	if len(ps) > 0 {
		return ps, nil
	}
	// No __init__ of its own: the one of the generic base, with the base's
	// type parameters replaced by what this class passes it.
	for _, b := range tmpl.node.Bases {
		sub, ok := b.(*pyast.Subscript)
		if !ok {
			continue
		}
		head, ok := sub.Value.(*pyast.Name)
		if !ok || g.genericClasses[head.ID] == nil {
			continue
		}
		base := g.genericClasses[head.ID]
		args := []pyast.Node{sub.Index}
		if t, ok := sub.Index.(*pyast.Tuple); ok {
			args = t.Elts
		}
		baseParams, err := g.constructorParams(base)
		if err != nil {
			return nil, err
		}
		mapping := map[string]pyast.Node{}
		for i, p := range base.params {
			if i < len(args) {
				mapping[p.name] = args[i]
			}
		}
		for _, bp := range baseParams {
			bp.ann = substAnnotation(bp.ann, mapping)
			ps = append(ps, bp)
		}
		return ps, nil
	}
	return ps, nil
}

// substAnnotation replaces the names in an annotation.
func substAnnotation(n pyast.Node, m map[string]pyast.Node) pyast.Node {
	switch x := n.(type) {
	case *pyast.Name:
		if r, ok := m[x.ID]; ok {
			return r
		}
	case *pyast.Subscript:
		return &pyast.Subscript{Value: substAnnotation(x.Value, m), Index: substAnnotation(x.Index, m), Lineno: x.Lineno}
	case *pyast.Tuple:
		elts := make([]pyast.Node, len(x.Elts))
		for i, e := range x.Elts {
			elts[i] = substAnnotation(e, m)
		}
		return &pyast.Tuple{Elts: elts, Lineno: x.Lineno}
	case *pyast.List:
		elts := make([]pyast.Node, len(x.Elts))
		for i, e := range x.Elts {
			elts[i] = substAnnotation(e, m)
		}
		return &pyast.List{Elts: elts, Lineno: x.Lineno}
	case *pyast.BinOp:
		return &pyast.BinOp{Left: substAnnotation(x.Left, m), Op: x.Op, Right: substAnnotation(x.Right, m), Lineno: x.Lineno}
	}
	return n
}

// parseAnnotation reads the text of a string annotation, "Pair[V, K]" or
// "Node | None", into the nodes an annotation is made of.
func parseAnnotation(s string, line int) (pyast.Node, error) {
	p := &annParser{src: s, line: line}
	n, err := p.union()
	if err != nil {
		return nil, err
	}
	p.skip()
	if p.pos != len(p.src) {
		return nil, unsupported("string annotation '"+s+"'", line)
	}
	return n, nil
}

type annParser struct {
	src  string
	pos  int
	line int
}

func (p *annParser) skip() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *annParser) peek() byte {
	p.skip()
	if p.pos < len(p.src) {
		return p.src[p.pos]
	}
	return 0
}

func (p *annParser) union() (pyast.Node, error) {
	left, err := p.atom()
	if err != nil {
		return nil, err
	}
	for p.peek() == '|' {
		p.pos++
		right, err := p.atom()
		if err != nil {
			return nil, err
		}
		left = &pyast.BinOp{Left: left, Op: "BitOr", Right: right, Lineno: p.line}
	}
	return left, nil
}

func (p *annParser) atom() (pyast.Node, error) {
	c := p.peek()
	switch {
	case c == '[':
		p.pos++
		var elts []pyast.Node
		for p.peek() != ']' {
			e, err := p.union()
			if err != nil {
				return nil, err
			}
			elts = append(elts, e)
			if p.peek() == ',' {
				p.pos++
			}
			if p.pos >= len(p.src) {
				return nil, unsupported("string annotation '"+p.src+"'", p.line)
			}
		}
		p.pos++
		return &pyast.List{Elts: elts, Lineno: p.line}, nil
	case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
		start := p.pos
		for p.pos < len(p.src) && (p.src[p.pos] == '_' || p.src[p.pos] == '.' || p.src[p.pos] >= 'a' && p.src[p.pos] <= 'z' || p.src[p.pos] >= 'A' && p.src[p.pos] <= 'Z' || p.src[p.pos] >= '0' && p.src[p.pos] <= '9') {
			p.pos++
		}
		name := p.src[start:p.pos]
		if name == "None" {
			return &pyast.Constant{PyType: "none", Lineno: p.line}, nil
		}
		var n pyast.Node = &pyast.Name{ID: name, Lineno: p.line}
		if p.peek() == '[' {
			p.pos++
			var args []pyast.Node
			for p.peek() != ']' {
				a, err := p.union()
				if err != nil {
					return nil, err
				}
				args = append(args, a)
				if p.peek() == ',' {
					p.pos++
				}
				if p.pos >= len(p.src) {
					return nil, unsupported("string annotation '"+p.src+"'", p.line)
				}
			}
			p.pos++
			var index pyast.Node = &pyast.Tuple{Elts: args, Lineno: p.line}
			if len(args) == 1 {
				index = args[0]
			}
			n = &pyast.Subscript{Value: n, Index: index, Lineno: p.line}
		}
		return n, nil
	case c == '"' || c == '\'':
		end := strings.IndexByte(p.src[p.pos+1:], c)
		if end < 0 {
			return nil, unsupported("string annotation '"+p.src+"'", p.line)
		}
		inner := p.src[p.pos+1 : p.pos+1+end]
		p.pos += end + 2
		return parseAnnotation(inner, p.line)
	case c == '.':
		// "..." in tuple[int, ...]
		if strings.HasPrefix(p.src[p.pos:], "...") {
			p.pos += 3
			return &pyast.Unsupported{Type: "ellipsis constant", Lineno: p.line}, nil
		}
	}
	return nil, unsupported("string annotation '"+p.src+"'", p.line)
}
