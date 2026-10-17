package link

import (
	"reflect"

	"github.com/amayakmt/py2go/internal/pyast"
)

// rewriteNew turns the __new__ of a class that derives from an immutable
// built-in type (str, int, float, tuple, bytes, frozenset) into an __init__:
// the class is compiled over a base that holds the value (see
// stdlib/_basetypes.py), so "obj = super().__new__(cls, value)" becomes
// "obj = self._set(value)", cls is type(self), and "return obj" returns
// nothing. A class with both __new__ and __init__ gets an __init__ that runs
// the two in turn.
func rewriteNew(x *pyast.ClassDef, bases map[string]string) {
	var newFn, initFn *pyast.FunctionDef
	for _, st := range x.Body {
		if fd, ok := st.(*pyast.FunctionDef); ok {
			switch fd.Name {
			case "__new__":
				newFn = fd
			case "__init__":
				initFn = fd
			}
		}
	}
	if newFn == nil || newFn.Args == nil || len(newFn.Args.Positional()) == 0 {
		return
	}
	// staticmethod(__new__) decorators are implicit; drop any.
	newFn.Decorators = nil
	newFn.HasDecorators = false
	params := newFn.Args.Positional()
	clsName := params[0].Name
	params[0].Name = "self"
	usesCls := false
	newFn.Body = rewriteNewBody(newFn.Body, clsName, bases, &usesCls)
	if usesCls {
		newFn.Body = append([]pyast.Node{&pyast.Assign{
			Targets: []pyast.Node{&pyast.Name{ID: clsName, Lineno: newFn.Lineno}},
			Value:   &pyast.Call{Func: &pyast.Name{ID: "type", Lineno: newFn.Lineno}, Args: []pyast.Node{&pyast.Name{ID: "self", Lineno: newFn.Lineno}}, Lineno: newFn.Lineno},
			Lineno:  newFn.Lineno,
		}}, newFn.Body...)
	}
	if initFn == nil {
		newFn.Name = "__init__"
		return
	}
	// Both: __init__(self, *args, **kwargs) runs the new part, then the old init.
	newFn.Name = "_basetype_new"
	initFn.Name = "_basetype_init"
	line := x.Lineno
	call := func(name string) pyast.Node {
		return &pyast.Expr{Value: &pyast.Call{
			Func:        &pyast.Attribute{Value: &pyast.Name{ID: "self", Lineno: line}, Attr: name, Lineno: line},
			Args:        []pyast.Node{&pyast.Starred{Value: &pyast.Name{ID: "args", Lineno: line}, Lineno: line}},
			Keywords:    []*pyast.Keyword{{Arg: "", Value: &pyast.Name{ID: "kwargs", Lineno: line}, Lineno: line}},
			HasKeywords: true,
			Lineno:      line,
		}, Lineno: line}
	}
	combined := &pyast.FunctionDef{
		Name: "__init__",
		Args: &pyast.Arguments{
			Args:   []*pyast.Arg{{Name: "self", Lineno: line}},
			Vararg: &pyast.Arg{Name: "args", Lineno: line},
			Kwarg:  &pyast.Arg{Name: "kwargs", Lineno: line},
		},
		HasParams: true,
		Body:      []pyast.Node{call("_basetype_new"), call("_basetype_init")},
		Lineno:    line,
	}
	x.Body = append(x.Body, combined)
}

// isBaseNewCall is super().__new__(cls, ...) or tuple.__new__(cls, ...).
func isBaseNewCall(n pyast.Node, clsName string, bases map[string]string) (*pyast.Call, bool) {
	c, ok := n.(*pyast.Call)
	if !ok {
		return nil, false
	}
	a, ok := c.Func.(*pyast.Attribute)
	if !ok || a.Attr != "__new__" || len(c.Args) == 0 {
		return nil, false
	}
	if first, ok := c.Args[0].(*pyast.Name); !ok || first.ID != clsName {
		return nil, false
	}
	switch v := a.Value.(type) {
	case *pyast.Call:
		if f, ok := v.Func.(*pyast.Name); ok && f.ID == "super" {
			return c, true
		}
	case *pyast.Name:
		if bases[v.ID] != "" {
			return c, true
		}
	}
	return nil, false
}

func rewriteNewBody(body []pyast.Node, clsName string, bases map[string]string, usesCls *bool) []pyast.Node {
	var out []pyast.Node
	for _, st := range body {
		switch x := st.(type) {
		case *pyast.Return:
			if x.Value == nil {
				out = append(out, x)
				continue
			}
			v := rewriteNewExpr(x.Value, clsName, bases, usesCls)
			if _, isName := v.(*pyast.Name); !isName {
				out = append(out, &pyast.Expr{Value: v, Lineno: x.Lineno})
			}
			out = append(out, &pyast.Return{Lineno: x.Lineno})
			continue
		case *pyast.If:
			x.Test = rewriteNewExpr(x.Test, clsName, bases, usesCls)
			x.Body = rewriteNewBody(x.Body, clsName, bases, usesCls)
			x.Orelse = rewriteNewBody(x.Orelse, clsName, bases, usesCls)
		case *pyast.For:
			x.Body = rewriteNewBody(x.Body, clsName, bases, usesCls)
			x.Orelse = rewriteNewBody(x.Orelse, clsName, bases, usesCls)
			rewriteNewChildren(x, clsName, bases, usesCls)
		case *pyast.While:
			x.Body = rewriteNewBody(x.Body, clsName, bases, usesCls)
			x.Orelse = rewriteNewBody(x.Orelse, clsName, bases, usesCls)
		case *pyast.Try:
			x.Body = rewriteNewBody(x.Body, clsName, bases, usesCls)
			for _, h := range x.Handlers {
				h.Body = rewriteNewBody(h.Body, clsName, bases, usesCls)
			}
			x.Orelse = rewriteNewBody(x.Orelse, clsName, bases, usesCls)
			x.Finalbody = rewriteNewBody(x.Finalbody, clsName, bases, usesCls)
		case *pyast.With:
			x.Body = rewriteNewBody(x.Body, clsName, bases, usesCls)
		default:
			rewriteNewChildren(st, clsName, bases, usesCls)
		}
		out = append(out, st)
	}
	return out
}

func rewriteNewExpr(n pyast.Node, clsName string, bases map[string]string, usesCls *bool) pyast.Node {
	if c, ok := isBaseNewCall(n, clsName, bases); ok {
		args := c.Args[1:]
		for i := range args {
			args[i] = rewriteNewExpr(args[i], clsName, bases, usesCls)
		}
		return &pyast.Call{Func: &pyast.Attribute{Value: &pyast.Name{ID: "self", Lineno: c.Lineno}, Attr: "_set", Lineno: c.Lineno}, Args: args, Keywords: c.Keywords, HasKeywords: c.HasKeywords, Lineno: c.Lineno}
	}
	if nm, ok := n.(*pyast.Name); ok && nm.ID == clsName {
		*usesCls = true
		return n
	}
	rewriteNewChildren(n, clsName, bases, usesCls)
	return n
}

// rewriteNewChildren rewrites, in place, the expressions held by a node.
func rewriteNewChildren(n pyast.Node, clsName string, bases map[string]string, usesCls *bool) {
	if n == nil {
		return
	}
	var visit func(v reflect.Value)
	visit = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return
			}
			if node, ok := v.Interface().(pyast.Node); ok && v.CanSet() {
				switch node.(type) {
				case *pyast.FunctionDef, *pyast.ClassDef, *pyast.Lambda:
					return
				}
				v.Set(reflect.ValueOf(rewriteNewExpr(node, clsName, bases, usesCls)))
				return
			}
		case reflect.Pointer:
			if !v.IsNil() {
				visit(v.Elem())
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					visit(v.Field(i))
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				visit(v.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(n).Elem())
}
