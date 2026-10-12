package codegen

import (
	"fmt"

	"github.com/amayakmt/py2go/internal/pyast"
)

// @dataclass is handled natively until the real dataclasses module is
// compiled from Lib/ (phase 12). The generated methods are members like any
// other, marked with synth so that the emitter writes their Go directly.

// dcInfo holds the options of a dataclass and its fields.
type dcInfo struct {
	frozen, order, eq, repr, init bool
	own                           []*dcField // declared in this class body
	fields                        []*dcField // all fields, in __init__ order
}

// dcField is one field of a dataclass.
type dcField struct {
	name    string
	typ     pyType
	def     pyast.Node // default value
	factory pyast.Node // default_factory
	init    bool
	repr    bool
	compare bool
	line    int
}

func (f *dcField) hasDefault() bool { return f.def != nil || f.factory != nil }

// dataclassOptions reads the decorators of a class. A class without
// decorators gets nil.
func (g *generator) dataclassOptions(c *classInfo) (*dcInfo, error) {
	var dc *dcInfo
	for _, d := range c.node.Decorators {
		var call *pyast.Call
		node := d
		if cl, ok := d.(*pyast.Call); ok {
			call, node = cl, cl.Func
		}
		name, ok := node.(*pyast.Name)
		if !ok || name.ID != "dataclass" || dc != nil {
			return nil, unsupported("class decorator (class decorators other than @dataclass arrive in phase 8)", d.Line())
		}
		if !g.dataclassImported {
			return nil, unsupported("@dataclass without 'from dataclasses import dataclass'", d.Line())
		}
		dc = &dcInfo{eq: true, repr: true, init: true}
		if call == nil {
			continue
		}
		if len(call.Args) > 0 {
			return nil, unsupported("positional argument to @dataclass", call.Line())
		}
		for _, kw := range call.Keywords {
			k, ok := kw.Value.(*pyast.Constant)
			if kw.Arg == "" || !ok || k.PyType != "bool" {
				return nil, unsupported("@dataclass option that is not a True/False literal", kw.Line())
			}
			on := k.Value.(bool)
			switch kw.Arg {
			case "frozen":
				dc.frozen = on
			case "order":
				dc.order = on
			case "eq":
				dc.eq = on
			case "repr":
				dc.repr = on
			case "init":
				dc.init = on
			case "slots", "match_args", "weakref_slot":
			default:
				return nil, unsupported("@dataclass option '"+kw.Arg+"'", kw.Line())
			}
		}
		if dc.order && !dc.eq {
			return nil, unsupported("@dataclass(order=True) needs eq=True (ValueError in Python)", call.Line())
		}
	}
	return dc, nil
}

// dataclassField reads "x: T", "x: T = default" and "x: T = field(...)".
func (g *generator) dataclassField(c *classInfo, name string, t pyType, n *pyast.AnnAssign) (*dcField, error) {
	f := &dcField{name: name, typ: t, init: true, repr: true, compare: true, line: n.Line()}
	if n.Value == nil {
		return f, nil
	}
	call, isCall := n.Value.(*pyast.Call)
	if !isCall {
		f.def = n.Value
		return f, g.checkImmutableDefault(f)
	}
	fn, ok := call.Func.(*pyast.Name)
	if !ok || fn.ID != "field" || g.isVariable("field") {
		f.def = n.Value
		return f, nil
	}
	if !g.dataclassImported {
		return nil, unsupported("field() without 'from dataclasses import field'", call.Line())
	}
	if len(call.Args) > 0 {
		return nil, unsupported("positional argument to field()", call.Line())
	}
	for _, kw := range call.Keywords {
		bool1 := func() (bool, error) {
			k, ok := kw.Value.(*pyast.Constant)
			if !ok || k.PyType != "bool" {
				return false, unsupported("field("+kw.Arg+"=...) that is not a True/False literal", kw.Line())
			}
			return k.Value.(bool), nil
		}
		var err error
		switch kw.Arg {
		case "default":
			f.def = kw.Value
		case "default_factory":
			f.factory = kw.Value
		case "init":
			f.init, err = bool1()
		case "repr":
			f.repr, err = bool1()
		case "compare":
			f.compare, err = bool1()
		case "hash", "metadata":
		default:
			return nil, unsupported("field() option '"+kw.Arg+"'", kw.Line())
		}
		if err != nil {
			return nil, err
		}
	}
	if f.def != nil && f.factory != nil {
		return nil, unsupported("field() with both default and default_factory (ValueError in Python)", call.Line())
	}
	return f, g.checkImmutableDefault(f)
}

// checkImmutableDefault rejects the defaults dataclasses refuse.
func (g *generator) checkImmutableDefault(f *dcField) error {
	switch f.def.(type) {
	case *pyast.List, *pyast.Dict, *pyast.Set:
		return unsupported(fmt.Sprintf("mutable default for field %s is not allowed: use default_factory", f.name), f.line)
	}
	return nil
}

// dataclassMembers collects the fields of a dataclass, base classes first,
// and adds the methods the decorator generates.
func (g *generator) dataclassMembers(c *classInfo) error {
	dc := c.dc
	var fields []*dcField
	index := map[string]int{}
	add := func(f *dcField) {
		if i, ok := index[f.name]; ok {
			fields[i] = f
			return
		}
		index[f.name] = len(fields)
		fields = append(fields, f)
	}
	for i := len(c.mro) - 1; i >= 1; i-- {
		if k := c.mro[i]; k.dc != nil {
			for _, f := range k.dc.own {
				add(f)
			}
		}
	}
	for _, f := range dc.own {
		add(f)
	}
	dc.fields = fields
	if c.matchArgs == nil {
		c.matchArgs = []string{}
		for _, f := range fields {
			if f.init {
				c.matchArgs = append(c.matchArgs, f.name)
			}
		}
	}

	seenDefault := false
	for _, f := range fields {
		if !f.init {
			continue
		}
		if f.hasDefault() {
			seenDefault = true
		} else if seenDefault {
			return unsupported(fmt.Sprintf("non-default argument '%s' follows default argument", f.name), f.line)
		}
	}

	own := func(name string) bool { m := c.members[name]; return m != nil && m.synth == "" }
	synth := func(name, kind string, params []param, ret pyType, special bool) {
		sig := &funcSig{name: name, qualname: c.name + "." + name, goName: name, line: c.line, pos: params, ret: ret, implicit: 1, selfName: "self"}
		sig.valueType = sig.goFuncType()
		c.addMember(&member{kind: mMethod, name: name, owner: c, sig: sig, special: special, synth: kind, line: c.line})
	}
	other := []param{{name: "other", typ: tObject}}

	if dc.init && !own("__init__") {
		var params []param
		for _, f := range fields {
			if !f.init {
				continue
			}
			p := param{name: f.name, typ: f.typ, hasDefault: f.hasDefault(), defNode: f.def, factory: f.factory, line: f.line}
			params = append(params, p)
		}
		sig := &funcSig{name: "__init__", qualname: c.name + ".__init__", goName: "__init__", line: c.line, pos: params, ret: tNone, implicit: 1, selfName: "self"}
		ps, _ := sig.defaultNodes()
		for i, p := range ps {
			p.defaultVar = fmt.Sprintf("_pyd_%s___init___%d", c.name, i)
		}
		sig.valueType = sig.goFuncType()
		c.addMember(&member{kind: mMethod, name: "__init__", owner: c, sig: sig, synth: "dcinit", line: c.line})
	}
	if dc.repr && !own("__repr__") {
		synth("__repr__", "dcrepr", nil, tStr, false)
	}
	if dc.eq && !own("__eq__") {
		synth("__eq__", "dceq", other, tBool, true)
	}
	if dc.order {
		for _, op := range []string{"lt", "le", "gt", "ge"} {
			if own("__" + op + "__") {
				return unsupported(fmt.Sprintf("@dataclass(order=True) on a class that defines __%s__ (TypeError in Python)", op), c.line)
			}
			synth("__"+op+"__", "dc"+op, other, tBool, true)
		}
	}
	if dc.eq && !own("__hash__") && c.members["__hash__"] == nil {
		if dc.frozen {
			synth("__hash__", "dchash", nil, tInt, false)
		} else {
			c.unhashable = true
		}
	}
	return nil
}
