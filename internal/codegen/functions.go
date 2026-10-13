package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// param is one parameter of a function.
type param struct {
	name       string
	typ        pyType
	hasDefault bool
	defaultVar string // Go variable that holds the default value
	line       int
	defNode    pyast.Node // the default expression, evaluated once when the def runs
	// factory makes a default per call instead: a dataclass field with
	// default_factory.
	factory pyast.Node
}

// funcSig is what is known about a def: its parameters, which can have
// defaults and keywords, its return type, and where it lives in the Go code.
type funcSig struct {
	name     string // the Python name
	qualname string // as Python spells __qualname__
	goName   string // what a call writes: the Go function or closure variable
	line     int

	pos      []param // positional-or-keyword parameters, positional-only first
	nPosOnly int
	kwonly   []param
	vararg   *param // typ is tuple[T, ...]
	kwarg    *param // typ is dict[str, V]
	ret      pyType

	doc    string
	hasDoc bool

	// implicit counts the parameters that error messages show but the call
	// does not write: the self of a bound method.
	implicit int
	selfName string // the name the method gives its first parameter
	// display is the name error messages use when it differs from name.
	display string
	// gen is set for a generator function, whose ret is the type of the
	// generator object it makes.
	gen *genInfo
	// leadArg is code the call passes before the real arguments: the number
	// of positional arguments an exception class was called with.
	leadArg string

	decorators []pyast.Node
	node       *pyast.FunctionDef
	nested     bool
	// valueType is the type of the name once the decorators have run, and
	// is the plain function type for an undecorated function.
	valueType pyType
}

// goParams are the parameters in the order of the Go function: positional,
// keyword-only, then *args and **kwargs.
func (s *funcSig) goParams() []param {
	out := append([]param{}, s.pos...)
	out = append(out, s.kwonly...)
	if s.vararg != nil {
		out = append(out, *s.vararg)
	}
	if s.kwarg != nil {
		out = append(out, *s.kwarg)
	}
	return out
}

// goFuncType is the Go signature as a function type. For a function with
// only plain parameters this is also its Python Callable type.
func (s *funcSig) goFuncType() pyType {
	var types []pyType
	for _, p := range s.goParams() {
		types = append(types, p.typ)
	}
	return funcOf(types, s.ret)
}

// plain reports whether the function takes only positional-or-keyword
// parameters, which is what a Callable value can describe.
func (s *funcSig) plain() bool {
	return len(s.kwonly) == 0 && s.vararg == nil && s.kwarg == nil
}

func (s *funcSig) minPositional() int {
	n := 0
	for _, p := range s.pos {
		if !p.hasDefault {
			n++
		}
	}
	return n
}

// docstring returns the docstring of a body: a first statement that is
// a string literal.
func docstring(body []pyast.Node) (string, bool) {
	if len(body) == 0 {
		return "", false
	}
	e, ok := body[0].(*pyast.Expr)
	if !ok {
		return "", false
	}
	c, ok := e.Value.(*pyast.Constant)
	if !ok || c.PyType != "str" {
		return "", false
	}
	return cleanDoc(c.Value.(string)), true
}

// cleanDoc strips the common indentation of a docstring, as the CPython 3.13
// compiler does: tabs expand, the first line is left alone, the margin is
// the smallest indent of the other non-blank lines, and blank lines at the
// ends are not removed.
func cleanDoc(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\t", "        "), "\n")
	if len(lines) == 1 {
		return s
	}
	margin := -1
	for _, l := range lines[1:] {
		body := strings.TrimLeft(l, " ")
		if body == "" {
			continue
		}
		if ind := len(l) - len(body); margin < 0 || ind < margin {
			margin = ind
		}
	}
	for i := 1; i < len(lines); i++ {
		if margin > 0 && len(lines[i]) >= margin {
			lines[i] = lines[i][margin:]
		} else if strings.TrimLeft(lines[i], " ") == "" {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

// buildSig validates a def line and reads its parameters and return type.
func (g *generator) buildSig(fn *pyast.FunctionDef, qualname string, nested bool) (*funcSig, error) {
	if !g.buildingMethod {
		if err := checkName(fn.Name, fn.Line()); err != nil {
			return nil, err
		}
	}
	if fn.HasTypeParams {
		return nil, unsupported("generic functions", fn.Line())
	}

	sig := &funcSig{name: fn.Name, qualname: qualname, goName: fn.Name, line: fn.Line(),
		decorators: fn.Decorators, node: fn, nested: nested}
	sig.doc, sig.hasDoc = docstring(fn.Body)

	a := fn.Args
	seen := map[string]bool{}
	read := func(arg *pyast.Arg, what string) (param, error) {
		if err := checkName(arg.Name, arg.Line()); err != nil {
			return param{}, err
		}
		if seen[arg.Name] {
			return param{}, unsupported(fmt.Sprintf("duplicate parameter '%s'", arg.Name), arg.Line())
		}
		seen[arg.Name] = true
		if arg.Annotation == nil {
			// Without a type hint the parameter takes anything.
			return param{name: arg.Name, typ: tObject, line: arg.Line()}, nil
		}
		t, err := g.annotationType(arg.Annotation, false)
		if err != nil {
			return param{}, err
		}
		return param{name: arg.Name, typ: t, line: arg.Line()}, nil
	}

	all := a.Positional()
	firstDefault := len(all) - len(a.Defaults)
	for i, arg := range all {
		p, err := read(arg, "")
		if err != nil {
			return nil, err
		}
		p.hasDefault = i >= firstDefault
		if p.hasDefault {
			p.defNode = a.Defaults[i-firstDefault]
		}
		sig.pos = append(sig.pos, p)
	}
	sig.nPosOnly = len(a.PosOnly)
	for i, arg := range a.Kwonly {
		p, err := read(arg, "keyword-only ")
		if err != nil {
			return nil, err
		}
		p.hasDefault = i < len(a.KwDefaults) && a.KwDefaults[i] != nil
		if p.hasDefault {
			p.defNode = a.KwDefaults[i]
		}
		sig.kwonly = append(sig.kwonly, p)
	}
	if a.Vararg != nil {
		p, err := read(a.Vararg, "*")
		if err != nil {
			return nil, err
		}
		p.typ = varTupleOf(p.typ)
		sig.vararg = &p
	}
	if a.Kwarg != nil {
		p, err := read(a.Kwarg, "**")
		if err != nil {
			return nil, err
		}
		p.typ = dictOf(tStr, p.typ)
		sig.kwarg = &p
	}

	var ret pyType
	if fn.Returns == nil {
		// Without a hint, a function that never returns a value returns
		// None, a generator yields anything, and any other returns anything.
		switch {
		case fn.IsAsync || isGeneratorBody(fn.Body):
			ret = genOf(tObject, tObject, tObject)
		case g.buildingMethod && protocolReturns[fn.Name] != "":
			ret = protocolReturns[fn.Name]
		case returnsValue(fn.Body):
			ret = tObject
		case g.buildingMethod && g.valueMethods[fn.Name] && !isDunder(fn.Name) && !g.buildingProperty:
			ret = tObject
		default:
			ret = tNone
		}
	} else {
		var err error
		ret, err = g.annotationType(fn.Returns, true)
		if err != nil {
			return nil, err
		}
	}
	sig.ret = ret
	if fn.IsAsync {
		if isGeneratorBody(fn.Body) {
			sig.gen = &genInfo{yield: tObject, send: tObject, ret: tObject, asyncGen: true}
			sig.ret = tObject
		} else {
			r := tObject
			if fn.Returns != nil {
				r = ret
				if r == tNone {
					r = tNoneType
				}
			}
			sig.gen = &genInfo{yield: tObject, send: tObject, ret: r, coro: true}
			sig.ret = genOf(tObject, tObject, r)
		}
	} else if isGeneratorBody(fn.Body) {
		switch {
		case ret.isIter():
			sig.gen = &genInfo{yield: ret.elem(), send: tNoneType, ret: tNoneType, iterator: true}
		case ret.isIterable():
			sig.gen = &genInfo{yield: ret.elem(), send: tNoneType, ret: tNoneType, iterator: true}
			sig.ret = iterOf(ret.elem())
		case ret.isGen():
			a := ret.typeArgs()
			sig.gen = &genInfo{yield: a[0], send: a[1], ret: a[2]}
		default:
			return nil, unsupported(fmt.Sprintf("generator function annotated %s (use Iterator[T], Iterable[T] or Generator[Y, S, R])", ret), fn.Line())
		}
	}
	sig.valueType = sig.goFuncType()
	return sig, nil
}

// defaultNodes returns the parameters that have a default evaluated when the
// def runs, with the expression of each.
func (s *funcSig) defaultNodes() (ps []*param, nodes []pyast.Node) {
	for i := range s.pos {
		if s.pos[i].defNode != nil {
			ps = append(ps, &s.pos[i])
			nodes = append(nodes, s.pos[i].defNode)
		}
	}
	for i := range s.kwonly {
		if s.kwonly[i].defNode != nil {
			ps = append(ps, &s.kwonly[i])
			nodes = append(nodes, s.kwonly[i].defNode)
		}
	}
	return ps, nodes
}

// collectSignatures records every top-level function's signature. A bad
// signature is stored, not returned, so errors still surface in source order
// when the offending def or a call to it is reached.
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
		if g.genericFns[fn.Name] != nil {
			continue
		}
		params, perr := g.functionTypeParams(fn)
		if perr != nil {
			g.sigErrs[fn.Name] = perr
			continue
		}
		if len(params) > 0 {
			if len(fn.Decorators) > 0 {
				g.sigErrs[fn.Name] = unsupported("decorator on a generic function", fn.Line())
				continue
			}
			g.genericFns[fn.Name] = &genericFn{name: fn.Name, node: fn, params: params}
			continue
		}
		sig, err := g.buildSig(fn, fn.Name, false)
		if err != nil {
			g.sigErrs[fn.Name] = err
			continue
		}
		g.sigs[fn.Name] = sig
		ps, _ := sig.defaultNodes()
		for i, p := range ps {
			p.defaultVar = fmt.Sprintf("_pyd_%s_%d", fn.Name, i)
		}
		if len(sig.decorators) > 0 {
			sig.goName = "_pyf_" + fn.Name
		}
	}
}

// decoratedTypes finds, before any code is generated, the type of every
// decorated top-level function, so that functions defined above it can
// call it. A decorator it cannot follow leaves the name unknown until its
// def is reached.
func (g *generator) decoratedTypes() {
	for name, sig := range g.sigs {
		if len(sig.decorators) == 0 {
			continue
		}
		t, err := g.decoratedType(sig)
		if err != nil {
			continue
		}
		sig.valueType = t
		g.globalTypes[name] = t
		g.globalLines[name] = sig.line
	}
}

// decoratedType is the type of the name after the decorators of sig ran,
// from the decorators' own signatures.
func (g *generator) decoratedType(sig *funcSig) (pyType, error) {
	cur := sig.goFuncType()
	if sig.plain() {
		cur = funcOf(sig.positionalTypes(), sig.ret)
	}
	for i := len(sig.decorators) - 1; i >= 0; i-- {
		d, err := g.typeOfDecorator(sig.decorators[i])
		if err != nil || !g.decoratorFits(d, cur) {
			if err == nil && !mentionsObject(d) && !mentionsObject(cur) {
				return "", unsupported(fmt.Sprintf("decorator of type %s applied to %s", d, cur), sig.decorators[i].Line())
			}
			cur = tObject
			continue
		}
		cur = d.funcRet()
	}
	return cur, nil
}

// decoratorFits reports whether a decorator of type d can be applied to a
// value of type cur with the types as they are.
func (g *generator) decoratorFits(d, cur pyType) bool {
	if !d.isFunc() || len(d.funcParams()) != 1 {
		return false
	}
	p := d.funcParams()[0]
	return p == cur || g.isSubtype(cur, p)
}

func (s *funcSig) positionalTypes() []pyType {
	out := make([]pyType, len(s.pos))
	for i, p := range s.pos {
		out[i] = p.typ
	}
	return out
}

// typeOfDecorator finds the type of a decorator expression from the
// top-level signatures alone: a function name, or a call of one.
func (g *generator) typeOfDecorator(n pyast.Node) (pyType, error) {
	switch n := n.(type) {
	case *pyast.Name:
		if sig, ok := g.sigs[n.ID]; ok && len(sig.decorators) == 0 && sig.plain() {
			return funcOf(sig.positionalTypes(), sig.ret), nil
		}
	case *pyast.Call:
		if callee, ok := n.Func.(*pyast.Name); ok {
			if sig, ok := g.sigs[callee.ID]; ok && len(sig.decorators) == 0 && sig.ret.isFunc() {
				return sig.ret, nil
			}
		}
	}
	return "", unsupported("decorator that is not a function or a call of one", n.Line())
}

// functionState is what changes when code generation enters a function.
type functionState struct {
	bigMode      bool
	gen          *genInfo
	genNone      bool
	method       *methodCtx
	specialRet   bool
	env          *scope
	fnName, top  string
	ret          pyType
	flow         *flowCtx
	handlers     []string
	fnGlobals    map[string]bool
	fnNonlocals  map[string]bool
	loopDepth    int
	compRoot     *scope
	callsUnknown bool
}

func (g *generator) saveFunctionState() functionState {
	return functionState{g.bigMode, g.gen, g.genNone, g.method, g.specialRet, g.env, g.fnName, g.topFn, g.ret, g.flow, g.handlers, g.fnGlobals, g.fnNonlocals,
		g.loopDepth, g.compRoot, g.callsUnknown}
}

func (g *generator) restoreFunctionState(s functionState) {
	g.bigMode = s.bigMode
	g.gen, g.genNone = s.gen, s.genNone
	g.method, g.specialRet = s.method, s.specialRet
	g.env, g.fnName, g.topFn, g.ret, g.flow, g.handlers = s.env, s.fnName, s.top, s.ret, s.flow, s.handlers
	g.fnGlobals, g.fnNonlocals, g.loopDepth, g.compRoot, g.callsUnknown = s.fnGlobals, s.fnNonlocals, s.loopDepth, s.compRoot, s.callsUnknown
}

// funcBody generates a function: a top-level function (header "func f(...)
// R") or a nested one (header "func(...) R", a function literal). It
// returns the text without a trailing newline.
func (g *generator) funcBody(fn *pyast.FunctionDef, sig *funcSig, header string, parent *scope, top bool, mc *methodCtx) (string, error) {
	saved := g.saveFunctionState()
	defer g.restoreFunctionState(saved)

	key := g.joinScope + "/" + fn.Name
	if mc != nil {
		key = mc.cls.name + "/" + mc.definer.name + "." + fn.Name
	}
	return g.withJoins(key, hasPlainBindings(fn.Body), func() (string, error) {
		return g.funcBodyOnce(fn, sig, header, parent, top, mc)
	})
}

// funcBodyOnce compiles the body once.
func (g *generator) funcBodyOnce(fn *pyast.FunctionDef, sig *funcSig, header string, parent *scope, top bool, mc *methodCtx) (string, error) {
	env := newScope(parent)
	env.boundary = true
	g.env = env
	g.bigMode = !hasAnnotations(fn)
	g.fnName, g.ret = fn.Name, sig.ret
	g.flow, g.handlers, g.loopDepth, g.compRoot = nil, nil, 0, nil
	g.fnGlobals, g.fnNonlocals = map[string]bool{}, map[string]bool{}
	g.callsUnknown = false
	if top {
		g.topFn = fn.Name
	}
	g.specialRet = false
	g.gen, g.genNone = sig.gen, false
	if sig.gen != nil {
		g.ret = sig.gen.ret
		if g.ret == tNoneType {
			g.ret, g.genNone = tNone, true
		}
	}
	if mc != nil {
		mc.selfVar = sig.selfName
		g.method = mc
		g.specialRet = mc.special
		g.topFn = ""
		switch mc.kind {
		case mClassMethod:
			env.declare(sig.selfName, classRefOf(mc.cls.name))
		case mStatic:
		default:
			env.declare(sig.selfName, mc.cls.typ())
		}
	}
	for _, p := range sig.goParams() {
		env.declare(p.name, p.typ)
	}

	var body strings.Builder
	stmts := fn.Body
	if sig.hasDoc {
		stmts = stmts[1:]
	}
	g.hoistBlockLocals(&body, fn, stmts)
	if err := g.predeclareDefs(&body, stmts); err != nil {
		return "", err
	}
	if err := g.stmtList(&body, stmts); err != nil {
		return "", err
	}
	// Go insists that a function with a result ends in a terminating
	// statement. Python would return None here, which breaks the annotation.
	if len(stmts) == 0 || func() bool { _, ok := stmts[len(stmts)-1].(*pyast.Return); return !ok }() {
		switch {
		case sig.gen != nil:
			// A generator that runs off its end returns None.
			if g.genNone {
				body.WriteString("return " + g.rt("NoneV") + "\n")
			} else {
				fmt.Fprintf(&body, "return %s\n", zeroValue(g, sig.gen.ret))
			}
		case sig.ret.isOptional():
			fmt.Fprintf(&body, "return %s[%s]()\n", g.rt("NoneOf"), g.goType(sig.ret.optElem()))
		case sig.ret == tObject:
			body.WriteString("return " + g.rt("NoneV") + "\n")
		case sig.ret != tNone:
			fmt.Fprintf(&body, "panic(%s(%q))\n", g.rt("MissingReturn"), fn.Name)
		}
	}

	var b strings.Builder
	if mc != nil && mc.kind == mMethod && sig.selfName == "self" && assignsName(fn.Body, "self") && strings.Contains(header, "(self *") {
		// The method rebinds self, which is a variable of the class type in
		// Python, not the pointer the receiver is: the receiver is renamed
		// and self is declared as the class.
		header = strings.Replace(header, "(self *", "(self_ *", 1)
		b.WriteString(header + " {\n")
		fmt.Fprintf(&b, "var self %s = self_\n_ = self\n", mc.cls.name)
	} else {
		b.WriteString(header + " {\n")
	}
	if top {
		b.WriteString(guardMarker(fn.Name) + "\n")
		if g.callsUnknown {
			if g.calls[fn.Name] == nil {
				g.calls[fn.Name] = map[string]bool{}
			}
			g.calls[fn.Name]["*"] = true
		}
	} else if g.callsUnknown || mc != nil {
		b.WriteString(g.rt("Enter") + "()\ndefer " + g.rt("Leave") + "()\n")
	}
	if sig.gen != nil {
		gi := sig.gen
		ctor := "return "
		tail := ""
		if gi.asyncGen {
			ctor, tail = "return "+g.rt("NewAsyncGen")+"(", ")"
		}
		fmt.Fprintf(&b, "%s%s[%s, %s, %s](func(%s *%s[%s, %s]) %s {\n%s})", ctor, g.rt("NewGen"),
			g.goType(gi.yield), g.goType(gi.send), g.goType(gi.ret), yielderVar, g.rt("Yielder"), g.goType(gi.yield), g.goType(gi.send), g.goType(gi.ret), body.String())
		if gi.iterator {
			b.WriteString(".Iterator()")
		}
		if gi.coro {
			fmt.Fprintf(&b, ".AsCoro(%q)", fn.Name)
		}
		b.WriteString(tail)
		b.WriteString("\n}")
		return b.String(), nil
	}
	b.WriteString(body.String())
	b.WriteString("}")
	return b.String(), nil
}

// goSignature spells the parameters and result of the Go function.
func (g *generator) goSignature(sig *funcSig) (params, result string) {
	ps := sig.goParams()
	parts := make([]string, len(ps))
	for i, p := range ps {
		parts[i] = p.name + " " + g.goType(p.typ)
	}
	if sig.ret != tNone {
		result = " " + g.goType(sig.ret)
	}
	return strings.Join(parts, ", "), result
}

// funcDef generates a top-level def. The function goes to package level,
// its defaults are evaluated where the def stands in the module, and its
// decorators run there.
func (g *generator) funcDef(funcs, main *strings.Builder, fn *pyast.FunctionDef) error {
	if g.genericFns[fn.Name] != nil {
		// The instantiations are written later; their defaults are
		// evaluated here.
		main.WriteString(lateMarker(fn.Name) + "\n")
		return nil
	}
	if g.defined[fn.Name] {
		return unsupported(fmt.Sprintf("redefinition of function '%s'", fn.Name), fn.Line())
	}
	if err := g.sigErrs[fn.Name]; err != nil {
		return err
	}
	sig := g.sigs[fn.Name]
	g.defined[fn.Name] = true

	if err := g.assignDefaults(main, sig, false); err != nil {
		return err
	}
	params, result := g.goSignature(sig)
	text, err := g.funcBody(fn, sig, fmt.Sprintf("func %s(%s)%s", sig.goName, params, result), nil, true, nil)
	if err != nil {
		return err
	}
	funcs.WriteString(text + "\n\n")

	if len(sig.decorators) == 0 {
		g.register(main, "func:"+sig.goName, func() string {
			return fmt.Sprintf("%s(%s, %s)\n", g.rt("RegisterFunc"), sig.goName, g.funcInfoLit(sig.name, sig))
		})
		if g.usesPickle {
			module := "__main__"
			if i := strings.LastIndex(sig.name, "_pm_"); i >= 0 {
				module = strings.ReplaceAll(sig.name[:i], "__", ".")
			}
			fmt.Fprintf(main, "%s(%q, %q, any(%s))\n", g.rt("RegisterGlobal"), module, pyName(sig.name), sig.goName)
		}
		return nil
	}
	cur, err := g.functionValue(sig)
	if err != nil {
		return err
	}
	applied, err := g.applyDecorators(sig, cur)
	if err != nil {
		return err
	}
	g.sigs[fn.Name].valueType = applied.typ
	g.globalTypes[fn.Name] = applied.typ
	g.declareGlobal(main, fn.Name, applied.typ, applied.code)
	return nil
}

// assignDefaults evaluates the default values of a def, once, at the place
// the def is executed. Top-level defaults live in package variables; nested
// ones in variables declared at the start of the block.
func (g *generator) assignDefaults(b *strings.Builder, sig *funcSig, nested bool) error {
	ps, nodes := sig.defaultNodes()
	for i, p := range ps {
		v, err := g.exprAs(nodes[i], p.typ)
		if err != nil {
			return err
		}
		v, err = g.coerce(v, p.typ, fmt.Sprintf("default of parameter '%s'", p.name), nodes[i].Line())
		if err != nil {
			return err
		}
		if !nested {
			g.defaultVars = append(g.defaultVars, defaultVar{p.defaultVar, p.typ})
		}
		fmt.Fprintf(b, "%s = %s\n", p.defaultVar, v.code)
	}
	return nil
}

type defaultVar struct {
	name string
	typ  pyType
}

// functionValue is a function used as a value: its name, typed as a Callable.
func (g *generator) functionValue(sig *funcSig) (value, error) {
	g.markValue("func:" + sig.goName)
	if !sig.plain() {
		// Usable only where a Callable is expected, through an adapter.
		return value{code: sig.goName, typ: sig.goFuncType(), prec: precAtom, fsig: sig}, nil
	}
	// A defaulted parameter is part of the type: the value must be called
	// with all of them.
	return value{code: sig.goName, typ: funcOf(sig.positionalTypes(), sig.ret), prec: precAtom, fsig: sig}, nil
}

// applyDecorators wraps a function value in its decorators, innermost
// first, as Python does.
func (g *generator) applyDecorators(sig *funcSig, cur value) (value, error) {
	for i := len(sig.decorators) - 1; i >= 0; i-- {
		d, err := g.typedExpr(sig.decorators[i])
		if err != nil {
			return value{}, err
		}
		if !g.decoratorFits(d.typ, cur.typ) {
			if !mentionsObject(d.typ) && !mentionsObject(cur.typ) {
				if d.typ.isFunc() && len(d.typ.funcParams()) == 1 {
					return value{}, unsupported(fmt.Sprintf("decorator expects %s, got %s", d.typ.funcParams()[0], cur.typ), sig.decorators[i].Line())
				}
				return value{}, unsupported(fmt.Sprintf("decorator of type %s", d.typ), sig.decorators[i].Line())
			}
			// Not a decorator the types follow: call it when the program runs.
			cur = g.dynValue(g.rt("DynCall")+"("+g.boxed(d)+", "+g.boxed(cur)+")", tObject)
			continue
		}
		cur = value{code: d.at(precAtom) + "(" + g.relabel(cur, d.typ.funcParams()[0]).code + ")", typ: d.typ.funcRet(), prec: precAtom}
	}
	return cur, nil
}

// predeclareDefs declares the functions defined directly in a block before
// the block's statements run, so that they can call each other and
// themselves whatever their order. The def statement assigns the closure.
func (g *generator) predeclareDefs(b *strings.Builder, stmts []pyast.Node) error {
	for _, st := range stmts {
		fn, ok := st.(*pyast.FunctionDef)
		if !ok {
			continue
		}
		qual := fn.Name
		if g.fnName != "" {
			qual = g.fnName + ".<locals>." + fn.Name
		}
		sig, err := g.buildSig(fn, qual, true)
		if err != nil {
			return err
		}
		if _, exists := g.env.lookupWithinFunction(fn.Name); exists && g.env.vars[fn.Name] != "" {
			return unsupported(fmt.Sprintf("redefinition of function '%s'", fn.Name), fn.Line())
		}
		ps, _ := sig.defaultNodes()
		for _, p := range ps {
			p.defaultVar = g.tmp("default_" + p.name)
			fmt.Fprintf(b, "var %s %s\n_ = %s\n", p.defaultVar, g.goType(p.typ), p.defaultVar)
		}
		t := sig.goFuncType()
		if len(sig.decorators) > 0 {
			dt, err := g.nestedDecoratedType(sig)
			if err != nil {
				return err
			}
			t = dt
		}
		sig.valueType = t
		g.declareFunc(b, fn.Name, t, sig)
	}
	return nil
}

// nestedDecoratedType types the decorators of a nested function.
func (g *generator) nestedDecoratedType(sig *funcSig) (pyType, error) {
	cur := sig.goFuncType()
	for i := len(sig.decorators) - 1; i >= 0; i-- {
		d, err := g.typedExpr(sig.decorators[i])
		if err != nil {
			return "", err
		}
		if !g.decoratorFits(d.typ, cur) {
			if !mentionsObject(d.typ) && !mentionsObject(cur) {
				return "", unsupported(fmt.Sprintf("decorator of type %s applied to %s", d.typ, cur), sig.decorators[i].Line())
			}
			cur = tObject
			continue
		}
		cur = d.typ.funcRet()
	}
	return cur, nil
}

// declareFunc declares the variable that holds a nested function.
func (g *generator) declareFunc(b *strings.Builder, name string, t pyType, sig *funcSig) {
	if g.atModuleLevel() {
		g.mainEnv.sigs[name] = sig
		g.declareGlobal(b, name, t, "")
		return
	}
	g.env.sigs[name] = sig
	g.declareLocal(b, name, t, "")
}

// nestedDef generates the def statement of a nested function: it
// evaluates the defaults, builds the closure and assigns it.
func (g *generator) nestedDef(b *strings.Builder, fn *pyast.FunctionDef) error {
	sig, _ := g.env.lookupSig(fn.Name)
	if sig == nil || sig.node != fn {
		return unsupported("function definition in an unexpected place", fn.Line())
	}
	if err := g.assignDefaults(b, sig, true); err != nil {
		return err
	}
	params, result := g.goSignature(sig)
	lit, err := g.funcBody(fn, sig, fmt.Sprintf("func(%s)%s", params, result), g.env, false, nil)
	if err != nil {
		return err
	}
	code := value{code: lit, typ: sig.goFuncType(), prec: precAtom}
	if len(sig.decorators) > 0 {
		code, err = g.applyDecorators(sig, code)
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(b, "%s = %s\n", fn.Name, code.code)
	if g.atModuleLevel() {
		g.markGlobalSet(b, fn.Name)
	}
	return nil
}

// lambda generates "lambda x, y: body" for the given parameter types. If
// ret is not empty the body must have that type.
func (g *generator) lambda(n *pyast.Lambda, params []pyType, ret pyType) (value, error) {
	a := n.Args
	if a.Vararg != nil || a.Kwarg != nil || len(a.Kwonly) > 0 || len(a.Defaults) > 0 {
		return value{}, unsupported("lambda with default, keyword-only, *args or **kwargs parameters", n.Line())
	}
	args := a.Positional()
	if len(args) != len(params) {
		return value{}, unsupported(fmt.Sprintf("lambda with %d parameters where %d are expected", len(args), len(params)), n.Line())
	}

	saved := g.saveFunctionState()
	defer g.restoreFunctionState(saved)
	env := newScope(g.env)
	env.boundary = true
	g.env = env
	g.flow, g.handlers, g.loopDepth = nil, nil, 0
	g.fnGlobals, g.fnNonlocals = map[string]bool{}, map[string]bool{}
	g.callsUnknown = false
	if g.fnName == "" {
		g.fnName = "<lambda>"
	}

	names := make([]string, len(args))
	for i, p := range args {
		if err := checkName(p.Name, p.Line()); err != nil {
			return value{}, err
		}
		if env.declare(p.Name, params[i]) == false {
			return value{}, unsupported(fmt.Sprintf("duplicate parameter '%s'", p.Name), p.Line())
		}
		names[i] = p.Name + " " + g.goType(params[i])
	}

	var body value
	var err error
	if ret == tNone || ret == "" {
		body, err = g.expr(n.Body)
	} else {
		body, err = g.exprAs(n.Body, ret)
	}
	if err != nil {
		return value{}, err
	}
	retType := body.typ
	if ret != "" {
		body, err = g.coerce(body, ret, "lambda result", n.Body.Line())
		if err != nil {
			return value{}, err
		}
		retType = ret
	}
	if retType == tExc {
		return value{}, unsupported("exception returned from a lambda", n.Line())
	}

	guard := ""
	if g.callsUnknown {
		guard = g.rt("Enter") + "()\ndefer " + g.rt("Leave") + "()\n"
	}
	var code string
	if retType == tNone {
		code = fmt.Sprintf("func(%s) {\n%s%s\n}", strings.Join(names, ", "), guard, body.code)
	} else {
		code = fmt.Sprintf("func(%s) %s {\n%sreturn %s\n}", strings.Join(names, ", "), g.goType(retType), guard, body.code)
	}
	return value{code: code, typ: funcOf(params, retType), prec: precAtom}, nil
}

// coerce checks that v can go where a value of type want is expected and
// converts it when Python's rules allow: any iterable of T goes where
// Iterable[T] is expected, and a tuple display goes where tuple[T, ...] is.
func (g *generator) coerce(v value, want pyType, dest string, line int) (value, error) {
	if v.typ == want {
		return v, nil
	}
	if g.isSubtype(v.typ, want) {
		if want == tObject {
			v = g.escape(v)
		}
		return g.relabel(v, want), nil
	}
	if v.typ == tObject && want != tNone && !want.isClassRef() && want != tExc {
		return value{code: g.rt("Cast") + "[" + g.goType(want) + "](" + v.code + ")", typ: want, prec: precAtom}, nil
	}
	if v.typ.isOptional() && g.isSubtype(v.typ.optElem(), want) && !want.isOptional() {
		// A variable that might be None goes where a value is needed, as in
		// Python: it fails when it is None, not when it is declared.
		inner := value{code: v.at(precAtom) + ".Unwrap(\"TypeError\", \"NoneType where a value is required\")", typ: v.typ.optElem(), prec: precAtom}
		return g.coerce(inner, want, dest, line)
	}
	if want.isIterable() {
		if code, elem, ok := g.iterableOf(v); ok && elem == want.elem() {
			return value{code: code, typ: want, prec: precAtom}, nil
		} else if ok && elem == tObject {
			return value{code: g.rt("CastIterable") + "[" + g.goType(want.elem()) + "](" + code + ")", typ: want, prec: precAtom}, nil
		}
	}
	if want.isFunc() && v.fsig != nil {
		if adapted, ok := g.adaptFunction(v, want); ok {
			return adapted, nil
		}
	}
	if want == tObject && v.typ == tExc {
		// An exception is an object like any other (a *PyError).
		return value{code: v.code, typ: tObject, prec: v.prec}, nil
	}
	return value{}, mismatch(dest, want, v.typ, line)
}

// adaptFunction lets a def with defaults, *args or keyword-only parameters
// stand where a Callable with fewer parameters is expected, by calling it
// with the defaults filled in:
//
//	func(a int64) int64 { return f(a, _pyd_f_0) }
func (g *generator) adaptFunction(v value, want pyType) (value, bool) {
	sig := v.fsig
	wp := want.funcParams()
	if len(wp) > len(sig.pos) || sig.ret != want.funcRet() {
		return value{}, false
	}
	var params, args []string
	for i, p := range sig.pos {
		if i < len(wp) {
			if p.typ != wp[i] {
				return value{}, false
			}
			name := fmt.Sprintf("_a%d", i+1)
			params = append(params, name+" "+g.goType(p.typ))
			args = append(args, name)
		} else if p.hasDefault {
			d, err := g.defaultArg(p)
			if err != nil {
				return value{}, false
			}
			args = append(args, d)
		} else {
			return value{}, false
		}
	}
	for _, p := range sig.kwonly {
		if !p.hasDefault {
			return value{}, false
		}
		args = append(args, p.defaultVar)
	}
	if sig.vararg != nil {
		args = append(args, g.rt("NewVarTuple")+"["+g.goType(sig.vararg.typ.elem())+"]()")
	}
	if sig.kwarg != nil {
		args = append(args, g.rt("NewDict")+"[string, "+g.goType(sig.kwarg.typ.valType())+"]()")
	}
	body := v.code + "(" + strings.Join(args, ", ") + ")"
	result := ""
	if want.funcRet() != tNone {
		result = " " + g.goType(want.funcRet())
		body = "return " + body
	}
	return value{code: "func(" + strings.Join(params, ", ") + ")" + result + " {\n" + body + "\n}", typ: want, prec: precAtom}, true
}

// builtinFunctions are the builtins that can stand for themselves as a
// function value, such as sorted(xs, key=len). Their parameter types come
// from the context, so each is wrapped in a lambda.
var builtinFunctions = map[string]bool{
	"len": true, "abs": true, "str": true, "int": true, "float": true, "bool": true, "repr": true,
	"round": true, "sum": true, "min": true, "max": true, "sorted": true, "list": true, "set": true,
	"tuple": true, "any": true, "all": true, "reversed": true, "dict": true, "print": true,
}

// builtinAsLambda returns "lambda a1, ..., an: name(a1, ..., an)" for a name
// that stands for a builtin, where a function of n parameters is expected.
func (g *generator) builtinAsLambda(n pyast.Node, want pyType) *pyast.Lambda {
	name, ok := n.(*pyast.Name)
	if !ok || !want.isFunc() || !builtinFunctions[name.ID] {
		return nil
	}
	if g.isVariable(name.ID) {
		return nil
	}
	if _, user := g.sigs[name.ID]; user {
		return nil
	}
	params := want.funcParams()
	var args []*pyast.Arg
	var names []pyast.Node
	for i := range params {
		id := fmt.Sprintf("_a%d", i+1)
		args = append(args, &pyast.Arg{Name: id, Lineno: name.Line()})
		names = append(names, &pyast.Name{ID: id, Lineno: name.Line()})
	}
	return &pyast.Lambda{
		Args:   &pyast.Arguments{Args: args},
		Body:   &pyast.Call{Func: &pyast.Name{ID: name.ID, Lineno: name.Line()}, Args: names, Lineno: name.Line()},
		Lineno: name.Line(),
	}
}

// functionAttribute reads __name__, __qualname__ and __doc__ of a function
// whose def is known.
func (g *generator) functionAttribute(sig *funcSig, attr string, line int) (value, error) {
	switch attr {
	case "__name__":
		return value{code: strconv.Quote(pyName(sig.name)), typ: tStr, prec: precAtom}, nil
	case "__qualname__":
		q := sig.qualname
		if i := strings.LastIndex(q, "."); i >= 0 {
			q = q[:i+1] + pyName(q[i+1:])
		} else {
			q = pyName(q)
		}
		return value{code: strconv.Quote(q), typ: tStr, prec: precAtom}, nil
	case "__doc__":
		if !sig.hasDoc {
			return value{}, unsupported(fmt.Sprintf("__doc__ of '%s', which has no docstring (it would be None)", sig.name), line)
		}
		return value{code: strconv.Quote(sig.doc), typ: tStr, prec: precAtom}, nil
	}
	return value{}, unsupported("attribute access '."+attr+"' on a function", line)
}

// protocolReturns are the results the special methods must have, which an
// unannotated method gets.
var protocolReturns = map[string]pyType{
	"__str__": tStr, "__repr__": tStr, "__format__": tStr,
	"__eq__": tBool, "__ne__": tBool, "__lt__": tBool, "__le__": tBool, "__gt__": tBool, "__ge__": tBool,
	"__bool__": tBool, "__contains__": tBool,
	"__len__": tInt, "__hash__": tInt, "__float__": tFloat,
}

// returnsValue reports whether a function body has a "return value" other than
// None, not counting the functions nested in it.
func returnsValue(body []pyast.Node) bool {
	found := false
	for _, st := range body {
		walkPrune(st, func(n pyast.Node) bool {
			switch r := n.(type) {
			case *pyast.Return:
				if r.Value != nil {
					if c, ok := r.Value.(*pyast.Constant); !ok || c.PyType != "none" {
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

// hasPropertyDecorator reports whether a method is a property or its setter.
func hasPropertyDecorator(fn *pyast.FunctionDef) bool {
	for _, d := range fn.Decorators {
		switch d := d.(type) {
		case *pyast.Name:
			if d.ID == "property" {
				return true
			}
		case *pyast.Attribute:
			return true
		}
	}
	return false
}

// mentionsObject reports whether a type has an object in it.
func mentionsObject(t pyType) bool { return strings.Contains(string(t), "object") }

// hasAnnotations reports whether a def has a type hint on a parameter or on
// its result. A def without any does its int arithmetic on objects.
func hasAnnotations(fn *pyast.FunctionDef) bool {
	if fn.Returns != nil {
		return true
	}
	a := fn.Args
	all := append(append([]*pyast.Arg{}, a.Positional()...), a.Kwonly...)
	if a.Vararg != nil {
		all = append(all, a.Vararg)
	}
	if a.Kwarg != nil {
		all = append(all, a.Kwarg)
	}
	for _, p := range all {
		if p.Annotation != nil {
			return true
		}
	}
	return false
}

// moduleAnnotated reports whether a module has a type hint anywhere.
func moduleAnnotated(mod *pyast.Module) bool {
	found := false
	walk(mod, func(n pyast.Node) {
		switch n := n.(type) {
		case *pyast.AnnAssign:
			found = true
		case *pyast.FunctionDef:
			if hasAnnotations(n) {
				found = true
			}
		}
	})
	return found
}

// hoistBlockLocals declares at the top of a function the variables that are
// first assigned inside a compound statement. In Python they belong to the
// function, so code after the block can read them; in Go a declaration in the
// block would end with it. The type is the join found for the variable.
func (g *generator) hoistBlockLocals(b *strings.Builder, fn *pyast.FunctionDef, stmts []pyast.Node) {
	if g.annotatedNames == nil {
		return
	}
	declared := map[string]bool{}
	for _, p := range paramNamesOf(fn.Args) {
		declared[p] = true
	}
	for _, name := range hasGlobalOrNonlocal(stmts) {
		declared[name] = true
	}
	var order []string
	seen := map[string]bool{}
	var visitStmts func(list []pyast.Node, depth int)
	bind := func(target pyast.Node, depth int) {
		var names []string
		var collect func(n pyast.Node)
		collect = func(n pyast.Node) {
			switch x := n.(type) {
			case *pyast.Name:
				names = append(names, x.ID)
			case *pyast.Tuple:
				for _, e := range x.Elts {
					collect(e)
				}
			case *pyast.List:
				for _, e := range x.Elts {
					collect(e)
				}
			case *pyast.Starred:
				collect(x.Value)
			}
		}
		collect(target)
		for _, nm := range names {
			if depth == 0 {
				declared[nm] = true
			} else if !declared[nm] && !seen[nm] {
				seen[nm] = true
				order = append(order, nm)
			}
		}
	}
	visitStmts = func(list []pyast.Node, depth int) {
		for _, st := range list {
			switch x := st.(type) {
			case *pyast.Assign:
				for _, t := range x.Targets {
					bind(t, depth)
				}
			case *pyast.AugAssign:
				bind(x.Target, depth)
			case *pyast.AnnAssign:
				if nm, ok := x.Target.(*pyast.Name); ok {
					declared[nm.ID] = true
				}
			case *pyast.For:
				bind(x.Target, depth)
				visitStmts(x.Body, depth+1)
				visitStmts(x.Orelse, depth+1)
			case *pyast.While:
				visitStmts(x.Body, depth+1)
				visitStmts(x.Orelse, depth+1)
			case *pyast.If:
				visitStmts(x.Body, depth+1)
				visitStmts(x.Orelse, depth+1)
			case *pyast.With:
				for _, it := range x.Items {
					if it.Vars != nil {
						bind(it.Vars, depth)
					}
				}
				visitStmts(x.Body, depth+1)
			case *pyast.Try:
				visitStmts(x.Body, depth+1)
				for _, h := range x.Handlers {
					visitStmts(h.Body, depth+1)
				}
				visitStmts(x.Orelse, depth+1)
				visitStmts(x.Finalbody, depth+1)
			case *pyast.Match:
				for _, c := range x.Cases {
					visitStmts(c.Body, depth+1)
				}
			case *pyast.FunctionDef:
				declared[x.Name] = true
			case *pyast.ClassDef:
				declared[x.Name] = true
			}
		}
	}
	visitStmts(stmts, 0)
	for _, name := range order {
		if g.annotatedNames[g.joinScope][name] {
			continue
		}
		t, ok := g.joins[g.joinScope][name]
		if !ok || t == "" || t == tNone || t == tExc {
			continue
		}
		if _, exists := g.env.lookup(name); exists {
			continue
		}
		g.env.declare(name, t)
		fmt.Fprintf(b, "var %s %s\n_ = %s\n", name, g.goType(t), name)
	}
}

func paramNamesOf(a *pyast.Arguments) []string {
	var out []string
	if a == nil {
		return out
	}
	for _, p := range a.Positional() {
		out = append(out, p.Name)
	}
	if a.Vararg != nil {
		out = append(out, a.Vararg.Name)
	}
	for _, p := range a.Kwonly {
		out = append(out, p.Name)
	}
	if a.Kwarg != nil {
		out = append(out, a.Kwarg.Name)
	}
	return out
}

// hasGlobalOrNonlocal lists the names a function body declares global or
// nonlocal.
func hasGlobalOrNonlocal(stmts []pyast.Node) []string {
	var out []string
	for _, st := range stmts {
		walkPrune(st, func(n pyast.Node) bool {
			switch x := n.(type) {
			case *pyast.Global:
				out = append(out, x.Names...)
			case *pyast.Nonlocal:
				out = append(out, x.Names...)
			case *pyast.FunctionDef, *pyast.ClassDef, *pyast.Lambda:
				return false
			}
			return true
		})
	}
	return out
}

// assignsName reports whether a function body assigns to the name, not
// counting the functions nested in it.
func assignsName(body []pyast.Node, name string) bool {
	found := false
	isName := func(n pyast.Node) bool { x, ok := n.(*pyast.Name); return ok && x.ID == name }
	for _, st := range body {
		walkPrune(st, func(n pyast.Node) bool {
			switch x := n.(type) {
			case *pyast.FunctionDef, *pyast.Lambda, *pyast.ClassDef:
				return false
			case *pyast.Assign:
				for _, t := range x.Targets {
					if isName(t) {
						found = true
					}
				}
			case *pyast.AugAssign:
				found = found || isName(x.Target)
			case *pyast.AnnAssign:
				found = found || isName(x.Target)
			case *pyast.For:
				found = found || isName(x.Target)
			}
			return !found
		})
	}
	return found
}
