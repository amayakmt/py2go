package codegen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/amayakmt/py2go/internal/pyast"
)

// A class that defines __setattr__ gets every "obj.x = v" of its instances
// sent to it; object.__setattr__(self, "x", v) and super().__setattr__("x", v)
// inside it store the attribute itself (the raw store, which declares the
// field like a plain assignment does).

// setattrHook is the user-defined __setattr__ of the class of the receiver, if
// the receiver is an instance of a class that has one.
func (g *generator) setattrHook(recvNode pyast.Node, hookName string) bool {
	if g.rawStore {
		return false
	}
	recv, err := g.typedExpr(recvNode)
	if err != nil || !recv.typ.isClass() {
		return false
	}
	c := g.classOf(recv.typ)
	if c == nil {
		return false
	}
	m := c.lookup(hookName)
	return m != nil && m.kind == mMethod && m.sig != nil
}

// hookedAssign is "obj.x = v" for an obj whose class defines __setattr__.
func (g *generator) hookedAssign(b *strings.Builder, target *pyast.Attribute, valueNode pyast.Node, line int) error {
	call := &pyast.Call{
		Func:   &pyast.Attribute{Value: target.Value, Attr: "__setattr__", Lineno: line},
		Args:   []pyast.Node{&pyast.Constant{PyType: "str", Value: target.Attr, Lineno: line}, valueNode},
		Lineno: line,
	}
	v, err := g.typedExpr(call)
	if err != nil {
		return err
	}
	if v.typ == tNone {
		b.WriteString(v.code + "\n")
	} else {
		b.WriteString("_ = " + v.code + "\n")
	}
	return nil
}

// rawSetattrCall recognizes object.__setattr__(obj, name, value) and
// object.__delattr__(obj, name) as statements.
func rawSetattrCall(n *pyast.Call) (recv pyast.Node, name pyast.Node, value pyast.Node, del bool, ok bool) {
	a, isAttr := n.Func.(*pyast.Attribute)
	if !isAttr || n.HasKeywords {
		return nil, nil, nil, false, false
	}
	base, isName := a.Value.(*pyast.Name)
	if !isName || base.ID != "object" {
		return nil, nil, nil, false, false
	}
	switch {
	case a.Attr == "__setattr__" && len(n.Args) == 3:
		return n.Args[0], n.Args[1], n.Args[2], false, true
	case a.Attr == "__delattr__" && len(n.Args) == 2:
		return n.Args[0], n.Args[1], nil, true, true
	}
	return nil, nil, nil, false, false
}

// rawStoreStmt compiles a raw attribute store: with a constant name it is a
// plain assignment that skips the hook; otherwise a dynamic one.
func (g *generator) rawStoreStmt(b *strings.Builder, recv, name, val pyast.Node, del bool, line int) error {
	if c, ok := name.(*pyast.Constant); ok && c.PyType == "str" {
		target := &pyast.Attribute{Value: recv, Attr: c.Value.(string), Lineno: line}
		saved := g.rawStore
		g.rawStore = true
		defer func() { g.rawStore = saved }()
		if del {
			return g.deleteStmt(b, &pyast.Delete{Targets: []pyast.Node{target}, Lineno: line})
		}
		return g.assignAttr(b, target, val, line)
	}
	r, err := g.typedExpr(recv)
	if err != nil {
		return err
	}
	nm, err := g.exprAs(name, tStr)
	if err != nil {
		return err
	}
	if nm.typ != tStr && nm.typ != tObject {
		return unsupported(fmt.Sprintf("attribute name of type %s", nm.typ), line)
	}
	nameCode := nm.code
	if nm.typ == tObject {
		nameCode = g.rt("Cast") + "[string](" + nm.code + ")"
	}
	if del {
		fmt.Fprintf(b, "%s(%s, %s)\n", g.rt("RawDelAttr"), g.boxed(r), nameCode)
		return nil
	}
	v, err := g.typedExpr(val)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, "%s(%s, %s, %s)\n", g.rt("RawSetAttr"), g.boxed(r), nameCode, g.boxed(v))
	return nil
}

// emitSetattrHook writes the run-time entry of a class's __setattr__ and
// __delattr__, for stores the compiler does not see (setattr(), DynSetAttr).
func (g *generator) emitSetattrHook(out *strings.Builder, c *classInfo) {
	recv := "(self *" + c.goStruct() + ")"
	for _, hook := range []string{"__setattr__", "__delattr__"} {
		m := c.lookup(hook)
		if m == nil || m.kind != mMethod || m.sig == nil {
			continue
		}
		params := m.sig.goParams()
		goName := "PySetAttrHook"
		args := []string{}
		if hook == "__delattr__" {
			goName = "PyDelAttrHook"
		}
		if len(params) >= 1 {
			args = append(args, g.rt("Cast")+"["+g.goType(params[0].typ)+"](name)")
		}
		if hook == "__setattr__" && len(params) >= 2 {
			args = append(args, g.rt("Cast")+"["+g.goType(params[1].typ)+"](v)")
		}
		sig := "name string, v any"
		if hook == "__delattr__" {
			sig = "name string"
		}
		fmt.Fprintf(out, "func %s %s(%s) {\n_ = self\nself.%s(%s)\n}\n", recv, goName, sig, methodGoName(m, c), strings.Join(args, ", "))
	}
}

var _ = strconv.Quote

// rawStoreExpr is a raw store used as an expression (super().__setattr__(...)):
// the statements go before it, and it is None.
func (g *generator) rawStoreExpr(recv, name, val pyast.Node, del bool, line int) (value, error) {
	var b strings.Builder
	if err := g.rawStoreStmt(&b, recv, name, val, del, line); err != nil {
		return value{}, err
	}
	return value{code: "func() " + g.rt("NoneT") + " {\n" + b.String() + "return " + g.rt("NoneV") + "\n}()", typ: tNone, prec: precAtom}, nil
}

// abstractMissing lists, sorted, the abstract methods of an abstract base
// class that c does not implement: instantiating c is then a TypeError, as
// ABCMeta makes it. Only classes under abc.ABC or metaclass=ABCMeta check.
func (g *generator) abstractMissing(c *classInfo) []string {
	isABC := false
	for _, k := range c.mro {
		if k.pyName() == "ABC" && !k.builtin || k.node != nil && k.node.ABCMeta {
			isABC = true
		}
	}
	if !isABC {
		return nil
	}
	seen := map[string]bool{}
	var missing []string
	for _, k := range c.mro {
		if k.node == nil {
			continue
		}
		for _, name := range k.node.Abstract {
			if seen[name] {
				continue
			}
			seen[name] = true
			m := c.lookup(name)
			if m == nil || m.owner == nil || m.owner.node == nil {
				continue
			}
			for _, a := range m.owner.node.Abstract {
				if a == name {
					missing = append(missing, name)
					break
				}
			}
		}
	}
	sort.Strings(missing)
	return missing
}
