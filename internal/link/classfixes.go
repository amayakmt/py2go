package link

import (
	"strconv"

	"github.com/amayakmt/py2go/internal/pyast"
)

// fixClasses rewrites, before linking, the class statements of a module that
// the compiler cannot take as they are:
//
//   - a class defined inside a class is moved before it, under the name
//     Outer_nested_Inner (Python still sees Inner), and the outer class body
//     gets "Inner = Outer_nested_Inner";
//   - "__class_getitem__ = classmethod(GenericAlias)" is dropped (C[int] is
//     only an annotation for the compiler);
//   - a class decorator other than the ones the compiler knows is called on
//     the class after its statement, and must give the class back.
func fixClasses(mod *pyast.Module) {
	moveDefproperty(mod)
	mod.Body = fixClassList(mod.Body)
}

// moveDefproperty turns the module-level calls defproperty(C, "name", doc=...)
// of xml.dom.minicompat into a read-only property of C that calls
// self._get_name().
func moveDefproperty(mod *pyast.Module) {
	classes := map[string]*pyast.ClassDef{}
	for _, st := range mod.Body {
		if cd, ok := st.(*pyast.ClassDef); ok {
			classes[cd.Name] = cd
		}
	}
	var out []pyast.Node
	for _, st := range mod.Body {
		if e, ok := st.(*pyast.Expr); ok {
			if c, ok := e.Value.(*pyast.Call); ok && decoratorName(c.Func) == "defproperty" && len(c.Args) >= 2 {
				cls, ok1 := c.Args[0].(*pyast.Name)
				name, ok2 := c.Args[1].(*pyast.Constant)
				if ok1 && ok2 && classes[cls.ID] != nil {
					if pname, isStr := name.Value.(string); isStr {
						classes[cls.ID].Body = append(classes[cls.ID].Body, propertyGetter(pname, "_get_"+pname, e.Lineno))
						continue
					}
				}
			}
		}
		out = append(out, st)
	}
	mod.Body = out
}

// propertyGetter is "@property def name(self): return self.getter()".
func propertyGetter(name, getter string, line int) pyast.Node {
	self := func() pyast.Node { return &pyast.Name{ID: "self", Lineno: line} }
	return &pyast.FunctionDef{
		Name:          name,
		Args:          &pyast.Arguments{Args: []*pyast.Arg{{Name: "self", Lineno: line}}},
		HasParams:     true,
		Decorators:    []pyast.Node{&pyast.Name{ID: "property", Lineno: line}},
		HasDecorators: true,
		Body: []pyast.Node{&pyast.Return{Value: &pyast.Call{
			Func: &pyast.Attribute{Value: self(), Attr: getter, Lineno: line}, Lineno: line,
		}, Lineno: line}},
		Lineno: line,
	}
}

// propertySetter is "@name.setter def name(self, value): self.setter(value)".
func propertySetter(name, setter string, line int) pyast.Node {
	return &pyast.FunctionDef{
		Name:          name,
		Args:          &pyast.Arguments{Args: []*pyast.Arg{{Name: "self", Lineno: line}, {Name: "value", Lineno: line}}},
		HasParams:     true,
		Decorators:    []pyast.Node{&pyast.Attribute{Value: &pyast.Name{ID: name, Lineno: line}, Attr: "setter", Lineno: line}},
		HasDecorators: true,
		Body: []pyast.Node{&pyast.Expr{Value: &pyast.Call{
			Func: &pyast.Attribute{Value: &pyast.Name{ID: "self", Lineno: line}, Attr: setter, Lineno: line},
			Args: []pyast.Node{&pyast.Name{ID: "value", Lineno: line}}, Lineno: line,
		}, Lineno: line}},
		Lineno: line,
	}
}

// classPropertyCall reads "name = property(getter[, setter])" with method names.
func classPropertyCall(v pyast.Node) (getter, setter string, ok bool) {
	c, isCall := v.(*pyast.Call)
	if !isCall {
		return "", "", false
	}
	if f, isName := c.Func.(*pyast.Name); !isName || f.ID != "property" || len(c.Args) == 0 || len(c.Args) > 2 {
		return "", "", false
	}
	g, isName := c.Args[0].(*pyast.Name)
	if !isName {
		return "", "", false
	}
	if len(c.Args) == 2 {
		s, isName := c.Args[1].(*pyast.Name)
		if !isName {
			return "", "", false
		}
		return g.ID, s.ID, true
	}
	return g.ID, "", true
}

func fixClassList(body []pyast.Node) []pyast.Node {
	var out []pyast.Node
	for _, st := range body {
		switch x := st.(type) {
		case *pyast.ClassDef:
			out = append(out, fixClass(x, "")...)
			continue
		case *pyast.FunctionDef:
			x.Body = fixClassList(x.Body)
		case *pyast.If:
			x.Body = fixClassList(x.Body)
			x.Orelse = fixClassList(x.Orelse)
		case *pyast.Try:
			x.Body = fixClassList(x.Body)
			for _, h := range x.Handlers {
				h.Body = fixClassList(h.Body)
			}
			x.Orelse = fixClassList(x.Orelse)
			x.Finalbody = fixClassList(x.Finalbody)
		}
		out = append(out, st)
	}
	return out
}

// knownClassDecorator is a decorator the compiler or the linker handles.
func knownClassDecorator(d pyast.Node) bool {
	fn := d
	if c, ok := d.(*pyast.Call); ok {
		fn = c.Func
	}
	name := ""
	switch f := fn.(type) {
	case *pyast.Name:
		name = f.ID
	case *pyast.Attribute:
		name = f.Attr
	}
	switch name {
	case "dataclass", "total_ordering", "skip", "skipIf", "skipUnless", "expectedFailure":
		return true
	}
	return false
}

// decoratorName is the last name of a decorator: "abstractmethod" for both
// @abstractmethod and @abc.abstractmethod.
func decoratorName(d pyast.Node) string {
	switch f := d.(type) {
	case *pyast.Name:
		return f.ID
	case *pyast.Attribute:
		return f.Attr
	}
	return ""
}

// abstractDecorators removes @abstractmethod and turns @abstractproperty,
// @abstractclassmethod and @abstractstaticmethod into the plain decorator,
// and tells whether the method was abstract.
func abstractDecorators(fd *pyast.FunctionDef) bool {
	abstract := false
	var keep []pyast.Node
	for _, d := range fd.Decorators {
		switch decoratorName(d) {
		case "abstractmethod":
			abstract = true
			continue
		case "abstractproperty":
			abstract = true
			keep = append(keep, &pyast.Name{ID: "property", Lineno: fd.Lineno})
			continue
		case "abstractclassmethod":
			abstract = true
			keep = append(keep, &pyast.Name{ID: "classmethod", Lineno: fd.Lineno})
			continue
		case "abstractstaticmethod":
			abstract = true
			keep = append(keep, &pyast.Name{ID: "staticmethod", Lineno: fd.Lineno})
			continue
		}
		keep = append(keep, d)
	}
	fd.Decorators = keep
	fd.HasDecorators = len(keep) > 0
	return abstract
}

func fixClass(x *pyast.ClassDef, prefix string) []pyast.Node {
	var kws []*pyast.Keyword
	for _, k := range x.Keywords {
		if k.Arg == "metaclass" && decoratorName(k.Value) == "ABCMeta" {
			x.ABCMeta = true
			continue
		}
		kws = append(kws, k)
	}
	x.Keywords = kws
	var before []pyast.Node
	var body []pyast.Node
	// "if __debug__:" in a class body: __debug__ is true (no -O), so its
	// body is the class's.
	var flat []pyast.Node
	for _, st := range x.Body {
		if ifs, ok := st.(*pyast.If); ok {
			if n, ok := ifs.Test.(*pyast.Name); ok && n.ID == "__debug__" {
				flat = append(flat, ifs.Body...)
				continue
			}
		}
		flat = append(flat, st)
	}
	x.Body = flat
	for _, st := range x.Body {
		switch y := st.(type) {
		case *pyast.ClassDef:
			inner := y.Name
			hoisted := prefix + x.Name + "_nested_" + inner
			display := y.Display
			if display == "" {
				display = inner
			}
			y.Name = hoisted
			y.Display = display
			before = append(before, fixClass(y, prefix+x.Name+"_nested_")...)
			body = append(body, &pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: inner, Lineno: y.Lineno}}, Value: &pyast.Name{ID: hoisted, Lineno: y.Lineno}, Lineno: y.Lineno})
			continue
		case *pyast.Assign:
			if len(y.Targets) == 1 {
				if t, ok := y.Targets[0].(*pyast.Name); ok && t.ID == "__class_getitem__" {
					continue
				}
			}
			if getter, setter, ok := classPropertyCall(y.Value); ok {
				for _, t := range y.Targets {
					if tn, isName := t.(*pyast.Name); isName {
						body = append(body, propertyGetter(tn.ID, getter, y.Lineno))
						if setter != "" {
							body = append(body, propertySetter(tn.ID, setter, y.Lineno))
						}
					}
				}
				continue
			}
			if len(y.Targets) > 1 {
				// a = b = value: a = value, then b = a.
				first := y.Targets[len(y.Targets)-1]
				body = append(body, &pyast.Assign{Targets: []pyast.Node{first}, Value: y.Value, Lineno: y.Lineno})
				for i := len(y.Targets) - 2; i >= 0; i-- {
					body = append(body, &pyast.Assign{Targets: []pyast.Node{y.Targets[i]}, Value: first, Lineno: y.Lineno})
				}
				continue
			}
		case *pyast.FunctionDef:
			if y.Name == "__class_getitem__" {
				continue
			}
			if abstractDecorators(y) {
				x.Abstract = append(x.Abstract, y.Name)
			}
		}
		body = append(body, st)
	}
	if len(body) == 0 {
		body = []pyast.Node{&pyast.Pass{Lineno: x.Lineno}}
	}
	x.Body = body
	var after []pyast.Node
	var keep []pyast.Node
	for i := len(x.Decorators) - 1; i >= 0; i-- {
		d := x.Decorators[i]
		if knownClassDecorator(d) {
			keep = append([]pyast.Node{d}, keep...)
			continue
		}
		line := x.Lineno
		call := &pyast.Call{Func: d, Args: []pyast.Node{&pyast.Name{ID: x.Name, Lineno: line}}, Lineno: line}
		msg := &pyast.Constant{PyType: "str", Value: "a class decorator that does not return the class it is given is not supported", Lineno: line}
		after = append(after, &pyast.If{
			Test:   &pyast.Compare{Left: call, Ops: []string{"IsNot"}, Comparators: []pyast.Node{&pyast.Name{ID: x.Name, Lineno: line}}, Lineno: line},
			Body:   []pyast.Node{&pyast.Raise{Exc: &pyast.Call{Func: &pyast.Name{ID: "TypeError", Lineno: line}, Args: []pyast.Node{msg}, Lineno: line}, Lineno: line}},
			Lineno: line,
		})
	}
	x.Decorators = keep
	out := append(before, x)
	return append(out, after...)
}

// fixBranchDefs renames a def that sits inside an if, for, while, try or with
// (in a module or a function body) to name_defN and assigns it to name right
// after: the compiler then sees an ordinary variable, which is visible after
// the block, as Python's def is.
func fixBranchDefs(mod *pyast.Module) {
	counter := 0
	var inBody func(body []pyast.Node, nested bool) []pyast.Node
	inBody = func(body []pyast.Node, nested bool) []pyast.Node {
		var out []pyast.Node
		for _, st := range body {
			switch x := st.(type) {
			case *pyast.FunctionDef:
				x.Body = inBody(x.Body, false)
				if nested {
					counter++
					orig := x.Name
					x.Name = orig + "_def" + strconv.Itoa(counter)
					out = append(out, x, &pyast.Assign{Targets: []pyast.Node{&pyast.Name{ID: orig, Lineno: x.Lineno}}, Value: &pyast.Name{ID: x.Name, Lineno: x.Lineno}, Lineno: x.Lineno})
					continue
				}
			case *pyast.ClassDef:
				for _, cst := range x.Body {
					if fd, ok := cst.(*pyast.FunctionDef); ok {
						fd.Body = inBody(fd.Body, false)
					}
				}
			case *pyast.If:
				x.Body = inBody(x.Body, true)
				x.Orelse = inBody(x.Orelse, true)
			case *pyast.For:
				x.Body = inBody(x.Body, true)
				x.Orelse = inBody(x.Orelse, true)
			case *pyast.While:
				x.Body = inBody(x.Body, true)
				x.Orelse = inBody(x.Orelse, true)
			case *pyast.Try:
				x.Body = inBody(x.Body, true)
				for _, h := range x.Handlers {
					h.Body = inBody(h.Body, true)
				}
				x.Orelse = inBody(x.Orelse, true)
				x.Finalbody = inBody(x.Finalbody, true)
			case *pyast.With:
				x.Body = inBody(x.Body, true)
			}
			out = append(out, st)
		}
		return out
	}
	mod.Body = inBody(mod.Body, false)
}
