package codegen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// The methods the runtime calls on an instance whose class it does not know:
// attribute access by name, operators, subscripts, calls and iteration. They
// forward to the typed methods of the class, converting each argument to the
// type the parameter has.

// binaryDunders maps an operator to its method and the reflected method.
var binaryDunders = []struct{ op, name, reflected string }{
	{"Add", "__add__", "__radd__"}, {"Sub", "__sub__", "__rsub__"}, {"Mult", "__mul__", "__rmul__"},
	{"Div", "__truediv__", "__rtruediv__"}, {"FloorDiv", "__floordiv__", "__rfloordiv__"},
	{"Mod", "__mod__", "__rmod__"}, {"Pow", "__pow__", "__rpow__"}, {"MatMult", "__matmul__", "__rmatmul__"},
	{"BitAnd", "__and__", "__rand__"}, {"BitOr", "__or__", "__ror__"}, {"BitXor", "__xor__", "__rxor__"},
	{"LShift", "__lshift__", "__rlshift__"}, {"RShift", "__rshift__", "__rrshift__"},
}

var unaryDunders = []struct{ op, name string }{
	{"USub", "__neg__"}, {"UAdd", "__pos__"}, {"Invert", "__invert__"}, {"Abs", "__abs__"},
}

// funcInfoLit spells the runtime description of a signature's parameters.
func (g *generator) funcInfoLit(name string, sig *funcSig) string {
	return g.funcInfoWith(name, sig, false)
}

// funcInfoSnapshot is funcInfoLit with the defaults read now: the values the
// parameters have at this moment are the ones the closure was defined with.
func (g *generator) funcInfoSnapshot(name string, sig *funcSig) string {
	return g.funcInfoWith(name, sig, true)
}

func (s *funcSig) hasDefaults() bool {
	for _, p := range append(append([]param{}, s.pos...), s.kwonly...) {
		if p.hasDefault {
			return true
		}
	}
	return false
}

func (g *generator) funcInfoWith(name string, sig *funcSig, snapshot bool) string {
	var params, defaults, kwonly []string
	for _, p := range sig.pos {
		params = append(params, strconv.Quote(pyName(p.name)))
	}
	for _, p := range sig.kwonly {
		params = append(params, strconv.Quote(pyName(p.name)))
		kwonly = append(kwonly, strconv.Quote(pyName(p.name))+": true")
	}
	for _, p := range append(append([]param{}, sig.pos...), sig.kwonly...) {
		if !p.hasDefault {
			continue
		}
		d, err := g.defaultArg(p)
		if err != nil || d == "" {
			continue
		}
		if snapshot {
			defaults = append(defaults, fmt.Sprintf("%s: func() func() any {\nv := any(%s)\nreturn func() any { return v }\n}()", strconv.Quote(pyName(p.name)), d))
		} else {
			defaults = append(defaults, fmt.Sprintf("%s: func() any { return %s }", strconv.Quote(pyName(p.name)), d))
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "&%s{Name: %s, Params: []string{%s}", g.rt("FuncInfo"), strconv.Quote(pyName(name)), strings.Join(params, ", "))
	if len(defaults) > 0 {
		fmt.Fprintf(&b, ", Defaults: map[string]func() any{%s}", strings.Join(defaults, ", "))
	}
	if sig.vararg != nil {
		fmt.Fprintf(&b, ", VarArgs: %s", strconv.Quote(pyName(sig.vararg.name)))
	}
	if sig.kwarg != nil {
		fmt.Fprintf(&b, ", KwArgs: %s", strconv.Quote(pyName(sig.kwarg.name)))
	}
	if len(kwonly) > 0 {
		fmt.Fprintf(&b, ", KwOnly: map[string]bool{%s}", strings.Join(kwonly, ", "))
	}
	if sig.gen != nil && (sig.gen.coro || sig.gen.asyncGen) {
		b.WriteString(", Async: true")
	}
	b.WriteString("}")
	return b.String()
}

// boundMethod is a Go function value that calls the method name of self.
func (g *generator) dynBound(c *classInfo, m *member) string {
	params, result := g.goSignature(m.sig)
	var args []string
	for _, p := range m.sig.goParams() {
		args = append(args, p.name)
	}
	call := fmt.Sprintf("self.M_%s(%s)", m.name, strings.Join(args, ", "))
	body := call
	if m.sig.ret != tNone {
		body = "return " + call
	}
	fn := fmt.Sprintf("func(%s)%s { %s }", params, result, body)
	return fmt.Sprintf("%s(%s, %s)", g.rt("WithInfo"), fn, g.funcInfoLit(c.pyName()+"."+m.name, m.sig))
}

// dynDunders are the special methods the runtime calls by name, as the
// built-ins round(), bytes() and the like do.
var dynDunders = map[string]bool{"__round__": true, "__bytes__": true, "__trunc__": true, "__floor__": true, "__ceil__": true, "__complex__": true, "__enter__": true, "__exit__": true, "__await__": true, "__aenter__": true, "__aexit__": true, "__aiter__": true, "__anext__": true}

func isDunder(name string) bool {
	return strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__")
}

// emitDynHooks writes the run-time hooks of a class.
func (g *generator) emitDynHooks(out *strings.Builder, c *classInfo) {
	recv := "(self *" + c.goStruct() + ")"
	var gets, sets []string
	for _, n := range c.visibleNames() {
		m := c.lookup(n)
		if m == nil || isDunder(n) && !dynDunders[n] {
			continue
		}
		switch {
		case m.isAttr():
			gets = append(gets, fmt.Sprintf("case %q:\nreturn self.Get_%s(), true", n, n))
			if m.settable() {
				sets = append(sets, fmt.Sprintf("case %q:\nself.Set_%s(%s[%s](v))\nreturn true", n, n, g.rt("Cast"), g.goType(m.typ)))
			}
		case m.isCallable() && m.sig != nil && !m.special:
			gets = append(gets, fmt.Sprintf("case %q:\nreturn %s, true", n, g.dynBound(c, m)))
		}
	}
	// Attributes, then the instance's dictionary, then methods.
	var attrGets, methodGets []string
	for _, g1 := range gets {
		if strings.Contains(g1, "self.Get_") {
			attrGets = append(attrGets, g1)
		} else {
			methodGets = append(methodGets, g1)
		}
	}
	fmt.Fprintf(out, "func %s PyGetAttr(name string) (any, bool) {\n", recv)
	if len(attrGets) > 0 {
		fmt.Fprintf(out, "switch name {\n%s\n}\n", strings.Join(attrGets, "\n"))
	}
	out.WriteString("if v, ok := self._pyd[name]; ok {\nreturn v, true\n}\n")
	if len(methodGets) > 0 {
		fmt.Fprintf(out, "switch name {\n%s\n}\n", strings.Join(methodGets, "\n"))
	}
	out.WriteString("return nil, false\n}\n")
	fmt.Fprintf(out, "func %s PySetAttr(name string, v any) bool {\n", recv)
	if len(sets) > 0 {
		fmt.Fprintf(out, "switch name {\n%s\n}\n", strings.Join(sets, "\n"))
	}
	if c.dc != nil && c.dc.frozen {
		fmt.Fprintf(out, "panic(%s(name))\n}\n", g.rt("FrozenInstance"))
	} else {
		out.WriteString("if self._pyd == nil {\nself._pyd = map[string]any{}\n}\nself._pyd[name] = v\nreturn true\n}\n")
	}
	fmt.Fprintf(out, "func %s PyDelAttr(name string) bool {\nif _, ok := self._pyd[name]; ok {\ndelete(self._pyd, name)\nreturn true\n}\nreturn false\n}\n", recv)
	if m := c.lookup("__getattr__"); m != nil && m.kind == mMethod && len(m.sig.pos) == 1 && !m.special {
		fmt.Fprintf(out, "func %s PyGetAttrHook(name string) (any, bool) {\nreturn self.M___getattr__(name), true\n}\n", recv)
	}

	// Binary operators, called with the instance on either side.
	var cases []string
	for _, d := range binaryDunders {
		for _, reflected := range []bool{false, true} {
			name := d.name
			if reflected {
				name = d.reflected
			}
			m := c.lookup(name)
			if m == nil || m.kind != mMethod || len(m.sig.pos) != 1 {
				continue
			}
			var b strings.Builder
			fmt.Fprintf(&b, "case op == %q && reflected == %v:\n", d.op, reflected)
			b.WriteString(g.assertParam("o", m.sig.pos[0].typ, "return nil, false"))
			if m.special {
				fmt.Fprintf(&b, "return self.M_%s(arg)", name)
			} else {
				fmt.Fprintf(&b, "return self.M_%s(arg), true", name)
			}
			cases = append(cases, b.String())
		}
	}
	if len(cases) > 0 {
		fmt.Fprintf(out, "func %s PyBinop(op string, o any, reflected bool) (any, bool) {\nswitch {\n%s\n}\nreturn nil, false\n}\n", recv, strings.Join(cases, "\n"))
	}
	cases = nil
	for _, d := range binaryDunders {
		name := "__i" + strings.TrimPrefix(d.name, "__")
		m := c.lookup(name)
		if m == nil || m.kind != mMethod || len(m.sig.pos) != 1 {
			continue
		}
		var b strings.Builder
		fmt.Fprintf(&b, "case %q:\n", d.op)
		b.WriteString(g.assertParam("o", m.sig.pos[0].typ, "return nil, false"))
		if m.special {
			fmt.Fprintf(&b, "return self.M_%s(arg)", name)
		} else {
			fmt.Fprintf(&b, "return self.M_%s(arg), true", name)
		}
		cases = append(cases, b.String())
	}
	if len(cases) > 0 {
		fmt.Fprintf(out, "func %s PyInplace(op string, o any) (any, bool) {\nswitch op {\n%s\n}\nreturn nil, false\n}\n", recv, strings.Join(cases, "\n"))
	}
	cases = nil
	for _, d := range unaryDunders {
		if m := c.lookup(d.name); m != nil && m.kind == mMethod && len(m.sig.pos) == 0 && !m.special {
			cases = append(cases, fmt.Sprintf("case %q:\nreturn self.M_%s(), true", d.op, d.name))
		}
	}
	if len(cases) > 0 {
		fmt.Fprintf(out, "func %s PyUnop(op string) (any, bool) {\nswitch op {\n%s\n}\nreturn nil, false\n}\n", recv, strings.Join(cases, "\n"))
	}

	one := func(name string, nargs int) *member {
		m := c.lookup(name)
		if m == nil || m.kind != mMethod || len(m.sig.pos) != nargs || m.special {
			return nil
		}
		return m
	}
	if m := one("__format__", 1); m != nil && (m.sig.pos[0].typ == tStr || m.sig.pos[0].typ == tObject) && m.sig.ret == tStr {
		fmt.Fprintf(out, "func %s PyFormat(spec string) string { return self.M___format__(spec) }\n", recv)
	}
	if m := one("__len__", 0); m != nil {
		fmt.Fprintf(out, "func %s PyLen() (int64, bool) { return self.M___len__(), true }\n", recv)
	}
	if m := one("__bool__", 0); m != nil {
		fmt.Fprintf(out, "func %s PyTruth() bool { return self.M___bool__() }\n", recv)
	} else if m := one("__len__", 0); m != nil {
		fmt.Fprintf(out, "func %s PyTruth() bool { return self.M___len__() != 0 }\n", recv)
	}
	if m := one("__int__", 0); m != nil {
		fmt.Fprintf(out, "func %s PyInt() (int64, bool) { return self.M___int__(), true }\n", recv)
	}
	if m := one("__float__", 0); m != nil {
		fmt.Fprintf(out, "func %s PyFloat() (float64, bool) { return self.M___float__(), true }\n", recv)
	}
	if m := one("__contains__", 1); m != nil {
		fmt.Fprintf(out, "func %s PyContains(item any) (bool, bool) {\n%sreturn self.M___contains__(arg), true\n}\n", recv, g.assertParam("item", m.sig.pos[0].typ, "return false, false"))
	}
	if m := one("__getitem__", 1); m != nil {
		fmt.Fprintf(out, "func %s PyGetItem(key any) (any, bool) {\n%sreturn self.M___getitem__(arg), true\n}\n", recv, g.assertParam("key", m.sig.pos[0].typ, "return nil, false"))
	}
	if m := one("__setitem__", 2); m != nil {
		fmt.Fprintf(out, "func %s PySetItem(key, v any) bool {\n%s%sself.M___setitem__(arg, arg2)\nreturn true\n}\n", recv,
			g.assertParam("key", m.sig.pos[0].typ, "return false"),
			strings.Replace(g.assertParam("v", m.sig.pos[1].typ, "return false"), "arg", "arg2", -1))
	}
	if m := one("__delitem__", 1); m != nil {
		fmt.Fprintf(out, "func %s PyDelItem(key any) bool {\n%sself.M___delitem__(arg)\nreturn true\n}\n", recv, g.assertParam("key", m.sig.pos[0].typ, "return false"))
	}
	if m := c.lookup("__call__"); m != nil && m.kind == mMethod && !m.special {
		fmt.Fprintf(out, "func %s PyCall(args []any, kw []%s) (any, bool) {\nreturn %s(%s, args, kw), true\n}\n", recv, g.rt("KW"), g.rt("DynCallKw"), g.dynBound(c, m))
	}
	if m := one("__iter__", 0); m != nil {
		if m.sig.ret == tObject {
			fmt.Fprintf(out, "func %s PyIterAny() (%s[any], bool) {\nreturn %s(self.M___iter__()), true\n}\n", recv, g.rt("Cursor"), g.rt("IterResult"))
		} else if code, elem, ok := g.userIterable(value{code: "self", typ: c.typ(), prec: precAtom}); ok {
			fmt.Fprintf(out, "func %s PyIterAny() (%s[any], bool) {\nreturn %s[%s](%s.Iter()), true\n}\n", recv, g.rt("Cursor"), g.rt("AnyCursor"), g.goType(elem), code)
		}
	}
	if m := one("__next__", 0); m != nil && m.sig.ret != tNone {
		fmt.Fprintf(out, "func %s PyNext() any {\nreturn self.M___next__()\n}\n", recv)
	}
	out.WriteString("\n")
}

var _ = sort.Strings

// register writes, where a def or class statement stands, the call that lets
// the runtime call it from an object. The line stays only if the function or
// class is used as a value somewhere.
func (g *generator) register(b *strings.Builder, key string, lines func() string) {
	saved := g.usesRuntime
	g.regLines[key] = lines()
	g.usesRuntime = saved
	b.WriteString("/*reg:" + key + "*/\n")
}

// markValue says that the function or class key is used as a value.
func (g *generator) markValue(key string) {
	if !g.recordJoins {
		g.regUsed[key] = true
	}
}

func (g *generator) placeRegistrations(text string) string {
	for key, lines := range g.regLines {
		repl := ""
		if g.regUsed[key] {
			repl = lines
			g.usesRuntime = true
		}
		text = strings.ReplaceAll(text, "/*reg:"+key+"*/\n", repl)
	}
	return text
}
