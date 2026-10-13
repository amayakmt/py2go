package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
	"github.com/amayakmt/py2go/runtime"
)

// Operators, protocols and builtins that reach the special methods of a
// class.

// dunderName maps an operator to the name of its special method.
var dunderName = map[string]string{
	"Add": "add", "Sub": "sub", "Mult": "mul", "Div": "truediv", "FloorDiv": "floordiv", "Mod": "mod",
	"Pow": "pow", "MatMult": "matmul", "BitAnd": "and", "BitOr": "or", "BitXor": "xor",
	"LShift": "lshift", "RShift": "rshift",
}

// operatorMethod finds a method of c that takes one operand of type arg.
func (g *generator) operatorMethod(c *classInfo, name string, arg pyType) *member {
	m := c.lookup(name)
	if m == nil || m.kind != mMethod || len(m.sig.pos) != 1 || m.sig.vararg != nil || m.sig.kwarg != nil || len(m.sig.kwonly) > 0 {
		return nil
	}
	if !g.isSubtype(arg, m.sig.pos[0].typ) {
		return nil
	}
	return m
}

// classBinary is "l op r" when an operand is an instance of a class: the
// left operand's method, then the reflected method of the right operand.
func (g *generator) classBinary(op string, l, r value, line int) (value, error) {
	name := dunderName[op]
	if op == "Mod" && l.typ == tStr {
		// "%s" % instance formats with the instance's str().
		return value{code: g.rt("Cast") + "[string](" + g.rt("DynBinary") + `("Mod", ` + l.typed().code + ", " + r.typed().code + "))", typ: tStr, prec: precAtom}, nil
	}
	if name == "" {
		return value{}, unsupported("operator "+opSymbols[op]+" on instances of classes", line)
	}
	lc, rc := g.classOf(l.typ), g.classOf(r.typ)
	var fwd, rev *member
	if lc != nil {
		fwd = g.operatorMethod(lc, "__"+name+"__", r.typ)
	}
	if rc != nil && l.typ != r.typ {
		rev = g.operatorMethod(rc, "__r"+name+"__", l.typ)
	}
	if fwd == nil && rev == nil {
		// The classes have the method, but not for this operand type. Python
		// finds out when the operator runs.
		var ret pyType
		if lc != nil {
			if m := lc.lookup("__" + name + "__"); m != nil && m.kind == mMethod {
				ret = m.sig.ret
			}
		}
		if rc != nil && ret == "" {
			if m := rc.lookup("__r" + name + "__"); m != nil && m.kind == mMethod {
				ret = m.sig.ret
			}
		}
		if ret == "" {
			// Neither class has the operator: the program fails with a
			// TypeError when it runs this. The type is a guess.
			ret = l.typ
		}
		return value{code: fmt.Sprintf("func() %s {\n_ = %s\n_ = %s\npanic(%s(%q, %q, %q))\n}()", g.goType(ret), l.typed().code, r.typed().code,
			g.rt("UnsupportedOperand"), opSymbols[op], g.typeNameOf(l.typ), g.typeNameOf(r.typ)), typ: ret, prec: precAtom}, nil
	}
	ret := pyType("")
	for _, m := range []*member{fwd, rev} {
		if m == nil {
			continue
		}
		if ret != "" && ret != m.sig.ret {
			return value{}, unsupported(fmt.Sprintf("operator %s whose methods return %s and %s", opSymbols[op], ret, m.sig.ret), line)
		}
		ret = m.sig.ret
	}
	var b strings.Builder
	fmt.Fprintf(&b, "func() %s {\nvar _l %s = %s\nvar _r %s = %s\n", g.goType(ret), g.goType(l.typ), l.typed().code, g.goType(r.typ), r.typed().code)
	if fwd != nil {
		fmt.Fprintf(&b, "if v, ok := _l.M___%s__(%s); ok {\nreturn v\n}\n", name, g.argAs("_r", r.typ, fwd.sig.pos[0].typ))
	}
	if rev != nil {
		fmt.Fprintf(&b, "if v, ok := _r.M___r%s__(%s); ok {\nreturn v\n}\n", name, g.argAs("_l", l.typ, rev.sig.pos[0].typ))
	}
	fmt.Fprintf(&b, "panic(%s(%q, %q, %q))\n}()", g.rt("UnsupportedOperand"), opSymbols[op], g.typeNameOf(l.typ), g.typeNameOf(r.typ))
	return value{code: b.String(), typ: ret, prec: precAtom}, nil
}

// argAs passes code of type from where a parameter of type to is expected.
func (g *generator) argAs(code string, from, to pyType) string {
	return code
}

// classInPlace is "x op= y" for an instance with the in-place method. ok is
// false when the class has none, and the binary operator is used instead.
func (g *generator) classInPlace(op string, cur, right value, line int) (value, bool, error) {
	name := dunderName[op]
	c := g.classOf(cur.typ)
	if name == "" || c == nil {
		return value{}, false, nil
	}
	m := g.operatorMethod(c, "__i"+name+"__", right.typ)
	if m == nil {
		return value{}, false, nil
	}
	fallback, err := g.classBinary(op, value{code: "_l", typ: cur.typ, prec: precAtom}, value{code: "_r", typ: right.typ, prec: precAtom}, line)
	if err != nil {
		// No fallback to the binary operator: the in-place method has to work.
		fallback = value{code: fmt.Sprintf("func() %s {\npanic(%s(%q, %q, %q))\n}()", g.goType(m.sig.ret), g.rt("UnsupportedOperand"), opSymbols[op]+"=", g.typeNameOf(cur.typ), g.typeNameOf(right.typ)), typ: m.sig.ret}
	}
	if fallback.typ != m.sig.ret {
		return value{}, true, unsupported(fmt.Sprintf("operator %s= whose methods return %s and %s", opSymbols[op], m.sig.ret, fallback.typ), line)
	}
	code := fmt.Sprintf("func() %s {\nvar _l %s = %s\nvar _r %s = %s\nif v, ok := _l.M___i%s__(_r); ok {\nreturn v\n}\nreturn %s\n}()",
		g.goType(m.sig.ret), g.goType(cur.typ), cur.code, g.goType(right.typ), right.typed().code, name, fallback.code)
	return value{code: code, typ: m.sig.ret, prec: precAtom}, true, nil
}

// classUnary is -x, +x and ~x on an instance.
func (g *generator) classUnary(op string, v value, line int) (value, error) {
	name := map[string]string{"USub": "neg", "UAdd": "pos", "Invert": "invert"}[op]
	c := g.classOf(v.typ)
	var m *member
	if c != nil && name != "" {
		m = c.lookup("__" + name + "__")
	}
	if m == nil || m.kind != mMethod || len(m.sig.pos) != 0 {
		return value{}, unsupported(fmt.Sprintf("bad operand type for unary %s: '%s'", opSymbols[op], g.typeNameOf(v.typ)), line)
	}
	return value{code: v.at(precAtom) + ".M___" + name + "__()", typ: m.sig.ret, prec: precAtom}, nil
}

// truth is a value's truth test as a Go bool.
func (g *generator) truthOptional(v value) value {
	return value{code: g.rt("OptTruth") + "(" + v.code + ")", typ: tBool, prec: precAtom}
}

func (g *generator) truth(v value) value {
	if v.typ.isOptional() {
		return g.truthOptional(v)
	}
	if v.typ == tNoneType {
		return value{code: "false", typ: tBool, prec: precAtom}
	}
	if c := g.classOf(v.typ); c != nil {
		if m := c.lookup("__bool__"); m != nil && m.kind == mMethod && len(m.sig.pos) == 0 {
			return value{code: v.at(precAtom) + ".M___bool__()", typ: tBool, prec: precAtom}
		}
		if m := c.lookup("__len__"); m != nil && m.kind == mMethod && len(m.sig.pos) == 0 {
			return value{code: v.at(precAtom) + ".M___len__() != 0", typ: tBool, prec: precCmp}
		}
		return value{code: "true", typ: tBool, prec: precAtom}
	}
	if v.typ == tOptExc {
		return value{code: v.at(precCmp+1) + " != nil", typ: tBool, prec: precCmp}
	}
	if v.typ == tObject {
		return value{code: g.rt("DynBool") + "(" + v.code + ")", typ: tBool, prec: precAtom}
	}
	return truth(v)
}

// classBuiltin is len(), abs(), int(), float(), bool() and hash() of an
// instance.
func (g *generator) classBuiltin(name string, n *pyast.Call, a value) (value, bool, error) {
	c := g.classOf(a.typ)
	method := map[string]string{"len": "__len__", "abs": "__abs__", "int": "__int__", "float": "__float__", "bool": "__bool__", "hash": "__hash__"}[name]
	if c == nil || method == "" {
		return value{}, false, nil
	}
	if name == "hash" {
		return value{code: g.rt("HashOf") + "(" + a.code + ")", typ: tInt, prec: precAtom}, true, nil
	}
	if name == "bool" {
		return g.truth(a), true, nil
	}
	m := c.lookup(method)
	if m == nil || m.kind != mMethod || len(m.sig.pos) != 0 {
		return value{}, true, unsupported(fmt.Sprintf("%s() of '%s'", name, c.name), n.Line())
	}
	want := map[string]pyType{"len": tInt, "int": tInt, "float": tFloat}[name]
	if want != "" && m.sig.ret != want {
		return value{}, true, unsupported(fmt.Sprintf("%s.%s returns %s, not %s", c.name, method, m.sig.ret, want), n.Line())
	}
	return value{code: a.at(precAtom) + ".M_" + method + "()", typ: m.sig.ret, prec: precAtom}, true, nil
}

// classContains is "x in obj" for an instance with __contains__.
func (g *generator) classContains(n *pyast.Compare, r value, negate bool) (value, error) {
	c := g.classOf(r.typ)
	m := c.lookup("__contains__")
	if m == nil || m.kind != mMethod || len(m.sig.pos) != 1 {
		// Without __contains__ the items are searched, if the class iterates.
		l, err := g.typedExpr(n.Left)
		if err != nil {
			return value{}, err
		}
		code := g.rt("DynContains") + "(" + g.boxed(r) + ", " + g.boxed(l) + ")"
		if negate {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, nil
	}
	l, err := g.exprAs(n.Left, m.sig.pos[0].typ)
	if err != nil {
		return value{}, err
	}
	l, err = g.coerce(l, m.sig.pos[0].typ, "argument of __contains__", n.Line())
	if err != nil {
		return value{}, err
	}
	code := r.at(precAtom) + ".M___contains__(" + l.code + ")"
	if m.sig.ret != tBool {
		return value{}, unsupported("__contains__ that does not return bool", n.Line())
	}
	if negate {
		return value{code: "!" + code, typ: tBool, prec: precUnary}, nil
	}
	return value{code: code, typ: tBool, prec: precAtom}, nil
}

// ---- subscripts ----

func (g *generator) itemMethod(c *classInfo, name string, params int) *member {
	m := c.lookup(name)
	if m == nil || m.kind != mMethod || len(m.sig.pos) != params || m.sig.vararg != nil || m.sig.kwarg != nil {
		return nil
	}
	return m
}

// classGetItem is obj[key].
func (g *generator) classGetItem(n *pyast.Subscript, base value) (value, error) {
	c := g.classOf(base.typ)
	m := g.itemMethod(c, "__getitem__", 1)
	if m == nil {
		return value{}, unsupported(fmt.Sprintf("'%s' object is not subscriptable", c.name), n.Line())
	}
	if _, isSlice := n.Index.(*pyast.Slice); isSlice {
		if m.sig.pos[0].typ != tObject {
			return value{}, unsupported("slice of an instance whose __getitem__ takes a "+string(m.sig.pos[0].typ), n.Line())
		}
		idx, err := g.dynIndexCode(n.Index)
		if err != nil {
			return value{}, err
		}
		return value{code: base.at(precAtom) + ".M___getitem__(" + idx + ")", typ: m.sig.ret, prec: precAtom}, nil
	}
	idx, err := g.exprAs(n.Index, m.sig.pos[0].typ)
	if err != nil {
		return value{}, err
	}
	idx, err = g.coerce(idx, m.sig.pos[0].typ, "argument of __getitem__", n.Line())
	if err != nil {
		return value{}, err
	}
	return value{code: base.at(precAtom) + ".M___getitem__(" + idx.code + ")", typ: m.sig.ret, prec: precAtom}, nil
}

// classSetItem is obj[key] = value; src evaluates the value.
func (g *generator) classSetItem(b *strings.Builder, target *pyast.Subscript, base value, valueNode pyast.Node, src string, srcType pyType, line int) error {
	c := g.classOf(base.typ)
	m := g.itemMethod(c, "__setitem__", 2)
	if m == nil {
		return unsupported(fmt.Sprintf("'%s' object does not support item assignment", c.name), line)
	}
	idx, err := g.exprAs(target.Index, m.sig.pos[0].typ)
	if err != nil {
		return err
	}
	idx, err = g.coerce(idx, m.sig.pos[0].typ, "key of __setitem__", line)
	if err != nil {
		return err
	}
	code := src
	if valueNode != nil {
		v, err := g.exprAs(valueNode, m.sig.pos[1].typ)
		if err != nil {
			return err
		}
		v, err = g.coerce(v, m.sig.pos[1].typ, "value of __setitem__", line)
		if err != nil {
			return err
		}
		code = v.code
	} else if srcType != m.sig.pos[1].typ {
		v, err := g.coerce(value{code: src, typ: srcType, prec: precAtom}, m.sig.pos[1].typ, "value of __setitem__", line)
		if err != nil {
			return err
		}
		code = v.code
	}
	fmt.Fprintf(b, "%s.M___setitem__(%s, %s)\n", base.at(precAtom), idx.code, code)
	return nil
}

// classDelItem is del obj[key].
func (g *generator) classDelItem(b *strings.Builder, sub *pyast.Subscript, base value, line int) error {
	c := g.classOf(base.typ)
	m := g.itemMethod(c, "__delitem__", 1)
	if m == nil {
		return unsupported(fmt.Sprintf("'%s' object doesn't support item deletion", c.name), line)
	}
	idx, err := g.exprAs(sub.Index, m.sig.pos[0].typ)
	if err != nil {
		return err
	}
	idx, err = g.coerce(idx, m.sig.pos[0].typ, "key of __delitem__", line)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, "%s.M___delitem__(%s)\n", base.at(precAtom), idx.code)
	return nil
}

// augClassItem is "obj[key] op= value".
func (g *generator) augClassItem(b *strings.Builder, n *pyast.AugAssign, target *pyast.Subscript, base value) error {
	c := g.classOf(base.typ)
	get := g.itemMethod(c, "__getitem__", 1)
	set := g.itemMethod(c, "__setitem__", 2)
	if get == nil || set == nil {
		return unsupported(fmt.Sprintf("augmented assignment to an item of '%s'", c.name), n.Line())
	}
	obj := base.at(precAtom)
	if !simpleOperand.MatchString(obj) {
		tmp := g.tmp("o")
		g.temp(b, tmp, base.typ, base.code)
		obj = tmp
	}
	idx, err := g.exprAs(target.Index, get.sig.pos[0].typ)
	if err != nil {
		return err
	}
	idx, err = g.coerce(idx, get.sig.pos[0].typ, "key of __getitem__", n.Line())
	if err != nil {
		return err
	}
	key := idx.code
	if !simpleOperand.MatchString(key) {
		tmp := g.tmp("i")
		g.temp(b, tmp, idx.typ, idx.code)
		key = tmp
	}
	right, err := g.exprAs(n.Value, get.sig.ret)
	if err != nil {
		return err
	}
	cur := value{code: obj + ".M___getitem__(" + key + ")", typ: get.sig.ret, prec: precAtom}
	res, err := g.augmented(n.Op, cur, right, n.Line())
	if err != nil {
		return err
	}
	res, err = g.coerce(res, set.sig.pos[1].typ, "value of __setitem__", n.Line())
	if err != nil {
		return err
	}
	fmt.Fprintf(b, "%s.M___setitem__(%s, %s)\n", obj, key, res.code)
	return nil
}

// ---- calls of instances ----

// callInstance is obj(args), which runs __call__.
func (g *generator) callInstance(n *pyast.Call, fv value) (value, error) {
	c := g.classOf(fv.typ)
	m := c.lookup("__call__")
	if m == nil || m.kind != mMethod {
		return value{}, unsupported(fmt.Sprintf("'%s' object is not callable", c.name), n.Line())
	}
	return g.bindCall(n, m.sig, fv.at(precAtom)+".M___call__", m.sig.qualname)
}

// ---- iteration ----

// userIterable adapts an instance with __iter__ and __next__ to a runtime
// Iterable.
func (g *generator) userIterable(v value) (code string, elem pyType, ok bool) {
	c := g.classOf(v.typ)
	if c == nil {
		return "", "", false
	}
	iter := c.lookup("__iter__")
	if iter == nil || iter.kind != mMethod || len(iter.sig.pos) != 0 {
		return "", "", false
	}
	if r := iter.sig.ret; r.isIter() || r.isGen() {
		return v.at(precAtom) + ".M___iter__()", r.elem(), true
	}
	if iter.sig.ret == tObject {
		return g.rt("DynIter") + "(" + v.code + ")", tObject, true
	}
	ic := g.classOf(iter.sig.ret)
	if ic == nil {
		return "", "", false
	}
	next := ic.lookup("__next__")
	if next == nil || next.kind != mMethod || len(next.sig.pos) != 0 {
		return "", "", false
	}
	elem = next.sig.ret
	tmp := v.code
	code = fmt.Sprintf("%s[%s]{Start: func() %s[%s] {\nreturn %s.M___iter__()\n}}",
		g.rt("UserIterable"), g.goType(elem), g.rt("Nexter"), g.goType(elem), tmp)
	return code, elem, true
}

// ---- isinstance, issubclass, type ----

// typeSpec is one class named in the second argument of isinstance.
type typeSpec struct {
	user    *classInfo // a class of the program
	exc     string     // a built-in exception class
	builtin string     // int, str, float, bool, list, ...
	object  bool
}

func (g *generator) typeSpecs(n pyast.Node) ([]typeSpec, error) {
	nodes := []pyast.Node{n}
	if t, ok := n.(*pyast.Tuple); ok {
		nodes = t.Elts
	}
	var specs []typeSpec
	for _, node := range nodes {
		name, ok := node.(*pyast.Name)
		if !ok || g.isVariable(name.ID) {
			return nil, unsupported("second argument of isinstance() that is not a class name", node.Line())
		}
		switch {
		case g.classes[name.ID] != nil:
			specs = append(specs, typeSpec{user: g.classes[name.ID]})
		case runtime.IsExceptionClass(name.ID):
			specs = append(specs, typeSpec{exc: name.ID})
		case name.ID == "object":
			specs = append(specs, typeSpec{object: true})
		case name.ID == "int" || name.ID == "str" || name.ID == "float" || name.ID == "bool" ||
			name.ID == "list" || name.ID == "dict" || name.ID == "set" || name.ID == "tuple" || name.ID == "range" ||
			name.ID == "complex" || name.ID == "bytes" || name.ID == "bytearray" || name.ID == "frozenset" || name.ID == "slice":
			specs = append(specs, typeSpec{builtin: name.ID})
		default:
			return nil, unsupported(fmt.Sprintf("isinstance() with '%s'", name.ID), node.Line())
		}
	}
	return specs, nil
}

// instanceTest is the Go expression for isinstance(x, spec), with x of the
// given static type, or a constant "true"/"false".
func (g *generator) instanceTest(x value, s typeSpec, line int) (string, error) {
	switch {
	case s.object:
		return "true", nil
	case s.user != nil:
		switch {
		case x.typ.isClass():
			if g.classOf(x.typ).isSubclassOf(s.user) {
				return "true", nil
			}
			return g.rt("InstanceOf") + "[" + s.user.name + "](" + x.code + ")", nil
		case x.typ == tObject:
			return g.rt("InstanceOf") + "[" + s.user.name + "](" + x.code + ")", nil
		case x.typ == tExc:
			if s.user.exc {
				return g.rt("IsSubclass") + "(" + x.at(precAtom) + ".Type, " + strconv.Quote(s.user.name) + ")", nil
			}
		}
		return "false", nil
	case s.exc != "":
		switch {
		case x.typ == tExc:
			return g.rt("IsSubclass") + "(" + x.at(precAtom) + ".Type, " + strconv.Quote(s.exc) + ")", nil
		case x.typ.isClass() && g.classOf(x.typ).exc:
			return g.rt("IsSubclass") + "(" + x.at(precAtom) + ".PyExc().Type, " + strconv.Quote(s.exc) + ")", nil
		case x.typ == tObject:
			return g.rt("IsExcInstance") + "(" + x.code + ", " + strconv.Quote(s.exc) + ")", nil
		}
		return "false", nil
	}
	// A built-in type.
	switch x.typ {
	case tObject:
		switch s.builtin {
		case "bool":
			return g.rt("InstanceOf") + "[bool](" + x.code + ")", nil
		case "str":
			return g.rt("InstanceOf") + "[string](" + x.code + ")", nil
		case "float":
			return g.rt("InstanceOf") + "[float64](" + x.code + ")", nil
		}
		return g.rt("IsInstanceName") + "(" + x.code + ", " + strconv.Quote(s.builtin) + ", true)", nil
	case tInt:
		return strconv.FormatBool(s.builtin == "int"), nil
	case tBool:
		return strconv.FormatBool(s.builtin == "bool" || s.builtin == "int"), nil
	case tStr, tFloat, tRange:
		return strconv.FormatBool(pyType(s.builtin) == x.typ), nil
	}
	return strconv.FormatBool(pythonTypeName(x.typ) == s.builtin), nil
}

// isinstanceCall is isinstance(x, C) and isinstance(x, (A, B)).
func (g *generator) isinstanceCall(n *pyast.Call) (value, error) {
	if n.HasKeywords || len(n.Args) != 2 {
		return value{}, unsupported(fmt.Sprintf("isinstance() with %d arguments", len(n.Args)), n.Line())
	}
	x, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	specs, err := g.typeSpecs(n.Args[1])
	if err != nil && strings.Contains(err.Error(), "that is not a class name") {
		// The classes are held in a variable: decided when the program runs.
		b, cerr := g.classesObject(n.Args[1])
		if cerr != nil {
			return value{}, cerr
		}
		return g.dynValue(g.rt("DynIsInstance")+"("+g.boxed(x)+", "+b+")", tBool), nil
	}
	if err != nil {
		return value{}, err
	}
	var parts []string
	for _, s := range specs {
		t, err := g.instanceTest(x, s, n.Line())
		if err != nil {
			return value{}, err
		}
		parts = append(parts, t)
	}
	code := strings.Join(parts, " || ")
	if len(parts) > 1 {
		code = "(" + code + ")"
	}
	if code == "true" || code == "false" {
		if simpleOperand.MatchString(x.code) || x.konst {
			return value{code: code, typ: tBool, prec: precAtom}, nil
		}
		return value{code: "func() bool {\n_ = " + x.code + "\nreturn " + code + "\n}()", typ: tBool, prec: precAtom}, nil
	}
	return value{code: code, typ: tBool, prec: precOr}, nil
}

// issubclassCall is issubclass(C, D) for two class names.
func (g *generator) issubclassCall(n *pyast.Call) (value, error) {
	if n.HasKeywords || len(n.Args) != 2 {
		return value{}, unsupported(fmt.Sprintf("issubclass() with %d arguments", len(n.Args)), n.Line())
	}
	first, ok := n.Args[0].(*pyast.Name)
	var sub *classInfo
	switch {
	case ok && g.classes[first.ID] != nil && !g.isVariable(first.ID):
		sub = g.classes[first.ID]
	case ok && runtime.IsExceptionClass(first.ID) && !g.isVariable(first.ID):
		sub = g.builtinClass(first.ID)
	default:
		// A class held in a variable: decided when the program runs.
		a, err := g.typedExpr(n.Args[0])
		if err != nil {
			return value{}, err
		}
		b, err := g.classesObject(n.Args[1])
		if err != nil {
			return value{}, err
		}
		return g.dynValue(g.rt("DynIsSubclass")+"("+g.boxed(a)+", "+b+")", tBool), nil
	}
	specs, err := g.typeSpecs(n.Args[1])
	if err != nil {
		return value{}, err
	}
	for _, s := range specs {
		switch {
		case s.object:
			return value{code: "true", typ: tBool, prec: precAtom}, nil
		case s.user != nil && sub.isSubclassOf(s.user):
			return value{code: "true", typ: tBool, prec: precAtom}, nil
		case s.exc != "" && sub.isSubclassOf(g.builtinClass(s.exc)):
			return value{code: "true", typ: tBool, prec: precAtom}, nil
		}
	}
	return value{code: "false", typ: tBool, prec: precAtom}, nil
}

// typeCall is type(x), which is only usable for its name and in
// comparisons. self gives an exactly known class.
func (g *generator) typeCall(n *pyast.Call) (value, error) {
	if n.HasKeywords || len(n.Args) != 1 {
		return value{}, unsupported(fmt.Sprintf("type() with %d arguments", len(n.Args)), n.Line())
	}
	x, err := g.typedExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	c := g.classOf(x.typ)
	if c == nil {
		return value{}, unsupported(fmt.Sprintf("type() of %s", x.typ), n.Line())
	}
	exact := false
	if name, ok := n.Args[0].(*pyast.Name); ok && g.method != nil && g.method.kind == mMethod && name.ID == g.selfName() && g.method.cls == c {
		exact = true
	}
	return value{code: x.code, typ: classRefOf(c.name), prec: x.prec, classRef: c, dynamicRef: !exact, exactRef: exact}, nil
}

// typeOperand reads one side of a comparison of classes: type(x), X.__class__,
// or the name of a class. It returns a Go expression for the class name.
func (g *generator) typeOperand(n pyast.Node) (code string, ok bool, err error) {
	switch n := n.(type) {
	case *pyast.Call:
		if name, _ := callName(n); name == "type" && !g.isVariable("type") {
			if len(n.Args) == 1 && !n.HasKeywords {
				arg, err := g.typedExpr(n.Args[0])
				if err != nil {
					return "", true, err
				}
				switch {
				case arg.typ == tObject:
					return g.rt("TypeName") + "(" + arg.code + ")", true, nil
				case arg.typ == tExc:
					return g.rt("BaseTypeName") + "(" + arg.at(precAtom) + ".Type)", true, nil
				case !arg.typ.isClass():
					return strconv.Quote(pythonTypeName(arg.typ)), true, nil
				}
			}
			x, err := g.typeCall(n)
			if err != nil {
				return "", true, err
			}
			if x.dynamicRef {
				return x.at(precAtom) + ".PyClassName()", true, nil
			}
			return strconv.Quote(x.classRef.pyName()), true, nil
		}
	case *pyast.Attribute:
		if n.Attr == "__class__" {
			x, err := g.typedExpr(n.Value)
			if err != nil {
				return "", true, err
			}
			if x.typ.isClass() {
				return x.at(precAtom) + ".PyClassName()", true, nil
			}
		}
	case *pyast.Name:
		if g.isVariable(n.ID) {
			return "", false, nil
		}
		if c := g.classes[n.ID]; c != nil {
			return strconv.Quote(c.pyName()), true, nil
		}
		switch n.ID {
		case "int", "str", "float", "bool":
			return strconv.Quote(n.ID), true, nil
		}
	}
	return "", false, nil
}

// typeCompare handles comparisons where a side is type(x), x.__class__ or a
// class name.
func (g *generator) typeCompare(n *pyast.Compare) (value, bool, error) {
	if len(n.Ops) != 1 {
		return value{}, false, nil
	}
	op := n.Ops[0]
	if op != "Is" && op != "IsNot" && op != "Eq" && op != "NotEq" {
		return value{}, false, nil
	}
	isTypeCall := func(x pyast.Node) bool {
		switch x := x.(type) {
		case *pyast.Call:
			name, _ := callName(x)
			return name == "type" && !g.isVariable("type")
		case *pyast.Attribute:
			return x.Attr == "__class__"
		}
		return false
	}
	if !isTypeCall(n.Left) && !isTypeCall(n.Comparators[0]) {
		return value{}, false, nil
	}
	l, lok, err := g.typeOperand(n.Left)
	if err != nil {
		return value{}, true, err
	}
	r, rok, err := g.typeOperand(n.Comparators[0])
	if err != nil {
		return value{}, true, err
	}
	if !lok || !rok {
		return value{}, true, unsupported("comparison of a class with something that is not a class", n.Line())
	}
	code := l + " == " + r
	if op == "IsNot" || op == "NotEq" {
		code = l + " != " + r
	}
	return value{code: code, typ: tBool, prec: precCmp}, true, nil
}

// objectCompare is ==, !=, <, <=, >, >= and is/is not where an operand is an
// instance of a class, or an object.
func (g *generator) objectCompare(op string, l, r value, line int) (value, bool, error) {
	involved := func(v value) bool { return v.typ.isClass() || v.typ == tObject }
	if op == "Is" || op == "IsNot" {
		// Identity is only meaningful for values that are objects in Python
		// and pointers here.
		identity := func(v value) bool {
			return involved(v) || v.typ.isList() || v.typ.isDict() || v.typ.isSet() || v.typ == tExc || v.typ == tOptExc
		}
		if (!identity(l) || !identity(r)) && g.boxable(l) && g.boxable(r) && l.typ != tNone && r.typ != tNone {
			// Numbers, strings and tuples: the same object when they are
			// equal, as CPython shares them in the common cases.
			code := g.rt("Is") + "(" + g.boxed(l) + ", " + g.boxed(r) + ")"
			if op == "IsNot" {
				return value{code: "!" + code, typ: tBool, prec: precUnary}, true, nil
			}
			return value{code: code, typ: tBool, prec: precAtom}, true, nil
		}
		if !identity(l) || !identity(r) {
			return value{}, true, unsupported(fmt.Sprintf("'%s' between %s and %s", opSymbols[op], l.typ, r.typ), line)
		}
		if l.typ != r.typ && !involved(l) && !involved(r) && !(l.typ == tExc || l.typ == tOptExc) {
			return value{}, true, unsupported(fmt.Sprintf("'%s' between %s and %s", opSymbols[op], l.typ, r.typ), line)
		}
		code := g.rt("Is") + "(" + g.identityBox(l) + ", " + g.identityBox(r) + ")"
		if op == "IsNot" {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, true, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, true, nil
	}
	if !involved(l) && !involved(r) {
		return value{}, false, nil
	}
	box := func(v value) string { return g.relabel(v, tObject).code }
	switch op {
	case "Is", "IsNot":
		if !(involved(l) && involved(r)) {
			return value{}, true, unsupported(fmt.Sprintf("'%s' between %s and %s", opSymbols[op], l.typ, r.typ), line)
		}
		code := g.rt("Is") + "(" + box(l) + ", " + box(r) + ")"
		if op == "IsNot" {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, true, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, true, nil
	case "Eq", "NotEq":
		code := g.rt("Eq") + "(" + box(l) + ", " + box(r) + ")"
		if op == "NotEq" {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, true, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, true, nil
	case "Lt", "LtE", "Gt", "GtE":
		fn := map[string]string{"Lt": "Lt", "LtE": "Le", "Gt": "Gt", "GtE": "Ge"}[op]
		if l.typ == tObject || r.typ == tObject {
			return g.dynCompare(op, l, r), true, nil
		}
		return value{code: g.rt(fn) + "(" + box(l) + ", " + box(r) + ")", typ: tBool, prec: precAtom}, true, nil
	}
	return value{}, false, nil
}

// zeroValue is the Go zero value of a type, written out.
func zeroValue(g *generator, t pyType) string {
	switch {
	case t == tInt:
		return "0"
	case t == tFloat:
		return "0"
	case t == tBool:
		return "false"
	case t == tStr:
		return `""`
	}
	return "*new(" + g.goType(t) + ")"
}

// callableCall is callable(x).
func (g *generator) callableCall(n *pyast.Call) (value, error) {
	if n.HasKeywords || len(n.Args) != 1 {
		return value{}, unsupported(fmt.Sprintf("callable() with %d arguments", len(n.Args)), n.Line())
	}
	x, err := g.recvExpr(n.Args[0])
	if err != nil {
		return value{}, err
	}
	if x.typ == tObject {
		return value{code: g.rt("DynCallable") + "(" + x.code + ")", typ: tBool, prec: precAtom}, nil
	}
	callable := x.typ.isFunc() || x.typ.isClassRef()
	if c := g.classOf(x.typ); c != nil {
		callable = c.lookup("__call__") != nil
	}
	return value{code: strconv.FormatBool(callable), typ: tBool, prec: precAtom}, nil
}

// identityBox passes a value to runtime.Is.
func (g *generator) identityBox(v value) string {
	if v.typ.isClass() || v.typ == tObject || v.typ.isList() || v.typ.isDict() || v.typ.isSet() || v.typ == tExc || v.typ == tOptExc {
		return v.code
	}
	return g.relabel(v, tObject).code
}

// noneCompare is "x is None" and "x is not None" for the one value that can be
// None before phase 8: e.__cause__ and e.__context__.
func (g *generator) noneCompare(n *pyast.Compare) (value, bool, error) {
	if len(n.Ops) != 1 {
		return value{}, false, nil
	}
	op := n.Ops[0]
	if op != "Is" && op != "IsNot" && op != "Eq" && op != "NotEq" {
		return value{}, false, nil
	}
	isNone := func(x pyast.Node) bool {
		c, ok := x.(*pyast.Constant)
		return ok && c.PyType == "none"
	}
	var other pyast.Node
	switch {
	case isNone(n.Comparators[0]):
		other = n.Left
	case isNone(n.Left):
		other = n.Comparators[0]
	default:
		return value{}, false, nil
	}
	v, err := g.typedExpr(other)
	if err != nil {
		return value{}, true, err
	}
	negate := op == "IsNot" || op == "NotEq"
	switch {
	case v.typ.isOptional():
		code := "!" + v.at(precAtom) + ".IsSome()"
		if negate {
			code = v.at(precAtom) + ".IsSome()"
		}
		return value{code: code, typ: tBool, prec: precUnary}, true, nil
	case v.typ == tObject:
		code := g.rt("IsNone") + "(" + v.code + ")"
		if negate {
			return value{code: "!" + code, typ: tBool, prec: precUnary}, true, nil
		}
		return value{code: code, typ: tBool, prec: precAtom}, true, nil
	case v.typ == tOptExc:
		code := v.at(precCmp+1) + " == nil"
		if negate {
			code = v.at(precCmp+1) + " != nil"
		}
		return value{code: code, typ: tBool, prec: precCmp}, true, nil
	}
	// Nothing else is None: the answer is known, and the operand still runs.
	answer := v.typ == tNoneType
	if negate {
		answer = !answer
	}
	code := strconv.FormatBool(answer)
	if !simpleOperand.MatchString(v.code) && !v.konst {
		code = "func() bool {\n_ = " + v.code + "\nreturn " + code + "\n}()"
	}
	return value{code: code, typ: tBool, prec: precAtom}, true, nil
}

// hashable reports whether a value of type t can be a dict key or set
// element: the scalars, tuples of hashable values, and instances of classes
// that hash by identity. A class with its own __eq__ or __hash__ would need
// the dict to call it, which the runtime's maps do not (phase 9).
func (g *generator) hashable(t pyType) bool {
	if t == tObject || t.hashable() {
		return true
	}
	if t.isTuple() {
		for _, e := range t.elems() {
			if !g.hashable(e) {
				return false
			}
		}
		return true
	}
	if c := g.classOf(t); c != nil {
		mode, _ := c.hashMode()
		return mode == "identity" || mode == "user"
	}
	return false
}

// userHashed reports whether t is a class with its own __hash__: its
// instances are dict keys and set elements as objects, which the runtime
// compares with __hash__ and __eq__.
func (g *generator) userHashed(t pyType) bool {
	c := g.classOf(t)
	if c == nil {
		return false
	}
	mode, _ := c.hashMode()
	return mode == "user"
}

// keyType is the type a dict or set holds for keys of type t.
func (g *generator) keyType(t pyType) pyType {
	if g.userHashed(t) {
		return tObject
	}
	return t
}

// sumInstances is sum() over instances: start + x0 + x1 + ..., with the
// operator methods. Without a start the first step is 0 + x0, which needs a
// __radd__ taking an int, as in Python.
func (g *generator) sumInstances(n *pyast.Call, list value, start *value) (value, error) {
	elem := list.typ.elem()
	dyn := func() (value, error) {
		startCode := "int64(0)"
		if start != nil {
			startCode = g.boxed(*start)
		}
		return g.dynValue(g.rt("DynSum")+"("+g.boxed(list)+", "+startCode+")", tObject), nil
	}
	var b strings.Builder
	var acc value
	if start != nil {
		acc = *start
		fmt.Fprintf(&b, "var _items %s = %s\nvar _acc %s = %s\n_ = _items\n", g.goType(list.typ), list.code, g.goType(acc.typ), acc.typed().code)
		acc = value{code: "_acc", typ: acc.typ, prec: precAtom}
	} else {
		first, err := g.classBinary("Add", value{code: "0", typ: tInt, prec: precAtom, konst: true}, value{code: "_items.Get(0)", typ: elem, prec: precAtom}, n.Line())
		if err != nil || first.typ == tObject {
			return dyn()
		}
		fmt.Fprintf(&b, "var _items %s = %s\nif _items.Len() == 0 {\npanic(%s(\"TypeError\", \"sum() of no instances needs a start value here\"))\n}\nvar _acc %s = %s\n",
			g.goType(list.typ), list.code, g.rt("NewError"), g.goType(first.typ), first.code)
		acc = value{code: "_acc", typ: first.typ, prec: precAtom}
	}
	step, err := g.classBinary("Add", acc, value{code: "_x", typ: elem, prec: precAtom}, n.Line())
	if err != nil || step.typ == tObject {
		return dyn()
	}
	if !g.isSubtype(step.typ, acc.typ) {
		return value{}, unsupported(fmt.Sprintf("sum() whose steps change the type from %s to %s", acc.typ, step.typ), n.Line())
	}
	skip := ""
	if start == nil {
		skip = "_skip := true\n"
	}
	fmt.Fprintf(&b, "%sfor _c := _items.Iter(); _c.Next(); {\nvar _x %s = _c.Value()\n", skip, g.goType(elem))
	if start == nil {
		b.WriteString("if _skip {\n_skip = false\ncontinue\n}\n")
	}
	fmt.Fprintf(&b, "_acc = %s\n}\nreturn _acc\n", step.code)
	return value{code: fmt.Sprintf("func() %s {\n%s}()", g.goType(acc.typ), b.String()), typ: acc.typ, prec: precAtom}, nil
}
