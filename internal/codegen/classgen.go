package codegen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// methodCtx is what code generation knows while it compiles a method for an
// exact class.
type methodCtx struct {
	cls     *classInfo // the class self belongs to, exactly
	definer *classInfo // the class whose body the method is written in
	kind    memberKind
	special bool // returns (result, implemented)
	name    string
	selfVar string // what the method calls its first parameter
}

// selfName is the Go name of the receiver of the method being compiled.
func (g *generator) selfName() string {
	if g.method == nil || g.method.kind == mStatic {
		return ""
	}
	return g.method.selfVar
}

// settable reports whether instances can assign the attribute.
func (m *member) settable() bool {
	switch m.kind {
	case mField:
		return true
	case mClassAttr:
		return m.shadow
	case mProperty:
		return m.setter != nil
	}
	return false
}

func (m *member) isAttr() bool {
	return m.kind == mField || m.kind == mClassAttr || m.kind == mProperty
}

// goAttrName is the Go identifier of the storage of an attribute.
func fieldName(name string) string { return "f_" + name }
func flagField(name string) string { return "s_" + name }

// classAttrVar is the package variable of a class attribute.
func classAttrVar(owner *classInfo, name string) string { return "_pyca_" + owner.name + "_" + name }

// definerMethod is the Go name of the method m of class x compiled for the
// exact class c.
func methodGoName(m *member, definer *classInfo) string {
	return "M_" + m.name + "__" + definer.name
}

// instanceMethodCallee is the Go expression that calls the version of m
// written in definer on the exact class c, with receiver code recv.
func (g *generator) memberCallee(recv string, c *classInfo, m *member) string {
	switch m.kind {
	case mStatic:
		return "_pys_" + m.owner.name + "_" + m.name
	case mClassMethod:
		return "_pycm_" + c.name + "_" + m.name + "__" + m.owner.name
	case mProperty:
		return recv + ".G_" + m.name + "__" + m.owner.name
	}
	return recv + "." + methodGoName(m, m.owner)
}

// resultSpelling is the result list of a Go function for a Python return type.
func (g *generator) resultSpelling(ret pyType, special bool) string {
	switch {
	case special:
		return " (" + g.goType(ret) + ", bool)"
	case ret == tNone:
		return ""
	}
	return " " + g.goType(ret)
}

// ifaceMethod is the line of a Go interface for a callable member.
func (g *generator) ifaceMethod(m *member) string {
	params, _ := g.goSignature(m.sig)
	return fmt.Sprintf("M_%s(%s)%s", m.name, params, g.resultSpelling(m.sig.ret, m.special))
}

// classDecl emits a class statement: the Go declarations to funcs and the
// execution of the class body (default values, class attributes, the
// registration of exception classes) to main.
func (g *generator) classDecl(funcs, main *strings.Builder, c *classInfo) error {
	if c.emitted {
		return nil
	}
	c.emitted = true
	g.refreshClassAttrs(c)

	// Defaults of the methods, evaluated where the class statement stands.
	for _, name := range c.order {
		m := c.members[name]
		if m.sig != nil {
			if err := g.assignDefaults(main, m.sig, false); err != nil {
				return err
			}
		}
		if m.setter != nil {
			if err := g.assignDefaults(main, m.setter, false); err != nil {
				return err
			}
		}
	}
	// Class attributes.
	g.classBodyCur, g.classBodySeen = c, map[string]bool{}
	for _, name := range c.order {
		m := c.members[name]
		if m.kind != mClassAttr {
			continue
		}
		v, err := g.exprAs(m.value, m.typ)
		if err != nil {
			return err
		}
		v, err = g.coerce(v, m.typ, fmt.Sprintf("class attribute '%s.%s'", c.name, name), m.line)
		if err != nil {
			return err
		}
		varName := classAttrVar(c, name)
		g.defaultVars = append(g.defaultVars, defaultVar{varName, m.typ})
		g.classBodySeen[name] = true
		fmt.Fprintf(main, "%s = %s\n", varName, v.code)
		fmt.Fprintf(main, "%s(%s, %s, func() any { return %s })\n", g.rt("RegisterClassAttr"), strconv.Quote(c.pyName()), strconv.Quote(pyName(name)), varName)
	}
	g.classBodyCur, g.classBodySeen = nil, nil
	bases := []string{strconv.Quote(c.pyName())}
	for _, k := range c.mro[1:] {
		if !k.builtin && k.name != "object" {
			bases = append(bases, strconv.Quote(k.pyName()))
		}
	}
	fmt.Fprintf(main, "%s(%s)\n", g.rt("RegisterClassBases"), strings.Join(bases, ", "))
	if g.usesPickle && c.generic == nil {
		fmt.Fprintf(main, "%s(%s, func() any { return &%s{} })\n", g.rt("RegisterAlloc"), strconv.Quote(c.pyName()), c.goStruct())
	}
	if i := strings.LastIndex(c.name, "_pm_"); i >= 0 {
		fmt.Fprintf(main, "%s(%s, %s)\n", g.rt("ClassModule"), strconv.Quote(c.pyName()), strconv.Quote(strings.ReplaceAll(c.name[:i], "__", ".")))
	}
	if c.exc {
		names := []string{strconv.Quote(c.name)}
		for _, k := range c.mro[1:] {
			if k.name != "object" {
				names = append(names, strconv.Quote(k.name))
			}
		}
		fmt.Fprintf(main, "%s(%s, %s)\n", g.rt("RegisterClass"), strconv.Quote(c.name), strings.Join(names, ", "))
		if c.generic == nil {
			// An exception class held in a variable can be called too.
			if g.usesPickle {
				g.markValue("class:" + c.name) // unpickling calls it with its arguments
			}
			g.register(main, "class:"+c.name, func() string {
				ctor := c.goNew()
				if c.ctorSig != nil {
					// The constructor takes the number of arguments given first.
					params, _ := g.goSignature(c.ctorSig)
					var names []string
					for _, p := range c.ctorSig.goParams() {
						names = append(names, p.name)
					}
					ctor = fmt.Sprintf("func(%s) %s { return %s(%d, %s) }", params, c.name, c.goNew(), len(c.ctorSig.pos), strings.Join(names, ", "))
				}
				return fmt.Sprintf("%s(%s, %s)\n%s(%s, %s)\n", g.rt("RegisterCtor"), strconv.Quote(c.pyName()), ctor,
					g.rt("RegisterCtorInfo"), strconv.Quote(c.pyName()), g.funcInfoLit(c.pyName(), g.ctorGoSig(c)))
			})
		}
	} else if c.generic == nil {
		// A class held in a variable can be called through the runtime.
		g.register(main, "class:"+c.name, func() string {
			var mro []string
			for _, k := range c.mro {
				if !k.builtin {
					mro = append(mro, strconv.Quote(k.name))
				}
			}
			return fmt.Sprintf("%s(%s, %s)\n%s(%s, %s)\n%s(%s, %s)\n", g.rt("RegisterCtor"), strconv.Quote(c.pyName()), c.goNew(),
				g.rt("RegisterCtorInfo"), strconv.Quote(c.pyName()), g.funcInfoLit(c.pyName(), g.ctorGoSig(c)),
				g.rt("RegisterClass"), strconv.Quote(c.name), strings.Join(mro, ", "))
		})
	}

	if err := g.checkOverrides(c); err != nil {
		return err
	}
	// Static methods are plain functions, compiled once.
	for _, name := range c.order {
		m := c.members[name]
		if m.kind != mStatic {
			continue
		}
		if err := g.emitStatic(funcs, c, m); err != nil {
			return err
		}
	}
	return g.emitClassTypes(funcs, c)
}

func (g *generator) emitStatic(out *strings.Builder, c *classInfo, m *member) error {
	params, result := g.goSignature(m.sig)
	header := fmt.Sprintf("func _pys_%s_%s(%s)%s", c.name, m.name, params, result)
	text, err := g.funcBody(m.node, m.sig, header, nil, false, &methodCtx{cls: c, definer: c, kind: mStatic, name: m.name})
	if err != nil {
		return err
	}
	out.WriteString(text + "\n\n")
	return nil
}

// emitClassTypes writes the interface, the struct, the constructor and every
// method of one class.
func (g *generator) emitClassTypes(out *strings.Builder, c *classInfo) error {
	names := c.visibleNames()

	// The interface.
	var iface []string
	iface = append(iface, "PyClassName() string", "PyRepr() string", "PyStr() string")
	if c.exc {
		iface = append(iface, "PyExc() *"+g.rt("PyError"))
	}
	for _, k := range c.mro {
		if !k.builtin {
			iface = append(iface, "Is_"+k.name+"()")
		}
	}
	for _, n := range names {
		m := c.lookup(n)
		switch {
		case m.isAttr():
			iface = append(iface, fmt.Sprintf("Get_%s() %s", n, g.goType(m.typ)))
			if m.settable() {
				iface = append(iface, fmt.Sprintf("Set_%s(v %s)", n, g.goType(m.typ)))
			}
		case m.isCallable() && n != "__init__":
			iface = append(iface, g.ifaceMethod(m))
		}
	}
	fmt.Fprintf(out, "type %s interface {\n%s\n}\n\n", c.name, strings.Join(iface, "\n"))

	// The struct.
	// The byte keeps the struct from being zero-sized, since Go may give
	// all pointers to zero-sized values one address, and instances must differ.
	fields := []string{"_pyid byte", "_pyd map[string]any"}
	if c.exc {
		fields = append(fields, g.rt("ExcBase"))
	}
	for _, n := range names {
		m := c.lookup(n)
		switch {
		case m.kind == mField:
			fields = append(fields, fieldName(n)+" "+g.goType(m.typ))
		case m.kind == mClassAttr && m.shadow:
			fields = append(fields, fieldName(n)+" "+g.goType(m.typ), flagField(n)+" bool")
		}
	}
	fmt.Fprintf(out, "type %s struct {\n%s\n}\n\n", c.goStruct(), strings.Join(fields, "\n"))

	g.emitConstructor(out, c)
	if err := g.emitAccessors(out, c, names); err != nil {
		return err
	}
	if err := g.emitMethods(out, c); err != nil {
		return err
	}
	g.emitHooks(out, c)
	return nil
}

// emitConstructor writes the function a call C(...) turns into.
func (g *generator) emitConstructor(out *strings.Builder, c *classInfo) {
	sig := g.ctorGoSig(c)
	params, _ := g.goSignature(sig)
	if c.exc && c.ctorSig != nil {
		params = strings.TrimSuffix("_given int, "+params, ", ")
	}
	fmt.Fprintf(out, "func %s(%s) %s {\nself := &%s{}\n", c.goNew(), params, c.name, c.goStruct())
	if c.exc {
		if c.ctorSig == nil {
			fmt.Fprintf(out, "self.Init(%s, self, args.Items()...)\n", strconv.Quote(c.name))
		} else {
			var raw []string
			for _, p := range c.ctorSig.pos {
				raw = append(raw, p.name)
			}
			all := "[]any{" + strings.Join(raw, ", ") + "}"
			if c.ctorSig.vararg != nil {
				all = "append(" + all + ", " + c.ctorSig.vararg.name + ".Boxed()...)"
			}
			fmt.Fprintf(out, "self.Init(%s, self, %s(_given, %s...)...)\n", strconv.Quote(c.name), g.rt("RawArgs"), all)
		}
	}
	if c.ctorSig != nil {
		var names []string
		for _, p := range c.ctorSig.goParams() {
			names = append(names, p.name)
		}
		fmt.Fprintf(out, "self.M___init____%s(%s)\n", c.ctorOwner.name, strings.Join(names, ", "))
	}
	out.WriteString("return self\n}\n\n")
}

// anyArgs spells ", a, b" for the variadic part of Init.
func anyArgs(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return ", " + strings.Join(names, ", ")
}

// ctorGoSig is the signature of the Go constructor: the __init__ parameters,
// or for an exception class without __init__ the arguments BaseException
// takes, or nothing.
func (g *generator) ctorGoSig(c *classInfo) *funcSig {
	if c.ctorSig != nil {
		return c.ctorSig
	}
	if c.exc {
		return &funcSig{name: c.name, vararg: &param{name: "args", typ: varTupleOf(tObject)}, ret: tNone}
	}
	return &funcSig{name: c.name, ret: tNone}
}

// emitAccessors writes Get_x and Set_x for every attribute of the class.
func (g *generator) emitAccessors(out *strings.Builder, c *classInfo, names []string) error {
	recv := "(self *" + c.goStruct() + ")"
	frozen := c.dc != nil && c.dc.frozen
	for _, n := range names {
		m := c.lookup(n)
		if !m.isAttr() {
			continue
		}
		t := g.goType(m.typ)
		switch m.kind {
		case mField:
			fmt.Fprintf(out, "func %s Get_%s() %s { return self.%s }\n", recv, n, t, fieldName(n))
			if frozen {
				fmt.Fprintf(out, "func %s Set_%s(v %s) { panic(%s(%q)) }\n", recv, n, t, g.rt("FrozenInstance"), n)
			} else {
				fmt.Fprintf(out, "func %s Set_%s(v %s) { self.%s = v }\n", recv, n, t, fieldName(n))
			}
		case mClassAttr:
			if m.shadow {
				fmt.Fprintf(out, "func %s Get_%s() %s {\nif self.%s {\nreturn self.%s\n}\nreturn %s\n}\n", recv, n, t, flagField(n), fieldName(n), classAttrVar(m.owner, n))
				fmt.Fprintf(out, "func %s Set_%s(v %s) { self.%s = v; self.%s = true }\n", recv, n, t, fieldName(n), flagField(n))
			} else {
				fmt.Fprintf(out, "func %s Get_%s() %s { return %s }\n", recv, n, t, classAttrVar(m.owner, n))
			}
		case mProperty:
			fmt.Fprintf(out, "func %s Get_%s() %s { return self.G_%s__%s() }\n", recv, n, t, n, m.owner.name)
			if m.setter != nil {
				fmt.Fprintf(out, "func %s Set_%s(v %s) { self.S_%s__%s(v) }\n", recv, n, t, n, m.owner.name)
			}
		}
	}
	out.WriteString("\n")
	return nil
}

// emitMethods compiles every method of every class in the MRO for the exact
// class c, and the forwarding methods that the interface lists.
func (g *generator) emitMethods(out *strings.Builder, c *classInfo) error {
	recv := "(self *" + c.goStruct() + ")"
	for _, x := range c.mro {
		if x.builtin {
			continue
		}
		for _, name := range x.order {
			m := x.members[name]
			if m.sig == nil {
				continue
			}
			if err := g.emitMember(out, c, x, m); err != nil {
				return err
			}
		}
	}
	// Forwarding methods: the name an instance call uses, to the version that
	// the MRO picks.
	for _, n := range c.visibleNames() {
		m := c.lookup(n)
		if !m.isCallable() || n == "__init__" {
			continue
		}
		sig := m.sig
		var names []string
		for _, p := range sig.goParams() {
			names = append(names, p.name)
		}
		params, _ := g.goSignature(sig)
		target := g.memberCallee("self", c, m)
		if m.kind == mMethod {
			target = "self." + methodGoName(m, m.owner)
		}
		ret := g.resultSpelling(sig.ret, m.special)
		body := target + "(" + strings.Join(names, ", ") + ")"
		if ret != "" {
			body = "return " + body
		}
		fmt.Fprintf(out, "func %s M_%s(%s)%s {\n%s\n}\n", recv, n, params, ret, body)
	}
	out.WriteString("\n")
	return nil
}

// emitMember compiles one member written in x for the exact class c.
func (g *generator) emitMember(out *strings.Builder, c, x *classInfo, m *member) error {
	selfName := m.sig.selfName
	if selfName == "" {
		selfName = "self"
	}
	recv := fmt.Sprintf("(%s *%s)", selfName, c.goStruct())
	if m.synth != "" {
		return g.emitSynth(out, c, x, m, selfName)
	}
	switch m.kind {
	case mStatic:
		return nil
	case mClassMethod:
		params, result := g.goSignature(m.sig)
		header := fmt.Sprintf("func _pycm_%s_%s__%s(%s)%s", c.name, m.name, x.name, params, result)
		text, err := g.funcBody(m.node, m.sig, header, nil, false, &methodCtx{cls: c, definer: x, kind: mClassMethod, name: m.name})
		if err != nil {
			return err
		}
		out.WriteString(text + "\n\n")
		return nil
	case mProperty:
		_, result := g.goSignature(m.sig)
		header := fmt.Sprintf("func %s G_%s__%s()%s", recv, m.name, x.name, result)
		text, err := g.funcBody(m.node, m.sig, header, nil, false, &methodCtx{cls: c, definer: x, kind: mProperty, name: m.name})
		if err != nil {
			return err
		}
		out.WriteString(text + "\n\n")
		if m.setter != nil {
			params, _ := g.goSignature(m.setter)
			header := fmt.Sprintf("func %s S_%s__%s(%s)", recv, m.name, x.name, params)
			text, err := g.funcBody(m.setNode, m.setter, header, nil, false, &methodCtx{cls: c, definer: x, kind: mProperty, name: m.name})
			if err != nil {
				return err
			}
			out.WriteString(text + "\n\n")
		}
		return nil
	}
	params, _ := g.goSignature(m.sig)
	header := fmt.Sprintf("func %s %s(%s)%s", recv, methodGoName(m, x), params, g.resultSpelling(m.sig.ret, m.special))
	text, err := g.funcBody(m.node, m.sig, header, nil, false, &methodCtx{cls: c, definer: x, kind: mMethod, special: m.special, name: m.name})
	if err != nil {
		return err
	}
	out.WriteString(text + "\n\n")
	return nil
}

// hashMode says how instances of c hash: "user" calls __hash__, "none" means
// unhashable, "identity" is the default.
func (c *classInfo) hashMode() (string, *member) {
	for _, k := range c.mro {
		if m := k.members["__hash__"]; m != nil {
			return "user", m
		}
		if k.unhashable {
			return "none", nil
		}
		if m := k.members["__eq__"]; m != nil && !k.builtin {
			return "none", nil
		}
	}
	return "identity", nil
}

// emitHooks writes the methods the runtime calls on every instance: the
// class name, repr, str, and when the class defines them ==, ordering and
// hash.
func (g *generator) emitHooks(out *strings.Builder, c *classInfo) {
	recv := "(self *" + c.goStruct() + ")"
	for _, k := range c.mro {
		if !k.builtin {
			fmt.Fprintf(out, "func %s Is_%s() {}\n", recv, k.name)
		}
	}
	fmt.Fprintf(out, "func %s PyClassName() string { return %q }\n", recv, c.pyName())

	switch m := c.lookup("__repr__"); {
	case m != nil:
		fmt.Fprintf(out, "func %s PyRepr() string { return self.M___repr____%s() }\n", recv, m.owner.name)
	case c.exc:
		fmt.Fprintf(out, "func %s PyRepr() string { return self.E.BaseRepr() }\n", recv)
	default:
		fmt.Fprintf(out, "func %s PyRepr() string { return %s(%q, self) }\n", recv, g.rt("DefaultRepr"), c.pyName())
	}
	switch m := c.lookup("__str__"); {
	case m != nil:
		fmt.Fprintf(out, "func %s PyStr() string { return self.M___str____%s() }\n", recv, m.owner.name)
	case c.exc:
		fmt.Fprintf(out, "func %s PyStr() string { return self.E.BaseStr() }\n", recv)
	default:
		fmt.Fprintf(out, "func %s PyStr() string { return self.PyRepr() }\n", recv)
	}

	if m := c.lookup("__eq__"); m != nil {
		p := m.sig.pos[0]
		fmt.Fprintf(out, "func %s PyEq(o any) (bool, bool) {\n", recv)
		out.WriteString(g.assertParam("o", p.typ, "return false, false"))
		fmt.Fprintf(out, "return self.M___eq____%s(arg)\n}\n", m.owner.name)
	}
	var ops []string
	for _, op := range []string{"lt", "le", "gt", "ge"} {
		if c.lookup("__"+op+"__") != nil {
			ops = append(ops, op)
		}
	}
	if len(ops) > 0 {
		fmt.Fprintf(out, "func %s PyOrder(op string, o any) (bool, bool) {\nswitch op {\n", recv)
		for _, op := range ops {
			m := c.lookup("__" + op + "__")
			fmt.Fprintf(out, "case %q:\n", op)
			out.WriteString(g.assertParam("o", m.sig.pos[0].typ, "return false, false"))
			fmt.Fprintf(out, "return self.M___%s____%s(arg)\n", op, m.owner.name)
		}
		out.WriteString("}\nreturn false, false\n}\n")
	}
	switch mode, m := c.hashMode(); mode {
	case "user":
		fmt.Fprintf(out, "func %s PyHash() int64 { return self.M___hash____%s() }\n", recv, m.owner.name)
		fmt.Fprintf(out, "func %s PyUserHash() {}\n", recv)
	case "none":
		fmt.Fprintf(out, "func %s PyHash() int64 { panic(%s(%q)) }\n", recv, g.rt("UnhashableType"), c.name)
	default:
		fmt.Fprintf(out, "func %s PyHash() int64 { return %s(self) }\n", recv, g.rt("IdentityHash"))
	}
	out.WriteString("\n")
	g.emitDynHooks(out, c)
}

// assertParam converts the any argument o to the Go type of a method
// parameter as "arg", running fail when it has another type.
func (g *generator) assertParam(o string, t pyType, fail string) string {
	if t == tObject {
		return fmt.Sprintf("arg := %s\n", o)
	}
	return fmt.Sprintf("arg, ok := %s.(%s)\nif !ok {\n%s\n}\n", o, g.goType(t), fail)
}

// emitSynth writes a method the transpiler makes up for a dataclass.
func (g *generator) emitSynth(out *strings.Builder, c, x *classInfo, m *member, selfName string) error {
	recv := fmt.Sprintf("(%s *%s)", selfName, c.goStruct())
	dc := x.dc
	var fields []*dcField
	for _, f := range dc.fields {
		fields = append(fields, f)
	}
	switch m.synth {
	case "dcinit":
		params, _ := g.goSignature(m.sig)
		fmt.Fprintf(out, "func %s %s(%s) {\n", recv, methodGoName(m, x), params)
		out.WriteString(g.rt("Enter") + "()\ndefer " + g.rt("Leave") + "()\n")
		for _, f := range dc.fields {
			switch {
			case f.init:
				fmt.Fprintf(out, "%s.%s = %s\n", selfName, fieldName(f.name), f.name)
			case f.def != nil:
				v, err := g.exprAs(f.def, f.typ)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s.%s = %s\n", selfName, fieldName(f.name), v.code)
			case f.factory != nil:
				v, err := g.factoryValue(f)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s.%s = %s\n", selfName, fieldName(f.name), v)
			}
		}
		if post := c.lookup("__post_init__"); post != nil && post.kind == mMethod {
			if len(post.sig.pos) > 0 || post.sig.vararg != nil {
				return unsupported("__post_init__ with parameters (InitVar arrives in phase 12)", post.line)
			}
			fmt.Fprintf(out, "%s.%s()\n", selfName, methodGoName(post, post.owner))
		}
		out.WriteString("}\n\n")
	case "dcrepr":
		fmt.Fprintf(out, "func %s %s() string {\n", recv, methodGoName(m, x))
		var parts []string
		for _, f := range fields {
			if f.repr {
				parts = append(parts, fmt.Sprintf("%q + %s(%s.%s)", f.name+"=", g.rt("Repr"), selfName, fieldName(f.name)))
			}
		}
		body := strconv.Quote(c.pyName()+"(") + " + "
		if len(parts) > 0 {
			body += strings.Join(parts, " + \", \" + ") + " + "
		}
		fmt.Fprintf(out, "return %s\")\"\n}\n\n", body)
	case "dceq":
		fmt.Fprintf(out, "func %s %s(other any) (bool, bool) {\n", recv, methodGoName(m, x))
		fmt.Fprintf(out, "o, ok := other.(%s)\n_, exact := other.(*%s)\nif !ok || !exact {\nreturn false, false\n}\n", x.name, c.goStruct())
		fmt.Fprintf(out, "return %s(%s, %s), true\n}\n\n", g.rt("Eq"), g.dcSeq(selfName+".", fields), g.dcSeq("o.Get_", fields, true))
	case "dclt", "dcle", "dcgt", "dcge":
		op := strings.TrimPrefix(m.synth, "dc")
		fn := map[string]string{"lt": "Lt", "le": "Le", "gt": "Gt", "ge": "Ge"}[op]
		fmt.Fprintf(out, "func %s %s(other any) (bool, bool) {\n", recv, methodGoName(m, x))
		fmt.Fprintf(out, "o, ok := other.(%s)\n_, exact := other.(*%s)\nif !ok || !exact {\nreturn false, false\n}\n", x.name, c.goStruct())
		fmt.Fprintf(out, "return %s(%s, %s), true\n}\n\n", g.rt(fn), g.dcSeq(selfName+".", fields), g.dcSeq("o.Get_", fields, true))
	case "dchash":
		fmt.Fprintf(out, "func %s %s() int64 {\nreturn %s(%s)\n}\n\n", recv, methodGoName(m, x), g.rt("HashOf"), g.dcSeq(selfName+".", fields))
	default:
		return unsupported("generated method "+m.synth, m.line)
	}
	return nil
}

// factoryValue is the call of a default_factory.
func (g *generator) factoryValue(f *dcField) (string, error) {
	call := &pyast.Call{Func: f.factory, Lineno: f.line}
	v, err := g.exprAs(call, f.typ)
	if err != nil {
		return "", err
	}
	v, err = g.coerce(v, f.typ, "default_factory of '"+f.name+"'", f.line)
	if err != nil {
		return "", err
	}
	return v.code, nil
}

// dcSeq spells a tuple of the compared fields as a runtime sequence.
func (g *generator) dcSeq(prefix string, fields []*dcField, accessor ...bool) string {
	var items []string
	for _, f := range fields {
		if !f.compare {
			continue
		}
		if len(accessor) > 0 && accessor[0] {
			items = append(items, prefix+f.name+"()")
		} else {
			items = append(items, prefix+fieldName(f.name))
		}
	}
	return g.rt("AnySeq") + "{" + strings.Join(items, ", ") + "}"
}

// sortedKeys returns the keys of a map in order, for deterministic output.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
