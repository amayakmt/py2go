package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// Expressions and statements that involve instances of classes.

func (g *generator) classOf(t pyType) *classInfo {
	if t.isClass() {
		return g.classes[string(t)]
	}
	return nil
}

// isSubtype reports whether a value of type got can go where want is
// expected: the same type, a subclass of a class, or anything for object.
func (g *generator) isSubtype(got, want pyType) bool {
	if got == want {
		return true
	}
	if want.isOptional() {
		return got == tNoneType || g.isSubtype(got, want.optElem())
	}
	if want == tObject {
		return got != tNone && got != "" && !got.isClassRef() && got != tExc
	}
	gc, wc := g.classOf(got), g.classOf(want)
	return gc != nil && wc != nil && gc.isSubclassOf(wc)
}

// relabel gives v the type want when v is a subclass instance, or any value
// where object is expected. The Go code needs no change, except that a
// constant gets its Go type before it is boxed.
func (g *generator) relabel(v value, want pyType) value {
	if want == "" || v.typ == want || !g.isSubtype(v.typ, want) {
		return v
	}
	if want.isOptional() {
		goElem := g.goType(want.optElem())
		if v.typ == tNoneType {
			none := g.rt("NoneOf") + "[" + goElem + "]()"
			if v.code != "" && v.code != g.rt("NoneV") {
				// The None came out of a call that had to run.
				none = "func() " + g.goType(want) + " {\n_ = " + v.code + "\nreturn " + none + "\n}()"
			}
			return value{code: none, typ: want, prec: precAtom}
		}
		inner := g.relabel(v, want.optElem())
		return value{code: g.rt("Some") + "[" + goElem + "](" + inner.typed().code + ")", typ: want, prec: precAtom}
	}
	if want == tObject && v.fsig != nil && v.typ.isFunc() && v.fsig.plain() && !v.fsig.hasDefaults() && v.fsig.name != "" {
		// A plain function keeps its name, for __name__.
		return value{code: g.rt("WithInfo") + "(" + v.code + ", " + g.funcInfoLit(v.fsig.name, v.fsig) + ")", typ: tObject, prec: precAtom}
	}
	if want == tObject && v.fsig != nil && v.typ.isFunc() && (len(v.fsig.kwonly) > 0 || v.fsig.vararg != nil || v.fsig.kwarg != nil || v.fsig.hasDefaults()) {
		// The function is called later with keywords or defaults: keep
		// its parameters with it.
		return value{code: "&" + g.rt("Callable") + "{Fn: " + v.code + ", Info: " + g.funcInfoSnapshot(v.fsig.name, v.fsig) + "}", typ: tObject, prec: precAtom}
	}
	if want == tObject {
		v = v.typed()
		if v.typ.isOptional() {
			// An Optional is not a Python object: None, or the value.
			v.code, v.prec = v.at(precAtom)+".Boxed()", precAtom
		}
		if v.konst && v.typ != tObject {
			v.code = g.goType(v.typ) + "(" + v.code + ")"
			v.prec = precAtom
			v.konst = false
		}
	}
	v.typ = want
	v.intLit, v.lit = false, false
	return v
}

// mangled applies the name mangling of private names (__x) inside the body
// of a class.
func (g *generator) mangled(attr string) string {
	if g.method == nil || !strings.HasPrefix(attr, "__") || strings.HasSuffix(attr, "__") {
		return attr
	}
	return "_" + strings.TrimLeft(g.method.definer.name, "_") + attr
}

// classNameValue is the value of a name that stands for a class.
func (g *generator) classNameValue(name string) (value, bool) {
	if g.isVariable(name) {
		return value{}, false
	}
	c := g.classes[name]
	if c == nil {
		return value{}, false
	}
	return value{typ: classRefOf(name), prec: precAtom, classRef: c}, true
}

// classRefValue is a variable that holds a class: cls in a classmethod.
func (g *generator) classRefValue(code string, t pyType) value {
	return value{code: code, typ: t, prec: precAtom, classRef: g.classes[t.classRefName()], exactRef: true}
}

// recvExpr generates the receiver of an attribute access or method call,
// which may be a class.
func (g *generator) recvExpr(n pyast.Node, attr ...string) (value, error) {
	v, err := g.expr(n)
	if err != nil {
		return v, err
	}
	if v.typ == tNone {
		return v, unsupported("use of a None value", n.Line())
	}
	if v.typ.isOptional() {
		// Using None as an object fails when it runs.
		code := v.at(precAtom) + ".Unwrap(\"TypeError\", \"'NoneType' object is not callable\")"
		if len(attr) > 0 {
			code = v.at(precAtom) + ".UnwrapAttr(" + strconv.Quote(attr[0]) + ")"
		}
		v = value{code: code, typ: v.typ.optElem(), prec: precAtom}
	}
	return v, nil
}

// constructorValue turns a class used as a value, as in map(Point, xs), into
// a function value.
func (g *generator) constructorValue(v value, line int) (value, error) {
	c := v.classRef
	sig := g.ctorGoSig(c)
	if !sig.plain() {
		return value{}, unsupported(fmt.Sprintf("class '%s' used as a function value (its __init__ has defaults or *args)", c.name), line)
	}
	return value{code: c.goNew(), typ: funcOf(sig.positionalTypes(), c.typ()), prec: precAtom, fsig: nil}, nil
}

// constructCall is C(args).
func (g *generator) constructCall(n *pyast.Call, c *classInfo) (value, error) {
	if c.ctorSig == nil && !c.exc {
		if len(n.Args) > 0 || n.HasKeywords {
			return value{}, unsupported(fmt.Sprintf("%s() takes no arguments", c.name), n.Line())
		}
		return value{code: c.goNew() + "()", typ: c.typ(), prec: precAtom}, nil
	}
	sig := g.ctorGoSig(c)
	fname := c.name
	if c.ctorSig != nil {
		fname = c.ctorOwner.name + ".__init__"
	}
	s2 := *sig
	s2.ret = c.typ()
	if c.exc && c.ctorSig != nil {
		s2.leadArg = "-1"
		if !hasSpread(n) {
			s2.leadArg = strconv.Itoa(len(n.Args))
		}
	}
	return g.bindCall(n, &s2, c.goNew(), fname)
}

// attrOf finds the member of c an attribute name refers to.
func (g *generator) attrOf(c *classInfo, attr string) *member {
	return c.lookup(g.mangled(attr))
}

// definedBySubclass finds, for an attribute that self of the method being
// compiled does not have, the member a subclass defines. Python allows a
// base class to use attributes that only its subclasses provide.
func (g *generator) definedBySubclass(recvNode pyast.Node, c *classInfo, name string) *member {
	nameNode, ok := recvNode.(*pyast.Name)
	if !ok || g.method == nil || g.method.kind != mMethod || nameNode.ID != g.selfName() || g.method.cls != c {
		return nil
	}
	for _, k := range g.classOrder {
		if k != c && k.isSubclassOf(c) {
			if m := k.lookup(name); m != nil {
				return m
			}
		}
	}
	// Or some other class does, as when a method only runs for instances
	// that have the attribute and the others never reach the call.
	for _, k := range g.classOrder {
		if k != c {
			if m := k.lookup(name); m != nil {
				return m
			}
		}
	}
	return nil
}

// instanceAttr reads recv.attr for an instance of c.
func (g *generator) instanceAttr(recv value, c *classInfo, attr string, line int) (value, error) {
	name := g.mangled(attr)
	m := c.lookup(name)
	switch {
	case m == nil:
		switch attr {
		case "__class__":
			return value{code: recv.code, typ: classRefOf(c.name), prec: recv.prec, classRef: c, dynamicRef: true}, nil
		case "__doc__":
			if c.hasDoc {
				return value{code: strconv.Quote(c.doc), typ: tStr, prec: precAtom}, nil
			}
			return value{}, unsupported(fmt.Sprintf("__doc__ of '%s', which has no docstring (it would be None)", c.name), line)
		}
		if hook := c.lookup("__getattr__"); hook != nil && hook.kind == mMethod && len(hook.sig.pos) == 1 && hook.sig.pos[0].typ == tStr {
			return value{code: recv.at(precAtom) + ".M___getattr__(" + strconv.Quote(attr) + ")", typ: hook.sig.ret, prec: precAtom}, nil
		}
		return g.dynValue(g.rt("DynGetAttr")+"("+recv.code+", "+strconv.Quote(attr)+")", tObject), nil
	case m.isAttr():
		return value{code: recv.at(precAtom) + ".Get_" + name + "()", typ: m.typ, prec: precAtom}, nil
	case m.isCallable():
		return g.boundMethod(recv, c, m, line)
	}
	return value{}, unsupported("attribute '"+attr+"'", line)
}

// boundMethod is obj.method used as a value, not called.
func (g *generator) boundMethod(recv value, c *classInfo, m *member, line int) (value, error) {
	if m.special {
		return value{}, unsupported("operator method '"+m.name+"' used as a value", line)
	}
	var code string
	switch m.kind {
	case mStatic:
		code = "_pys_" + m.owner.name + "_" + m.name
	default:
		code = recv.at(precAtom) + ".M_" + m.name
	}
	sig := m.sig
	t := sig.goFuncType()
	if sig.plain() {
		t = funcOf(sig.positionalTypes(), sig.ret)
	}
	return value{code: code, typ: t, prec: precAtom, fsig: sig}, nil
}

// classAttrRead is C.x for a class C.
func (g *generator) classAttrRead(c *classInfo, attr string, exactRef bool, line int) (value, error) {
	name := g.mangled(attr)
	switch attr {
	case "__name__", "__qualname__":
		return value{code: strconv.Quote(c.pyName()), typ: tStr, prec: precAtom}, nil
	case "__doc__":
		if c.hasDoc {
			return value{code: strconv.Quote(c.doc), typ: tStr, prec: precAtom}, nil
		}
		return value{}, unsupported(fmt.Sprintf("__doc__ of '%s', which has no docstring (it would be None)", c.name), line)
	case "__module__":
		return value{code: `"__main__"`, typ: tStr, prec: precAtom}, nil
	}
	m := c.lookup(name)
	if m == nil {
		return value{}, unsupported(fmt.Sprintf("attribute '%s' of class '%s'", attr, c.name), line)
	}
	switch m.kind {
	case mClassAttr:
		return value{code: classAttrVar(m.owner, m.name), typ: m.typ, prec: precAtom}, nil
	case mStatic:
		return g.boundMethod(value{}, c, m, line)
	case mClassMethod:
		sig := m.sig
		t := sig.goFuncType()
		if sig.plain() {
			t = funcOf(sig.positionalTypes(), sig.ret)
		}
		return value{code: "_pycm_" + c.name + "_" + m.name + "__" + m.owner.name, typ: t, prec: precAtom, fsig: sig}, nil
	case mField:
		return value{}, unsupported(fmt.Sprintf("'%s.%s' is an instance attribute", c.name, attr), line)
	case mProperty:
		return value{}, unsupported(fmt.Sprintf("'%s.%s' is a property, read it on an instance", c.name, attr), line)
	}
	if m.kind == mMethod && m.sig != nil && !m.special {
		return g.unboundMethod(c, m), nil
	}
	return value{}, unsupported(fmt.Sprintf("unbound method '%s.%s' used as a value", c.name, attr), line)
}

// unboundMethod is C.method used as a value: a function of the instance and
// the method's own parameters.
func (g *generator) unboundMethod(c *classInfo, m *member) value {
	params, result := g.goSignature(m.sig)
	self := "self " + c.name
	if params != "" {
		self += ", " + params
	}
	var args []string
	for _, p := range m.sig.goParams() {
		args = append(args, p.name)
	}
	call := fmt.Sprintf("self.M_%s(%s)", m.name, strings.Join(args, ", "))
	if m.sig.ret != tNone {
		call = "return " + call
	}
	fn := fmt.Sprintf("func(%s)%s { %s }", self, result, call)
	withSelf := *m.sig
	withSelf.pos = append([]param{{name: "self", typ: c.typ()}}, m.sig.pos...)
	info := g.funcInfoSnapshot(c.pyName()+"."+m.name, &withSelf)
	return value{code: g.rt("WithInfo") + "(" + fn + ", " + info + ")", typ: tObject, prec: precAtom}
}

// classAttribute handles n.Attr for a receiver that is a class or an
// instance of one. ok is false when the receiver is neither.
func (g *generator) classAttribute(n *pyast.Attribute) (v value, ok bool, err error) {
	// super().name
	if sup, isSuper := g.superTarget(n.Value); isSuper {
		v, err = g.superAttr(sup, n.Attr, n.Line())
		return v, true, err
	}
	recvNode := n.Value
	// type(x).__name__ and x.__class__.__name__ read the dynamic class.
	if call, isCall := recvNode.(*pyast.Call); isCall && n.Attr == "__name__" {
		if name, _ := callName(call); name == "type" && len(call.Args) == 1 && !call.HasKeywords && !g.isVariable("type") {
			arg, err := g.typedExpr(call.Args[0])
			if err != nil {
				return value{}, true, err
			}
			if arg.typ.isClass() {
				return value{code: arg.at(precAtom) + ".PyClassName()", typ: tStr, prec: precAtom}, true, nil
			}
			if arg.typ == tObject || arg.typ.isGen() {
				return value{code: g.rt("TypeName") + "(" + g.boxed(arg) + ")", typ: tStr, prec: precAtom}, true, nil
			}
			if arg.typ == tExc {
				return value{code: g.rt("BaseTypeName") + "(" + arg.at(precAtom) + ".Type)", typ: tStr, prec: precAtom}, true, nil
			}
			if arg.typ == tOptExc {
				return value{code: g.rt("ExcTypeName") + "(" + arg.code + ")", typ: tStr, prec: precAtom}, true, nil
			}
			return value{code: strconv.Quote(pythonTypeName(arg.typ)), typ: tStr, prec: precAtom}, true, nil
		}
	}
	if name, isName := recvNode.(*pyast.Name); isName {
		if cv, isClass := g.classNameValue(name.ID); isClass {
			v, err := g.classAttrRead(cv.classRef, n.Attr, false, n.Line())
			return v, true, err
		}
	}
	recv, err := g.recvExpr(recvNode, n.Attr)
	if err != nil {
		return value{}, true, err
	}
	switch {
	case recv.typ.isClassRef():
		c := recv.classRef
		if recv.dynamicRef {
			if n.Attr == "__name__" {
				return value{code: recv.at(precAtom) + ".PyClassName()", typ: tStr, prec: precAtom}, true, nil
			}
			return value{}, true, unsupported(fmt.Sprintf("attribute '%s' of the class of a value that is not exactly known", n.Attr), n.Line())
		}
		v, err := g.classAttrRead(c, n.Attr, recv.exactRef, n.Line())
		return v, true, err
	case recv.typ.isClass() && g.classOf(recv.typ).exc && isExcSlot(n.Attr) && g.classOf(recv.typ).lookup(g.mangled(n.Attr)) == nil:
		v, err := g.excAttr(value{code: recv.at(precAtom) + ".PyExc()", typ: tExc, prec: precAtom}, n.Attr, n.Line())
		return v, true, err
	case recv.typ.isClass():
		v, err := g.instanceAttr(recv, g.classOf(recv.typ), n.Attr, n.Line())
		return v, true, err
	case recv.typ == tExc:
		v, err := g.excAttr(recv, n.Attr, n.Line())
		return v, true, err
	case recv.typ == tObject:
		return g.dynValue(g.rt("DynGetAttr")+"("+recv.code+", "+strconv.Quote(n.Attr)+")", tObject), true, nil
	}
	return value{}, false, nil
}

// pythonTypeName is the name type(x).__name__ shows for a static type.
func pythonTypeName(t pyType) string {
	switch {
	case t.isList():
		return "list"
	case t.isTuple() || t.isVarTuple():
		return "tuple"
	case t.isDict():
		return "dict"
	case t.isSet():
		return "set"
	case t.isFunc():
		return "function"
	case t == tRange:
		return "range"
	case t.isGen():
		return "generator"
	}
	return string(t)
}

// excAttr reads an attribute of a caught exception.
func (g *generator) excAttr(recv value, attr string, line int) (value, error) {
	switch attr {
	case "exceptions":
		return value{code: g.rt("GroupExceptions") + "(" + recv.code + ")", typ: varTupleOf(tExc), prec: precAtom}, nil
	case "message":
		return value{code: g.rt("GroupMessage") + "(" + recv.code + ")", typ: tStr, prec: precAtom}, nil
	case "value":
		return value{code: g.rt("StopValue") + "(" + recv.at(precAtom) + ")", typ: tObject, prec: precAtom}, nil
	case "args":
		return value{code: g.rt("NewVarTuple") + "[any](" + recv.at(precAtom) + ".Args...)", typ: varTupleOf(tObject), prec: precAtom}, nil
	case "__cause__":
		return value{code: recv.at(precAtom) + ".Cause", typ: tOptExc, prec: precAtom}, nil
	case "__context__":
		return value{code: recv.at(precAtom) + ".Context", typ: tOptExc, prec: precAtom}, nil
	case "__suppress_context__":
		return value{code: recv.at(precAtom) + ".SuppressContext", typ: tBool, prec: precAtom}, nil
	}
	switch attr {
	case "code", "errno", "strerror", "filename", "filename2", "name", "path", "encoding", "object", "start", "end", "reason", "msg", "pos", "lineno", "colno", "doc", "pattern":
		return g.dynValue(g.rt("ExcAttr")+"("+recv.at(precAtom)+", "+strconv.Quote(attr)+")", tObject), nil
	}
	return value{}, unsupported("attribute '."+attr+"' of an exception", line)
}

// superInfo describes a super() call.
type superInfo struct {
	cls     *classInfo // the exact class of self
	definer *classInfo
	self    string
	static  bool // called from a classmethod: no instance
}

// superTarget recognises super() and super(C, self).
func (g *generator) superTarget(n pyast.Node) (superInfo, bool) {
	call, ok := n.(*pyast.Call)
	if !ok || call.HasKeywords {
		return superInfo{}, false
	}
	name, ok := call.Func.(*pyast.Name)
	if !ok || name.ID != "super" || g.isVariable("super") || g.method == nil {
		return superInfo{}, false
	}
	if len(call.Args) != 0 && len(call.Args) != 2 {
		return superInfo{}, false
	}
	mc := g.method
	return superInfo{cls: mc.cls, definer: mc.definer, self: g.selfCode(), static: mc.kind == mClassMethod}, true
}

// selfCode is the Go name of the receiver of the method being compiled.
func (g *generator) selfCode() string { return g.selfName() }

// superAfter finds, in the MRO of the exact class, the first class after the
// definer that has the name.
func (s superInfo) find(name string) *member {
	pos := -1
	for i, k := range s.cls.mro {
		if k == s.definer {
			pos = i
		}
	}
	for _, k := range s.cls.mro[pos+1:] {
		if m := k.members[name]; m != nil {
			return m
		}
	}
	return nil
}

// superAttr is super().attr without a call.
func (g *generator) superAttr(s superInfo, attr string, line int) (value, error) {
	m := s.find(g.mangled(attr))
	if m == nil {
		return value{}, unsupported("super()."+attr, line)
	}
	switch m.kind {
	case mProperty:
		return value{code: s.self + ".G_" + m.name + "__" + m.owner.name + "()", typ: m.typ, prec: precAtom}, nil
	case mClassAttr:
		return value{code: classAttrVar(m.owner, m.name), typ: m.typ, prec: precAtom}, nil
	}
	return value{}, unsupported("super()."+attr+" used as a value", line)
}

// superCall is super().name(args).
func (g *generator) superCall(n *pyast.Call, s superInfo, name string) (value, error) {
	m := s.find(g.mangled(name))
	if m != nil && m.isCallable() {
		var callee string
		switch m.kind {
		case mMethod:
			if s.static {
				return value{}, unsupported("super() call of an instance method from a classmethod", n.Line())
			}
			callee = s.self + "." + methodGoName(m, m.owner)
		case mClassMethod:
			callee = "_pycm_" + s.cls.name + "_" + m.name + "__" + m.owner.name
		case mStatic:
			callee = "_pys_" + m.owner.name + "_" + m.name
		}
		v, err := g.bindCall(n, m.sig, callee, m.sig.qualname)
		if err != nil {
			return v, err
		}
		if m.special {
			v = value{code: "func() " + g.goType(m.sig.ret) + " {\nr, _ := " + v.code + "\nreturn r\n}()", typ: m.sig.ret, prec: precAtom}
		}
		return v, nil
	}
	return g.superBuiltin(n, s, name)
}

// superBuiltin is a super() call that reaches object or a built-in
// exception class.
func (g *generator) superBuiltin(n *pyast.Call, s superInfo, name string) (value, error) {
	if s.static {
		return value{}, unsupported("super()."+name+"() in a classmethod", n.Line())
	}
	argValues := func() ([]string, error) {
		if n.HasKeywords {
			return nil, unsupported("keyword arguments in super()."+name+"()", n.Line())
		}
		out := make([]string, len(n.Args))
		for i, a := range n.Args {
			v, err := g.exprAs(a, tObject)
			if err != nil {
				return nil, err
			}
			v = g.relabel(v, tObject)
			out[i] = v.code
		}
		return out, nil
	}
	atom := func(code string, t pyType) (value, error) { return value{code: code, typ: t, prec: precAtom}, nil }
	switch name {
	case "__init__":
		args, err := argValues()
		if err != nil {
			return value{}, err
		}
		if s.cls.exc {
			return atom(s.self+".SetArgs("+strings.Join(args, ", ")+")", tNone)
		}
		if len(args) > 0 {
			return value{}, unsupported("object.__init__() takes exactly one argument (the instance to initialize)", n.Line())
		}
		return atom("func() {}()", tNone)
	case "__str__":
		if s.cls.exc {
			return atom(s.self+".E.BaseStr()", tStr)
		}
		return atom(s.self+".PyRepr()", tStr)
	case "__repr__":
		if s.cls.exc {
			return atom(s.self+".E.BaseRepr()", tStr)
		}
		return atom(g.rt("DefaultRepr")+"("+strconv.Quote(s.cls.name)+", "+s.self+")", tStr)
	case "__eq__":
		args, err := argValues()
		if err != nil || len(args) != 1 {
			return value{}, unsupported("super().__eq__ needs one argument", n.Line())
		}
		return atom(g.rt("Is")+"("+s.self+", "+args[0]+")", tBool)
	case "__hash__":
		return atom(g.rt("IdentityHash")+"("+s.self+")", tInt)
	}
	return value{}, unsupported("super()."+name+"()", n.Line())
}

// instanceCall is obj.name(args) for an instance of a class.
func (g *generator) instanceCall(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	c := g.classOf(recv.typ)
	name := g.mangled(m.Attr)
	mem := c.lookup(name)
	if mem == nil {
		if sub := g.definedBySubclass(m.Value, c, name); sub != nil && sub.isCallable() {
			// A method the base class calls on self and only subclasses
			// define. An instance of this exact class has no such attribute.
			ret := sub.sig.ret
			text := fmt.Sprintf("panic(%s(%q, %q))", g.rt("NoAttribute"), c.name, m.Attr)
			code := "func() {\n" + text + "\n}()"
			if ret != tNone {
				code = "func() " + g.goType(ret) + " {\n" + text + "\n}()"
			}
			return value{code: code, typ: ret, prec: precAtom}, nil
		}
		// An attribute that holds a function, or __getattr__.
		fv, err := g.instanceAttr(recv, c, m.Attr, n.Line())
		if err != nil {
			return value{}, err
		}
		if fv.typ.isClassRef() && fv.classRef != nil {
			return g.dynCall(n, g.classObject(fv))
		}
		if !fv.typ.isFunc() && g.boxable(fv) {
			return g.dynCall(n, fv)
		}
		if !fv.typ.isFunc() {
			return value{}, unsupported(fmt.Sprintf("call of attribute '%s' of type %s", m.Attr, fv.typ), n.Line())
		}
		return g.callValue(n, fv)
	}
	if mem.isCallable() {
		if mem.special {
			return value{}, unsupported(fmt.Sprintf("explicit call of operator method %s.%s (use the operator)", c.name, m.Attr), n.Line())
		}
		callee := recv.at(precAtom) + ".M_" + mem.name
		if mem.kind == mStatic {
			callee = "_pys_" + mem.owner.name + "_" + mem.name
		}
		return g.bindCall(n, mem.sig, callee, mem.sig.qualname)
	}
	fv, err := g.instanceAttr(recv, c, m.Attr, n.Line())
	if err != nil {
		return value{}, err
	}
	if fv.typ.isClassRef() && fv.classRef != nil {
		return g.dynCall(n, g.classObject(fv))
	}
	if !fv.typ.isFunc() && g.boxable(fv) {
		return g.dynCall(n, fv)
	}
	if !fv.typ.isFunc() {
		return value{}, unsupported(fmt.Sprintf("call of attribute '%s' of type %s", m.Attr, fv.typ), n.Line())
	}
	return g.callValue(n, fv)
}

// classLevelCall is C.name(args) or cls.name(args).
func (g *generator) classLevelCall(n *pyast.Call, m *pyast.Attribute, recv value) (value, error) {
	c := recv.classRef
	if recv.dynamicRef {
		return value{}, unsupported(fmt.Sprintf("method '%s' of the class of a value that is not exactly known", m.Attr), n.Line())
	}
	name := g.mangled(m.Attr)
	mem := c.lookup(name)
	if mem == nil {
		return value{}, unsupported(fmt.Sprintf("attribute '%s' of class '%s'", m.Attr, c.name), n.Line())
	}
	switch mem.kind {
	case mStatic:
		return g.bindCall(n, mem.sig, "_pys_"+mem.owner.name+"_"+mem.name, mem.sig.qualname)
	case mClassMethod:
		return g.bindCall(n, mem.sig, "_pycm_"+c.name+"_"+mem.name+"__"+mem.owner.name, mem.sig.qualname)
	case mMethod:
		return g.unboundCall(n, c, mem)
	}
	fv, err := g.classAttrRead(c, m.Attr, recv.exactRef, n.Line())
	if err != nil {
		return value{}, err
	}
	if fv.typ.isClassRef() && fv.classRef != nil {
		return g.dynCall(n, g.classObject(fv))
	}
	if !fv.typ.isFunc() && g.boxable(fv) {
		return g.dynCall(n, fv)
	}
	if !fv.typ.isFunc() {
		return value{}, unsupported(fmt.Sprintf("call of attribute '%s' of type %s", m.Attr, fv.typ), n.Line())
	}
	return g.callValue(n, fv)
}

// unboundCall is Base.method(self, args), the old way to call a base class.
func (g *generator) unboundCall(n *pyast.Call, c *classInfo, mem *member) (value, error) {
	if len(n.Args) == 0 {
		return value{}, unsupported(fmt.Sprintf("%s() missing 1 required positional argument: 'self'", mem.sig.qualname), n.Line())
	}
	first, ok := n.Args[0].(*pyast.Name)
	if !ok || g.method == nil || g.method.kind != mMethod || first.ID != g.selfName() || !g.method.cls.isSubclassOf(c) {
		return value{}, unsupported(fmt.Sprintf("call of the unbound method %s.%s on something other than self", c.name, mem.name), n.Line())
	}
	if mem.special {
		return value{}, unsupported("explicit call of operator method "+mem.name, n.Line())
	}
	rest := *n
	rest.Args = n.Args[1:]
	return g.bindCall(&rest, mem.sig, g.selfName()+"."+methodGoName(mem, mem.owner), mem.sig.qualname)
}

// ---- assignment to attributes ----

// attrStore describes where an assignment to recv.attr goes.
type attrStore struct {
	typ    pyType
	member *member                 // the attribute, when it is a field of the class
	set    func(src string) string // the Go statement that stores src
	read   func() string           // the Go expression that reads the current value
}

// resolveStore finds the destination of "recv.attr = ...". The receiver is
// evaluated once: when recvCode is not a plain name it is stored in a
// temporary written to b.
func (g *generator) resolveStore(b *strings.Builder, target *pyast.Attribute) (*attrStore, error) {
	line := target.Line()
	// C.x = v and cls.x = v
	if name, isName := target.Value.(*pyast.Name); isName {
		if cv, isClass := g.classNameValue(name.ID); isClass {
			return g.classStore(cv.classRef, target.Attr, line)
		}
	}
	if sup, isSuper := g.superTarget(target.Value); isSuper {
		_ = sup
		return nil, unsupported("assignment through super()", line)
	}
	recv, err := g.recvExpr(target.Value, target.Attr)
	if err != nil {
		return nil, err
	}
	switch {
	case recv.typ.isClassRef():
		if recv.dynamicRef {
			return nil, unsupported("assignment to an attribute of a class that is not exactly known", line)
		}
		return g.classStore(recv.classRef, target.Attr, line)
	case recv.typ.isClass():
	case recv.typ == tObject:
		recvCode := recv.at(precAtom)
		if !simpleOperand.MatchString(recvCode) {
			tmp := g.tmp("o")
			g.temp(b, tmp, recv.typ, recv.code)
			recvCode = tmp
		}
		attr := strconv.Quote(target.Attr)
		return &attrStore{
			typ:  tObject,
			set:  func(src string) string { return g.rt("DynSetAttr") + "(" + recvCode + ", " + attr + ", " + src + ")" },
			read: func() string { return g.rt("DynGetAttr") + "(" + recvCode + ", " + attr + ")" },
		}, nil
	default:
		return nil, unsupported(fmt.Sprintf("assignment to attribute '%s' of %s", target.Attr, recv.typ), line)
	}
	c := g.classOf(recv.typ)
	name := g.mangled(target.Attr)
	m := c.lookup(name)

	// The receiver is evaluated once.
	recvCode := recv.at(precAtom)
	if !simpleOperand.MatchString(recvCode) {
		tmp := g.tmp("o")
		g.temp(b, tmp, recv.typ, recv.code)
		recvCode = tmp
	}

	if m == nil && g.infer {
		// The first assignment to self.x in a method declares the attribute.
		if nameNode, ok := target.Value.(*pyast.Name); ok && g.method != nil && g.method.kind == mMethod && nameNode.ID == g.selfName() {
			return &attrStore{typ: "", set: func(src string) string { return "" }, read: func() string { return "" }}, nil
		}
	}
	if m == nil {
		// An attribute the class does not declare lives in the instance's
		// dictionary.
		attr := strconv.Quote(target.Attr)
		return &attrStore{
			typ:  tObject,
			set:  func(src string) string { return g.rt("DynSetAttr") + "(" + recvCode + ", " + attr + ", " + src + ")" },
			read: func() string { return g.rt("DynGetAttr") + "(" + recvCode + ", " + attr + ")" },
		}, nil
	}
	switch m.kind {
	case mField, mClassAttr:
		if m.kind == mClassAttr && !m.shadow {
			return nil, unsupported(fmt.Sprintf("assignment to class attribute '%s.%s' through an instance", c.name, target.Attr), line)
		}
	case mProperty:
		if m.setter == nil {
			return &attrStore{typ: m.typ,
				set: func(src string) string {
					return fmt.Sprintf("panic(%s(%q, %q))", g.rt("NoSetter"), c.name, target.Attr)
				},
				read: func() string { return recvCode + ".Get_" + name + "()" }}, nil
		}
	default:
		return nil, unsupported(fmt.Sprintf("assignment to method '%s' of class '%s'", target.Attr, c.name), line)
	}
	return &attrStore{
		typ:    m.typ,
		member: m,
		set:    func(src string) string { return fmt.Sprintf("%s.Set_%s(%s)", recvCode, name, src) },
		read:   func() string { return recvCode + ".Get_" + name + "()" },
	}, nil
}

// classStore is C.x = v.
func (g *generator) classStore(c *classInfo, attr string, line int) (*attrStore, error) {
	m := c.lookup(g.mangled(attr))
	if m == nil || m.kind != mClassAttr {
		return nil, unsupported(fmt.Sprintf("assignment to attribute '%s' of class '%s' (only existing class attributes can be assigned)", attr, c.name), line)
	}
	if m.owner != c {
		return nil, unsupported(fmt.Sprintf("assignment to '%s.%s', which is inherited from '%s' (it would create a new attribute)", c.name, attr, m.owner.name), line)
	}
	v := classAttrVar(m.owner, m.name)
	return &attrStore{
		typ:  m.typ,
		set:  func(src string) string { return v + " = " + src },
		read: func() string { return v },
	}, nil
}

// registerField adds the attribute that the first assignment to self.x in a
// method declares, while the classes are being inferred.
func (g *generator) registerField(attr string, t pyType, line int) {
	c := g.method.definer
	name := g.mangled(attr)
	key := c.name + "." + name
	if t == tNoneType {
		// self.x = None: x may hold something else too, and then it is
		// Optional. It is declared once an assignment gives it a type.
		g.noneFields[key] = true
		if m := c.lookup(name); m != nil && m.kind == mField && !m.typ.isOptional() {
			m.typ = optionalOf(m.typ)
			g.inferGrew = true
		}
		return
	}
	if m := c.lookup(name); m != nil {
		return
	}
	if g.noneFields[key] {
		t = optionalOf(t)
	}
	c.addMember(&member{kind: mField, name: name, owner: c, typ: t, line: line})
	g.inferGrew = true
}

// assignAttr is "recv.attr = value".
func (g *generator) assignAttr(b *strings.Builder, target *pyast.Attribute, valueNode pyast.Node, line int) error {
	store, err := g.resolveStore(b, target)
	if err != nil {
		return err
	}
	if store.typ == "" { // inference
		v, err := g.typedExpr(valueNode)
		if err != nil {
			return err
		}
		if v.typ == tExc {
			return unsupported("attribute that holds an exception", line)
		}
		g.registerField(target.Attr, v.typ, line)
		return nil
	}
	if g.infer && g.bigMode && store.member != nil && store.member.kind == mField && store.typ != tObject {
		// Code without annotations: an attribute that is given a value of
		// another type, or a value this analysis cannot type, holds any.
		if vt, err := g.typedExpr(valueNode); err == nil && vt.typ != store.typ && !g.isSubtype(vt.typ, store.typ) && ((vt.typ == tObject && store.typ == tInt) || (vt.typ.isScalar() && store.typ.isScalar()) || (vt.typ == tNoneType && store.typ.isScalar())) {
			store.member.typ = tObject
			g.inferGrew = true
			return nil
		}
	}
	v, err := g.exprAs(valueNode, store.typ)
	if err != nil {
		return err
	}
	v, err = g.coerce(v, store.typ, fmt.Sprintf("attribute '%s'", target.Attr), line)
	if err != nil {
		return err
	}
	b.WriteString(store.set(v.code) + "\n")
	return nil
}

// assignAttrValue stores an already evaluated value, from unpacking or from
// an annotated assignment.
func (g *generator) assignAttrValue(b *strings.Builder, target *pyast.Attribute, src string, t pyType, line int) error {
	store, err := g.resolveStore(b, target)
	if err != nil {
		return err
	}
	if store.typ == "" {
		g.registerField(target.Attr, t, line)
		return nil
	}
	if store.typ != t {
		sub := value{code: src, typ: t, prec: precAtom}
		sub, err = g.coerce(sub, store.typ, fmt.Sprintf("attribute '%s'", target.Attr), line)
		if err != nil {
			return err
		}
		src = sub.code
	}
	b.WriteString(store.set(src) + "\n")
	return nil
}

// annAssignAttr is "self.x: int = value".
func (g *generator) annAssignAttr(b *strings.Builder, n *pyast.AnnAssign, target *pyast.Attribute) error {
	t, err := g.annotationType(n.Annotation, false)
	if err != nil {
		return err
	}
	if n.Value == nil {
		return unsupported("annotation of an attribute without a value", n.Line())
	}
	if g.infer {
		if nameNode, ok := target.Value.(*pyast.Name); ok && g.method != nil && g.method.kind == mMethod && nameNode.ID == g.selfName() {
			if existing := g.method.definer.lookup(g.mangled(target.Attr)); existing == nil {
				g.registerField(target.Attr, t, n.Line())
			}
		}
	}
	v, err := g.exprAs(n.Value, t)
	if err != nil {
		return err
	}
	v, err = g.coerce(v, t, fmt.Sprintf("attribute '%s'", target.Attr), n.Line())
	if err != nil {
		return err
	}
	store, err := g.resolveStore(b, target)
	if err != nil {
		return err
	}
	if store.typ == "" {
		return nil
	}
	if store.typ != t {
		return mismatch(fmt.Sprintf("attribute '%s'", target.Attr), store.typ, t, n.Line())
	}
	b.WriteString(store.set(v.code) + "\n")
	return nil
}

// augAttr is "recv.attr += value".
func (g *generator) augAttr(b *strings.Builder, n *pyast.AugAssign, target *pyast.Attribute) error {
	store, err := g.resolveStore(b, target)
	if err != nil {
		return err
	}
	if store.typ == "" {
		return unsupported("augmented assignment to an attribute that is not assigned before", n.Line())
	}
	right, err := g.exprAs(n.Value, store.typ)
	if err != nil {
		return err
	}
	cur := value{code: store.read(), typ: store.typ, prec: precAtom}
	res, err := g.augmented(n.Op, cur, right, n.Line())
	if err != nil {
		return err
	}
	if res.typ != store.typ {
		res, err = g.coerce(res, store.typ, fmt.Sprintf("attribute '%s'", target.Attr), n.Line())
		if err != nil {
			return err
		}
	}
	if method, inPlace := inPlaceMethod(store.typ, n.Op); inPlace {
		fmt.Fprintf(b, "%s.%s(%s)\n", cur.at(precAtom), method, right.code)
		return nil
	}
	b.WriteString(store.set(res.code) + "\n")
	return nil
}

// augmented applies "op=" to two values: a class with the in-place method
// uses it, anything else is the binary operator.
func (g *generator) augmented(op string, cur, right value, line int) (value, error) {
	if cur.typ.isClass() {
		if v, ok, err := g.classInPlace(op, cur, right, line); ok || err != nil {
			return v, err
		}
	}
	if involvesObject(cur, right) {
		return g.dynValue(g.rt("DynInplace")+`("`+op+`", `+g.boxed(cur)+", "+g.boxed(right)+")", tObject), nil
	}
	return g.binary(op, cur, right, line)
}

// unwrapTypeName is the Python name of a static type for TypeError messages.
func (g *generator) typeNameOf(t pyType) string {
	if t.isClass() || t == tInt || t == tFloat || t == tStr || t == tBool {
		return pyName(string(t))
	}
	return pythonTypeName(t)
}

var _ = runtime.IsExceptionClass

// isExcSlot reports the attributes every exception has.
func isExcSlot(attr string) bool {
	return attr == "exceptions" || attr == "message" || attr == "value" || attr == "args" || attr == "__cause__" || attr == "__context__" || attr == "__suppress_context__"
}

// commonClass is the nearest class both a and b derive from.
func (g *generator) commonClass(a, b pyType) (pyType, bool) {
	ca, cb := g.classOf(a), g.classOf(b)
	if ca == nil || cb == nil {
		return "", false
	}
	for _, k := range ca.mro {
		if !k.builtin && cb.isSubclassOf(k) {
			return k.typ(), true
		}
	}
	return "", false
}

// classObject is a class used as a value, as a runtime.ClassRef.
func (g *generator) classObject(v value) value {
	if v.classRef != nil {
		g.markValue("class:" + v.classRef.name)
	}
	if v.dynamicRef {
		// Any class of the program may be the one behind the value.
		for name := range g.classes {
			g.markValue("class:" + name)
		}
	}
	code := g.rt("ClassOf") + "(" + strconv.Quote(v.classRef.pyName()) + ")"
	if v.dynamicRef {
		code = g.rt("ClassOf") + "(" + v.at(precAtom) + ".PyClassName())"
	}
	v.code, v.prec = code, precAtom
	if v.typ.isClassRef() {
		v.typ = tObject
	}
	return v
}
